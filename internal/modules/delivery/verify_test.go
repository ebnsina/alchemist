package delivery

import "testing"

// assetPrefix decides what a signature covers. It must agree exactly with
// assetPrefix() in deploy/edge/playback_auth.js; internal/signing/parity_test.go
// asserts that agreement against the real script.
func TestAssetPrefix(t *testing.T) {
	cases := []struct {
		path string
		want string
		ok   bool
	}{
		{"/playback/t1/a1/master.m3u8", "/playback/t1/a1", true},
		{"/playback/t1/a1/720p.cmfv", "/playback/t1/a1", true},
		{"/playback/t1/a1/key", "/playback/t1/a1", true},
		{"/playback/t1/a1/sprite.vtt", "/playback/t1/a1", true},
		// A deeper path still resolves to the same asset prefix.
		{"/playback/t1/a1/nested/thing", "/playback/t1/a1", true},
		// Anything that is not a playback path must not authorize.
		{"/playback/t1/a1", "", false},
		{"/playback/t1", "", false},
		{"/playback", "", false},
		{"/v1/assets/a1", "", false},
		{"/", "", false},
		{"", "", false},
		{"/playback//a1/file", "", false},
		{"/playback/t1//file", "", false},
	}
	for _, tc := range cases {
		got, ok := assetPrefix(tc.path)
		if ok != tc.ok || got != tc.want {
			t.Errorf("assetPrefix(%q) = (%q, %v), want (%q, %v)",
				tc.path, got, ok, tc.want, tc.ok)
		}
	}
}
