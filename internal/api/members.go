package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/ebnsina/alchemist/internal/platform/httpx"
	"github.com/ebnsina/alchemist/internal/platform/passwd"
)

// An invite is valid for a week. Long enough to survive a holiday, short enough that
// a link forwarded out of someone's inbox a month later is already dead.
const inviteTTL = 7 * 24 * time.Hour

var knownRoles = map[string]bool{"owner": true, "admin": true, "member": true}

type member struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	Role        string  `json:"role"`
	CreatedAt   string  `json:"created_at"`
	LastLoginAt *string `json:"last_login_at"`
	You         bool    `json:"you"`
}

type invite struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
	ExpiresAt string `json:"expires_at"`
}

// callerRole reads the signed-in user's role. Every mutation here is gated on it,
// and a missing row means the session outlived the user it belonged to.
func (s *Server) callerRole(r *http.Request, tenantID, userID string) (string, error) {
	var role string
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(), `select role from users where id = $1`, userID).Scan(&role)
	})
	return role, err
}

func (s *Server) canAdminister(w http.ResponseWriter, r *http.Request, tenantID, userID string) bool {
	role, err := s.callerRole(r, tenantID, userID)
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return false
	}
	if role != "owner" && role != "admin" {
		writeErrFor(w, r, http.StatusForbidden, "not_permitted",
			"Only an owner or an admin can change who is on the team.")
		return false
	}
	return true
}

func (s *Server) listMembers(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)
	userID, _ := r.Context().Value(userKey).(string)

	page, ok := httpx.ParseList(w, r, []string{"created_at", "email", "role"}, "created_at")
	if !ok {
		return
	}
	role := r.URL.Query().Get("role")
	if role != "" && !knownRoles[role] {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_filter",
			"Filter by role owner, admin or member.")
		return
	}

	// Pending invites are a short list beside the team and are not paged: a page of
	// members with only some of the invites is harder to read, not easier.
	const where = `from users
	  where ($1::text = '' or email::text ilike '%' || $1::text || '%')
	    and ($2::text = '' or role = $2::text)`

	members := []member{}
	invites := []invite{}
	var total int
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(),
			`select id::text, email::text, role,
			        to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
			        to_char(last_login_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
			   `+where+`
			  order by `+page.OrderBy()+` limit $3 offset $4`,
			page.Q, role, page.Limit, page.Offset)
		if err != nil {
			return err
		}
		for rows.Next() {
			var m member
			if err := rows.Scan(&m.ID, &m.Email, &m.Role, &m.CreatedAt, &m.LastLoginAt); err != nil {
				rows.Close()
				return err
			}
			m.You = m.ID == userID
			members = append(members, m)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}
		if err := tx.QueryRow(r.Context(), `select count(*) `+where, page.Q, role).
			Scan(&total); err != nil {
			return err
		}

		irows, err := tx.Query(r.Context(),
			`select id::text, email::text, role,
			        to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
			        to_char(expires_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
			   from invites
			  where accepted_at is null and expires_at > now()
			  order by created_at desc`)
		if err != nil {
			return err
		}
		defer irows.Close()
		for irows.Next() {
			var i invite
			if err := irows.Scan(&i.ID, &i.Email, &i.Role, &i.CreatedAt, &i.ExpiresAt); err != nil {
				return err
			}
			invites = append(invites, i)
		}
		return irows.Err()
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"members": members, "invites": invites, "total": total,
	})
}

type createInviteRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (s *Server) createInvite(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)
	userID, _ := r.Context().Value(userKey).(string)
	if !s.canAdminister(w, r, tenantID, userID) {
		return
	}

	var req createInviteRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&req); err != nil {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request",
			"Something in that request did not come through.")
		return
	}
	email := strings.TrimSpace(req.Email)
	if email == "" || !strings.Contains(email, "@") {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_email",
			"That email address does not look right.")
		return
	}
	role := req.Role
	if role == "" {
		role = "member"
	}
	// Only an owner can mint another owner: an admin promoting itself sideways into
	// ownership is the whole point of having two levels.
	if !knownRoles[role] || role == "owner" {
		callerRole, _ := s.callerRole(r, tenantID, userID)
		if !knownRoles[role] || callerRole != "owner" {
			writeErrFor(w, r, http.StatusBadRequest, "invalid_role",
				"Pick admin or member. Only an owner can add another owner.")
			return
		}
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	sum := sha256.Sum256([]byte(token))
	expires := time.Now().UTC().Add(inviteTTL)

	var id string
	var taken bool
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(r.Context(),
			`select exists (select 1 from users where email = $1)`, email).Scan(&taken); err != nil {
			return err
		}
		if taken {
			return nil
		}
		// A resend replaces the pending invite rather than stacking a second one, so
		// the address only ever has one live link.
		return tx.QueryRow(r.Context(),
			`insert into invites (tenant_id, email, role, token_hash, invited_by, expires_at)
			 values ($1, $2, $3, $4, $5, $6)
			 on conflict (tenant_id, email) where accepted_at is null
			 do update set role = excluded.role, token_hash = excluded.token_hash,
			               invited_by = excluded.invited_by, expires_at = excluded.expires_at,
			               created_at = now()
			 returning id::text`,
			tenantID, email, role, sum[:], userID, expires).Scan(&id)
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"We couldn't send that invite. Please try again.")
		return
	}
	if taken {
		writeErrFor(w, r, http.StatusConflict, "already_a_member",
			"That person is already on the team.")
		return
	}

	// The token is returned once. There is no mail sender in this repo yet, so the
	// dashboard shows the link and whoever invited them passes it on.
	writeJSON(w, http.StatusCreated, map[string]any{
		"id": id, "email": email, "role": role,
		"token":      token,
		"expires_at": expires.Format(time.RFC3339),
	})
}

func (s *Server) deleteInvite(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)
	userID, _ := r.Context().Value(userKey).(string)
	if !s.canAdminister(w, r, tenantID, userID) {
		return
	}
	id := chi.URLParam(r, "id")

	var tag int64
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		ct, err := tx.Exec(r.Context(),
			`delete from invites where id = $1 and accepted_at is null`, id)
		if err != nil {
			return err
		}
		tag = ct.RowsAffected()
		return nil
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	if tag == 0 {
		writeErrFor(w, r, http.StatusNotFound, "not_found", "We couldn't find that invite.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type updateMemberRequest struct {
	Role string `json:"role"`
}

func (s *Server) updateMember(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)
	userID, _ := r.Context().Value(userKey).(string)
	if !s.canAdminister(w, r, tenantID, userID) {
		return
	}
	id := chi.URLParam(r, "id")

	var req updateMemberRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&req); err != nil || !knownRoles[req.Role] {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_role", "Pick owner, admin or member.")
		return
	}
	callerRole, err := s.callerRole(r, tenantID, userID)
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	if req.Role == "owner" && callerRole != "owner" {
		writeErrFor(w, r, http.StatusForbidden, "not_permitted",
			"Only an owner can make someone else an owner.")
		return
	}

	var lastOwner, missing bool
	err = s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		var current string
		if err := tx.QueryRow(r.Context(),
			`select role from users where id = $1`, id).Scan(&current); err != nil {
			if err == pgx.ErrNoRows {
				missing = true
				return nil
			}
			return err
		}
		// Demoting the only owner leaves an account nobody can administer, and
		// nothing in the API can put it back.
		if current == "owner" && req.Role != "owner" {
			var owners int
			if err := tx.QueryRow(r.Context(),
				`select count(*) from users where role = 'owner'`).Scan(&owners); err != nil {
				return err
			}
			if owners <= 1 {
				lastOwner = true
				return nil
			}
		}
		_, err := tx.Exec(r.Context(), `update users set role = $2 where id = $1`, id, req.Role)
		return err
	})
	switch {
	case err != nil:
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
	case missing:
		writeErrFor(w, r, http.StatusNotFound, "not_found", "We couldn't find that person.")
	case lastOwner:
		writeErrFor(w, r, http.StatusConflict, "last_owner",
			"Make someone else an owner first. An account cannot be left without one.")
	default:
		writeJSON(w, http.StatusOK, map[string]string{"id": id, "role": req.Role})
	}
}

