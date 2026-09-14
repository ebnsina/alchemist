package signing

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// The edge validates signatures in njs, the control plane mints them in Go. If those
// two implementations ever diverge, every playback URL 403s at the edge while looking
// perfectly valid at the origin -- a failure that would only show up in production,
// on the BDIX nodes, under real traffic.
//
// This runs the real deploy/edge/playback_auth.js under Node and asserts it produces
// byte-identical output to compute() here.
func TestEdgeScriptMatchesGo(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not installed")
	}

	script, err := filepath.Abs("../../../deploy/edge/playback_auth.js")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(script); err != nil {
		t.Fatalf("edge script missing: %v", err)
	}

	// Node decides module kind from package.json, so stage the real script beside one.
	dir := t.TempDir()
	src, err := os.ReadFile(script)
	if err != nil {
		t.Fatal(err)
	}
	write := func(name, content string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("package.json", `{"type":"module"}`)
	write("playback_auth.js", string(src))
	write("harness.mjs", `
import mod from './playback_auth.js';
const cases = JSON.parse(process.argv[2]);
console.log(JSON.stringify(cases.map(c => ({
  sig: mod.compute(c.secret, c.prefix, String(c.exp)),
  prefix: mod.assetPrefix(c.uri),
}))));
`)

	type testCase struct {
		Secret string `json:"secret"`
		Prefix string `json:"prefix"`
		Exp    int64  `json:"exp"`
		URI    string `json:"uri"`
	}
	cases := []testCase{
		{"first-generation-secret-at-least-32by", "/playback/t1/a1", 1789372237,
			"/playback/t1/a1/master.m3u8"},
		{"second-generation-secret-at-least-32b", "/playback/t1/a1", 1,
			"/playback/t1/a1/720p.cmfv"},
		{"a-secret-with-unicode-বাংলা-in-it!!", "/playback/tenant/asset", 2000000000,
			"/playback/tenant/asset/key"},
		{"third-secret-that-is-also-32-bytes-ok", "/playback/x/y", 1735689600,
			"/playback/x/y/sprite.vtt"},
	}
	payload, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}

	out, err := exec.Command(node, filepath.Join(dir, "harness.mjs"), string(payload)).Output()
	if err != nil {
		t.Fatalf("running edge script under node: %v", err)
	}

	var got []struct {
		Sig    string `json:"sig"`
		Prefix string `json:"prefix"`
	}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("harness output %q: %v", out, err)
	}
	if len(got) != len(cases) {
		t.Fatalf("got %d results, want %d", len(got), len(cases))
	}

	for i, c := range cases {
		want := compute([]byte(c.Secret), c.Prefix, c.Exp)
		if got[i].Sig != want {
			t.Errorf("case %d signature mismatch:\n  njs %s\n  go  %s", i, got[i].Sig, want)
		}
		if got[i].Prefix != c.Prefix {
			t.Errorf("case %d prefix derivation: njs %q, want %q", i, got[i].Prefix, c.Prefix)
		}
	}
}
