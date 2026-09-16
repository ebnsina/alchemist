package live

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
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

// newStreamKey mints a key and the hash that is all we keep of it.
func newStreamKey() (string, []byte, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", nil, err
	}
	key := hex.EncodeToString(raw)
	sum := sha256.Sum256([]byte(key))
	return key, sum[:], nil
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
	// older hardware encoders speak nothing else, and camera is a browser publishing
	// its own webcam for a customer who has no encoder at all.
	if req.Protocol == "" {
		req.Protocol = "srt"
	}
	if req.Protocol != "srt" && req.Protocol != "rtmp" && req.Protocol != "camera" {
		httpx.ErrorFor(w, r, http.StatusBadRequest, "invalid_protocol",
			"Choose camera, srt or rtmp.")
		return
	}

	key, sum, err := newStreamKey()
	if err != nil {
		httpx.ErrorFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}

	var out stream
	err = m.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`insert into live_streams (tenant_id, name, key_hash, protocol)
			 values ($1,$2,$3,$4)
			 returning id::text, name, protocol, state, created_at`,
			tenantID, req.Name, sum, req.Protocol).
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

	page, ok := httpx.ParseList(w, r, []string{"created_at", "name", "state"}, "created_at")
	if !ok {
		return
	}
	states := r.URL.Query()["state"]
	protocol := r.URL.Query().Get("protocol")

	const where = `from live_streams
	  where ($1::text = '' or name ilike '%' || $1::text || '%'
	                       or id::text like lower($1::text) || '%')
	    and (coalesce(cardinality($2::text[]), 0) = 0 or state = any($2::text[]))
	    and ($3::text = '' or protocol = $3::text)`

	out := []stream{}
	var total int
	err := m.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(),
			`select id::text, name, protocol, state, created_at `+where+
				` order by `+page.OrderBy()+` limit $4 offset $5`,
			page.Q, states, protocol, page.Limit, page.Offset)
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
		if err := rows.Err(); err != nil {
			return err
		}
		return tx.QueryRow(r.Context(), `select count(*) `+where,
			page.Q, states, protocol).Scan(&total)
	})
	if err != nil {
		httpx.ErrorFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"live_streams": out, "total": total})
}

// patchStream renames a stream, at any time, including mid broadcast.
//
// Nothing in flight reads the name: the encoder is authorised by the key hash and
// the broadcast is watched at its own asset, so a rename is a label change and
// refusing one during a live class would only make the dashboard lie.
func (m *Module) patchStream(w http.ResponseWriter, r *http.Request) {
	tenantID := httpx.Tenant(r)

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&req); err != nil {
		httpx.ErrorFor(w, r, http.StatusBadRequest, "invalid_request",
			"Send a JSON body with a name for the stream.")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || len(req.Name) > 120 {
		httpx.ErrorFor(w, r, http.StatusBadRequest, "invalid_request",
			"Give the stream a name of up to 120 characters.")
		return
	}

	var l stream
	err := m.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`update live_streams set name = $2 where id = $1
			 returning id::text, name, protocol, state, created_at`,
			chi.URLParam(r, "id"), req.Name).
			Scan(&l.ID, &l.Name, &l.Protocol, &l.State, &l.CreatedAt)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.ErrorFor(w, r, http.StatusNotFound, "stream_not_found",
			"We couldn't find that stream.")
		return
	}
	if err != nil {
		httpx.ErrorFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	httpx.JSON(w, http.StatusOK, l)
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

	key, sum, err := newStreamKey()
	if err != nil {
		httpx.ErrorFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}

	var state string
	err = m.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`update live_streams set key_hash = $2, updated_at = now()
			  where id = $1 returning state`, streamID, sum).Scan(&state)
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

	// A browser is the encoder for a camera stream, so it needs a key it can send --
	// and the one minted at creation was shown once and never stored. Nobody pasted
	// it into an encoder either, so a fresh key per broadcast costs nothing and keeps
	// the credential in page script alive for exactly one session.
	var publishToken string
	var sum []byte
	if protocol == "camera" {
		publishToken, sum, err = newStreamKey()
		if err != nil {
			httpx.ErrorFor(w, r, http.StatusInternalServerError, "internal_error",
				"We couldn't start that stream.")
			return
		}
	}

	var sessionID string
	err = m.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		if _, err := tx.Exec(r.Context(),
			`update live_streams set state = 'armed', updated_at = now(),
			        key_hash = coalesce($2, key_hash)
			  where id = $1`, streamID, sum); err != nil {
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
	out := map[string]any{
		"stream_id":  streamID,
		"session_id": sessionID,
		"asset_id":   assetID,
		"ingest_url": publishURL(protocol, m.ingestHost, streamID),
	}
	if protocol == "camera" {
		// WHIP reads credentials from the Authorization header and nowhere else, so
		// this is sent as "Bearer publisher:<token>" rather than joined into the URL.
		out["publish_token"] = publishToken
	} else {
		// OBS and most encoders split this into two fields and join them with a
		// slash. Handing over one URL gets the key appended a second time, which
		// publishes to a path nothing authorised -- so the two halves are named.
		out["ingest_server"] = publishURL(protocol, m.ingestHost, streamID)
		out["ingest_stream_key"] = ""
	}
	httpx.JSON(w, http.StatusAccepted, out)
}

// stopStream ends a broadcast that is on air.
//
// It writes the intent rather than doing the work: the thing holding ffmpeg is the
// worker, in another process, and the API has no handle on it. The worker reads this
// on its next poll -- within a second -- and goes down its ordinary ending path, so a
// stopped broadcast converts into its recording exactly like one whose encoder hung
// up. Nothing here touches a session that has already ended, which is what keeps a
// recording mid-conversion from being stranded.
func (m *Module) stopStream(w http.ResponseWriter, r *http.Request) {
	tenantID := httpx.Tenant(r)
	streamID := chi.URLParam(r, "id")

	// The stream is looked up first so a typo in the id is a 404 rather than "not
	// broadcasting", which would send someone hunting for a broadcast that never
	// existed. Both reads are in one transaction: a stream deleted between them
	// would otherwise report nothing to stop instead of gone.
	var sessionID string
	var known bool
	err := m.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(r.Context(),
			`select exists (select 1 from live_streams where id = $1)`,
			streamID).Scan(&known); err != nil {
			return err
		}
		if !known {
			return pgx.ErrNoRows
		}
		return tx.QueryRow(r.Context(),
			`update live_sessions set stop_requested_at = now()
			  where id = (select id from live_sessions
			               where stream_id = $1 and state in ('waiting', 'live')
			               order by created_at desc limit 1)
			 returning id::text`, streamID).Scan(&sessionID)
	})
	if errors.Is(err, pgx.ErrNoRows) && !known {
		httpx.ErrorFor(w, r, http.StatusNotFound, "stream_not_found",
			"We couldn't find that stream.")
		return
	}
	if errors.Is(err, pgx.ErrNoRows) {
		// The stream exists but is not on air: never started, or already over with
		// its recording converting. Both are "there is nothing to stop".
		httpx.ErrorFor(w, r, http.StatusConflict, "not_broadcasting",
			"That stream isn't on air, so there's nothing to stop.")
		return
	}
	if err != nil {
		httpx.ErrorFor(w, r, http.StatusInternalServerError, "internal_error",
			"We couldn't stop that broadcast.")
		return
	}

	httpx.JSON(w, http.StatusAccepted, map[string]any{
		"stream_id":  streamID,
		"session_id": sessionID,
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
