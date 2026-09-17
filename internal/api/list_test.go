package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/ebnsina/alchemist/internal/platform/db"
)

// These run against a real database because the things worth proving here are the
// database's: that a page and its total agree, that a search matches what it claims,
// and that another tenant's row is invisible rather than forbidden.

type listFixture struct {
	db       *db.DB
	admin    *pgx.Conn
	tenantA  string
	tenantB  string
	assetIDs []string
}

func newFixture(t *testing.T) *listFixture {
	t.Helper()
	url := os.Getenv("ALCHEMIST_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("ALCHEMIST_TEST_DATABASE_URL not set")
	}
	ctx := context.Background()

	d, err := db.New(ctx, url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	admin, err := pgx.Connect(ctx, os.Getenv("ALCHEMIST_TEST_ADMIN_URL"))
	if err != nil {
		d.Close()
		t.Fatalf("admin connect: %v", err)
	}

	f := &listFixture{db: d, admin: admin}
	for _, dst := range []*string{&f.tenantA, &f.tenantB} {
		if err := admin.QueryRow(ctx,
			`insert into tenants (name, ladder_profile) values ('list-test', 'bd-mobile')
			 returning id::text`).Scan(dst); err != nil {
			t.Fatalf("seed tenant: %v", err)
		}
	}
	t.Cleanup(func() {
		_, _ = admin.Exec(ctx, `delete from tenants where id = any($1::uuid[])`,
			[]string{f.tenantA, f.tenantB})
		_ = admin.Close(ctx)
		d.Close()
	})
	return f
}

// seedAssets creates titled assets for tenant A, oldest first.
func (f *listFixture) seedAssets(t *testing.T, titles ...string) {
	t.Helper()
	for _, title := range titles {
		var id string
		if err := f.admin.QueryRow(context.Background(),
			`insert into assets (tenant_id, ladder_profile, state, title)
			 values ($1, 'bd-mobile', 'ready', $2) returning id::text`,
			f.tenantA, title).Scan(&id); err != nil {
			t.Fatalf("seed asset %q: %v", title, err)
		}
		f.assetIDs = append(f.assetIDs, id)
	}
}

// get runs one list request as a tenant and decodes the envelope.
func (f *listFixture) get(t *testing.T, tenantID, target string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, target, nil)
	r = r.WithContext(context.WithValue(r.Context(), tenantKey, tenantID))
	rec := httptest.NewRecorder()
	(&Server{db: f.db}).listAssets(rec, r)

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("%s: response is not JSON: %v", target, err)
	}
	return rec, body
}

func assetTitles(t *testing.T, body map[string]any) []string {
	t.Helper()
	rows, _ := body["assets"].([]any)
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		m, _ := row.(map[string]any)
		title, _ := m["title"].(string)
		out = append(out, title)
	}
	return out
}

// A table that cannot trust its own pages shows the same video twice and loses
// another; a total that moves between pages makes "1-25 of 340" a lie.
func TestListAssetsPagesDisjointlyWithAStableTotal(t *testing.T) {
	f := newFixture(t)
	f.seedAssets(t, "one", "two", "three", "four", "five")

	seen := map[string]bool{}
	var firstTotal float64
	for offset := 0; offset < 4; offset += 2 {
		rec, body := f.get(t, f.tenantA,
			"/v1/assets?limit=2&sort=created_at&order=asc&offset="+strconv.Itoa(offset))
		if rec.Code != http.StatusOK {
			t.Fatalf("offset %d: status %d: %s", offset, rec.Code, rec.Body)
		}
		total, _ := body["total"].(float64)
		if offset == 0 {
			firstTotal = total
		} else if total != firstTotal {
			t.Errorf("total moved between pages: %v then %v", firstTotal, total)
		}
		for _, title := range assetTitles(t, body) {
			if seen[title] {
				t.Errorf("%q appeared on two pages", title)
			}
			seen[title] = true
		}
	}
	if firstTotal != 5 {
		t.Errorf("total = %v, want 5 -- the count must ignore limit and offset", firstTotal)
	}
	if len(seen) != 4 {
		t.Errorf("two pages of two returned %d distinct videos, want 4", len(seen))
	}
}

