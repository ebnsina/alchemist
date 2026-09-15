package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/ebnsina/alchemist/internal/pipeline"
)

type createUploadResponse struct {
	AssetID   string `json:"asset_id"`
	UploadURL string `json:"upload_url"`
	ExpiresIn int    `json:"expires_in_seconds"`
}

// createUpload returns a presigned target so source bytes go straight to object
// storage and never pass through this API.
func (s *Server) createUpload(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	if msg, err := s.checkIngestQuota(r.Context(), tenantID); err != nil {
		if errors.Is(err, errQuotaExceeded) {
			writeErrFor(w, r, http.StatusTooManyRequests, "quota_exceeded", msg)
			return
		}
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}

	var assetID, profile string
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`insert into assets (tenant_id, state, ladder_profile)
			 select $1, 'uploading', ladder_profile from tenants where id = $1
			 returning id::text, ladder_profile`, tenantID).Scan(&assetID, &profile)
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"We couldn't start your upload. Please try again.")
		return
	}

	key := fmt.Sprintf("src/%s/%s/original", tenantID, assetID)
	const ttl = 6 * time.Hour
	url, err := s.store.PresignPut(r.Context(), key, ttl)
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"We couldn't start your upload. Please try again.")
		return
	}

	_ = s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(r.Context(),
			`update assets set source_key = $2 where id = $1`, assetID, key)
		return err
	})

	writeJSON(w, http.StatusCreated, createUploadResponse{
		AssetID: assetID, UploadURL: url, ExpiresIn: int(ttl.Seconds()),
	})
}

// completeUpload is the customer's signal that bytes have landed. It enqueues the
// transcode rather than doing any work inline.
func (s *Server) completeUpload(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)
	assetID := chi.URLParam(r, "id")

	var state string
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`update assets set state = 'uploaded', updated_at = now()
			  where id = $1 and state = 'uploading'
			 returning state`, assetID).Scan(&state)
	})
	if err == pgx.ErrNoRows {
		writeErrFor(w, r, http.StatusNotFound, "asset_not_found",
			"We couldn't find an upload waiting on that ID.")
		return
	}
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}

	if _, err := s.river.Insert(r.Context(),
		pipeline.TranscodeArgs{AssetID: assetID, TenantID: tenantID}, nil); err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"We couldn't queue your video for processing.")
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"asset_id": assetID, "state": "uploaded"})
}

type playbackURLs struct {
	HLS        string `json:"hls"`
	DASH       string `json:"dash"`
	Poster     string `json:"poster"`
	Thumbnails string `json:"thumbnails"`
}

// rendition is one size of one video. Chunk counts are what makes progress
// legible: an encode is a pile of chunks, and "12 of 40" says more than a state name.
type rendition struct {
	Height      int    `json:"height"`
	Codec       string `json:"codec"`
	BitrateBps  int    `json:"bitrate_bps"`
	State       string `json:"state"`
	ChunksDone  int    `json:"chunks_done"`
	ChunksTotal int    `json:"chunks_total"`
	Bytes       *int64 `json:"bytes"`
	Lazy        bool   `json:"lazy"`
}

type assetResponse struct {
	ID          string        `json:"id"`
	State       string        `json:"state"`
	ErrorCode   *string       `json:"error_code,omitempty"`
	DurationSec *float64      `json:"duration_seconds,omitempty"`
	Width       *int          `json:"width,omitempty"`
	Height      *int          `json:"height,omitempty"`
	SourceBytes *int64        `json:"source_bytes,omitempty"`
	CreatedAt   string        `json:"created_at,omitempty"`
	Renditions  []rendition   `json:"renditions"`
	Playback    *playbackURLs `json:"playback,omitempty"`
}

// bindable is what a viewer id or watermark label may contain. Restricted to
// characters that pass through a URL untouched, so the bytes the edge hashes are the
// bytes that were signed -- percent-encoding would make njs and Go disagree and every
// bound link would 403 at the edge only.
var bindable = regexp.MustCompile(`^[A-Za-z0-9._~@-]+$`)

func bindingOK(v string, max int) bool {
	return v == "" || (len(v) <= max && bindable.MatchString(v))
}

