package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net"
	"net/http"
	"net/mail"
	"strings"
	"sync"
	"time"

	"github.com/ebnsina/alchemist/internal/platform/passwd"
)

// Self-service accounts. Everything before this point assumed an operator had
// already created the tenant and handed over a key; these three endpoints are what
// let somebody start without us.
//
// The session cookie is HttpOnly, so the token never reaches page script and an XSS
// cannot read it. The API key minted at signup is returned once, in the response
// body, because it is the thing the customer automates with — it is not a session.

const (
	sessionCookie = "alchemist_session"
	sessionTTL    = 30 * 24 * time.Hour
)

type signupRequest struct {
	Org      string `json:"org"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// authRoutes are open to a browser, so they carry their own CORS and their own rate
// limit. The rest of /v1 is machine-to-machine and needs neither.
func (s *Server) authEnabled() bool { return len(s.webOrigins) > 0 }

func (s *Server) postSignup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&req); err != nil {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request",
			"Send a JSON body with org, email and password.")
		return
	}
	req.Org = strings.TrimSpace(req.Org)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Org == "" || len(req.Org) > 120 {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_org",
			"Tell us what to call your organisation.")
		return
	}
	if !validEmail(req.Email) {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_email",
			"That email address does not look right.")
		return
	}
	hash, err := passwd.Hash(req.Password)
	if err != nil {
		writeErrFor(w, r, http.StatusBadRequest, "weak_password",
			"Use a password of at least 10 characters.")
		return
	}

	// The first API key is minted with the account: a customer who has to come back
	// and ask an operator for one has not really self-served.
	key, keyHash, err := newAPIKey()
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side. Please try again.")
		return
	}

	var userID, tenantID, keyID string
	err = s.db.Pool().QueryRow(r.Context(),
		`select user_id::text, tenant_id::text, api_key_id::text
		   from auth_signup($1, $2, $3, $4, $5, $6)`,
		req.Org, req.Email, hash, "bd-mobile", "First key", keyHash[:]).
		Scan(&userID, &tenantID, &keyID)
	if err != nil {
		// A duplicate email is the one failure the caller can act on. Anything else
		// is ours, and saying which is which is not worth leaking the schema.
		if strings.Contains(err.Error(), "users_email_key") {
			writeErrFor(w, r, http.StatusConflict, "email_taken",
				"There is already an account with that email.")
			return
		}
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side. Please try again.")
		return
	}

	if err := s.startSession(w, r, userID); err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Your account exists, but signing you in failed. Try logging in.")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"tenant_id": tenantID,
		"org":       req.Org,
		"email":     req.Email,
		// Shown once. We store only its hash, so we cannot show it again.
		"api_key": key,
	})
}

func (s *Server) postLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&req); err != nil {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request",
			"Send a JSON body with email and password.")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	var userID, tenantID, hash string
	err := s.db.Pool().QueryRow(r.Context(),
		`select id::text, tenant_id::text, password_hash from auth_find_user($1)`, req.Email).
		Scan(&userID, &tenantID, &hash)
	if err != nil {
		// Verify against a dummy hash anyway: skipping the work here is what makes an
		// unknown address answer faster than a wrong password, which is a free way to
		// enumerate customers.
		_ = passwd.Verify(req.Password, passwd.Dummy)
		writeErrFor(w, r, http.StatusUnauthorized, "invalid_credentials",
			"That email and password do not match.")
		return
	}
	if err := passwd.Verify(req.Password, hash); err != nil {
		writeErrFor(w, r, http.StatusUnauthorized, "invalid_credentials",
			"That email and password do not match.")
		return
	}
	if err := s.startSession(w, r, userID); err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side. Please try again.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"tenant_id": tenantID, "email": req.Email})
}

func (s *Server) postLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookie); err == nil {
		sum := sha256.Sum256([]byte(c.Value))
		_, _ = s.db.Pool().Exec(r.Context(), `select auth_end_session($1)`, sum[:])
	}
	// Clear it whether or not the token was live, so a stale cookie cannot linger.
	http.SetCookie(w, s.sessionCookie("", -time.Hour))
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) getSession(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(sessionCookie)
	if err != nil || c.Value == "" {
		writeErrFor(w, r, http.StatusUnauthorized, "no_session", "You are not signed in.")
		return
	}
	sum := sha256.Sum256([]byte(c.Value))
	var userID, tenantID, email, org string
	err = s.db.Pool().QueryRow(r.Context(),
		`select user_id::text, tenant_id::text, email::text, org_name from auth_session($1)`, sum[:]).
		Scan(&userID, &tenantID, &email, &org)
	if err != nil {
		writeErrFor(w, r, http.StatusUnauthorized, "no_session", "You are not signed in.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"user_id": userID, "tenant_id": tenantID, "email": email, "org": org,
	})
}

func (s *Server) startSession(w http.ResponseWriter, r *http.Request, userID string) error {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	sum := sha256.Sum256([]byte(token))
	expires := time.Now().Add(sessionTTL)
	if _, err := s.db.Pool().Exec(r.Context(),
		`select auth_start_session($1, $2, $3)`, userID, sum[:], expires); err != nil {
		return err
	}
	http.SetCookie(w, s.sessionCookie(token, sessionTTL))
	return nil
}

func (s *Server) sessionCookie(value string, ttl time.Duration) *http.Cookie {
	c := &http.Cookie{
		Name:     sessionCookie,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.sessionSecure,
		// Lax, not None: the site and the API are the same registrable domain in
		// production, and None would make the cookie usable from any other site.
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(ttl),
		MaxAge:   int(ttl.Seconds()),
	}
	if s.sessionDomain != "" {
		c.Domain = s.sessionDomain
	}
	return c
}

func newAPIKey() (string, [32]byte, error) {
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return "", [32]byte{}, err
	}
	key := "alch_" + base64.RawURLEncoding.EncodeToString(raw)
	return key, sha256.Sum256([]byte(key)), nil
}

func validEmail(addr string) bool {
	if len(addr) > 254 || !strings.Contains(addr, "@") {
		return false
	}
	_, err := mail.ParseAddress(addr)
	return err == nil
}

// cors answers preflight and marks the response for the configured origins only.
// A wildcard is impossible here by construction: credentialed requests need a
// specific origin, and an unlisted one gets no header at all.
func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := false
		for _, o := range s.webOrigins {
			if o == origin {
				allowed = true
				break
			}
		}
		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// authLimiter is a fixed window per client address. Deliberately in-process and
// deliberately crude: it stops a password-guessing loop from one host, which is the
// attack this surface actually invites.
//
// ponytail: per-process counters, so N replicas allow N windows. Move to a shared
// counter in Postgres or Redis when there is more than one instance.
type authLimiter struct {
	mu      sync.Mutex
	hits    map[string]int
	resetAt time.Time
	limit   int
	window  time.Duration
}

func newAuthLimiter(limit int, window time.Duration) *authLimiter {
	return &authLimiter{hits: map[string]int{}, resetAt: time.Now().Add(window),
		limit: limit, window: window}
}

func (l *authLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if time.Now().After(l.resetAt) {
		l.hits = map[string]int{}
		l.resetAt = time.Now().Add(l.window)
	}
	l.hits[key]++
	return l.hits[key] <= l.limit
}

func (s *Server) rateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}
		if !s.authLimiter.allow(host) {
			writeErrFor(w, r, http.StatusTooManyRequests, "too_many_attempts",
				"Too many attempts. Wait a minute and try again.")
			return
		}
		next.ServeHTTP(w, r)
	})
}
