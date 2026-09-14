// Package config loads all runtime configuration from the environment.
// A missing variable is a boot failure, never a silent default.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

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

	// Optional, and only for local harnesses: exact "ip:port" pairs that bypass the
	// SSRF address rules. Never set this in production.
	if v := os.Getenv("ALCHEMIST_FETCH_ALLOWLIST"); strings.TrimSpace(v) != "" {
		c.FetchAllowlist = strings.Split(v, ",")
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