func (s *Server) removeMember(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)
	userID, _ := r.Context().Value(userKey).(string)
	if !s.canAdminister(w, r, tenantID, userID) {
		return
	}
	id := chi.URLParam(r, "id")
	if id == userID {
		writeErrFor(w, r, http.StatusConflict, "cannot_remove_self",
			"You cannot remove yourself. Ask another owner to do it.")
		return
	}

	var lastOwner, missing bool
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		var role string
		if err := tx.QueryRow(r.Context(),
			`select role from users where id = $1`, id).Scan(&role); err != nil {
			if err == pgx.ErrNoRows {
				missing = true
				return nil
			}
			return err
		}
		if role == "owner" {
			var owners int
			if err := tx.QueryRow(r.Context(),
				`select count(*) from users where role = 'owner'`).Scan(&owners); err != nil {
				return err
			}
			if owners <= 1 {
				lastOwner = true
				return nil
			}
		}
		// Sessions cascade with the user, so removing someone signs them out
		// everywhere rather than leaving a live cookie behind.
		_, err := tx.Exec(r.Context(), `delete from users where id = $1`, id)
		return err
	})
	switch {
	case err != nil:
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
	case missing:
		writeErrFor(w, r, http.StatusNotFound, "not_found", "We couldn't find that person.")
	case lastOwner:
		writeErrFor(w, r, http.StatusConflict, "last_owner",
			"Make someone else an owner first. An account cannot be left without one.")
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

type acceptInviteRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// getInvite shows who an invite is for, before there is any account to authenticate.
// Unknown and expired are the same answer: a probe must not be able to tell which
// addresses have been invited to which organisation.
func (s *Server) getInvite(w http.ResponseWriter, r *http.Request) {
	sum := sha256.Sum256([]byte(r.URL.Query().Get("token")))

	var email, role, org string
	err := s.db.Pool().QueryRow(r.Context(),
		`select email::text, role, org_name from invite_preview($1)`, sum[:]).
		Scan(&email, &role, &org)
	if err != nil {
		writeErrFor(w, r, http.StatusNotFound, "invite_not_found",
			"That invite has expired or already been used. Ask for a new one.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"email": email, "role": role, "org": org})
}

// acceptInvite turns a token into an account inside an existing tenant. It is the
// one way a second user is ever created, and it signs them straight in.
func (s *Server) acceptInvite(w http.ResponseWriter, r *http.Request) {
	var req acceptInviteRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&req); err != nil {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request",
			"Send a JSON body with token and password.")
		return
	}
	hash, err := passwd.Hash(req.Password)
	if err != nil {
		writeErrFor(w, r, http.StatusBadRequest, "weak_password",
			"Use a password of at least 10 characters.")
		return
	}
	sum := sha256.Sum256([]byte(req.Token))

	var userID, tenantID, email string
	err = s.db.Pool().QueryRow(r.Context(),
		`select user_id::text, tenant_id::text, email::text from invite_redeem($1, $2)`,
		sum[:], hash).Scan(&userID, &tenantID, &email)
	if err != nil {
		if strings.Contains(err.Error(), "users_email_key") {
			writeErrFor(w, r, http.StatusConflict, "email_taken",
				"There is already an account with that email.")
			return
		}
		writeErrFor(w, r, http.StatusNotFound, "invite_not_found",
			"That invite has expired or already been used. Ask for a new one.")
		return
	}

	if err := s.startSession(w, r, userID); err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Your account exists, but signing you in failed. Try logging in.")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"tenant_id": tenantID, "email": email})
}
