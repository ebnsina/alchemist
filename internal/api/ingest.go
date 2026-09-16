package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/ebnsina/alchemist/internal/pipeline"
	"github.com/ebnsina/alchemist/internal/platform/httpx"
)

type createAssetRequest struct {
	URL   string  `json:"url"`
	Title *string `json:"title"`
}

// createAssetFromURL starts an ingest from a URL the customer supplies. This is how
// libraries migrate off another provider, so it is a first-class path, not a shortcut.
//
// The URL is validated again at fetch time, on every redirect hop. Anything checked
// only here would be a TOCTOU window.
func (s *Server) createAssetFromURL(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	var req createAssetRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&req); err != nil {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request",
			"Send a JSON body with a \"url\" field.")
		return
	}
	req.URL = strings.TrimSpace(req.URL)
	if req.URL == "" {
		writeErrFor(w, r, http.StatusBadRequest, "missing_url", "Include the video's URL.")
		return
	}
	u, err := url.Parse(req.URL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_url",
			"That doesn't look like a valid http or https link.")
		return
	}

	title, ok := cleanTitle(req.Title)
	if !ok {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_title",
			"That name is too long. Keep it under 200 characters.")
		return
	}
	// An import with no title takes the file name off the URL: a library migrated in
	// bulk otherwise arrives as several hundred rows of uuid.
	if title == nil {
		if t := titleFromPath(u.Path); t != "" && len(t) <= 200 {
			title = &t
		}
	}

	if msg, qerr := s.checkIngestQuota(r.Context(), tenantID); qerr != nil {
		if errors.Is(qerr, errQuotaExceeded) {
			writeErrFor(w, r, http.StatusTooManyRequests, "quota_exceeded", msg)
			return
		}
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}

	var assetID string
	err = s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`insert into assets (tenant_id, state, ladder_profile, source_url, title)
			 select $1, 'uploaded', ladder_profile, $2, $3 from tenants where id = $1
			 returning id::text`, tenantID, req.URL, title).Scan(&assetID)
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"We couldn't start that import. Please try again.")
		return
	}

	if _, err := s.river.Insert(r.Context(),
		pipeline.TranscodeArgs{AssetID: assetID, TenantID: tenantID}, nil); err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"We couldn't queue your video for processing.")
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{
		"asset_id": assetID, "state": "uploaded",
	})
}

type createWebhookRequest struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
}

type createWebhookResponse struct {
	ID     string   `json:"id"`
	URL    string   `json:"url"`
	Events []string `json:"events"`
	Secret string   `json:"secret"`
}

// The live lifecycle is three events, not one. A customer integrating live needs to
// know the encoder arrived, that the broadcast is over, and that it never started --
// and asset.ready already tells them the recording converted, because a broadcast is
// an asset and its recording becomes an ordinary VOD one. See docs/06-live.md.
var validEvents = map[string]bool{
	"asset.ready": true, "asset.failed": true, "rendition.ready": true,
	"live.started": true, "live.ended": true, "live.failed": true,
}

func (s *Server) createWebhook(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	var req createWebhookRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&req); err != nil {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request",
			"Send a JSON body with a \"url\" field.")
		return
	}
	u, err := url.Parse(strings.TrimSpace(req.URL))
	if err != nil || u.Scheme != "https" || u.Host == "" {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_url",
			"Webhook URLs must be https.")
		return
	}
	if len(req.Events) == 0 {
		req.Events = []string{"asset.ready", "asset.failed"}
	}
	for _, e := range req.Events {
		if !validEvents[e] {
			writeErrFor(w, r, http.StatusBadRequest, "unknown_event",
				"We don't send an event called \""+e+"\".")
			return
		}
	}

	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}

	var id string
	err = s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`insert into webhook_endpoints (tenant_id, url, secret, events)
			 values ($1,$2,$3,$4) returning id::text`,
			tenantID, req.URL, secret, req.Events).Scan(&id)
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"We couldn't save that webhook.")
		return
	}

	// The signing secret is shown once, here. It is never returned again.
	writeJSON(w, http.StatusCreated, createWebhookResponse{
		ID: id, URL: req.URL, Events: req.Events,
		Secret: hex.EncodeToString(secret),
	})
}

