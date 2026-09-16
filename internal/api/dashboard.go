package api

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/ebnsina/alchemist/internal/platform/httpx"
)

// What the customer's own dashboard needs and an integration does not: a list of
// assets to look at, and keys to hand to their code.
//
// Both authenticate the same way as the rest of /v1 — an API key, or a session
// cookie from a configured origin.

// tenantFromSession resolves the session cookie, but only for a request that came
// from an origin we published the dashboard on. Without that check the cookie would
// authorise any site that can make the browser send it.
func (s *Server) tenantFromSession(r *http.Request) (tenant, user string, ok bool) {
	if !s.authEnabled() {
		return "", "", false
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		// Same-origin GETs arrive without Origin in some browsers. Sec-Fetch-Site is
		// the modern signal and cannot be set by script.
		if site := r.Header.Get("Sec-Fetch-Site"); site != "same-origin" && site != "none" {
			return "", "", false
		}
	} else {
		listed := false
		for _, o := range s.webOrigins {
			if o == origin {
				listed = true
				break
			}
		}
		if !listed {
			return "", "", false
		}
	}

	c, err := r.Cookie(sessionCookie)
	if err != nil || c.Value == "" {
		return "", "", false
	}
	sum := sha256.Sum256([]byte(c.Value))
	var userID, tenantID, email, org string
	if err := s.db.Pool().QueryRow(r.Context(),
		`select user_id::text, tenant_id::text, email::text, org_name from auth_session($1)`,
		sum[:]).Scan(&userID, &tenantID, &email, &org); err != nil {
		return "", "", false
	}
	return tenantID, userID, true
}

// requireSession rejects an API key on the endpoints that administer the account
// itself. A key is something a server holds; handing it the power to add a teammate
// turns a leaked key into a permanent way back in.
func (s *Server) requireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.Context().Value(userKey).(string); !ok {
			writeErrFor(w, r, http.StatusForbidden, "session_required",
				"Sign in to change this. An API key cannot.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

type assetRow struct {
	ID        string   `json:"id"`
	Title     *string  `json:"title"`
	State     string   `json:"state"`
	ErrorCode *string  `json:"error_code"`
	Duration  *float64 `json:"duration_sec"`
	SizeBytes *int64   `json:"source_bytes"`
	Height    *int     `json:"height"`
	CreatedAt string   `json:"created_at"`
}

func (s *Server) listAssets(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	page, ok := httpx.ParseList(w, r,
		[]string{"created_at", "duration_sec", "source_bytes", "state"}, "created_at")
	if !ok {
		return
	}
	states := r.URL.Query()["state"]

	// The page and the count read the same predicate, so a filter can never be
	// applied to one and forgotten on the other.
	const where = `from assets
	  where ($1::text = '' or title ilike '%' || $1::text || '%'
	                       or id::text like lower($1::text) || '%')
	    and (coalesce(cardinality($2::text[]), 0) = 0 or state::text = any($2::text[]))`

	out := []assetRow{}
	var total int
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		// No tenant predicate: row-level security is the boundary, and writing one
		// here would suggest it is not.
		rows, err := tx.Query(r.Context(),
			`select id::text, title, state::text, error_code, duration_sec, source_bytes,
			        height, created_at `+where+
				` order by `+page.OrderBy()+` limit $3 offset $4`,
			page.Q, states, page.Limit, page.Offset)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var a assetRow
			var created time.Time
			if err := rows.Scan(&a.ID, &a.Title, &a.State, &a.ErrorCode, &a.Duration,
				&a.SizeBytes, &a.Height, &created); err != nil {
				return err
			}
			a.CreatedAt = created.UTC().Format(time.RFC3339)
			out = append(out, a)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		return tx.QueryRow(r.Context(), `select count(*) `+where, page.Q, states).Scan(&total)
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"assets": out, "total": total})
}