// A search that matches more than it claims is how a customer concludes the filter
// is broken and stops using it.
func TestListAssetsSearchMatchesTitleAndIDPrefixOnly(t *testing.T) {
	f := newFixture(t)
	f.seedAssets(t, "Physics lecture 1", "Chemistry lab", "physics revision")

	_, body := f.get(t, f.tenantA, "/v1/assets?q=physics")
	got := assetTitles(t, body)
	if len(got) != 2 || body["total"].(float64) != 2 {
		t.Fatalf("q=physics returned %v (total %v), want the two physics videos",
			got, body["total"])
	}
	for _, title := range got {
		if title == "Chemistry lab" {
			t.Error("q=physics matched a chemistry video")
		}
	}

	// The id prefix is the other half of the claim: pasting the short id the
	// dashboard shows has to find the video.
	prefix := f.assetIDs[1][:8]
	_, body = f.get(t, f.tenantA, "/v1/assets?q="+prefix)
	if got := assetTitles(t, body); len(got) != 1 || got[0] != "Chemistry lab" {
		t.Errorf("q=%s returned %v, want just the video with that id", prefix, got)
	}

	// And nothing that is neither.
	_, body = f.get(t, f.tenantA, "/v1/assets?q=biology")
	if total := body["total"].(float64); total != 0 {
		t.Errorf("q=biology matched %v videos", total)
	}
}

func TestListAssetsRefusesAnUnknownSort(t *testing.T) {
	f := newFixture(t)
	rec, body := f.get(t, f.tenantA, "/v1/assets?sort=tenant_id")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400 -- an unknown sort must be refused, not ignored", rec.Code)
	}
	errObj, _ := body["error"].(map[string]any)
	if errObj["code"] != "invalid_sort" {
		t.Errorf("code %v, want invalid_sort", errObj["code"])
	}
}

// The boundary is the database. Another tenant's video must be invisible to the
// update, not refused by a check in Go that someone can later forget to write.
func TestPatchAssetFindsNothingForAnotherTenant(t *testing.T) {
	f := newFixture(t)
	f.seedAssets(t, "A's lecture")

	r := httptest.NewRequest(http.MethodPatch, "/v1/assets/"+f.assetIDs[0],
		strings.NewReader(`{"title":"stolen"}`))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", f.assetIDs[0])
	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
	// Tenant B, naming tenant A's asset by its real id.
	r = r.WithContext(context.WithValue(ctx, tenantKey, f.tenantB))
	rec := httptest.NewRecorder()
	(&Server{db: f.db}).patchAsset(rec, r)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d, want 404: RLS should make the row invisible", rec.Code)
	}

	var title *string
	if err := f.db.AsTenant(context.Background(), f.tenantA, func(tx pgx.Tx) error {
		return tx.QueryRow(context.Background(),
			`select title from assets where id = $1`, f.assetIDs[0]).Scan(&title)
	}); err != nil {
		t.Fatalf("re-reading the asset: %v", err)
	}
	if title == nil || *title != "A's lecture" {
		t.Errorf("title is now %v; another tenant changed it", title)
	}
}

