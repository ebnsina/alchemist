package delivery

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// verifyPlayback authorizes a playback request without returning any content. It
// exists for edges and CDNs that cannot run njs, which call it as a subrequest
// (nginx auth_request) before consulting their cache.
//
// This is security-critical rather than a convenience. The media cache key
// deliberately excludes the signature -- including it would give every viewer their
// own copy and collapse the hit rate to nothing -- so a cached object is served to
// anyone who asks for that URI. Authorization therefore MUST happen before the cache
// lookup. Without it, the first valid request warms the cache and every later
// unsigned request is served straight out of it.
func (m *Module) verifyPlayback(w http.ResponseWriter, r *http.Request) {
	// The original request path arrives via X-Original-URI (nginx auth_request sets
	// $request_uri there); fall back to this request's own query for direct calls.
	target := r.Header.Get("X-Original-URI")
	if target == "" {
		target = r.URL.RequestURI()
	}

	path, rawQuery, _ := strings.Cut(target, "?")
	prefix, ok := assetPrefix(path)
	if !ok {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	q, err := url.ParseQuery(rawQuery)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	lock := q.Get("org")
	if !m.verify(prefix, q.Get("kid"), q.Get("sig"), q.Get("exp"), q.Get("vid"), q.Get("wm"), lock) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	// The subrequest carries the viewer's own headers, so the lock is enforced here --
	// before the cache lookup, which is the only place it can be enforced at all for a
	// slice that is already warm.
	if !originAllowed(r, lock) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// assetPrefix reduces /playback/{tenant}/{asset}/{file} to the prefix a signature
// covers, so one token authorizes a manifest and everything beneath it. This must
// agree with assetPrefix() in deploy/edge/playback_auth.js.
func assetPrefix(path string) (string, bool) {
	parts := strings.Split(path, "/")
	if len(parts) < 5 || parts[1] != "playback" || parts[2] == "" || parts[3] == "" {
		return "", false
	}
	return fmt.Sprintf("/playback/%s/%s", parts[2], parts[3]), true
}