// patchAsset names a video. Title is the only field: everything else about an asset
// is produced by the pipeline, not chosen.
func (s *Server) patchAsset(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	var req struct {
		Title *string `json:"title"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&req); err != nil {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request",
			"Send a JSON body with a \"title\".")
		return
	}
	title, ok := cleanTitle(req.Title)
	if !ok {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_title",
			"That name is too long. Keep it under 200 characters.")
		return
	}

	var out assetRow
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		// RLS scopes the update, so another tenant's id matches nothing and the
		// handler answers 404 -- the same answer as an id that never existed.
		return tx.QueryRow(r.Context(),
			`update assets set title = $2, updated_at = now() where id = $1
			 returning id::text, title, state::text`, chi.URLParam(r, "id"), title).
			Scan(&out.ID, &out.Title, &out.State)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		writeErrFor(w, r, http.StatusNotFound, "asset_not_found", "We couldn't find that video.")
		return
	}
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id": out.ID, "title": out.Title, "state": out.State,
	})
}

// cleanTitle trims a supplied name. ok is false only when it is too long to store;
// absent and empty both mean unnamed, which renders as the short id.
func cleanTitle(v *string) (*string, bool) {
	if v == nil {
		return nil, true
	}
	t := strings.TrimSpace(*v)
	switch {
	case len(t) > 200:
		return nil, false
	case t == "":
		return nil, true
	}
	return &t, true
}

type keyRow struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	CreatedAt string  `json:"created_at"`
	RevokedAt *string `json:"revoked_at"`
}

func (s *Server) listKeys(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	page, ok := httpx.ParseList(w, r, []string{"created_at", "name"}, "created_at")
	if !ok {
		return
	}
	revoked, ok := httpx.Flag(w, r, "revoked")
	if !ok {
		return
	}

	const where = `from api_keys
	  where ($1::text = '' or name ilike '%' || $1::text || '%')
	    and ($2::bool is null or (revoked_at is not null) = $2::bool)`

	out := []keyRow{}
	var total int
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(),
			`select id::text, name, created_at, revoked_at `+where+
				` order by `+page.OrderBy()+` limit $3 offset $4`,
			page.Q, revoked, page.Limit, page.Offset)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var k keyRow
			var created time.Time
			var revoked *time.Time
			if err := rows.Scan(&k.ID, &k.Name, &created, &revoked); err != nil {
				return err
			}
			k.CreatedAt = created.UTC().Format(time.RFC3339)
			if revoked != nil {
				v := revoked.UTC().Format(time.RFC3339)
				k.RevokedAt = &v
			}
			out = append(out, k)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		return tx.QueryRow(r.Context(), `select count(*) `+where, page.Q, revoked).Scan(&total)
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	// Never the key itself: only its hash is stored, so there is nothing to return.
	writeJSON(w, http.StatusOK, map[string]any{"keys": out, "total": total})
}

// patchKey renames a key. The secret is untouched, so code holding it keeps working:
// the name is only how a person tells one key from another.
func (s *Server) patchKey(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&req); err != nil {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request",
			"Send a JSON body with a \"name\".")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || len(req.Name) > 60 {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request",
			"Give the key a name of up to 60 characters.")
		return
	}

	var k keyRow
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`update api_keys set name = $2 where id = $1 returning id::text, name`,
			chi.URLParam(r, "id"), req.Name).Scan(&k.ID, &k.Name)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		writeErrFor(w, r, http.StatusNotFound, "key_not_found", "We couldn't find that key.")
		return
	}
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": k.ID, "name": k.Name})
}

func (s *Server) createKey(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	var req struct {
		Name string `json:"name"`
	}
	_ = json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&req)
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		req.Name = "Untitled key"
	}
	if len(req.Name) > 60 {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request", "That name is too long.")
		return
	}

	key, hash, err := newAPIKey()
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	var id string
	err = s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`insert into api_keys (tenant_id, name, key_hash)
			 values ($1::uuid, $2, $3) returning id::text`, tenantID, req.Name, hash[:]).Scan(&id)
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{
		"id": id, "name": req.Name, "api_key": key,
	})
}

func (s *Server) deleteKey(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)
	id := chi.URLParam(r, "id")

	var revoked bool
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		// RLS scopes the update, so another tenant's id simply matches nothing.
		tag, err := tx.Exec(r.Context(),
			`update api_keys set revoked_at = now()
			  where id = $1::uuid and revoked_at is null`, id)
		if err != nil {
			return err
		}
		revoked = tag.RowsAffected() > 0
		return nil
	})
	if err != nil {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request", "That key id is not valid.")
		return
	}
	if !revoked {
		writeErrFor(w, r, http.StatusNotFound, "not_found", "We couldn't find that key.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
