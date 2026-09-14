package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type qoeBeacon struct {
	SessionID     string  `json:"session_id"`
	AssetID       string  `json:"asset_id"`
	StartupMS     *int    `json:"startup_ms"`
	RebufferCount *int    `json:"rebuffer_count"`
	RebufferMS    *int    `json:"rebuffer_ms"`
	AvgBitrateBPS *int    `json:"avg_bitrate_bps"`
	ErrorCode     *string `json:"error_code"`
	Network       *string `json:"network"`
	Country       *string `json:"country"`
}

// postBeacon records playback quality from a viewer's player.
//
// Authorized by the playback signature rather than an API key: the sender is a viewer,
// who has no key. Everything in the body is attacker-controlled, so values are clamped
// rather than trusted, and the response is always 204 — a player must never retry or
// surface an error because telemetry failed.
func (s *Server) postBeacon(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenant")
	assetID := chi.URLParam(r, "asset")

	// Beacons are cross-origin by definition: the player runs on the customer's page
	// or in an iframe on a different host. Today's sender uses a text/plain blob,
	// which is a "simple request" and skips preflight -- but any future change that
	// needs a real content type or a readable response would silently start failing
	// without these.
	setBeaconCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	q := r.URL.Query()
	prefix := "/playback/" + tenantID + "/" + assetID
	if !s.delivery.VerifyPlayback(prefix, q.Get("kid"), q.Get("sig"), q.Get("exp")) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var b qoeBeacon
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&b); err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if strings.TrimSpace(b.SessionID) == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_ = s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(r.Context(),
			`insert into qoe_events (tenant_id, asset_id, session_id, startup_ms,
			     rebuffer_count, rebuffer_ms, avg_bitrate_bps, error_code, network, country)
			 values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			tenantID, assetID, clampStr(b.SessionID, 64),
			clampInt(b.StartupMS, 0, 600_000),
			clampInt(b.RebufferCount, 0, 10_000),
			clampInt(b.RebufferMS, 0, 86_400_000),
			clampInt(b.AvgBitrateBPS, 0, 100_000_000),
			clampPtrStr(b.ErrorCode, 64),
			clampPtrStr(b.Network, 32),
			clampPtrStr(b.Country, 8))
		return err
	})
	w.WriteHeader(http.StatusNoContent)
}

func setBeaconCORS(w http.ResponseWriter) {
	h := w.Header()
	h.Set("Access-Control-Allow-Origin", "*")
	h.Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	h.Set("Access-Control-Allow-Headers", "Content-Type")
	h.Set("Access-Control-Max-Age", "86400")
}

func clampInt(v *int, lo, hi int) *int {
	if v == nil {
		return nil
	}
	n := *v
	if n < lo {
		n = lo
	}
	if n > hi {
		n = hi
	}
	return &n
}

func clampStr(s string, max int) string {
	if len(s) > max {
		return s[:max]
	}
	return s
}

func clampPtrStr(s *string, max int) *string {
	if s == nil {
		return nil
	}
	v := clampStr(strings.TrimSpace(*s), max)
	if v == "" {
		return nil
	}
	return &v
}
