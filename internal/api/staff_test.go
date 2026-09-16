package api

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Impersonation is the whole platform-admin feature, and every one of its failure
// modes is a database fact: who the session acts as, whether the flag is still set,
// and what a customer is never shown. So these run against a real database.

const testOrigin = "https://app.example"

func staffServer(f *listFixture) *Server {
	return &Server{db: f.db, webOrigins: []string{testOrigin},
		authLimiter: newAuthLimiter(100, time.Minute)}
}

// seedUser creates an account in one tenant, staff or not.
func (f *listFixture) seedUser(t *testing.T, tenant, email string, admin bool) string {
	t.Helper()
	var id string
	if err := f.admin.QueryRow(context.Background(),
		`insert into users (tenant_id, email, password_hash, role, platform_admin)
		 values ($1, $2, 'x', 'owner', $3) returning id::text`, tenant, email, admin).
		Scan(&id); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return id
}

// seedSession returns the raw cookie value; only its hash is stored.
func (f *listFixture) seedSession(t *testing.T, userID string) string {
	t.Helper()
	token := "tok-" + userID
	sum := sha256.Sum256([]byte(token))
	if _, err := f.admin.Exec(context.Background(),
		`insert into sessions (user_id, token_hash, expires_at)
		 values ($1, $2, now() + interval '1 hour')`, userID, sum[:]); err != nil {
		t.Fatalf("seed session: %v", err)
	}
	return token
}

func (f *listFixture) seedAssetIn(t *testing.T, tenant, title string) string {
	t.Helper()
	var id string
	if err := f.admin.QueryRow(context.Background(),
		`insert into assets (tenant_id, ladder_profile, state, title)
		 values ($1, 'bd-mobile', 'failed', $2) returning id::text`, tenant, title).
		Scan(&id); err != nil {
		t.Fatalf("seed asset: %v", err)
	}
	return id
}

// seedFailedJob writes a queue failure carrying the kind of text a customer must
// never see: a real path from a real machine.
func (f *listFixture) seedFailedJob(t *testing.T, assetID, text string) {
	t.Helper()
	ctx := context.Background()
	if _, err := f.admin.Exec(ctx,
		`insert into river_job (args, kind, state, attempt, max_attempts, queue,
		                        finalized_at, errors)
		 values (jsonb_build_object('asset_id', $1::text), 'transcode', 'discarded', 1, 1,
		         'encode', now(),
		         array[jsonb_build_object('at', now(), 'attempt', 1, 'error', $2::text)])`,
		assetID, text); err != nil {
		t.Fatalf("seed job: %v", err)
	}
	t.Cleanup(func() {
		_, _ = f.admin.Exec(ctx, `delete from river_job where args->>'asset_id' = $1`, assetID)
	})
}

// call runs one request through the real router, cookie and origin included, so the
// chain under test is authenticate -> requireSession -> staffOnly, not a handler
// called directly.
func call(t *testing.T, srv *Server, token, method, target, body string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, target, nil)
	} else {
		r = httptest.NewRequest(method, target, strings.NewReader(body))
	}
	r.Header.Set("Origin", testOrigin)
	r.Header.Set("Content-Type", "application/json")
	r.AddCookie(&http.Cookie{Name: sessionCookie, Value: token})
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, r)

	var decoded map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &decoded)
	return rec, decoded
}

func errorCode(body map[string]any) string {
	e, _ := body["error"].(map[string]any)
	code, _ := e["code"].(string)
	return code
}

// A refusal, never an empty list and never a 404: somebody debugging at 2am has to
// be able to tell "no data" from "not allowed".
func TestStaffRoutesRefuseANonStaffSession(t *testing.T) {
	f := newFixture(t)
	srv := staffServer(f)
	token := f.seedSession(t, f.seedUser(t, f.tenantA, "customer@example.test", false))
	asset := f.seedAssetIn(t, f.tenantA, "theirs")

	for _, c := range []struct{ method, target, body string }{
		{http.MethodGet, "/v1/staff/tenants", ""},
		{http.MethodPost, "/v1/staff/impersonate", `{"tenant_id":"` + f.tenantB + `"}`},
		{http.MethodDelete, "/v1/staff/impersonate", ""},
		{http.MethodGet, "/v1/assets/" + asset + "/diagnostics", ""},
		{http.MethodGet, "/v1/live-sessions/" + asset + "/diagnostics", ""},
	} {
		rec, body := call(t, srv, token, c.method, c.target, c.body)
		if rec.Code != http.StatusForbidden || errorCode(body) != "staff_only" {
			t.Errorf("%s %s: status %d code %q, want 403 staff_only",
				c.method, c.target, rec.Code, errorCode(body))
		}
	}
}

// The point of impersonation: the customer's own endpoints answer for the customer's
// account, with no second code path to drift.
func TestImpersonationSwitchesTheTenantAndBack(t *testing.T) {
	f := newFixture(t)
	srv := staffServer(f)
	staff := f.seedUser(t, f.tenantA, "founder@alchemist.test", true)
	token := f.seedSession(t, staff)
	f.seedAssetIn(t, f.tenantB, "a customer's video")

	if rec, body := call(t, srv, token, http.MethodPost, "/v1/staff/impersonate",
		`{"tenant_id":"`+f.tenantB+`"}`); rec.Code != http.StatusOK {
		t.Fatalf("impersonate: status %d, body %v", rec.Code, body)
	}

	rec, body := call(t, srv, token, http.MethodGet, "/v1/assets", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("assets while impersonating: status %d", rec.Code)
	}
	if got := assetTitles(t, body); len(got) != 1 || got[0] != "a customer's video" {
		t.Fatalf("assets while impersonating = %v, want the other tenant's one video", got)
	}

	if rec, _ := call(t, srv, token, http.MethodDelete, "/v1/staff/impersonate", ""); rec.Code != http.StatusOK {
		t.Fatalf("stop impersonating: status %d", rec.Code)
	}
	_, body = call(t, srv, token, http.MethodGet, "/v1/assets", "")
	if got := assetTitles(t, body); len(got) != 0 {
		t.Fatalf("assets after stopping = %v, want the staff account's own (none)", got)
	}
}

