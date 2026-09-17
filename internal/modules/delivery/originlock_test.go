package delivery

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// The edge and the origin must agree on what counts as an allowed request, not only on
// the hash. njs reads r.headersIn and Go reads http.Header; a disagreement here 403s at
// the edge while passing locally, which is the failure mode the signing parity test
// exists to prevent and this is the same seam one layer up.
func TestOriginLockMatchesEdgeScript(t *testing.T) {
	cases := []struct {
		Want    string `json:"want"`
		Origin  string `json:"origin"`
		Referer string `json:"referer"`
		allowed bool
	}{
		{"https://app.school.example", "https://app.school.example", "", true},
		{"https://app.school.example", "HTTPS://APP.SCHOOL.EXAMPLE", "", true},
		{"https://app.school.example", "https://evil.example", "", false},
		// Safari sends no Origin on a non-CORS media request, so Referer decides.
		{"https://app.school.example", "", "https://app.school.example/lesson/4?x=1", true},
		{"https://app.school.example", "", "https://evil.example/steal", false},
		// Neither header: the whole bypass if it were allowed.
		{"https://app.school.example", "", "", false},
		// A port is part of an origin. Dropping it lets :8080 play a :8443 link.
		{"https://app.example:8443", "https://app.example", "", false},
	}

	for i, c := range cases {
		r := httptest.NewRequest(http.MethodGet, "/playback/t/a/720p.cmfv", nil)
		if c.Origin != "" {
			r.Header.Set("Origin", c.Origin)
		}
		if c.Referer != "" {
			r.Header.Set("Referer", c.Referer)
		}
		if got := originAllowed(r, c.Want); got != c.allowed {
			t.Errorf("case %d go: allowed=%v, want %v", i, got, c.allowed)
		}
	}

	// An unlocked token allows anything, which is every link that exists today.
	if !originAllowed(httptest.NewRequest(http.MethodGet, "/x", nil), "") {
		t.Error("an unlocked token must not require an Origin")
	}

	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not installed; Go side checked")
	}
	src, err := os.ReadFile(filepath.Join("..", "..", "..", "deploy", "edge", "playback_auth.js"))
	if err != nil {
		t.Fatalf("edge script missing: %v", err)
	}

	dir := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("package.json", `{"type":"module"}`)
	write("playback_auth.js", string(src))
	write("origin.mjs", `
import mod from './playback_auth.js';
console.log(JSON.stringify(JSON.parse(process.argv[2]).map(c => {
  const headers = {};
  if (c.origin) headers['Origin'] = c.origin;
  if (c.referer) headers['Referer'] = c.referer;
  return mod.originAllowed({ headersIn: headers }, c.want);
})));
`)

	payload, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(node, filepath.Join(dir, "origin.mjs"), string(payload)).Output()
	if err != nil {
		t.Fatalf("running edge script under node: %v", err)
	}
	var njs []bool
	if err := json.Unmarshal(out, &njs); err != nil {
		t.Fatalf("harness output %q: %v", out, err)
	}
	if len(njs) != len(cases) {
		t.Fatalf("got %d answers from njs, want %d", len(njs), len(cases))
	}
	for i, c := range cases {
		if njs[i] != c.allowed {
			t.Errorf("case %d njs: allowed=%v, want %v", i, njs[i], c.allowed)
		}
	}
}

func TestNormalizeOrigin(t *testing.T) {
	for _, c := range []struct {
		in   string
		want string
		ok   bool
	}{
		{"https://app.example", "https://app.example", true},
		{"HTTPS://App.Example/", "https://app.example", true},
		{"https://app.example:8443/lesson", "https://app.example:8443", true},
		{"app.example", "", false}, // no scheme: not what a browser sends
		{"ftp://app.example", "", false},
		{"", "", false},
	} {
		got, ok := NormalizeOrigin(c.in)
		if got != c.want || ok != c.ok {
			t.Errorf("NormalizeOrigin(%q) = %q,%v; want %q,%v", c.in, got, ok, c.want, c.ok)
		}
	}
}
