package api

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/ebnsina/alchemist/internal/platform/httpx"
)

// The platform administrator: one Alchemist account that supports a customer by
// acting as them, not a second application that reads every tenant at once.
//
// Impersonation is resolved in auth_session, where a cookie already becomes a
// tenant, so every existing endpoint keeps working unchanged and a support engineer
// sees exactly what the customer sees -- a parallel admin view would drift from it
// the first week. While impersonating the session is read-only: the guard is in
// authenticate, so no endpoint can opt out of it.
//
// The one thing staff see that a customer does not is the raw queue error. A
// failure carries whatever ffmpeg or the storage client said -- paths, hostnames,
// internal arguments -- so /v1/assets/{id}/activity withholds it and returns the
// stable error_code; the diagnostics endpoints here return it untouched.

var uuidLike = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// staffOnly refuses with 403 and a code, never 404 and never an empty list: at 2am
// the difference between "no data" and "not allowed" is the whole debugging session.
// The flag comes from the session read on this request, so revoking it takes effect
// immediately rather than when the session expires.
func (s *Server) staffOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if admin, _ := r.Context().Value(staffKey).(bool); !admin {
			writeErrFor(w, r, http.StatusForbidden, "staff_only",
				"This is for Alchemist staff.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

type staffTenantRow struct {
	ID          string `json:"tenant_id"`
	Name        string `json:"name"`
	Assets      int64  `json:"assets"`
	LiveStreams int64  `json:"live_streams"`
	Members     int64  `json:"members"`
	CreatedAt   string `json:"created_at"`
}

// staffTenants is the picker impersonation starts from: names and counts, enough to
// recognise an account, and no customer content at all.
func (s *Server) staffTenants(w http.ResponseWriter, r *http.Request) {
	page, ok := httpx.ParseList(w, r, []string{"created_at", "name", "assets"}, "created_at")
	if !ok {
		return
	}
	caller, _ := r.Context().Value(userKey).(string)

	const from = `from staff_tenants($1, $2)`
	rows, err := s.db.Pool().Query(r.Context(),
		`select id::text, name, assets, live_streams, members, created_at `+from+
			` order by `+page.OrderBy()+` limit $3 offset $4`,
		caller, page.Q, page.Limit, page.Offset)
	if err != nil {
		s.staffFailed(w, r)
		return
	}
	defer rows.Close()

	out := []staffTenantRow{}
	for rows.Next() {
		var t staffTenantRow
		var created time.Time
		if err := rows.Scan(&t.ID, &t.Name, &t.Assets, &t.LiveStreams, &t.Members,
			&created); err != nil {
			s.staffFailed(w, r)
			return
		}
		t.CreatedAt = created.UTC().Format(time.RFC3339)
		out = append(out, t)
	}
	if rows.Err() != nil {
		s.staffFailed(w, r)
		return
	}

	// The page and the count read the same function with the same arguments, so a
	// search can never be applied to one and forgotten on the other.
	var total int
	if err := s.db.Pool().QueryRow(r.Context(),
		`select count(*) `+from, caller, page.Q).Scan(&total); err != nil {
		s.staffFailed(w, r)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tenants": out, "total": total})
}

// startImpersonation writes the acting tenant onto the session row. It is stored
// server-side rather than sent per request so nothing a browser holds can widen
// what a support session sees.
func (s *Server) startImpersonation(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TenantID string `json:"tenant_id"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&req); err != nil {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request",
			"Send a JSON body with a \"tenant_id\".")
		return
	}
	if !uuidLike.MatchString(req.TenantID) {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request",
			"Send a JSON body with a \"tenant_id\".")
		return
	}
	if !s.setActingTenant(w, r, &req.TenantID) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"tenant_id": req.TenantID, "impersonating": true, "read_only": true,
	})
}

func (s *Server) stopImpersonation(w http.ResponseWriter, r *http.Request) {
	if !s.setActingTenant(w, r, nil) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"impersonating": false})
}

func (s *Server) setActingTenant(w http.ResponseWriter, r *http.Request, tenantID *string) bool {
	c, err := r.Cookie(sessionCookie)
	if err != nil || c.Value == "" {
		writeErrFor(w, r, http.StatusForbidden, "session_required",
			"Sign in to change this. An API key cannot.")
		return false
	}
	sum := sha256.Sum256([]byte(c.Value))

	var done bool
	err = s.db.Pool().QueryRow(r.Context(),
		`select coalesce(staff_impersonate($1, $2), false)`, sum[:], tenantID).Scan(&done)
	if err != nil {
		s.staffFailed(w, r)
		return false
	}
	if !done {
		// The flag was verified this request, so the only thing left is the tenant.
		writeErrFor(w, r, http.StatusNotFound, "tenant_not_found",
			"We couldn't find that account.")
		return false
	}
	return true
}

// jobRow is a river_job row as it is, error text included.
type jobRow struct {
	ID          int64           `json:"id"`
	Kind        string          `json:"kind"`
	State       string          `json:"state"`
	Attempt     int             `json:"attempt"`
	MaxAttempts int             `json:"max_attempts"`
	Queue       string          `json:"queue"`
	QueuedAt    string          `json:"queued_at"`
	ScheduledAt *string         `json:"scheduled_at"`
	StartedAt   *string         `json:"started_at"`
	FinishedAt  *string         `json:"finished_at"`
	Errors      json.RawMessage `json:"errors"`
}

// assetJobs reads every queue row for one asset. river_job is not tenant-scoped, so
// nothing here may run until the asset has been read through RLS.
func (s *Server) assetJobs(r *http.Request, assetID string) ([]jobRow, error) {
	rows, err := s.db.Pool().Query(r.Context(),
		`select id, kind, state::text, attempt, max_attempts, queue, created_at,
		        scheduled_at, attempted_at, finalized_at,
		        coalesce(to_jsonb(errors), '[]'::jsonb)
		   from river_job
		  where args->>'asset_id' = $1
		  order by created_at`, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []jobRow{}
	for rows.Next() {
		var j jobRow
		var queued time.Time
		var scheduled, started, finished *time.Time
		if err := rows.Scan(&j.ID, &j.Kind, &j.State, &j.Attempt, &j.MaxAttempts, &j.Queue,
			&queued, &scheduled, &started, &finished, &j.Errors); err != nil {
			return nil, err
		}
		j.QueuedAt = queued.UTC().Format(time.RFC3339)
		j.ScheduledAt, j.StartedAt, j.FinishedAt =
			rfc3339(scheduled), rfc3339(started), rfc3339(finished)
		out = append(out, j)
	}
	return out, rows.Err()
}

// assetDiagnostics is listActivity with the error text left in. Same tenant scope as
// any other endpoint, so impersonation is what points it at a customer's video.
func (s *Server) assetDiagnostics(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)
	assetID := chi.URLParam(r, "id")
	if !uuidLike.MatchString(assetID) {
		writeErrFor(w, r, http.StatusNotFound, "asset_not_found", "We couldn't find that video.")
		return
	}

	var state string
	var errorCode, dedup, mediaPrefix *string
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`select state::text, error_code, deduplicated_from::text, media_prefix
			   from assets where id = $1`, assetID).
			Scan(&state, &errorCode, &dedup, &mediaPrefix)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		writeErrFor(w, r, http.StatusNotFound, "asset_not_found", "We couldn't find that video.")
		return
	}
	if err != nil {
		s.staffFailed(w, r)
		return
	}

	jobs, err := s.assetJobs(r, assetID)
	if err != nil {
		s.staffFailed(w, r)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"asset": map[string]any{
			"id": assetID, "tenant_id": tenantID, "state": state, "error_code": errorCode,
			"deduplicated_from": dedup, "media_prefix": mediaPrefix,
		},
		"jobs": jobs,
	})
}

type sessionDiagnostics struct {
	ID              string  `json:"id"`
	StreamID        string  `json:"stream_id"`
	StreamName      string  `json:"stream_name"`
	StreamProtocol  string  `json:"stream_protocol"`
	StreamState     string  `json:"stream_state"`
	AssetID         string  `json:"asset_id"`
	AssetState      string  `json:"asset_state"`
	State           string  `json:"state"`
	ErrorCode       *string `json:"error_code"`
	StartedAt       *string `json:"started_at"`
	EndedAt         *string `json:"ended_at"`
	LastSeenAt      *string `json:"last_seen_at"`
	StopRequestedAt *string `json:"stop_requested_at"`
	SweptAt         *string `json:"swept_at"`
	CreatedAt       string  `json:"created_at"`
}

// liveSessionDiagnostics answers the question a broadcast that failed always raises:
// did the encoder ever arrive. last_seen_at is when a segment last landed, which is
// the difference between "nothing was sent" and "it stopped mid-match".
func (s *Server) liveSessionDiagnostics(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)
	id := chi.URLParam(r, "id")
	if !uuidLike.MatchString(id) {
		writeErrFor(w, r, http.StatusNotFound, "session_not_found",
			"We couldn't find that broadcast.")
		return
	}

	var v sessionDiagnostics
	var started, ended, seen, stop, swept *time.Time
	var created time.Time
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`select s.id::text, s.stream_id::text, ls.name, ls.protocol, ls.state,
			        s.asset_id::text, a.state::text, s.state, s.error_code,
			        s.started_at, s.ended_at, s.last_seen_at, s.stop_requested_at,
			        s.swept_at, s.created_at
			   from live_sessions s
			   join live_streams ls on ls.id = s.stream_id
			   join assets a on a.id = s.asset_id
			  where s.id = $1`, id).
			Scan(&v.ID, &v.StreamID, &v.StreamName, &v.StreamProtocol, &v.StreamState,
				&v.AssetID, &v.AssetState, &v.State, &v.ErrorCode, &started, &ended,
				&seen, &stop, &swept, &created)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		writeErrFor(w, r, http.StatusNotFound, "session_not_found",
			"We couldn't find that broadcast.")
		return
	}
	if err != nil {
		s.staffFailed(w, r)
		return
	}
	v.StartedAt, v.EndedAt, v.LastSeenAt = rfc3339(started), rfc3339(ended), rfc3339(seen)
	v.StopRequestedAt, v.SweptAt = rfc3339(stop), rfc3339(swept)
	v.CreatedAt = created.UTC().Format(time.RFC3339)

	jobs, err := s.assetJobs(r, v.AssetID)
	if err != nil {
		s.staffFailed(w, r)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"session": v, "jobs": jobs})
}

func rfc3339(t *time.Time) *string {
	if t == nil {
		return nil
	}
	v := t.UTC().Format(time.RFC3339)
	return &v
}

func (s *Server) staffFailed(w http.ResponseWriter, r *http.Request) {
	writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
		"Something went wrong on our side.")
}
