package api

import (
	"errors"
	"fmt"
	"net/http"
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

type assetResponse struct {
	ID          string        `json:"id"`
	State       string        `json:"state"`
	ErrorCode   *string       `json:"error_code,omitempty"`
	DurationSec *float64      `json:"duration_seconds,omitempty"`
	Width       *int          `json:"width,omitempty"`
	Height      *int          `json:"height,omitempty"`
	Playback    *playbackURLs `json:"playback,omitempty"`
}

func (s *Server) getAsset(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)
	assetID := chi.URLParam(r, "id")

	var resp assetResponse
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`select id::text, state::text, error_code, duration_sec, width, height
			   from assets where id = $1`, assetID).
			Scan(&resp.ID, &resp.State, &resp.ErrorCode, &resp.DurationSec,
				&resp.Width, &resp.Height)
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

	if resp.State == "ready" || resp.State == "partially_ready" {
		exp := time.Now().Add(4 * time.Hour).Unix()
		base := fmt.Sprintf("/playback/%s/%s", tenantID, assetID)
		kid, sig := s.delivery.SignPlayback(base, exp)
		q := fmt.Sprintf("exp=%d&kid=%s&sig=%s", exp, kid, sig)
		resp.Playback = &playbackURLs{
			HLS:        fmt.Sprintf("%s/master.m3u8?%s", base, q),
			DASH:       fmt.Sprintf("%s/manifest.mpd?%s", base, q),
			Poster:     fmt.Sprintf("%s/poster.jpg?%s", base, q),
			Thumbnails: fmt.Sprintf("%s/sprite.vtt?%s", base, q),
		}
	}
	writeJSON(w, http.StatusOK, resp)
}
