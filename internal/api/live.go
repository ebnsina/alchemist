package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/ebnsina/alchemist/internal/pipeline"
	"github.com/ebnsina/alchemist/internal/platform/media"
)

// Live. A stream is an asset from the moment it is armed, so playback URLs, the
// signed prefix, the origin and the player need nothing live-specific. See
// docs/06-live.md.

// Live carries what the live surface needs. An empty host leaves it unmounted.
type Live struct {
	IngestHost string
}

// liveEnabled is about this deployment: with no ingest host there is nowhere for an
// encoder to connect, so the endpoints are not served at all rather than served and
// always failing.
func (s *Server) liveEnabled() bool { return s.live.IngestHost != "" }

// requireLive is about this tenant. Live is a separate product, so a VOD-only
// customer reaching these endpoints gets a clear "not on your plan" rather than a
// stream they were never sold.
//
// Absent limits row means absent entitlement: enabling live has to be deliberate, or
// deploying an ingest host would quietly hand it to every tenant on the box.
func (s *Server) requireLive(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID, _ := r.Context().Value(tenantKey).(string)

		var enabled bool
		err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
			err := tx.QueryRow(r.Context(),
				`select live_enabled from tenant_limits`).Scan(&enabled)
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			return err
		})
		if err != nil {
			writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
				"Something went wrong on our side.")
			return
		}
		if !enabled {
			writeErrFor(w, r, http.StatusForbidden, "live_not_enabled",
				"Live streaming isn't part of your plan yet. Talk to us and we'll turn it on.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

type liveStream struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Protocol  string    `json:"protocol"`
	State     string    `json:"state"`
	AssetID   string    `json:"asset_id,omitempty"`
	IngestURL string    `json:"ingest_url,omitempty"`
	StreamKey string    `json:"stream_key,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type createLiveStreamRequest struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
}

// createLiveStream mints the stream key and returns it exactly once, like an API key.
func (s *Server) createLiveStream(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	var req createLiveStreamRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&req); err != nil {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request",
			"Send a JSON body with a name for the stream.")
		return
	}
	if req.Name == "" {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request", "Give the stream a name.")
		return
	}
	// SRT is the default because it survives a lossy uplink; RTMP is there because
	// older hardware encoders speak nothing else.
	if req.Protocol == "" {
		req.Protocol = "srt"
	}
	if req.Protocol != "srt" && req.Protocol != "rtmp" {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_protocol",
			"Choose srt or rtmp.")
		return
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	key := hex.EncodeToString(raw)
	sum := sha256.Sum256([]byte(key))

	var out liveStream
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`insert into live_streams (tenant_id, name, key_hash, protocol)
			 values ($1,$2,$3,$4)
			 returning id::text, name, protocol, state, created_at`,
			tenantID, req.Name, sum[:], req.Protocol).
			Scan(&out.ID, &out.Name, &out.Protocol, &out.State, &out.CreatedAt)
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"We couldn't create that stream.")
		return
	}
	// Shown once and never again: only its hash is stored.
	out.StreamKey = key
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) listLiveStreams(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	out := []liveStream{}
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(),
			`select id::text, name, protocol, state, created_at
			   from live_streams order by created_at desc limit 200`)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var l liveStream
			if err := rows.Scan(&l.ID, &l.Name, &l.Protocol, &l.State, &l.CreatedAt); err != nil {
				return err
			}
			out = append(out, l)
		}
		return rows.Err()
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"live_streams": out})
}

func (s *Server) getLiveStream(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	var l liveStream
	var assetID *string
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`select l.id::text, l.name, l.protocol, l.state, l.created_at,
			        (select s.asset_id::text from live_sessions s
			          where s.stream_id = l.id order by s.created_at desc limit 1)
			   from live_streams l where l.id = $1`, chi.URLParam(r, "id")).
			Scan(&l.ID, &l.Name, &l.Protocol, &l.State, &l.CreatedAt, &assetID)
	})
	if err == pgx.ErrNoRows {
		writeErrFor(w, r, http.StatusNotFound, "stream_not_found", "We couldn't find that stream.")
		return
	}
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	if assetID != nil {
		l.AssetID = *assetID
	}
	writeJSON(w, http.StatusOK, l)
}

// startLiveStream arms a stream: it takes a port, creates the asset the broadcast
// will be watched at, and queues the worker that waits for the encoder.
//
// Arming is explicit so a port and a worker slot are held only for a stream somebody
// intends to use, rather than for every stream ever created.
func (s *Server) startLiveStream(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)
	streamID := chi.URLParam(r, "id")

	var state, protocol, profile string
	var assetID, sessionID string
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(r.Context(),
			`select l.state, l.protocol, t.ladder_profile
			   from live_streams l, tenants t where l.id = $1`, streamID).
			Scan(&state, &protocol, &profile); err != nil {
			return err
		}
		if state == "armed" || state == "live" {
			return nil
		}
		if err := tx.QueryRow(r.Context(),
			`insert into assets (tenant_id, ladder_profile, state)
			 values ($1, $2, 'live') returning id::text`,
			tenantID, profile).Scan(&assetID); err != nil {
			return err
		}
		if _, err := tx.Exec(r.Context(),
			`update live_streams set state = 'armed', updated_at = now()
			  where id = $1`, streamID); err != nil {
			return err
		}
		return tx.QueryRow(r.Context(),
			`insert into live_sessions (stream_id, tenant_id, asset_id)
			 values ($1,$2,$3) returning id::text`, streamID, tenantID, assetID).Scan(&sessionID)
	})
	switch {
	case err == pgx.ErrNoRows && state == "":
		writeErrFor(w, r, http.StatusNotFound, "stream_not_found", "We couldn't find that stream.")
		return
	case err != nil:
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"We couldn't start that stream.")
		return
	case sessionID == "":
		writeErrFor(w, r, http.StatusConflict, "stream_busy",
			"That stream is already waiting for an encoder.")
		return
	}

	if _, err := s.river.Insert(r.Context(),
		pipeline.LiveArgs{SessionID: sessionID, TenantID: tenantID}, nil); err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"We couldn't start that stream.")
		return
	}

	// The ingest server checks the key against this exact path before it accepts a
	// publisher, so the URL carries no secret and is safe to show and to log.
	writeJSON(w, http.StatusAccepted, map[string]any{
		"stream_id":  streamID,
		"session_id": sessionID,
		"asset_id":   assetID,
		"ingest_url": media.LivePublishURL(protocol, s.live.IngestHost, streamID),
	})
}

func (s *Server) deleteLiveStream(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	var tag int64
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		ct, err := tx.Exec(r.Context(), `delete from live_streams where id = $1`,
			chi.URLParam(r, "id"))
		tag = ct.RowsAffected()
		return err
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	if tag == 0 {
		writeErrFor(w, r, http.StatusNotFound, "stream_not_found", "We couldn't find that stream.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