func (s *Server) getAsset(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)
	assetID := chi.URLParam(r, "id")

	// The customer's own id for whoever is watching, and the label their player
	// draws on screen. Both stay opaque to us: they are signed, echoed and never
	// stored, so we never learn which student a link belongs to.
	viewer := r.URL.Query().Get("viewer")
	label := r.URL.Query().Get("watermark")
	if !bindingOK(viewer, 64) || !bindingOK(label, 48) {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_viewer",
			"A viewer id or watermark can use letters, numbers and - . _ ~ @ only, "+
				"up to 64 and 48 characters.")
		return
	}

	var resp assetResponse
	resp.Renditions = []rendition{}
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		var created time.Time
		if err := tx.QueryRow(r.Context(),
			`select id::text, state::text, error_code, duration_sec, width, height,
			        source_bytes, created_at
			   from assets where id = $1`, assetID).
			Scan(&resp.ID, &resp.State, &resp.ErrorCode, &resp.DurationSec,
				&resp.Width, &resp.Height, &resp.SourceBytes, &created); err != nil {
			return err
		}
		resp.CreatedAt = created.UTC().Format(time.RFC3339)

		rows, err := tx.Query(r.Context(),
			`select height, codec, bitrate_bps, state, chunks_done, chunks_total, bytes, lazy
			   from renditions
			  where asset_id = (select coalesce(deduplicated_from, id)
			                      from assets where id = $1)
			  order by height desc, codec`, assetID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var d rendition
			if err := rows.Scan(&d.Height, &d.Codec, &d.BitrateBps, &d.State,
				&d.ChunksDone, &d.ChunksTotal, &d.Bytes, &d.Lazy); err != nil {
				return err
			}
			resp.Renditions = append(resp.Renditions, d)
		}
		return rows.Err()
	})
	if err == pgx.ErrNoRows {
		writeErrFor(w, r, http.StatusNotFound, "asset_not_found", "We couldn't find that video.")
		return
	}
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}

	// A broadcast is watchable while it is happening, and through the window where
	// it has ended but the recording has not been converted yet.
	switch resp.State {
	case "ready", "partially_ready", "live", "live_ended":
		exp := time.Now().Add(4 * time.Hour).Unix()
		base := fmt.Sprintf("/playback/%s/%s", tenantID, assetID)
		kid, sig := s.delivery.SignPlayback(base, exp, viewer, label)
		q := fmt.Sprintf("exp=%d&kid=%s&sig=%s", exp, kid, sig)
		if viewer != "" {
			q += "&vid=" + viewer
		}
		if label != "" {
			q += "&wm=" + label
		}
		resp.Playback = &playbackURLs{
			HLS:        fmt.Sprintf("%s/master.m3u8?%s", base, q),
			DASH:       fmt.Sprintf("%s/manifest.mpd?%s", base, q),
			Poster:     fmt.Sprintf("%s/poster.jpg?%s", base, q),
			Thumbnails: fmt.Sprintf("%s/sprite.vtt?%s", base, q),
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

// deleteAsset removes the video and queues its objects for reclamation.
//
// The media is reclaimed only when nothing else plays it. A deduplicated asset shares
// the canonical asset's files, so deleting those would leave every duplicate reporting
// ready while every byte range 404s -- the failure this platform has already had once.
func (s *Server) deleteAsset(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)
	assetID := chi.URLParam(r, "id")

	var mediaPrefix string
	var sourceKey, mezzKey *string
	var owns bool
	var heir *string
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(r.Context(),
			`select coalesce(media_prefix, 'cmaf/' || tenant_id::text || '/' ||
			          coalesce(deduplicated_from, id)::text),
			        deduplicated_from is null, source_key, mezzanine_key
			   from assets where id = $1`, assetID).
			Scan(&mediaPrefix, &owns, &sourceKey, &mezzKey); err != nil {
			return err
		}
		// The oldest asset still playing this media, if any. Everything below hangs
		// off whether one exists.
		if err := tx.QueryRow(r.Context(),
			`select (select id::text from assets
			          where id <> $1 and (deduplicated_from = $1 or media_prefix = $2)
			          order by created_at limit 1)`, assetID, mediaPrefix).Scan(&heir); err != nil {
			return err
		}
		if heir != nil {
			if err := promoteHeir(r.Context(), tx, assetID, *heir, mediaPrefix); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(r.Context(), `delete from assets where id = $1`, assetID); err != nil {
			return err
		}
		// Queued in the same transaction as the delete: a row that is gone with no
		// job behind it is an object nobody will ever reclaim.
		_, err := s.river.InsertTx(r.Context(), tx,
			reclaimFor(mediaPrefix, sourceKey, mezzKey, owns && heir == nil), nil)
		return err
	})
	if errors.Is(err, pgx.ErrNoRows) {
		writeErrFor(w, r, http.StatusNotFound, "asset_not_found", "We couldn't find that video.")
		return
	}
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// promoteHeir moves what the departing asset owned on behalf of the group onto one
// survivor, so every other duplicate keeps resolving through deduplicated_from exactly
// as it did before.
//
// Nulling the pointers instead is not enough and the failure is silent: content_keys
// and renditions are children of assets with on delete cascade, so the key every
// duplicate decrypts with and the rendition rows the API reads for them would vanish
// with the parent. Playback would keep working off the surviving objects while /key
// returned 404 and the asset reported no renditions at all.
func promoteHeir(ctx context.Context, tx pgx.Tx, assetID, heir, mediaPrefix string) error {
	// Every survivor gets the prefix written out: it is named after an id that is
	// about to stop existing, and only the heir would otherwise resolve correctly.
	for _, q := range []string{
		`update assets set media_prefix = $3
		  where id <> $1 and (deduplicated_from = $1 or media_prefix = $3)`,
		`update content_keys set asset_id = $2 where asset_id = $1`,
		`update renditions set asset_id = $2 where asset_id = $1`,
		`update assets set deduplicated_from = null where id = $2`,
		`update assets set deduplicated_from = $2 where deduplicated_from = $1 and id <> $2`,
	} {
		if _, err := tx.Exec(ctx, q, assetID, heir, mediaPrefix); err != nil {
			return err
		}
	}
	return nil
}

// reclaimFor lists what this asset alone was keeping alive. The source is always its
// own; the media and the mezzanine belong to it only when it is nobody else's source.
func reclaimFor(mediaPrefix string, sourceKey, mezzKey *string, ownsMedia bool) pipeline.ReclaimArgs {
	var args pipeline.ReclaimArgs
	if sourceKey != nil && *sourceKey != "" {
		args.Keys = append(args.Keys, *sourceKey)
	}
	if ownsMedia {
		if mezzKey != nil && *mezzKey != "" {
			args.Keys = append(args.Keys, *mezzKey)
		}
		args.Prefixes = append(args.Prefixes, mediaPrefix+"/")
	}
	return args
}
