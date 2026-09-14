package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// The contact form on the public site. Unauthenticated by definition — the point is
// to hear from people who do not have an account — so it sits behind the same origin
// check and rate limit as signup, and stores nothing it was not given.

type contactRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Org     string `json:"org"`
	Message string `json:"message"`
	// Honeypot. A real form leaves it empty; most bots fill every field they find.
	Website string `json:"website"`
}

func (s *Server) postContact(w http.ResponseWriter, r *http.Request) {
	var req contactRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&req); err != nil {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request",
			"Send a JSON body with name, email and message.")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Org = strings.TrimSpace(req.Org)
	req.Message = strings.TrimSpace(req.Message)

	if req.Website != "" {
		// Answer as though it worked. Telling a bot it was caught only teaches it.
		writeJSON(w, http.StatusAccepted, map[string]string{"status": "received"})
		return
	}
	if req.Name == "" || len(req.Name) > 120 {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request", "Tell us your name.")
		return
	}
	if !validEmail(req.Email) {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_email",
			"That email address does not look right.")
		return
	}
	if len(req.Message) < 10 || len(req.Message) > 4000 {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_message",
			"Tell us a little more — a sentence or two is plenty.")
		return
	}
	if len(req.Org) > 120 {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request", "That organisation name is too long.")
		return
	}

	var id string
	if err := s.db.Pool().QueryRow(r.Context(),
		`select contact_submit($1, $2, $3, $4, $5)::text`,
		req.Name, req.Email, req.Org, req.Message, r.Header.Get("Origin")).Scan(&id); err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"We could not send that just now. Please try again.")
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "received"})
}

type contactRow struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Email     string  `json:"email"`
	Org       *string `json:"org"`
	Message   string  `json:"message"`
	Source    *string `json:"source"`
	CreatedAt string  `json:"created_at"`
	HandledAt *string `json:"handled_at"`
}

// listContact is operator-only: these are other people's enquiries, and no customer
// key should reach them.
func (s *Server) listContact(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	rows, err := s.db.Pool().Query(r.Context(), `select * from admin_list_contact($1)`, limit)
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	defer rows.Close()

	out := []contactRow{}
	for rows.Next() {
		var c contactRow
		var created time.Time
		var handled *time.Time
		if err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.Org, &c.Message, &c.Source,
			&created, &handled); err != nil {
			writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
				"Something went wrong on our side.")
			return
		}
		c.CreatedAt = created.UTC().Format(time.RFC3339)
		if handled != nil {
			v := handled.UTC().Format(time.RFC3339)
			c.HandledAt = &v
		}
		out = append(out, c)
	}
	writeJSON(w, http.StatusOK, map[string]any{"requests": out})
}
