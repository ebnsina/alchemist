package api

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"strings"
)

// Publish authorization for the ingest server.
//
// The ingest server terminates SRT and RTMP and asks before accepting a publisher.
// That indirection is the entire point: ffmpeg's own listener accepts one connection
// on a port and never exposes SRT's streamid, so it cannot check a stream key -- which
// left the port itself as the credential, and anyone who reached an armed port could
// publish to a customer's stream.
//
// Unauthenticated by design, like /internal/verify-playback: the caller holds no API
// key. It must be bound to a private interface. See deploy/README.md.
//
// The rule and the wire shape are separated deliberately. Ingest servers disagree
// about how to ask and how to be answered -- MediaMTX reads 204 as yes and 401 as no,
// while SRS requires 200 with a body of "0" and treats a bare 204 as a failure -- so
// one endpoint cannot serve both. Each gets a small handler; they share the rule.

// publishAllowed is the rule, and it is the whole of it.
//
// Two things must hold, not one: the key has to resolve to a live stream, and that
// stream has to be the path being published to. Checking only the key would let a
// customer with one valid key publish over every other stream on the server.
func (s *Server) publishAllowed(ctx context.Context, key, path string) bool {
	if key == "" || path == "" {
		return false
	}
	sum := sha256.Sum256([]byte(key))

	// Runs with no tenant in scope, so it goes through the definer function; RLS is
	// forced on live_streams and a plain select would return no rows.
	var streamID string
	err := s.db.Pool().QueryRow(ctx,
		`select coalesce(resolve_stream_key($1)::text, '')`, sum[:]).Scan(&streamID)
	if err != nil || streamID == "" {
		return false
	}
	return strings.EqualFold(strings.Trim(path, "/"), streamID)
}

// mediamtxAuthRequest is MediaMTX's payload. Only the fields that decide the answer
// are read; the rest are ignored rather than rejected, so a newer version that adds a
// field does not start failing every publish.
type mediamtxAuthRequest struct {
	Password string `json:"password"`
	Token    string `json:"token"`
	Action   string `json:"action"`
	Path     string `json:"path"`
	Protocol string `json:"protocol"`
	IP       string `json:"ip"`
}

// authorizeIngest answers MediaMTX: 204 to allow, 401 to refuse.
func (s *Server) authorizeIngest(w http.ResponseWriter, r *http.Request) {
	var req mediamtxAuthRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&req); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Playback never comes from the ingest server -- it is served by the origin, with
	// a signed URL. Anything but a publish is refused whatever key it carries.
	if req.Action != "publish" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	key := req.Password
	if key == "" {
		key = req.Token
	}
	if !s.publishAllowed(r.Context(), key, req.Path) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