// Read-only is what makes impersonation safe enough to have: a support session must
// not be able to delete a customer's video by clicking the wrong row.
func TestImpersonationRefusesEveryWrite(t *testing.T) {
	f := newFixture(t)
	srv := staffServer(f)
	token := f.seedSession(t, f.seedUser(t, f.tenantA, "founder2@alchemist.test", true))
	asset := f.seedAssetIn(t, f.tenantB, "not yours to delete")

	call(t, srv, token, http.MethodPost, "/v1/staff/impersonate", `{"tenant_id":"`+f.tenantB+`"}`)

	for _, c := range []struct{ method, target, body string }{
		{http.MethodDelete, "/v1/assets/" + asset, ""},
		{http.MethodPatch, "/v1/assets/" + asset, `{"title":"renamed"}`},
		{http.MethodPost, "/v1/keys", `{"name":"mine"}`},
	} {
		rec, body := call(t, srv, token, c.method, c.target, c.body)
		if rec.Code != http.StatusForbidden || errorCode(body) != "read_only_session" {
			t.Errorf("%s %s: status %d code %q, want 403 read_only_session",
				c.method, c.target, rec.Code, errorCode(body))
		}
	}
}

// A flag taken away has to stop working on the next request. Anything else means
// revoking staff access waits for a session to expire, which is up to thirty days.
func TestRevokingTheFlagEndsImpersonationImmediately(t *testing.T) {
	f := newFixture(t)
	srv := staffServer(f)
	staff := f.seedUser(t, f.tenantA, "founder3@alchemist.test", true)
	token := f.seedSession(t, staff)
	f.seedAssetIn(t, f.tenantB, "still theirs")

	call(t, srv, token, http.MethodPost, "/v1/staff/impersonate", `{"tenant_id":"`+f.tenantB+`"}`)
	if _, err := f.admin.Exec(context.Background(),
		`update users set platform_admin = false where id = $1`, staff); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	_, body := call(t, srv, token, http.MethodGet, "/v1/assets", "")
	if got := assetTitles(t, body); len(got) != 0 {
		t.Fatalf("assets after the flag was revoked = %v, want the session's own tenant", got)
	}
	rec, body := call(t, srv, token, http.MethodGet, "/v1/staff/tenants", "")
	if rec.Code != http.StatusForbidden || errorCode(body) != "staff_only" {
		t.Fatalf("staff route after revoke: status %d code %q, want 403 staff_only",
			rec.Code, errorCode(body))
	}
}

// The asymmetry the whole feature exists for, proved from both sides at once.
func TestActivityHidesTheErrorTextAndDiagnosticsShowIt(t *testing.T) {
	f := newFixture(t)
	srv := staffServer(f)
	const secret = "ffmpeg: /var/lib/alchemist/work/rung-720/mezzanine.mp4: no such file"
	asset := f.seedAssetIn(t, f.tenantB, "broken")
	f.seedFailedJob(t, asset, secret)

	customer := f.seedSession(t, f.seedUser(t, f.tenantB, "owner@customer.test", false))
	rec, _ := call(t, srv, customer, http.MethodGet, "/v1/assets/"+asset+"/activity", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("activity: status %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), secret) || strings.Contains(rec.Body.String(), "ffmpeg") {
		t.Fatalf("activity leaked the raw error text to the customer: %s", rec.Body.String())
	}

	staff := f.seedSession(t, f.seedUser(t, f.tenantA, "founder4@alchemist.test", true))
	call(t, srv, staff, http.MethodPost, "/v1/staff/impersonate", `{"tenant_id":"`+f.tenantB+`"}`)
	rec, _ = call(t, srv, staff, http.MethodGet, "/v1/assets/"+asset+"/diagnostics", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("diagnostics: status %d, body %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), secret) {
		t.Fatalf("diagnostics withheld the error text staff exist to read: %s", rec.Body.String())
	}
}

// A tenant owner who could grant this would own every other tenant too, so the
// customer surface must have no path to the column at all.
func TestTenantOwnerCannotGrantPlatformAdmin(t *testing.T) {
	f := newFixture(t)
	srv := staffServer(f)
	owner := f.seedUser(t, f.tenantA, "owner2@customer.test", false)
	mate := f.seedUser(t, f.tenantA, "mate@customer.test", false)
	token := f.seedSession(t, owner)

	rec, _ := call(t, srv, token, http.MethodPatch, "/v1/members/"+mate,
		`{"role":"admin","platform_admin":true}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("patch member: status %d, body %s", rec.Code, rec.Body.String())
	}

	var admin bool
	if err := f.admin.QueryRow(context.Background(),
		`select platform_admin from users where id = $1`, mate).Scan(&admin); err != nil {
		t.Fatalf("read flag: %v", err)
	}
	if admin {
		t.Fatal("a tenant owner granted platform_admin through /v1/members")
	}
}
