// Package config loads all runtime configuration from the environment.
// A missing variable is a boot failure, never a silent default.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// minSecretBytes is the floor for anything used as key material. Shorter secrets
// weaken every playback signature and every wrapped content key at once.
const minSecretBytes = 32

type Config struct {
	HTTPAddr       string
	DatabaseURL    string
	S3Endpoint     string
	S3Region       string
	S3Bucket       string
	S3AccessKey    string
	S3SecretKey    string
	PlaybackKeys   []string
	WorkDir        string
	KEK            string
	AdminKey       string
	FetchAllowlist []string
	EncodeWorkers  int
}

func Load() (*Config, error) {
	var missing []string
	get := func(k string) string {
		v := os.Getenv(k)
		if strings.TrimSpace(v) == "" {
			missing = append(missing, k)
		}
		return v
	}

	c := &Config{
		HTTPAddr:    get("ALCHEMIST_HTTP_ADDR"),
		DatabaseURL: get("ALCHEMIST_DATABASE_URL"),
		S3Endpoint:  get("ALCHEMIST_S3_ENDPOINT"),
		S3Region:    get("ALCHEMIST_S3_REGION"),
		S3Bucket:    get("ALCHEMIST_S3_BUCKET"),
		S3AccessKey: get("ALCHEMIST_S3_ACCESS_KEY"),
		S3SecretKey: get("ALCHEMIST_S3_SECRET_KEY"),
		WorkDir:     get("ALCHEMIST_WORK_DIR"),
		KEK:         get("ALCHEMIST_KEK"),
	}

	// "kid:secret,kid:secret" -- the first signs, all are accepted, so a key can be
	// rolled out to every edge before it starts being used.
	c.PlaybackKeys = strings.Split(get("ALCHEMIST_PLAYBACK_KEYS"), ",")

	// Optional. When unset the admin surface returns 404 rather than existing
	// unprotected, so forgetting to configure it cannot expose tenant creation.
	c.AdminKey = os.Getenv("ALCHEMIST_ADMIN_KEY")
	if c.AdminKey != "" && len(c.AdminKey) < minSecretBytes {
		return nil, fmt.Errorf("ALCHEMIST_ADMIN_KEY must be at least %d bytes, got %d",
			minSecretBytes, len(c.AdminKey))
	}

	// Optional, and only for local harnesses: exact "ip:port" pairs that bypass the
	// SSRF address rules. Never set this in production.
	if v := os.Getenv("ALCHEMIST_FETCH_ALLOWLIST"); strings.TrimSpace(v) != "" {
		c.FetchAllowlist = strings.Split(v, ",")
	}

	// Secret strength is checked here rather than only where the secrets are used.
	// Config is the boot contract: everything that makes the process refuse to start
	// belongs in one place, before anything else is constructed. The checks in
	// keys.NewWrapper and signing.NewKeyring stay as defence in depth.
	if len(c.KEK) < minSecretBytes {
		return nil, fmt.Errorf("ALCHEMIST_KEK must be at least %d bytes, got %d",
			minSecretBytes, len(c.KEK))
	}
	for _, entry := range c.PlaybackKeys {
		_, secret, ok := strings.Cut(strings.TrimSpace(entry), ":")
		if !ok || len(secret) < minSecretBytes {
			return nil, fmt.Errorf(
				"ALCHEMIST_PLAYBACK_KEYS entries must be kid:secret with a secret of "+
					"at least %d bytes", minSecretBytes)
		}
	}

	workers, err := strconv.Atoi(get("ALCHEMIST_ENCODE_WORKERS"))
	if err == nil && workers > 0 {
		c.EncodeWorkers = workers
	} else if len(missing) == 0 {
		return nil, fmt.Errorf("ALCHEMIST_ENCODE_WORKERS must be a positive integer")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return c, nil
}
