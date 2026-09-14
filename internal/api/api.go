// Package api is the customer-facing HTTP surface. Errors leave here as stable
// codes, never as raw database or ffmpeg text.
package api

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"

	"github.com/ebnsina/alchemist/internal/modules/delivery"
	"github.com/ebnsina/alchemist/internal/platform/db"
	"github.com/ebnsina/alchemist/internal/platform/httpx"
	"github.com/ebnsina/alchemist/internal/platform/keys"
	"github.com/ebnsina/alchemist/internal/platform/storage"
)

type Server struct {
	db       *db.DB
	store    *storage.Store
	river    *river.Client[pgx.Tx]
	delivery *delivery.Module
	keys     *keys.Wrapper
}

func New(database *db.DB, store *storage.Store, rc *river.Client[pgx.Tx], d *delivery.Module, kw *keys.Wrapper) *Server {
	return &Server{db: database, store: store, river: rc, delivery: d, keys: kw}
}

type ctxKey string

const tenantKey ctxKey = "tenant_id"

func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer)

	r.Get("/healthz", s.health)

	// Playback is public by signature, not by API key: a viewer has no key.
	// Delivery owns the whole playback surface. Mounting it as a unit is what lets
	// a standalone origin service mount exactly this and nothing else.
	s.delivery.Routes(r)

	// Viewers post telemetry, authorized by the playback signature, not an API key.
	r.Post("/playback/{tenant}/{asset}/beacon", s.postBeacon)
	r.Options("/playback/{tenant}/{asset}/beacon", s.postBeacon)

	r.Route("/v1", func(r chi.Router) {
		r.Use(s.authenticate)
		r.Get("/whoami", s.whoami)
		r.Post("/uploads", s.createUpload)
		r.Post("/assets/{id}/complete", s.completeUpload)
		r.Get("/assets/{id}", s.getAsset)
		r.Post("/assets", s.createAssetFromURL)
		r.Post("/webhooks", s.createWebhook)
		r.Get("/webhooks", s.listWebhooks)
		r.Get("/usage", s.getUsage)
		r.Post("/bucket-sources", s.createBucketSource)
		r.Get("/bucket-sources", s.listBucketSources)
	})
	return r
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if err := s.db.Pool().Ping(r.Context()); err != nil {
		writeErrFor(w, r, http.StatusServiceUnavailable, "database_unavailable",
			"The service is temporarily unavailable. Please try again shortly.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// authenticate resolves a bearer API key to a tenant. The key is stored hashed;
// a plaintext compare never happens.
func (s *Server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || key == "" {
			writeErrFor(w, r, http.StatusUnauthorized, "missing_api_key",
				"Include your API key as a Bearer token.")
			return
		}
		sum := sha256.Sum256([]byte(key))

		// The resolver returns NULL rather than no rows when nothing matches.
		var tenantID *string
		err := s.db.Pool().QueryRow(r.Context(),
			`select resolve_api_key($1)::text`, sum[:]).Scan(&tenantID)
		if err == nil && tenantID == nil {
			writeErrFor(w, r, http.StatusUnauthorized, "invalid_api_key",
				"That API key is not valid.")
			return
		}
		if err != nil {
			writeErrFor(w, r, http.StatusServiceUnavailable, "database_unavailable",
				"The service is temporarily unavailable. Please try again shortly.")
			return
		}

		ctx := context.WithValue(r.Context(), tenantKey, *tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) whoami(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	var name, profile string
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`select name, ladder_profile from tenants`).Scan(&name, &profile)
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"tenant_id": tenantID, "name": name, "ladder_profile": profile,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeErr is kept for handlers without a request in scope; writeErrFor is preferred
// because it can answer in the caller's language.
func writeErr(w http.ResponseWriter, status int, code, message string) {
	httpx.Error(w, status, code, message)
}

func writeErrFor(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	httpx.ErrorFor(w, r, status, code, message)
}
