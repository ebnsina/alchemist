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

type fakeKeys struct{ id, key []byte }

func (f fakeKeys) Get(context.Context, string, string) ([]byte, []byte, error) {
	if f.key == nil {
		return nil, nil, ErrNotFound
	}
	return f.id, f.key, nil
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
	kid, sig := kr.Sign(prefix, exp, viewer, label, "")
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

// The player is built against this exact body. A field renamed or padded base64 here
// is a licence no browser accepts, and it fails as a black screen.
func TestClearKeyLicenceShape(t *testing.T) {
	kr := keyring(t)
	id := []byte{0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88,
		0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11, 0x00}
	m := New(fakeStore{}, fakeKeys{id: id, key: id}, kr)
	r, _ := origin(t, m)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet,
		signedURL(t, kr, "/playback/t1/a1/key", "", ""), nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type %q, want application/json", ct)
	}
	if cc := w.Header().Get("Cache-Control"); cc != "no-store" {
		t.Errorf("Cache-Control %q, want no-store", cc)
	}
	want := `{"keys":[{"kty":"oct","kid":"_-7dzLuqmYh3ZlVEMyIRAA","k":"_-7dzLuqmYh3ZlVEMyIRAA"}],"type":"temporary"}`
	if got := strings.TrimSpace(w.Body.String()); got != want {
		t.Errorf("licence body\n got %s\nwant %s", got, want)
	}
}

// shaka requests the licence with POST and a challenge body. Serving GET only makes
// every encrypted asset fail with a 405 while the origin looks healthy.
func TestClearKeyLicenceOverPOST(t *testing.T) {
	kr := keyring(t)
	m := New(fakeStore{}, fakeKeys{id: []byte{1, 2}, key: []byte{3, 4}}, kr)
	r, _ := origin(t, m)

	req := httptest.NewRequest(http.MethodPost,
		signedURL(t, kr, "/playback/t1/a1/key", "", ""), strings.NewReader("challenge"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"type":"temporary"`) {
		t.Errorf("POST did not return a licence: %q", w.Body.String())
	}
	if allow := w.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(allow, "POST") {
		t.Errorf("Allow-Methods %q does not admit the licence POST", allow)
	}
}

// Safari has no Clear Key, only FairPlay. An Apple viewer gets a reason instead of a
// player that never starts.
func TestSafariToldWhyItCannotPlay(t *testing.T) {
	kr := keyring(t)
	m := New(fakeStore{}, fakeKeys{id: []byte{1}, key: []byte{2}}, kr)
	r, _ := origin(t, m)

	req := httptest.NewRequest(http.MethodGet,
		signedURL(t, kr, "/playback/t1/a1/key", "", ""), nil)
	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 "+
			"(KHTML, like Gecko) Version/17.4 Safari/605.1.15")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden || !strings.Contains(w.Body.String(), "browser_not_supported") {
		t.Fatalf("status %d body %q, want 403 browser_not_supported", w.Code, w.Body.String())
	}
}

func TestAppleOnlyFairPlay(t *testing.T) {
	cases := map[string]bool{
		"Mozilla/5.0 (Macintosh) AppleWebKit/605.1.15 Version/17.4 Safari/605.1.15":      true,
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_4) AppleWebKit/605.1.15 CriOS/122 Mobile": true,
		"Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 Chrome/122.0 Safari/537.36":    false,
		"Mozilla/5.0 (X11; Linux x86_64; rv:124.0) Gecko/20100101 Firefox/124.0":         false,
		"Mozilla/5.0 (Macintosh) AppleWebKit/537.36 Chrome/122.0 Safari/537.36 Edg/122":  false,
		// An old Android WebView says "Mobile Safari" and carries no Chrome token.
		// Blocking it would refuse exactly the low-end handset this product targets.
		"Mozilla/5.0 (Linux; U; Android 4.4.2; SM-G7102) AppleWebKit/534.30 Version/4.0 Mobile Safari/534.30": false,
		"Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 Chrome/120.0 Mobile Safari/537.36":             false,
		"": false,
	}
	for ua, want := range cases {
		if got := appleOnlyFairPlay(ua); got != want {
			t.Errorf("appleOnlyFairPlay(%q) = %v, want %v", ua, got, want)
		}
	}
}
