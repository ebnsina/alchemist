package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func accountServer() *Server {
	return &Server{
		webOrigins:  []string{"https://app.example"},
		authLimiter: newAuthLimiter(3, time.Minute),
	}
}

// Preflight has to answer before any handler runs, or a browser never sends the
// real request and signup fails with no error anyone can see.
func TestPreflightAnswersForAllowedOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/v1/auth/signup", nil)
	req.Header.Set("Origin", "https://app.example")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	accountServer().Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("preflight status = %d, want 204", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example" {
		t.Fatalf("allow-origin = %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("allow-credentials = %q, want true", got)
	}
}

// An unlisted origin must get no CORS header at all. A credentialed wildcard is the
// one mistake that turns every account endpoint into a cross-site one.
func TestUnlistedOriginGetsNoCORSHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/v1/auth/login", nil)
	req.Header.Set("Origin", "https://evil.example")
	rec := httptest.NewRecorder()
	accountServer().Routes().ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("allow-origin = %q, want empty for an unlisted origin", got)
	}
}

// With no configured origin the surface must not exist. Signup reachable by default
// would mean any deployment that forgot the variable is open for account creation.
func TestAccountRoutesAbsentWithoutOrigins(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/signup",
		strings.NewReader(`{"org":"x","email":"a@b.co","password":"0123456789"}`))
	rec := httptest.NewRecorder()
	(&Server{}).Routes().ServeHTTP(rec, req)

	if rec.Code == http.StatusCreated || rec.Code == http.StatusOK {
		t.Fatalf("signup answered %d with no configured origin", rec.Code)
	}
}

func TestRateLimitStopsGuessing(t *testing.T) {
	s := accountServer()
	var last int
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login",
			strings.NewReader(`{"email":"a@b.co","password":"nope"}`))
		req.RemoteAddr = "203.0.113.9:1234"
		rec := httptest.NewRecorder()
		// The handler would need a database; the limiter answers before it.
		func() {
			defer func() { _ = recover() }()
			s.Routes().ServeHTTP(rec, req)
		}()
		last = rec.Code
	}
	if last != http.StatusTooManyRequests {
		t.Fatalf("after 5 attempts status = %d, want 429", last)
	}
}

func TestLimiterIsPerAddress(t *testing.T) {
	l := newAuthLimiter(2, time.Minute)
	for i := 0; i < 2; i++ {
		if !l.allow("10.0.0.1") {
			t.Fatal("first two attempts must pass")
		}
	}
	if l.allow("10.0.0.1") {
		t.Fatal("third attempt from the same address must be refused")
	}
	if !l.allow("10.0.0.2") {
		t.Fatal("a different address must not inherit the refusal")
	}
}

// A session must not authorise a request that arrives from somewhere we did not
// publish the dashboard. The cookie is attached by the browser regardless of who
// asked, which is the whole shape of CSRF.
func TestSessionRejectedFromUnlistedOrigin(t *testing.T) {
	s := accountServer()
	req := httptest.NewRequest(http.MethodGet, "/v1/keys", nil)
	req.Header.Set("Origin", "https://evil.example")
	req.AddCookie(&http.Cookie{Name: sessionCookie, Value: "whatever"})

	if _, ok := s.tenantFromSession(req); ok {
		t.Fatal("session accepted from an unlisted origin")
	}
}

// A cross-site GET carries Sec-Fetch-Site: cross-site and no Origin. Accepting it
// because Origin was absent would undo the check above.
func TestSessionRejectedForCrossSiteFetch(t *testing.T) {
	s := accountServer()
	req := httptest.NewRequest(http.MethodGet, "/v1/keys", nil)
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	req.AddCookie(&http.Cookie{Name: sessionCookie, Value: "whatever"})

	if _, ok := s.tenantFromSession(req); ok {
		t.Fatal("session accepted for a cross-site fetch")
	}
}

// With no accounts surface configured, a cookie means nothing at all.
func TestSessionIgnoredWhenAccountsDisabled(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/keys", nil)
	req.Header.Set("Origin", "https://app.example")
	req.AddCookie(&http.Cookie{Name: sessionCookie, Value: "whatever"})

	if _, ok := (&Server{}).tenantFromSession(req); ok {
		t.Fatal("session accepted with no configured origin")
	}
}
