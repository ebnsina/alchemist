package live

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Everything that can be refused without touching the database must be, because the
// refusals are the point: this endpoint is unauthenticated, and each of these cases
// is someone reaching it who should get nothing back.
//
// The key-resolves-to-this-path check needs a live database and is covered by the
// journey harness.
func TestIngestAuthRefusesBeforeTouchingTheDatabase(t *testing.T) {
	// A nil database makes this honest: any case that reaches a query panics rather
	// than quietly passing, so a refusal here is a refusal on logic alone.
	m := &Module{}

	cases := []struct {
		name string
		body string
	}{
		{"not JSON at all", `{`},
		{"reading, not publishing", `{"action":"read","path":"abc","password":"k"}`},
		{"playback, not publishing", `{"action":"playback","path":"abc","password":"k"}`},
		{"no action", `{"path":"abc","password":"k"}`},
		{"no key", `{"action":"publish","path":"abc"}`},
		{"empty key", `{"action":"publish","path":"abc","password":""}`},
		{"no path", `{"action":"publish","password":"k"}`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/internal/live/authorize",
				strings.NewReader(c.body))
			w := httptest.NewRecorder()
			m.authorizeIngest(w, r)
			if w.Code != http.StatusUnauthorized {
				t.Errorf("got %d, want 401 — this would have let a publisher through",
					w.Code)
			}
		})
	}
}
