package delivery

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"path"
	"strings"

	"github.com/ebnsina/alchemist/internal/platform/httpx"

	"github.com/go-chi/chi/v5"
)

// servePlayback is the origin. Getting these headers right is what makes the edge
// cache and any commercial CDN behave identically in front of it.
func (m *Module) servePlayback(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenant")
	assetID := chi.URLParam(r, "asset")
	file := path.Base(chi.URLParam(r, "file"))

	prefix := fmt.Sprintf("/playback/%s/%s", tenantID, assetID)
	q := r.URL.Query()
	if !m.verify(prefix, q.Get("kid"), q.Get("sig"), q.Get("exp")) {
		httpx.ErrorFor(w, r, http.StatusForbidden, "playback_not_authorized",
			"This playback link has expired or is not valid.")
		return
	}

	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	key := m.prefix(r.Context(), tenantID, assetID) + "/" + file
	obj, err := m.store.GetPassthrough(r.Context(), key, r.Header.Get("Range"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.ErrorFor(w, r, http.StatusNotFound, "not_found", "We couldn't find that file.")
			return
		}
		if errors.Is(err, ErrRangeNotSatisfiable) {
			httpx.ErrorFor(w, r, http.StatusRequestedRangeNotSatisfiable, "range_not_satisfiable",
				"That part of the file does not exist.")
			return
		}
		httpx.ErrorFor(w, r, http.StatusServiceUnavailable, "storage_unavailable",
			"The video is temporarily unavailable. Please try again shortly.")
		return
	}
	defer obj.Body.Close()

	// Media is content-addressed and never mutates, so it is cacheable forever.
	// Rewritten files get a short TTL rather than none, so an edge still absorbs a
	// burst -- caching them for a year would serve one viewer's signature to everyone.
	if isRewritten(file) {
		w.Header().Set("Cache-Control", "public, max-age=2")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	}
	w.Header().Set("Content-Type", contentTypeFor(file))
	w.Header().Set("Accept-Ranges", "bytes")
	if obj.ETag != "" {
		w.Header().Set("ETag", obj.ETag)
	}
	if obj.ContentRange != "" {
		w.Header().Set("Content-Range", obj.ContentRange)
	}
	if obj.ContentLength > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", obj.ContentLength))
	}

	// A conditional request that still matches costs no bytes at all.
	if match := r.Header.Get("If-None-Match"); match != "" && match == obj.ETag && !isRewritten(file) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// A master playlist request is the first thing a player does, so it is the
	// signal that this asset is actually being watched and its deferred renditions
	// are worth generating.
	if file == "master.m3u8" && m.observer != nil && r.Method == http.MethodGet {
		m.observer.OnPlaybackStarted(r.Context(), tenantID, assetID)
	}

	// A playlist read is one viewer present now: players re-read it every segment
	// duration, and it is the only request every viewer makes and shares with nobody.
	if isManifest(file) && r.Method == http.MethodGet {
		m.meterViewer(tenantID, assetID, viewerKey(r))
	}

	// Manifests and the scrubbing index are small and must be rewritten so their
	// relative URIs stay authorized; media is streamed through untouched.
	if isRewritten(file) {
		body, err := io.ReadAll(obj.Body)
		if err != nil {
			httpx.ErrorFor(w, r, http.StatusServiceUnavailable, "storage_unavailable",
				"The video is temporarily unavailable. Please try again shortly.")
			return
		}
		signed := signManifest(body, r.URL.RawQuery, file)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(signed)))
		w.WriteHeader(http.StatusOK)
		if r.Method != http.MethodHead {
			_, _ = w.Write(signed)
			m.meterEgress(tenantID, int64(len(signed)))
		}
		return
	}

	status := http.StatusOK
	if obj.ContentRange != "" {
		status = http.StatusPartialContent
	}
	w.WriteHeader(status)
	if r.Method != http.MethodHead {
		n, _ := io.Copy(w, obj.Body)
		m.meterEgress(tenantID, n)
	}
}

// isRewritten lists the files whose contents reference other files by relative path,
// and so have to carry the signature forward.
func isRewritten(name string) bool {
	return strings.HasSuffix(name, ".m3u8") || strings.HasSuffix(name, ".mpd") ||
		strings.HasSuffix(name, ".vtt")
}

func setCORS(w http.ResponseWriter) {
	h := w.Header()
	h.Set("Access-Control-Allow-Origin", "*")
	h.Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
	h.Set("Access-Control-Allow-Headers", "Range, If-None-Match")
	// Players need these to reason about byte ranges.
	h.Set("Access-Control-Expose-Headers", "Content-Length, Content-Range, ETag, Accept-Ranges")
	h.Set("Access-Control-Max-Age", "86400")
}

func contentTypeFor(name string) string {
	switch {
	case strings.HasSuffix(name, ".m3u8"):
		return "application/vnd.apple.mpegurl"
	case strings.HasSuffix(name, ".mpd"):
		return "application/dash+xml"
	case strings.HasSuffix(name, ".jpg"):
		return "image/jpeg"
	case strings.HasSuffix(name, ".vtt"):
		return "text/vtt"
	default:
		return "video/mp4"
	}
}

// isManifest is the subset of rewritten files a player polls: the playlists. The
// scrubbing index is fetched once and is not a sign anyone is still watching.
func isManifest(name string) bool {
	return strings.HasSuffix(name, ".m3u8") || strings.HasSuffix(name, ".mpd")
}

// viewerKey identifies one viewer well enough to count them, without storing anything
// that identifies a person: the address and user agent are hashed together and the
// digest is all that is ever held.
//
// ponytail: carrier-grade NAT puts many mobile viewers behind one address, so this
// under-counts on exactly the network most BD viewers use. The player's beacon already
// carries a real per-viewer session_id (internal/api/qoe.go) -- feed that in when the
// number has to be exact rather than indicative.
func viewerKey(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	sum := sha256.Sum256([]byte(host + "\x00" + r.Header.Get("User-Agent")))
	return hex.EncodeToString(sum[:16])
}
