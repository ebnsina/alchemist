// Package api is the customer-facing HTTP surface. Errors leave here as stable
// codes, never as raw database or ffmpeg text.
package api

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"

	"github.com/ebnsina/alchemist/internal/modules/delivery"
	"github.com/ebnsina/alchemist/internal/platform/db"
	"github.com/ebnsina/alchemist/internal/platform/httpx"
	"github.com/ebnsina/alchemist/internal/platform/keys"
	"github.com/ebnsina/alchemist/internal/platform/metrics"
	"github.com/ebnsina/alchemist/internal/platform/storage"
)

type Server struct {
	db            *db.DB
	store         *storage.Store
	river         *river.Client[pgx.Tx]
	delivery      *delivery.Module
	keys          *keys.Wrapper
	adminKey      string
	metrics       *metrics.Registry
	webOrigins    []string
	sessionDomain string
	sessionSecure bool
	authLimiter   *authLimiter
}

// Accounts carries what the browser-facing signup and login surface needs. Zero
// origins leaves that surface unmounted.
type Accounts struct {
	WebOrigins    []string
	SessionDomain string
	SessionSecure bool
}

func New(database *db.DB, store *storage.Store, rc *river.Client[pgx.Tx], d *delivery.Module, kw *keys.Wrapper, adminKey string, m *metrics.Registry, acc Accounts) *Server {
	return &Server{db: database, store: store, river: rc, delivery: d, keys: kw,
		adminKey: adminKey, metrics: m,
		webOrigins: acc.WebOrigins, sessionDomain: acc.SessionDomain,
		sessionSecure: acc.SessionSecure,
		// Ten attempts a minute from one address: generous for a person, useless for
		// a dictionary.
		authLimiter: newAuthLimiter(10, time.Minute)}
}

type ctxKey string

const tenantKey ctxKey = "tenant_id"

func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer)

	r.Get("/healthz", s.health)

	// Scrape endpoint. Not behind the customer API key: it carries no tenant data,
	// and a scraper has no key. Bind it to an internal interface in production.
	if s.metrics != nil {
		r.Method("GET", "/metrics", s.metrics.Handler())
	}

	// Playback is public by signature, not by API key: a viewer has no key.
	// Delivery owns the whole playback surface. Mounting it as a unit is what lets
	// a standalone origin service mount exactly this and nothing else.
	s.delivery.Routes(r)

	// Viewers post telemetry, authorized by the playback signature, not an API key.
	r.Post("/playback/{tenant}/{asset}/beacon", s.postBeacon)
	r.Options("/playback/{tenant}/{asset}/beacon", s.postBeacon)

	// Operator surface, behind a separate credential so a leaked customer key cannot
	// mint tenants or more keys.
	r.Route("/admin", func(r chi.Router) {
		r.Use(s.adminOnly)
		r.Post("/tenants", s.createTenant)
		r.Get("/tenants", s.listTenants)
		r.Post("/tenants/{id}/keys", s.issueKey)
		r.Delete("/keys/{keyID}", s.revokeKey)
		r.Get("/contact", s.listContact)
	})

	// Accounts. Outside the API-key middleware by necessity — signing up is how you
	// get a key — and mounted only when a browser origin is configured.
	if s.authEnabled() {
		r.Route("/v1/auth", func(r chi.Router) {
			r.Use(s.cors)
			// Only the two endpoints that take a password are rate limited. Session
			// and logout are not: the dashboard asks who it is talking to on every
			// page, and a 429 there logs the customer out of their own account.
			r.Group(func(r chi.Router) {
				r.Use(s.rateLimit)
				r.Post("/signup", s.postSignup)
				r.Post("/login", s.postLogin)
			})
			r.Post("/logout", s.postLogout)
			r.Get("/session", s.getSession)
		})
		// The contact form is not an account endpoint, but it is the same shape:
		// a browser, no credential, and a reason to rate limit.
		r.Group(func(r chi.Router) {
			r.Use(s.cors, s.rateLimit)
			r.Post("/v1/contact", s.postContact)
		})
	}

	r.Route("/v1", func(r chi.Router) {
		if s.authEnabled() {
			r.Use(s.cors)
		}
		r.Use(s.authenticate)
		r.Get("/whoami", s.whoami)
		r.Post("/uploads", s.createUpload)
		r.Post("/assets/{id}/complete", s.completeUpload)
		r.Get("/assets/{id}", s.getAsset)
		r.Get("/assets", s.listAssets)
		r.Post("/assets", s.createAssetFromURL)
		r.Post("/webhooks", s.createWebhook)
		r.Get("/webhooks", s.listWebhooks)
		r.Get("/usage", s.getUsage)
		r.Post("/bucket-sources", s.createBucketSource)
		r.Get("/bucket-sources", s.listBucketSources)
		r.Get("/keys", s.listKeys)
		r.Post("/keys", s.createKey)
		r.Delete("/keys/{id}", s.deleteKey)
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
			// No key: this may be the dashboard, which holds a cookie rather than a
			// key. A session is accepted only from a configured origin, because a
			// cookie is sent by the browser on any site's behalf and that is what
			// CSRF is.
			if tenantID, ok := s.tenantFromSession(r); ok {
				ctx := context.WithValue(r.Context(), tenantKey, tenantID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
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
