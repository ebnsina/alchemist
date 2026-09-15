package delivery

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ebnsina/alchemist/internal/platform/httpx"

	"github.com/go-chi/chi/v5"
)

// clearKey is an EME Clear Key licence, the shape a browser's key session expects
// back. Chrome, Firefox and Edge all implement it, and it needs no licence vendor.
type clearKey struct {
	Keys []clearKeyEntry `json:"keys"`
	Type string          `json:"type"`
}

type clearKeyEntry struct {
	KTY string `json:"kty"`
	KID string `json:"kid"`
	K   string `json:"k"`
}

// serveContentKey hands the decryption key to an authorized player.
//
// It is deliberately the same signature that authorizes the manifest and segments:
// a viewer who cannot fetch the media cannot fetch the key either. That is also the
// ceiling -- the key reaches the browser in the clear, so this is encryption, not
// DRM, and what it buys is that a lifted bucket or backup decodes to nothing.
func (m *Module) serveContentKey(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenant")
	assetID := chi.URLParam(r, "asset")

	prefix := fmt.Sprintf("/playback/%s/%s", tenantID, assetID)
	q := r.URL.Query()
	if !m.verify(prefix, q.Get("kid"), q.Get("sig"), q.Get("exp"), q.Get("vid"), q.Get("wm")) {
		httpx.ErrorFor(w, r, http.StatusForbidden, "playback_not_authorized",
			"This playback link has expired or is not valid.")
		return
	}

	setKeyCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// shaka POSTs the EME licence challenge here. Clear Key needs nothing from it --
	// the answer is the same key either way -- but it is read bounded rather than
	// left for the connection to stall on.
	if r.Body != nil {
		_, _ = io.Copy(io.Discard, http.MaxBytesReader(w, r.Body, 64<<10))
	}

	keyID, key, err := m.keys.Get(r.Context(), tenantID, assetID)
	if errors.Is(err, ErrNotFound) {
		httpx.ErrorFor(w, r, http.StatusNotFound, "not_found", "We couldn't find that file.")
		return
	}
	if err != nil {
		httpx.ErrorFor(w, r, http.StatusServiceUnavailable, "storage_unavailable",
			"The video is temporarily unavailable. Please try again shortly.")
		return
	}

	// WebKit has no Clear Key at all: FairPlay is its only key system. Saying so is
	// the whole point -- an Apple viewer would otherwise get a black screen and no
	// reason for it.
	//
	// ponytail: the ceiling is that Apple devices cannot play encrypted assets here.
	// FairPlay needs an Apple certificate and a licence server; when one exists,
	// answer this request with an SPC/CKC exchange instead of refusing.
	if appleOnlyFairPlay(r.Header.Get("User-Agent")) {
		httpx.ErrorFor(w, r, http.StatusForbidden, "browser_not_supported",
			"This video needs Chrome, Firefox or Edge. Safari cannot play it yet.")
		return
	}

	body, err := json.Marshal(clearKey{
		Keys: []clearKeyEntry{{
			KTY: "oct",
			KID: base64.RawURLEncoding.EncodeToString(keyID),
			K:   base64.RawURLEncoding.EncodeToString(key),
		}},
		Type: "temporary",
	})
	if err != nil {
		httpx.ErrorFor(w, r, http.StatusServiceUnavailable, "storage_unavailable",
			"The video is temporarily unavailable. Please try again shortly.")
		return
	}

	// Never cached: the key's protection is the short-lived signature on the request.
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		_, _ = w.Write(body)
	}
}

// setKeyCORS differs from the media one by POST: shaka sends the licence request as
// a JSON POST from the customer's page, which is a real preflight rather than a
// simple request, so OPTIONS has to answer for it or every encrypted asset fails to
// play with the origin looking healthy.
func setKeyCORS(w http.ResponseWriter) {
	h := w.Header()
	h.Set("Access-Control-Allow-Origin", "*")
	h.Set("Access-Control-Allow-Methods", "GET, HEAD, POST, OPTIONS")
	h.Set("Access-Control-Allow-Headers", "Content-Type")
	h.Set("Access-Control-Max-Age", "86400")
}

// appleOnlyFairPlay matches the engines whose only key system is FairPlay. Every iOS
// browser is WebKit, Chrome on an iPhone included, so the platform decides this and
// the browser name does not.
func appleOnlyFairPlay(ua string) bool {
	if strings.Contains(ua, "iPhone") || strings.Contains(ua, "iPad") || strings.Contains(ua, "iPod") {
		return true
	}
	// Android first: an old WebView says "Mobile Safari" with no Chrome token, and
	// that is the low-end BD handset this product exists for. Blocking it would be
	// the worst possible false positive.
	if strings.Contains(ua, "Android") {
		return false
	}
	if !strings.Contains(ua, "Safari") {
		return false
	}
	for _, other := range []string{"Chrome", "Chromium", "Edg/", "OPR/", "Firefox"} {
		if strings.Contains(ua, other) {
			return false
		}
	}
	return true
}
