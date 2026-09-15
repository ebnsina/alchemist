package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// The whole point of requireSession: an API key is a machine credential, and a
// machine must never be able to invite a person or change a role. A leaked key that
// could add an owner would be a permanent way back into the account.
func TestRequireSessionRejectsAPIKeyAuth(t *testing.T) {
	reached := false
	h := (&Server{}).requireSession(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		reached = true
	}))

	// authenticate puts a tenant in context for a key, and a tenant *and* a user for
	// a session. A key therefore looks exactly like this.
	req := httptest.NewRequest(http.MethodGet, "/v1/members", nil)
	req = req.WithContext(context.WithValue(req.Context(), tenantKey, "t-1"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if reached {
		t.Fatal("handler ran for a request authenticated with an API key")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestRequireSessionAllowsSessionAuth(t *testing.T) {
	reached := false
	h := (&Server{}).requireSession(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		reached = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/members", nil)
	ctx := context.WithValue(req.Context(), tenantKey, "t-1")
	ctx = context.WithValue(ctx, userKey, "u-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req.WithContext(ctx))

	if !reached {
		t.Fatalf("handler did not run for a session request; status %d", rec.Code)
	}
}

// A declared Content-Type is the client's claim. Accepting it on trust is how a
// file that is not an image ends up served from our origin as one.
func TestLogoSniffRejectsMismatchedBytes(t *testing.T) {
	png := []byte("\x89PNG\r\n\x1a\nrest")
	jpg := []byte{0xFF, 0xD8, 0xFF, 0x00}
	webp := append([]byte("RIFF....WEBP"), 0x00)
	html := []byte("<svg onload=alert(1)>")

	for _, tc := range []struct {
		name string
		body []byte
		ext  string
		want bool
	}{
		{"png as png", png, "png", true},
		{"jpg as jpg", jpg, "jpg", true},
		{"webp as webp", webp, "webp", true},
		{"png claimed as webp", png, "webp", false},
		{"markup claimed as png", html, "png", false},
		{"empty", nil, "png", false},
	} {
		if got := sniffMatches(tc.body, tc.ext); got != tc.want {
			t.Errorf("%s: sniffMatches = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// SVG is a document that can carry script. It must not be an accepted logo type.
func TestLogoTypesExcludeSVG(t *testing.T) {
	if _, ok := logoTypes["image/svg+xml"]; ok {
		t.Fatal("image/svg+xml is accepted as a logo type")
	}
}
