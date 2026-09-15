// Package adapters binds platform implementations to the narrow interfaces modules
// declare.
//
// This is the seam. Modules never import concrete infrastructure, so extracting one
// into its own service means copying the module plus the handful of adapters it
// uses, and nothing else. If a module ever imports a platform package directly, that
// extraction stops being mechanical -- internal/modules/boundary_test.go fails the
// build when it happens.
package adapters

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"

	"github.com/ebnsina/alchemist/internal/modules/delivery"
	"github.com/ebnsina/alchemist/internal/pipeline"
	"github.com/ebnsina/alchemist/internal/platform/db"
	"github.com/ebnsina/alchemist/internal/platform/keys"
	"github.com/ebnsina/alchemist/internal/platform/storage"
)

// ObjectStore adapts the S3 store to delivery's view of it.
type ObjectStore struct{ Store *storage.Store }

func (o ObjectStore) GetPassthrough(ctx context.Context, key, rangeHeader string) (*delivery.Object, error) {
	obj, err := o.Store.GetPassthrough(ctx, key, rangeHeader)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrNotFound):
			return nil, delivery.ErrNotFound
		case errors.Is(err, storage.ErrRangeNotSatisfiable):
			return nil, delivery.ErrRangeNotSatisfiable
		}
		return nil, err
	}
	return &delivery.Object{
		Body:          obj.Body,
		ETag:          obj.ETag,
		ContentLength: obj.ContentLength,
		ContentRange:  obj.ContentRange,
	}, nil
}

// ContentKeys stores content keys wrapped under the platform KEK. Both the delivery
// module (which reads) and the transcoder (which writes) go through this, so neither
// knows how the other persists them.
type ContentKeys struct {
	DB      *db.DB
	Wrapper *keys.Wrapper
}

// Get resolves through deduplicated_from, the same way playback does.
//
// A duplicate owns no media and therefore no key: it plays the canonical asset's
// encrypted bytes, so asking for its own key returns nothing and the player fails to
// decrypt with no error anywhere but a 404 on /key. The canonical id is also what the
// key was wrapped under, so it has to come back from the same query.
func (c ContentKeys) Get(ctx context.Context, tenantID, assetID string) ([]byte, []byte, error) {
	var keyID, wrapped, nonce []byte
	var canonical string
	err := c.DB.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`select k.key_id, k.wrapped_key, k.nonce, k.asset_id::text
			   from assets a
			   join content_keys k on k.asset_id = coalesce(a.deduplicated_from, a.id)
			  where a.id = $1`, assetID).Scan(&keyID, &wrapped, &nonce, &canonical)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, delivery.ErrNotFound
	}
	if err != nil {
		return nil, nil, err
	}
	key, err := c.Wrapper.Unwrap(wrapped, nonce, canonical)
	return keyID, key, err
}

func (c ContentKeys) Put(ctx context.Context, tenantID, assetID string, keyID, key []byte) error {
	wrapped, nonce, err := c.Wrapper.Wrap(key, assetID)
	if err != nil {
		return err
	}
	return c.DB.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`insert into content_keys (asset_id, tenant_id, key_id, wrapped_key, nonce)
			 values ($1,$2,$3,$4,$5)
			 on conflict (asset_id) do update set
			   key_id = excluded.key_id, wrapped_key = excluded.wrapped_key,
			   nonce = excluded.nonce`,
			assetID, tenantID, keyID, wrapped, nonce)
		return err
	})
}

// LazyRenditions triggers on-demand generation when an asset is first played.
type LazyRenditions struct {
	DB    *db.DB
	River *river.Client[pgx.Tx]
}

// OnPlaybackStarted queues any rendition the profile defers and nobody has asked for
// yet. It is deliberately fire-and-forget: a viewer must never wait on it, and the
// low ladder is already playing.
func (l LazyRenditions) OnPlaybackStarted(ctx context.Context, tenantID, assetID string) {
	type pending struct {
		assetID string
		height  int
		codec   string
	}
	var want []pending

	// Resolved through deduplicated_from: a duplicate shares the canonical asset's
	// media, so a rung generated under the duplicate's own id is published where no
	// playback request will ever look for it.
	if err := l.DB.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx,
			`update renditions set requested_at = now()
			  where asset_id = (select coalesce(deduplicated_from, id)
			                      from assets where id = $1)
			    and lazy and state = 'pending' and requested_at is null
			 returning asset_id::text, height, codec`, assetID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var p pending
			if err := rows.Scan(&p.assetID, &p.height, &p.codec); err != nil {
				return err
			}
			want = append(want, p)
		}
		return rows.Err()
	}); err != nil {
		return
	}

	for _, p := range want {
		// Unique-by-args on the job means a burst of viewers produces one encode.
		_, _ = l.River.Insert(ctx, pipeline.JITArgs{
			AssetID: p.assetID, TenantID: tenantID, Height: p.height, Codec: p.codec,
		}, nil)
	}
}

// DedupResolver points a deduplicated asset at the media it shares.
type DedupResolver struct{ DB *db.DB }

// StoragePrefix follows deduplicated_from so several assets can reference one copy of
// the media. Resolving at read time rather than copying objects is the entire saving:
// duplicating the files would make dedup pointless.
//
// media_prefix overrides it, and is set only when the asset the prefix was named
// after has been deleted while others still play its bytes.
//
// A broadcast resolves to live/ instead, which is the only thing about live that the
// origin has to know: those segments are transient and swept once the recording is
// converted, so they must not share a prefix with the VOD library.
func (d DedupResolver) StoragePrefix(ctx context.Context, tenantID, assetID string) (string, error) {
	var prefix, canonical, state string
	err := d.DB.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`select coalesce(media_prefix, 'cmaf/' || tenant_id::text || '/' ||
			          coalesce(deduplicated_from, id)::text),
			        coalesce(deduplicated_from, id)::text, state::text
			   from assets where id = $1`, assetID).Scan(&prefix, &canonical, &state)
	})
	if err != nil {
		return "", err
	}
	if state == "live" || state == "live_ended" {
		return "live/" + tenantID + "/" + canonical, nil
	}
	return prefix, nil
}
