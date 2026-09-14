package config

import (
	"strings"
	"testing"
)

func validEnv() map[string]string {
	return map[string]string{
		"ALCHEMIST_HTTP_ADDR":      ":8080",
		"ALCHEMIST_DATABASE_URL":   "postgres://u@localhost/db",
		"ALCHEMIST_S3_ENDPOINT":    "http://localhost:9000",
		"ALCHEMIST_S3_REGION":      "us-east-1",
		"ALCHEMIST_S3_BUCKET":      "alchemist",
		"ALCHEMIST_S3_ACCESS_KEY":  "key",
		"ALCHEMIST_S3_SECRET_KEY":  "secret",
		"ALCHEMIST_PLAYBACK_KEYS":  "k1:a-playback-key-that-is-32-bytes-ok",
		"ALCHEMIST_KEK":            "a-key-encryption-key-of-32-bytes!!",
		"ALCHEMIST_WORK_DIR":       "/tmp/alchemist",
		"ALCHEMIST_ENCODE_WORKERS": "4",
		// Optional, but pinned empty so the test does not inherit whatever the
		// developer's shell happens to export.
		"ALCHEMIST_FETCH_ALLOWLIST": "",
	}
}

func setEnv(t *testing.T, env map[string]string) {
	t.Helper()
	for k, v := range env {
		t.Setenv(k, v)
	}
}

func TestLoadValid(t *testing.T) {
	setEnv(t, validEnv())
	c, err := Load()
	if err != nil {
		t.Fatalf("valid configuration rejected: %v", err)
	}
	if c.EncodeWorkers != 4 {
		t.Errorf("EncodeWorkers = %d, want 4", c.EncodeWorkers)
	}
	if len(c.PlaybackKeys) != 1 {
		t.Errorf("PlaybackKeys = %v, want one entry", c.PlaybackKeys)
	}
}

// A missing variable must stop the process at boot. Falling back to a default here
// means a service running against the wrong database or signing with a key nobody
// configured -- failures that surface much later and much worse.
func TestEveryVariableIsRequired(t *testing.T) {
	for missing := range validEnv() {
		if missing == "ALCHEMIST_FETCH_ALLOWLIST" {
			continue // genuinely optional
		}
		env := validEnv()
		delete(env, missing)
		// t.Setenv cannot unset, so an empty value stands in; Load treats blank as
		// absent, which is the behaviour that matters.
		env[missing] = ""
		setEnv(t, env)

		_, err := Load()
		if err == nil {
			t.Errorf("%s was missing and Load succeeded", missing)
			continue
		}
		if !strings.Contains(err.Error(), missing) {
			t.Errorf("%s missing, but the error does not name it: %v", missing, err)
		}
	}
}

// A short key silently weakens every signature and every wrapped content key, so it
// is rejected rather than accepted and quietly padded.
func TestShortSecretsRejected(t *testing.T) {
	for _, tc := range []struct{ key, value, want string }{
		{"ALCHEMIST_KEK", "too-short", "ALCHEMIST_KEK"},
		{"ALCHEMIST_PLAYBACK_KEYS", "k1:short", "ALCHEMIST_PLAYBACK_KEYS"},
	} {
		env := validEnv()
		env[tc.key] = tc.value
		setEnv(t, env)

		_, err := Load()
		if err == nil {
			t.Errorf("%s = %q was accepted", tc.key, tc.value)
			continue
		}
		if !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: error does not name the variable: %v", tc.key, err)
		}
	}
}

func TestEncodeWorkersMustBePositiveInteger(t *testing.T) {
	for _, v := range []string{"0", "-2", "many", "3.5"} {
		env := validEnv()
		env["ALCHEMIST_ENCODE_WORKERS"] = v
		setEnv(t, env)

		if _, err := Load(); err == nil {
			t.Errorf("ALCHEMIST_ENCODE_WORKERS = %q was accepted", v)
		}
	}
}

// The allowlist disables an SSRF defence, so it must stay opt-in and empty by default.
func TestFetchAllowlistDefaultsEmpty(t *testing.T) {
	setEnv(t, validEnv())
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(c.FetchAllowlist) != 0 {
		t.Errorf("FetchAllowlist = %v, want empty when unset", c.FetchAllowlist)
	}
}
