package live

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

	"github.com/ebnsina/alchemist/internal/platform/httpx"
)

// requireLive is about this tenant. Live is a separate product, so a VOD-only
// customer reaching these endpoints gets a clear "not on your plan" rather than a
// stream they were never sold.
//
// Absent limits row means absent entitlement: enabling live has to be deliberate, or
// deploying an ingest host would quietly hand it to every tenant on the box.
func (m *Module) requireLive(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := httpx.Tenant(r)

		var enabled bool
		err := m.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
			err := tx.QueryRow(r.Context(),
				`select live_enabled from tenant_limits`).Scan(&enabled)
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			return err
		})
		if err != nil {
			httpx.ErrorFor(w, r, http.StatusInternalServerError, "internal_error",
				"Something went wrong on our side.")
			return
		}
		if !enabled {
			httpx.ErrorFor(w, r, http.StatusForbidden, "live_not_enabled",
				"Live streaming isn't part of your plan yet. Talk to us and we'll turn it on.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

type stream struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Protocol  string    `json:"protocol"`
	State     string    `json:"state"`
	AssetID   string    `json:"asset_id,omitempty"`
	IngestURL string    `json:"ingest_url,omitempty"`
	StreamKey string    `json:"stream_key,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type createStreamRequest struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
}

// createStream mints the stream key and returns it exactly once, like an API key.
func (m *Module) createStream(w http.ResponseWriter, r *http.Request) {
	tenantID := httpx.Tenant(r)

	var req createStreamRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&req); err != nil {
		httpx.ErrorFor(w, r, http.StatusBadRequest, "invalid_request",
			"Send a JSON body with a name for the stream.")
		return
	}
	if req.Name == "" {
		httpx.ErrorFor(w, r, http.StatusBadRequest, "invalid_request", "Give the stream a name.")
		return
	}
	// SRT is the default because it survives a lossy uplink; RTMP is there because
	// older hardware encoders speak nothing else.
	if req.Protocol == "" {
		req.Protocol = "srt"
	}
	if req.Protocol != "srt" && req.Protocol != "rtmp" {
		httpx.ErrorFor(w, r, http.StatusBadRequest, "invalid_protocol",
			"Choose srt or rtmp.")
		return
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		httpx.ErrorFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	key := hex.EncodeToString(raw)
	sum := sha256.Sum256([]byte(key))

	var out stream
	err := m.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`insert into live_streams (tenant_id, name, key_hash, protocol)
			 values ($1,$2,$3,$4)
			 returning id::text, name, protocol, state, created_at`,
			tenantID, req.Name, sum[:], req.Protocol).
			Scan(&out.ID, &out.Name, &out.Protocol, &out.State, &out.CreatedAt)
	})
	if err != nil {
		httpx.ErrorFor(w, r, http.StatusInternalServerError, "internal_error",
			"We couldn't create that stream.")
		return
	}
	// Shown once and never again: only its hash is stored.
	out.StreamKey = key
	httpx.JSON(w, http.StatusCreated, out)
}

func (m *Module) listStreams(w http.ResponseWriter, r *http.Request) {
	tenantID := httpx.Tenant(r)

	out := []stream{}
	err := m.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(),
			`select id::text, name, protocol, state, created_at
			   from live_streams order by created_at desc limit 200`)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var l stream
			if err := rows.Scan(&l.ID, &l.Name, &l.Protocol, &l.State, &l.CreatedAt); err != nil {
				return err
			}
			out = append(out, l)
		}
		return rows.Err()
	})
	if err != nil {
		httpx.ErrorFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"live_streams": out})
}

func (m *Module) getStream(w http.ResponseWriter, r *http.Request) {
	tenantID := httpx.Tenant(r)

	var l stream
	var assetID *string
	err := m.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`select l.id::text, l.name, l.protocol, l.state, l.created_at,
			        (select s.asset_id::text from live_sessions s
			          where s.stream_id = l.id order by s.created_at desc limit 1)
			   from live_streams l where l.id = $1`, chi.URLParam(r, "id")).
			Scan(&l.ID, &l.Name, &l.Protocol, &l.State, &l.CreatedAt, &assetID)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.ErrorFor(w, r, http.StatusNotFound, "stream_not_found", "We couldn't find that stream.")
		return
	}
	if err != nil {
		httpx.ErrorFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	if assetID != nil {
		l.AssetID = *assetID
	}
	httpx.JSON(w, http.StatusOK, l)
}

// startStream arms a stream: it creates the asset the broadcast will be watched at
// and queues the worker that waits for the encoder.
//
// Arming is explicit so a worker slot is held only for a stream somebody intends to
// use, rather than for every stream ever created.
//
// The asset is created between the two steps rather than inside one transaction,
// because live does not own assets: the busy check comes first so a refused start
// never leaves one behind.
// replaceKey mints a new stream key and forgets the old one.
//
// There is no "show me the key again": only its hash is stored, exactly like an API
// key, so a key nobody wrote down is gone. Replacing it is the honest answer, and it
// doubles as the fix for a leaked one.
func (m *Module) replaceKey(w http.ResponseWriter, r *http.Request) {
	tenantID := httpx.Tenant(r)
	streamID := chi.URLParam(r, "id")

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		httpx.ErrorFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	key := hex.EncodeToString(raw)
	sum := sha256.Sum256([]byte(key))

	var state string
	err := m.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`update live_streams set key_hash = $2, updated_at = now()
			  where id = $1 returning state`, streamID, sum[:]).Scan(&state)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.ErrorFor(w, r, http.StatusNotFound, "stream_not_found", "We couldn't find that stream.")
		return
	}
	if err != nil {
		httpx.ErrorFor(w, r, http.StatusInternalServerError, "internal_error",
			"We couldn't replace that key.")
		return
	}

	// An encoder already publishing keeps going: the ingest server checked the key at
	// handshake and does not re-check mid-connection. It is the next connection that
	// needs the new one, which is what the copy says.
	httpx.JSON(w, http.StatusOK, map[string]any{
		"stream_id":  streamID,
		"stream_key": key,
		"state":      state,
	})
}

func (m *Module) startStream(w http.ResponseWriter, r *http.Request) {
	tenantID := httpx.Tenant(r)
	streamID := chi.URLParam(r, "id")

	var state, protocol string
	err := m.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`select state, protocol from live_streams where id = $1`, streamID).
			Scan(&state, &protocol)
	})
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		httpx.ErrorFor(w, r, http.StatusNotFound, "stream_not_found", "We couldn't find that stream.")
		return
	case err != nil:
		httpx.ErrorFor(w, r, http.StatusInternalServerError, "internal_error",
			"We couldn't start that stream.")
		return
	case state == "armed" || state == "live":
		httpx.ErrorFor(w, r, http.StatusConflict, "stream_busy",
			"That stream is already waiting for an encoder.")
		return
	}

	assetID, err := m.assets.CreateForBroadcast(r.Context(), tenantID)
	if err != nil {
		httpx.ErrorFor(w, r, http.StatusInternalServerError, "internal_error",
			"We couldn't start that stream.")
		return
	}

	var sessionID string
	err = m.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		if _, err := tx.Exec(r.Context(),
			`update live_streams set state = 'armed', updated_at = now()
			  where id = $1`, streamID); err != nil {
			return err
		}
		return tx.QueryRow(r.Context(),
			`insert into live_sessions (stream_id, tenant_id, asset_id)
			 values ($1,$2,$3) returning id::text`, streamID, tenantID, assetID).Scan(&sessionID)
	})
	if err != nil {
		httpx.ErrorFor(w, r, http.StatusInternalServerError, "internal_error",
			"We couldn't start that stream.")
		return
	}

	if err := m.queue.EnqueueSession(r.Context(), sessionID, tenantID); err != nil {
		httpx.ErrorFor(w, r, http.StatusInternalServerError, "internal_error",
			"We couldn't start that stream.")
		return
	}

	// The ingest server checks the key against this exact path before it accepts a
	// publisher, so the URL carries no secret and is safe to show and to log.
	httpx.JSON(w, http.StatusAccepted, map[string]any{
		"stream_id":  streamID,
		"session_id": sessionID,
		"asset_id":   assetID,
		"ingest_url": publishURL(protocol, m.ingestHost, streamID),
	})
}

func (m *Module) deleteStream(w http.ResponseWriter, r *http.Request) {
	tenantID := httpx.Tenant(r)

	var tag int64
	err := m.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		ct, err := tx.Exec(r.Context(), `delete from live_streams where id = $1`,
			chi.URLParam(r, "id"))
		tag = ct.RowsAffected()
		return err
	})
	if err != nil {
		httpx.ErrorFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	if tag == 0 {
		httpx.ErrorFor(w, r, http.StatusNotFound, "stream_not_found", "We couldn't find that stream.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
