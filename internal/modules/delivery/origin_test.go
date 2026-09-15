package delivery

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ebnsina/alchemist/internal/platform/signing"
)

type fakeStore struct{ body string }

func (f fakeStore) GetPassthrough(context.Context, string, string) (*Object, error) {
	return &Object{Body: io.NopCloser(strings.NewReader(f.body)), ContentLength: int64(len(f.body))}, nil
}

type fakeKeys struct{}

func (fakeKeys) Get(context.Context, string, string) ([]byte, error) {
	return nil, ErrNotFound
}
func (fakeKeys) Put(context.Context, string, string, []byte, []byte) error { return nil }

// fakeMeter refuses every device it has not already seen, and records whether it was
// asked at all.
type fakeMeter struct {
	asked bool
	allow bool
}

func (m *fakeMeter) RecordEgress(string, int64)          {}
func (m *fakeMeter) RecordViewer(string, string, string) {}
func (m *fakeMeter) AllowViewer(context.Context, string, string, string) bool {
	m.asked = true
	return m.allow
}

func origin(t *testing.T, m *Module) (*chi.Mux, *signing.Keyring) {
	t.Helper()
	r := chi.NewRouter()
	m.Routes(r)
	return r, nil
}

func signedURL(t *testing.T, kr *signing.Keyring, path, viewer, label string) string {
	t.Helper()
	exp := time.Now().Add(time.Hour).Unix()
	prefix, _ := assetPrefix(path)
	kid, sig := kr.Sign(prefix, exp, viewer, label)
	q := "exp=" + itoa(exp) + "&kid=" + kid + "&sig=" + sig
	if viewer != "" {
		q += "&vid=" + viewer
	}
	if label != "" {
		q += "&wm=" + label
	}
	return path + "?" + q
}

func itoa(i int64) string {
	return time.Unix(i, 0).UTC().Format("") + strings.TrimSpace(jsonNumber(i))
}

func jsonNumber(i int64) string {
	b, _ := json.Marshal(i)
	return string(b)
}

func keyring(t *testing.T) *signing.Keyring {
	t.Helper()
	kr, err := signing.NewKeyring([]string{"k1:a-playback-secret-of-at-least-32-bytes"})
	if err != nil {
		t.Fatal(err)
	}
	return kr
}

// The device cap is what answers one login shared with a class, so a refusal has to
// reach the viewer as a stable code rather than as a broken player.
func TestDeviceCapRefusesBoundPlaylist(t *testing.T) {
	kr := keyring(t)
	meter := &fakeMeter{allow: false}
	m := New(fakeStore{"#EXTM3U\n"}, fakeKeys{}, kr).WithMeter(meter)
	r, _ := origin(t, m)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet,
		signedURL(t, kr, "/playback/t1/a1/master.m3u8", "student-1", ""), nil))

	if w.Code != http.StatusForbidden {
		t.Fatalf("status %d, want 403", w.Code)
	}
	if !strings.Contains(w.Body.String(), "viewer_limit_reached") {
		t.Errorf("body %q does not carry the stable code", w.Body.String())
	}
}

// An unbound link must not consult the cap at all: it identifies nobody, so there is
// nothing to count, and charging it against some other viewer's allowance would lock
// people out of links that were never bound.
func TestUnboundPlaylistSkipsTheCap(t *testing.T) {
	kr := keyring(t)
	meter := &fakeMeter{allow: false}
	m := New(fakeStore{"#EXTM3U\n"}, fakeKeys{}, kr).WithMeter(meter)
	r, _ := origin(t, m)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet,
		signedURL(t, kr, "/playback/t1/a1/master.m3u8", "", ""), nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", w.Code)
	}
	if meter.asked {
		t.Error("an unbound link was counted against a viewer cap")
	}
}

// A signature covers the binding, so a viewer who edits their id out of the URL to
// escape the cap is left with a link the origin refuses outright.
func TestStrippedBindingIsRefused(t *testing.T) {
	kr := keyring(t)
	m := New(fakeStore{"#EXTM3U\n"}, fakeKeys{}, kr).WithMeter(&fakeMeter{allow: true})
	r, _ := origin(t, m)

	bound := signedURL(t, kr, "/playback/t1/a1/master.m3u8", "student-1", "")
	stripped := strings.Replace(bound, "&vid=student-1", "", 1)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, stripped, nil))
	if w.Code != http.StatusForbidden {
		t.Fatalf("status %d, want 403 for a link with its binding removed", w.Code)
	}
}