// "Deleted" has to mean deliveries stop, including ones already queued -- otherwise
// a customer who removed an endpoint keeps receiving callbacks at it.
func TestDeleteWebhookStopsDeliveries(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	var id string
	if err := f.admin.QueryRow(ctx,
		`insert into webhook_endpoints (tenant_id, url, secret, events)
		 values ($1, 'https://example.test/hook', '\x00', array['asset.ready'])
		 returning id::text`, f.tenantA).Scan(&id); err != nil {
		t.Fatalf("seed endpoint: %v", err)
	}

	r := httptest.NewRequest(http.MethodDelete, "/v1/webhooks/"+id, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	r = r.WithContext(context.WithValue(
		context.WithValue(r.Context(), chi.RouteCtxKey, rctx), tenantKey, f.tenantA))
	rec := httptest.NewRecorder()
	(&Server{db: f.db}).deleteWebhook(rec, r)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status %d, want 204: %s", rec.Code, rec.Body)
	}

	// Both places a delivery can come from: the fan-out that picks endpoints for a
	// new event, and the job that loads one endpoint just before it posts.
	err := f.db.AsTenant(ctx, f.tenantA, func(tx pgx.Tx) error {
		var queued int
		if err := tx.QueryRow(ctx,
			`select count(*) from webhook_endpoints
			  where active and 'asset.ready' = any(events)`).Scan(&queued); err != nil {
			return err
		}
		if queued != 0 {
			t.Errorf("a new asset.ready would still fan out to %d endpoints", queued)
		}
		var url string
		err := tx.QueryRow(ctx,
			`select url from webhook_endpoints where id = $1`, id).Scan(&url)
		if !errors.Is(err, pgx.ErrNoRows) {
			t.Errorf("a queued delivery would still resolve the endpoint to %q (%v)", url, err)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("checking deliveries: %v", err)
	}
}

// Every list endpoint composes its own SQL from a shared predicate. A typo in one of
// them is invisible until a customer opens that page, so ask each for a filtered,
// searched, sorted page and require a 200.
func TestEveryListEndpointAnswersAPage(t *testing.T) {
	f := newFixture(t)
	s := &Server{db: f.db}

	for name, tc := range map[string]struct {
		query   string
		handler http.HandlerFunc
	}{
		"assets":         {"q=x&state=ready&state=failed&sort=state&order=asc", s.listAssets},
		"keys":           {"q=x&revoked=false&sort=name&order=asc", s.listKeys},
		"webhooks":       {"q=x&active=true&sort=url&order=asc", s.listWebhooks},
		"bucket-sources": {"q=x&state=paused&sort=bucket&order=asc", s.listBucketSources},
		"migrations":     {"q=x&state=importing&sort=state&order=asc", s.listMigrations},
		"members":        {"q=x&role=admin&sort=email&order=asc", s.listMembers},
		"edits":          {"q=abc&state=queued&sort=state&order=asc", s.listEdits},
	} {
		r := httptest.NewRequest(http.MethodGet, "/v1/"+name+"?limit=5&offset=0&"+tc.query, nil)
		ctx := context.WithValue(r.Context(), tenantKey, f.tenantA)
		rec := httptest.NewRecorder()
		tc.handler(rec, r.WithContext(context.WithValue(ctx, userKey, "00000000-0000-0000-0000-000000000000")))
		if rec.Code != http.StatusOK {
			t.Errorf("%s: status %d: %s", name, rec.Code, rec.Body)
			continue
		}
		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		if _, ok := body["total"]; !ok {
			t.Errorf("%s: no total in the response; a table cannot say how many there are", name)
		}
	}
}

// A PATCH built with coalesce sends a typed NULL for every field the caller left
// out. Postgres cannot infer a parameter no column constrains, and the whole update
// 500s -- which only shows up against a real database.
func TestPatchWebhookLeavesOmittedFieldsAlone(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	var id string
	if err := f.admin.QueryRow(ctx,
		`insert into webhook_endpoints (tenant_id, url, secret, events)
		 values ($1, 'https://example.test/hook', '\x00', array['asset.ready'])
		 returning id::text`, f.tenantA).Scan(&id); err != nil {
		t.Fatalf("seed endpoint: %v", err)
	}

	patch := func(body string) map[string]any {
		t.Helper()
		r := httptest.NewRequest(http.MethodPatch, "/v1/webhooks/"+id, strings.NewReader(body))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", id)
		r = r.WithContext(context.WithValue(
			context.WithValue(r.Context(), chi.RouteCtxKey, rctx), tenantKey, f.tenantA))
		rec := httptest.NewRecorder()
		(&Server{db: f.db}).patchWebhook(rec, r)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status %d: %s", body, rec.Code, rec.Body)
		}
		var out map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("%s: %v", body, err)
		}
		return out
	}

	if got := patch(`{"active":false}`); got["active"] != false ||
		got["url"] != "https://example.test/hook" {
		t.Errorf("pausing changed the URL too: %v", got)
	}
	if got := patch(`{"url":"https://example.test/other"}`); got["active"] != false ||
		got["url"] != "https://example.test/other" {
		t.Errorf("changing the URL resumed a paused endpoint: %v", got)
	}
}

// The ceiling is the account's setting and ?ttl= may only shorten it. A caller able to
// lengthen it would be setting the policy rather than working inside it -- and nothing
// at the edge re-checks what went into a token.
func TestPlaybackTTLBounds(t *testing.T) {
	const max = 2 * time.Hour

	for _, c := range []struct {
		asked string
		want  time.Duration
		ok    bool
	}{
		{"", max, true},                 // no ask: the account's own setting
		{"900", 15 * time.Minute, true}, // shorter is fine
		{"7200", max, true},             // exactly the ceiling
		{"7201", 0, false},              // one second over it
		{"30", 0, false},                // under the floor a player cannot retry in
		{"0", 0, false},
		{"-60", 0, false},
		{"soon", 0, false},
	} {
		got, err := resolveTTL(max, c.asked)
		if (err == nil) != c.ok {
			t.Errorf("ttl=%q: err=%v, want ok=%v", c.asked, err, c.ok)
			continue
		}
		if c.ok && got != c.want {
			t.Errorf("ttl=%q: got %v, want %v", c.asked, got, c.want)
		}
	}
}
