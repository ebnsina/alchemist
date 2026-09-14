package api

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/ebnsina/alchemist/internal/platform/httpx"
)

// Admin endpoints create tenants and issue API keys. They are separated from the
// customer API by a different credential, because a leaked customer key must never
// be able to mint more tenants or keys.
//
// Deliberately a shared secret rather than a user system: this is operator tooling,
// and an operator has shell access anyway. A real console would put a person behind it.
func (s *Server) adminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.adminKey == "" {
			httpx.Error(w, http.StatusNotFound, "not_found", "We couldn't find that.")
			return
		}
		key, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		// Constant-time: a timing oracle here leaks the key that mints every other key.
		if !ok || subtle.ConstantTimeCompare([]byte(key), []byte(s.adminKey)) != 1 {
			httpx.Error(w, http.StatusUnauthorized, "invalid_admin_key",
				"That admin key is not valid.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

type createTenantRequest struct {
	Name          string `json:"name"`
	LadderProfile string `json:"ladder_profile"`
}

func (s *Server) createTenant(w http.ResponseWriter, r *http.Request) {
	var req createTenantRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request",
			"Send a JSON body with a \"name\" field.")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", "Include a tenant name.")
		return
	}
	if req.LadderProfile == "" {
		req.LadderProfile = "bd-mobile"
	}

	var id string
	err := s.db.Pool().QueryRow(r.Context(),
		`select admin_create_tenant($1, $2)::text`, req.Name, req.LadderProfile).Scan(&id)
	if err != nil {
		// An unknown profile is the likely cause and is the customer's to fix.
		httpx.Error(w, http.StatusBadRequest, "invalid_ladder_profile",
			"That ladder profile does not exist.")
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]string{
		"tenant_id": id, "name": req.Name, "ladder_profile": req.LadderProfile,
	})
}

func (s *Server) listTenants(w http.ResponseWriter, r *http.Request) {
	type item struct {
		ID            string `json:"tenant_id"`
		Name          string `json:"name"`
		LadderProfile string `json:"ladder_profile"`
		CreatedAt     string `json:"created_at"`
	}
	items := []item{}

	rows, err := s.db.Pool().Query(r.Context(),
		`select id::text, name, ladder_profile,
		        to_char(created_at,'YYYY-MM-DD"T"HH24:MI:SS"Z"') from admin_list_tenants()`)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	defer rows.Close()
	for rows.Next() {
		var it item
		if err := rows.Scan(&it.ID, &it.Name, &it.LadderProfile, &it.CreatedAt); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "internal_error",
				"Something went wrong on our side.")
			return
		}
		items = append(items, it)
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"tenants": items})
}

type issueKeyRequest struct {
	Name string `json:"name"`
}

// issueKey mints an API key. The plaintext is returned once and only the hash is
// stored, so a database dump does not yield working credentials.
func (s *Server) issueKey(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "id")

	var req issueKeyRequest
	_ = json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&req)
	if strings.TrimSpace(req.Name) == "" {
		req.Name = "default"
	}

	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	// URL-safe and prefixed, so a leaked key is recognisable in logs and scanners.
	plaintext := "alc_" + base64.RawURLEncoding.EncodeToString(raw)
	sum := sha256.Sum256([]byte(plaintext))

	var keyID string
	err := s.db.Pool().QueryRow(r.Context(),
		`select admin_issue_key($1, $2, $3)::text`, tenantID, req.Name, sum[:]).Scan(&keyID)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "tenant_not_found", "We couldn't find that tenant.")
		return
	}

	httpx.JSON(w, http.StatusCreated, map[string]string{
		"key_id": keyID, "tenant_id": tenantID, "name": req.Name,
		"api_key": plaintext,
		"note":    "Store this now. It is not shown again.",
	})
}

func (s *Server) revokeKey(w http.ResponseWriter, r *http.Request) {
	var revoked bool
	err := s.db.Pool().QueryRow(r.Context(),
		`select coalesce(admin_revoke_key($1), false)`, chi.URLParam(r, "keyID")).Scan(&revoked)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		httpx.Error(w, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	if !revoked {
		httpx.Error(w, http.StatusNotFound, "key_not_found",
			"We couldn't find an active key with that ID.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
