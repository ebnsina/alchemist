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
	"io"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"

	"github.com/ebnsina/alchemist/internal/modules/delivery"
	"github.com/ebnsina/alchemist/internal/pipeline"
	"github.com/ebnsina/alchemist/internal/platform/db"
	"github.com/ebnsina/alchemist/internal/platform/keys"
	"github.com/ebnsina/alchemist/internal/platform/media"
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
	// live_armed resolves here too: nothing is written yet, so the honest answer is a
	// 404 for a segment that does not exist, not a hit in the VOD library.
	if state == "live_armed" || state == "live" || state == "live_ended" {
		return "live/" + tenantID + "/" + canonical, nil
	}
	return prefix, nil
}

// Put lets the live module publish a segment without seeing the storage client. The
// same adapter serves delivery's reads, so one binding covers both directions.
func (o ObjectStore) Put(ctx context.Context, key string, body io.Reader, contentType string) error {
	return o.Store.Put(ctx, key, body, contentType)
}

// LiveAssets is live's view of the asset a broadcast is watched at. The asset table
// belongs to the video side, so this is the file that becomes an HTTP call the day
// live moves out -- and the only one.
type LiveAssets struct{ DB *db.DB }

// CreateForBroadcast mints the asset on the tenant's own ladder profile, which is
// what the worker then resolves its single realtime rung from.
//
// live_armed, not live: arming is not broadcasting. An asset created in 'live' shows
// as ON AIR the moment somebody presses Start and hands out playback URLs for
// segments nothing has written yet. MarkLive moves it on the first segment.
func (l LiveAssets) CreateForBroadcast(ctx context.Context, tenantID string) (string, error) {
	var assetID string
	err := l.DB.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`insert into assets (tenant_id, ladder_profile, state)
			 select $1, t.ladder_profile, 'live_armed' from tenants t
			 returning id::text`, tenantID).Scan(&assetID)
	})
	return assetID, err
}

func (l LiveAssets) MarkLive(ctx context.Context, tenantID, assetID string) error {
	return l.setState(ctx, tenantID, assetID, "live", nil)
}

func (l LiveAssets) MarkEnded(ctx context.Context, tenantID, assetID string) error {
	return l.setState(ctx, tenantID, assetID, "live_ended", nil)
}

func (l LiveAssets) MarkFailed(ctx context.Context, tenantID, assetID, code string) error {
	return l.setState(ctx, tenantID, assetID, "failed", &code)
}

// State is what tells the sweep whether the recording still lives in the live prefix.
func (l LiveAssets) State(ctx context.Context, tenantID, assetID string) (string, error) {
	var state string
	err := l.DB.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`select state::text from assets where id = $1`, assetID).Scan(&state)
	})
	return state, err
}

func (l LiveAssets) setState(ctx context.Context, tenantID, assetID, state string, code *string) error {
	return l.DB.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`update assets set state = $2::asset_state,
			        error_code = coalesce($3, error_code), updated_at = now()
			  where id = $1`, assetID, state, code)
		return err
	})
}

// LiveLadder resolves the rungs a broadcast encodes at. ladder_profiles is shared
// reference data, so live reads it here rather than joining to it itself.
type LiveLadder struct{ DB *db.DB }

func (l LiveLadder) Rungs(ctx context.Context, tenantID, assetID string) ([]media.Rung, error) {
	var raw []byte
	if err := l.DB.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`select p.rungs from assets a
			   join ladder_profiles p on p.name = a.ladder_profile
			  where a.id = $1`, assetID).Scan(&raw)
	}); err != nil {
		return nil, err
	}
	return media.ParseLadder(raw)
}

// LiveQueue schedules the broadcast job. The module names what it wants queued; the
// job type and the queue it lands on are this side of the seam.
type LiveQueue struct{ River *river.Client[pgx.Tx] }

func (q LiveQueue) EnqueueSession(ctx context.Context, sessionID, tenantID string) error {
	_, err := q.River.Insert(ctx, pipeline.LiveArgs{SessionID: sessionID, TenantID: tenantID}, nil)
	return err
}

// ConvertRecording is the ordinary transcode job: the recording reads its source back
// out of the live prefix, and everything after that -- ladder, encryption, CMAF,
// rendition rows, asset.ready -- is the path every upload already takes.
//
// Unique by args because both the worker that ended the broadcast and the reaper that
// noticed it ended ask for this, and converting one recording twice would publish the
// same objects under two encodes.
func (q LiveQueue) ConvertRecording(ctx context.Context, tenantID, assetID string) error {
	_, err := q.River.Insert(ctx,
		pipeline.TranscodeArgs{AssetID: assetID, TenantID: tenantID},
		&river.InsertOpts{UniqueOpts: river.UniqueOpts{ByArgs: true}})
	return err
}

func (q LiveQueue) ReclaimPrefix(ctx context.Context, prefix string) error {
	_, err := q.River.Insert(ctx, pipeline.ReclaimArgs{Prefixes: []string{prefix}}, nil)
	return err
}

// LiveEvents publishes the broadcast lifecycle through the same webhook path every
// other event uses.
type LiveEvents struct {
	DB    *db.DB
	River *river.Client[pgx.Tx]
}

func (e LiveEvents) Emit(ctx context.Context, tenantID, event string, data map[string]any) error {
	return pipeline.Emit(ctx, e.DB, e.River, tenantID, event, data)
}
