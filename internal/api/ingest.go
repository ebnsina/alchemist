package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/ebnsina/alchemist/internal/pipeline"
)

type createAssetRequest struct {
	URL string `json:"url"`
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
			`insert into assets (tenant_id, state, ladder_profile, source_url)
			 select $1, 'uploaded', ladder_profile, $2 from tenants where id = $1
			 returning id::text`, tenantID, req.URL).Scan(&assetID)
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

var validEvents = map[string]bool{
	"asset.ready": true, "asset.failed": true, "rendition.ready": true,
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

func (s *Server) listWebhooks(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	type item struct {
		ID     string   `json:"id"`
		URL    string   `json:"url"`
		Events []string `json:"events"`
		Active bool     `json:"active"`
	}
	items := []item{}
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(),
			`select id::text, url, events, active from webhook_endpoints
			  order by created_at desc`)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var it item
			if err := rows.Scan(&it.ID, &it.URL, &it.Events, &it.Active); err != nil {
				return err
			}
			items = append(items, it)
		}
		return rows.Err()
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"webhooks": items})
}
