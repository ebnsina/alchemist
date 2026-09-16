package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// A list endpoint that quietly ignores what it was asked for is worse than one that
// refuses: the table shows a page nobody requested and looks right while doing it.
func TestParseListRefusesRatherThanClamps(t *testing.T) {
	sortable := []string{"created_at", "name"}

	for _, tc := range []struct{ query, code string }{
		{"limit=101", "invalid_limit"},
		{"limit=0", "invalid_limit"},
		{"limit=lots", "invalid_limit"},
		{"offset=-1", "invalid_offset"},
		{"offset=1.5", "invalid_offset"},
		{"sort=secret", "invalid_sort"},
		{"sort=tenant_id", "invalid_sort"},
		{"order=sideways", "invalid_order"},
	} {
		rec := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/v1/things?"+tc.query, nil)
		if _, ok := ParseList(rec, r, sortable, "created_at"); ok {
			t.Errorf("%s was accepted", tc.query)
			continue
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status %d, want 400", tc.query, rec.Code)
		}
		var body struct {
			Error struct{ Code string } `json:"error"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s: response is not an error envelope: %v", tc.query, err)
		}
		if body.Error.Code != tc.code {
			t.Errorf("%s: code %q, want %q", tc.query, body.Error.Code, tc.code)
		}
	}
}

func TestParseListDefaultsAndOrderBy(t *testing.T) {
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/v1/things", nil)
	l, ok := ParseList(rec, r, []string{"created_at", "name"}, "created_at")
	if !ok {
		t.Fatalf("a request with no parameters was refused: %s", rec.Body)
	}
	if l.Limit != 25 || l.Offset != 0 || l.Q != "" {
		t.Errorf("defaults are %+v, want limit 25, offset 0, no search", l)
	}
	if l.OrderBy() != "created_at desc" {
		t.Errorf("OrderBy = %q, want \"created_at desc\"", l.OrderBy())
	}

	rec = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/v1/things?sort=name&order=asc&q=+bengali+", nil)
	l, ok = ParseList(rec, r, []string{"created_at", "name"}, "created_at")
	if !ok {
		t.Fatalf("a valid request was refused: %s", rec.Body)
	}
	if l.OrderBy() != "name asc" || l.Q != "bengali" {
		t.Errorf("got %q and q %q, want \"name asc\" and \"bengali\"", l.OrderBy(), l.Q)
	}
}

func TestFlagRefusesNonBoolean(t *testing.T) {
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/v1/things?revoked=maybe", nil)
	if _, ok := Flag(rec, r, "revoked"); ok || rec.Code != http.StatusBadRequest {
		t.Fatalf("revoked=maybe was accepted (status %d)", rec.Code)
	}

	rec = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/v1/things", nil)
	v, ok := Flag(rec, r, "revoked")
	if !ok || v != nil {
		t.Fatalf("an absent filter should mean no filter, got %v", v)
	}
}
