package api

import (
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"strings"
)

// Publish authorization for the ingest server.
//
// The ingest server terminates SRT and RTMP and asks this endpoint whether a
// publisher may proceed. That indirection is the entire point: ffmpeg's own listener
// accepts one connection on a port and never exposes SRT's streamid, so it cannot
// check a stream key -- which left the port itself as the credential, and anyone who
// reached an armed port could publish to a customer's stream.
//
// Unauthenticated by design, like /internal/verify-playback: the caller is the ingest
// server, which holds no API key. It must be bound to a private interface. See
// deploy/README.md.

// ingestAuthRequest is the ingest server's published payload. Only the fields that
// decide the answer are read; the rest are ignored rather than rejected, so a newer
// ingest server that adds a field does not start failing every publish.
type ingestAuthRequest struct {
	Password string `json:"password"`
	Token    string `json:"token"`
	Action   string `json:"action"`
	Path     string `json:"path"`
	Protocol string `json:"protocol"`
	IP       string `json:"ip"`
}

// authorizeIngest answers whether this publisher may write to this path.
//
// Two things must hold, not one: the key has to resolve to a live stream, and that
// stream has to be the path being published to. Checking only the key would let a
// customer with one valid key publish over any other stream on the server.
func (s *Server) authorizeIngest(w http.ResponseWriter, r *http.Request) {
	var req ingestAuthRequest
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
	if key == "" || req.Path == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	sum := sha256.Sum256([]byte(key))

	// Runs with no tenant in scope, so it goes through the definer function; RLS is
	// forced on live_streams and a plain select would return no rows.
	var streamID string
	err := s.db.Pool().QueryRow(r.Context(),
		`select coalesce(resolve_stream_key($1)::text, '')`, sum[:]).Scan(&streamID)
	if err != nil || streamID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if !strings.EqualFold(strings.Trim(req.Path, "/"), streamID) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
