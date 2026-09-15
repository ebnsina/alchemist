package delivery

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ebnsina/alchemist/internal/platform/httpx"

	"github.com/go-chi/chi/v5"
)

// serveContentKey hands the decryption key to an authorized player.
//
// It is deliberately the same signature that authorizes the manifest and segments:
// a viewer who cannot fetch the media cannot fetch the key either, binding included.
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

	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	key, err := m.keys.Get(r.Context(), tenantID, assetID)
	if errors.Is(err, ErrNotFound) {
		httpx.ErrorFor(w, r, http.StatusNotFound, "not_found", "We couldn't find that file.")
		return
	}
	if err != nil {
		httpx.ErrorFor(w, r, http.StatusServiceUnavailable, "storage_unavailable",
			"The video is temporarily unavailable. Please try again shortly.")
		return
	}

	// Never cached: the key's protection is the short-lived signature on the request.
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(key)))
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		_, _ = w.Write(key)
	}
}