type webhookRow struct {
	ID     string   `json:"id"`
	URL    string   `json:"url"`
	Events []string `json:"events"`
	Active bool     `json:"active"`
}

func (s *Server) listWebhooks(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	page, ok := httpx.ParseList(w, r, []string{"created_at", "url"}, "created_at")
	if !ok {
		return
	}
	active, ok := httpx.Flag(w, r, "active")
	if !ok {
		return
	}

	const where = `from webhook_endpoints
	  where ($1::text = '' or url ilike '%' || $1::text || '%')
	    and ($2::bool is null or active = $2::bool)`

	items := []webhookRow{}
	var total int
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(),
			`select id::text, url, events, active `+where+
				` order by `+page.OrderBy()+` limit $3 offset $4`,
			page.Q, active, page.Limit, page.Offset)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var it webhookRow
			if err := rows.Scan(&it.ID, &it.URL, &it.Events, &it.Active); err != nil {
				return err
			}
			items = append(items, it)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		return tx.QueryRow(r.Context(), `select count(*) `+where, page.Q, active).Scan(&total)
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"webhooks": items, "total": total})
}

func (s *Server) getWebhook(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	var it webhookRow
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`select id::text, url, events, active from webhook_endpoints where id = $1`,
			chi.URLParam(r, "id")).Scan(&it.ID, &it.URL, &it.Events, &it.Active)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		writeErrFor(w, r, http.StatusNotFound, "webhook_not_found",
			"We couldn't find that webhook.")
		return
	}
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	// The secret is not here and never will be: only the endpoint keeps a copy.
	writeJSON(w, http.StatusOK, it)
}

// patchWebhook changes where deliveries go, or pauses them. The signing secret is
// untouched, so a paused endpoint resumed later verifies with the same secret.
func (s *Server) patchWebhook(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	var req struct {
		URL    *string `json:"url"`
		Active *bool   `json:"active"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&req); err != nil {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request",
			"Send a JSON body with a \"url\" or \"active\".")
		return
	}
	if req.URL != nil {
		u, err := url.Parse(strings.TrimSpace(*req.URL))
		if err != nil || u.Scheme != "https" || u.Host == "" {
			writeErrFor(w, r, http.StatusBadRequest, "invalid_url",
				"Webhook URLs must be https.")
			return
		}
		trimmed := u.String()
		req.URL = &trimmed
	}

	var it webhookRow
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`update webhook_endpoints
			    set url = coalesce($2, url), active = coalesce($3, active)
			  where id = $1
			 returning id::text, url, events, active`,
			chi.URLParam(r, "id"), req.URL, req.Active).
			Scan(&it.ID, &it.URL, &it.Events, &it.Active)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		writeErrFor(w, r, http.StatusNotFound, "webhook_not_found",
			"We couldn't find that webhook.")
		return
	}
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	writeJSON(w, http.StatusOK, it)
}

// deleteWebhook stops deliveries for good. A job already queued for this endpoint
// loads it before posting and finds nothing, so nothing in flight is delivered
// either -- which is the difference between deleting and pausing.
func (s *Server) deleteWebhook(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	var n int64
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		// The delivery log is a child of the endpoint and goes with it; pause with
		// PATCH active=false instead to keep the history.
		tag, err := tx.Exec(r.Context(),
			`delete from webhook_endpoints where id = $1`, chi.URLParam(r, "id"))
		if err != nil {
			return err
		}
		n = tag.RowsAffected()
		return nil
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	if n == 0 {
		writeErrFor(w, r, http.StatusNotFound, "webhook_not_found",
			"We couldn't find that webhook.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
