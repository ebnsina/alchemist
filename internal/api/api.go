// Package api is the customer-facing HTTP surface. Errors leave here as stable
// codes, never as raw database or ffmpeg text.
package api

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"

	"github.com/ebnsina/alchemist/internal/modules/delivery"
	"github.com/ebnsina/alchemist/internal/modules/live"
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
	playerURL     string
	live          *live.Module
}

// Accounts carries what the browser-facing signup and login surface needs. Zero
// origins leaves that surface unmounted.
type Accounts struct {
	WebOrigins    []string
	SessionDomain string
	SessionSecure bool
}

func New(database *db.DB, store *storage.Store, rc *river.Client[pgx.Tx], d *delivery.Module, kw *keys.Wrapper, adminKey string, m *metrics.Registry, acc Accounts, lv *live.Module, playerURL string) *Server {
	return &Server{db: database, store: store, river: rc, delivery: d, keys: kw,
		adminKey: adminKey, metrics: m, live: lv, playerURL: playerURL,
		webOrigins: acc.WebOrigins, sessionDomain: acc.SessionDomain,
		sessionSecure: acc.SessionSecure,
		// Ten attempts a minute from one address: generous for a person, useless for
		// a dictionary.
		authLimiter: newAuthLimiter(10, time.Minute)}
}

type ctxKey string

const (
	// The tenant key is httpx's, so a module can read who is calling without
	// importing the control plane.
	tenantKey = httpx.TenantKey
	// Set only when the caller authenticated with a session cookie. An API key is a
	// machine credential: it must never be able to invite a person or change a role.
	userKey ctxKey = "user_id"
)

// liveEnabled is about this deployment: with no ingest host there is nowhere for an
// encoder to connect, so cmd/ builds no module and the endpoints are not served at
// all rather than served and always failing.
func (s *Server) liveEnabled() bool { return s.live != nil }

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

	// The ingest server asks whether a publisher may write to a stream. It holds no
	// API key, so this is unauthenticated and must be bound to a private interface --
	// the stream key inside the request is the credential. See deploy/README.md.
	if s.liveEnabled() {
		s.live.IngestRoutes(r)
	}

	// Operator surface, behind a separate credential so a leaked customer key cannot
	// mint tenants or more keys.
	r.Route("/admin", func(r chi.Router) {
		r.Use(s.adminOnly)
		r.Post("/tenants", s.createTenant)
		r.Get("/tenants", s.listTenants)
		r.Post("/tenants/{id}/keys", s.issueKey)
		r.Delete("/keys/{keyID}", s.revokeKey)
		r.Put("/tenants/{id}/live", s.setTenantLive)
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
			// Redeeming an invite is signup for an account that already exists, and
			// it takes a password, so it is rate limited with the other two.
			r.Get("/invite", s.getInvite)
			r.With(s.rateLimit).Post("/invite", s.acceptInvite)
		})
		// The contact form is not an account endpoint, but it is the same shape:
		// a browser, no credential, and a reason to rate limit.
		r.Group(func(r chi.Router) {
			r.Use(s.cors, s.rateLimit)
			r.Post("/v1/contact", s.postContact)
		})
	}

	// Public: a viewer with no account and no key still has to see the logo.
	r.Get("/brand/{tenant}/logo", s.serveBrandLogo)

	// The embed. Authorised by the playback signature, because the caller is a
	// viewer's browser and holds no key.
	if s.embedEnabled() {
		r.Get("/e/{tenant}/{asset}", s.serveEmbed)
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
		r.Patch("/assets/{id}", s.patchAsset)
		r.Delete("/assets/{id}", s.deleteAsset)
		r.Get("/assets", s.listAssets)
		r.Get("/assets/{id}/chunks", s.listChunks)
		r.Get("/assets/{id}/activity", s.listActivity)
		r.Post("/assets", s.createAssetFromURL)
		r.Post("/webhooks", s.createWebhook)
		r.Get("/webhooks", s.listWebhooks)
		r.Get("/webhooks/{id}", s.getWebhook)
		r.Patch("/webhooks/{id}", s.patchWebhook)
		r.Delete("/webhooks/{id}", s.deleteWebhook)
		r.Get("/usage", s.getUsage)
		r.Post("/bucket-sources", s.createBucketSource)
		r.Get("/bucket-sources", s.listBucketSources)
		r.Get("/bucket-sources/{id}", s.getBucketSource)
		r.Patch("/bucket-sources/{id}", s.patchBucketSource)
		r.Delete("/bucket-sources/{id}", s.deleteBucketSource)
		r.Get("/keys", s.listKeys)
		r.Post("/keys", s.createKey)
		r.Patch("/keys/{id}", s.patchKey)
		r.Delete("/keys/{id}", s.deleteKey)
		r.Get("/branding", s.getBranding)
		r.Get("/ladder-profiles", s.listProfiles)
		r.Get("/playback-settings", s.getPlayback)
		r.Get("/migration-providers", s.listProviders)
		r.Get("/migrations", s.listMigrations)
		r.Get("/migrations/{id}/items", s.listMigrationItems)
		// Live is mounted only when an ingest host is configured: without somewhere
		// for an encoder to connect, the endpoints could only ever fail. Whether this
		// particular tenant bought it is a separate question, asked per request.
		if s.liveEnabled() {
			s.live.Routes(r)
		}

		r.Post("/edits", s.createEdit)
		r.Get("/edits", s.listEdits)
		r.Delete("/edits/{id}", s.deleteEdit)

		// Account administration. Session only — see requireSession.
		r.Group(func(r chi.Router) {
			r.Use(s.requireSession)
			r.Get("/members", s.listMembers)
			r.Post("/members/invites", s.createInvite)
			r.Delete("/members/invites/{id}", s.deleteInvite)
			r.Patch("/members/{id}", s.updateMember)
			r.Delete("/members/{id}", s.removeMember)
			r.Put("/ladder-profile", s.setProfile)
			r.Put("/playback-settings", s.setPlayback)
			// Handing over another service's API key is account administration, not
			// something a machine credential should be able to do.
			r.Post("/migrations", s.createMigration)
			r.Post("/migrations/{id}/confirm", s.confirmMigration)
			r.Post("/migrations/{id}/pause", s.pauseMigration)
			r.Post("/migrations/{id}/resume", s.resumeMigration)
			r.Delete("/migrations/{id}", s.deleteMigration)
			r.Put("/branding/logo", s.putBrandingLogo)
			r.Delete("/branding/logo", s.deleteBrandingLogo)
		})
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
			if tenantID, userID, ok := s.tenantFromSession(r); ok {
				ctx := context.WithValue(r.Context(), tenantKey, tenantID)
				ctx = context.WithValue(ctx, userKey, userID)
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
	var live bool
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(r.Context(),
			`select name, ladder_profile from tenants`).Scan(&name, &profile); err != nil {
			return err
		}
		// Both have to be true to mean anything: a tenant who bought Live still cannot
		// use it on a deployment with no ingest host, and a dashboard that offered it
		// would be sending them at endpoints that are not mounted.
		err := tx.QueryRow(r.Context(),
			`select live_enabled from tenant_limits`).Scan(&live)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"tenant_id": tenantID, "name": name, "ladder_profile": profile,
		"live_enabled": live && s.liveEnabled(),
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
