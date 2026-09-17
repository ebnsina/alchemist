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
	// WebOrigins are the browser origins allowed to call the account endpoints with
	// credentials. Empty means the account surface is off, not open to everyone.
	WebOrigins []string
	// SessionDomain scopes the session cookie. Empty means host-only, which is
	// correct for localhost and wrong the moment the site and the API differ.
	SessionDomain string
	// SessionSecure marks the cookie Secure. Off only for a local http harness.
	SessionSecure bool
	// LiveIngestHost is the address encoders publish to. Empty leaves the live
	// surface unmounted, the same way no web origin leaves the account surface off.
	LiveIngestHost string
	// PlayerURL is the module URL of the player bundle the embed page loads, e.g.
	// https://cdn.example/player/v1/alchemist-player.js. Empty leaves /e/ unmounted:
	// a page with no player could only ever render an empty box.
	PlayerURL string
	// LivePullBase is where the transcoder reads a published stream from, e.g.
	// rtsp://127.0.0.1:8554. The ingest server terminates SRT and RTMP, checks the
	// stream key against the API, and this is the private side of it.
	LivePullBase string
	// LiveMaxStreams bounds concurrent broadcasts on this box. Cores, not ports:
	// one rung is roughly one core held for the length of the broadcast.
	LiveMaxStreams int

	// Payment gateways. Each is optional and each is a pair: half a gateway is worse
	// than none, because a webhook with no secret to verify against fails closed and
	// every payment silently goes uncredited. A gateway with no keys leaves invoices
	// visible and unpayable online, which is the honest state for a deployment that
	// has not been given any.
	StripeSecretKey     string
	StripeWebhookSecret string
	SSLCommerzStoreID   string
	SSLCommerzStorePass string
	SSLCommerzSandbox   bool
	// BillingURL is where a customer lands after paying, and what a checkout link
	// returns to. Required once any gateway is configured: a redirect to nowhere is
	// how a paid invoice looks like a failed one.
	BillingURL string
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

	// Optional. Unset means no browser origin may call the account endpoints, which
	// is the safe default: a wildcard with credentials is not something to arrive at
	// by forgetting a variable.
	if v := os.Getenv("ALCHEMIST_WEB_ORIGINS"); strings.TrimSpace(v) != "" {
		for _, o := range strings.Split(v, ",") {
			if o = strings.TrimSpace(o); o != "" {
				c.WebOrigins = append(c.WebOrigins, o)
			}
		}
	}
	c.SessionDomain = strings.TrimSpace(os.Getenv("ALCHEMIST_SESSION_DOMAIN"))
	// Defaults to on. Turning the Secure flag off has to be a deliberate act.
	c.SessionSecure = strings.TrimSpace(os.Getenv("ALCHEMIST_SESSION_INSECURE")) == ""

	// Optional as a pair. Live is off unless both are set; one without the other is a
	// boot failure rather than a surface that half exists.
	c.PlayerURL = strings.TrimSpace(os.Getenv("ALCHEMIST_PLAYER_URL"))

	c.StripeSecretKey = strings.TrimSpace(os.Getenv("ALCHEMIST_STRIPE_SECRET_KEY"))
	c.StripeWebhookSecret = strings.TrimSpace(os.Getenv("ALCHEMIST_STRIPE_WEBHOOK_SECRET"))
	if (c.StripeSecretKey == "") != (c.StripeWebhookSecret == "") {
		return nil, fmt.Errorf(
			"ALCHEMIST_STRIPE_SECRET_KEY and ALCHEMIST_STRIPE_WEBHOOK_SECRET must be set together")
	}
	c.SSLCommerzStoreID = strings.TrimSpace(os.Getenv("ALCHEMIST_SSLCOMMERZ_STORE_ID"))
	c.SSLCommerzStorePass = strings.TrimSpace(os.Getenv("ALCHEMIST_SSLCOMMERZ_STORE_PASSWORD"))
	if (c.SSLCommerzStoreID == "") != (c.SSLCommerzStorePass == "") {
		return nil, fmt.Errorf(
			"ALCHEMIST_SSLCOMMERZ_STORE_ID and ALCHEMIST_SSLCOMMERZ_STORE_PASSWORD must be set together")
	}
	// Live money by default. A sandbox store that is treated as live refuses every
	// payment; a live store treated as sandbox charges real cards against a test
	// endpoint. Opting into sandbox has to be the deliberate act.
	c.SSLCommerzSandbox = strings.TrimSpace(os.Getenv("ALCHEMIST_SSLCOMMERZ_SANDBOX")) != ""
	c.BillingURL = strings.TrimRight(strings.TrimSpace(os.Getenv("ALCHEMIST_BILLING_URL")), "/")
	if c.BillingURL == "" && (c.StripeSecretKey != "" || c.SSLCommerzStoreID != "") {
		return nil, fmt.Errorf(
			"ALCHEMIST_BILLING_URL is required when a payment gateway is configured")
	}

	c.LiveIngestHost = strings.TrimSpace(os.Getenv("ALCHEMIST_LIVE_INGEST_HOST"))
	c.LivePullBase = strings.TrimSpace(os.Getenv("ALCHEMIST_LIVE_PULL_BASE"))
	if (c.LiveIngestHost == "") != (c.LivePullBase == "") {
		return nil, fmt.Errorf(
			"ALCHEMIST_LIVE_INGEST_HOST and ALCHEMIST_LIVE_PULL_BASE must be set together")
	}
	if c.LivePullBase != "" {
		if !strings.HasPrefix(c.LivePullBase, "rtsp://") {
			return nil, fmt.Errorf(
				"ALCHEMIST_LIVE_PULL_BASE must be an rtsp:// address, got %q", c.LivePullBase)
		}
		c.LivePullBase = strings.TrimSuffix(c.LivePullBase, "/")
		c.LiveMaxStreams = 2
		if v := strings.TrimSpace(os.Getenv("ALCHEMIST_LIVE_MAX_STREAMS")); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n < 1 {
				return nil, fmt.Errorf(
					"ALCHEMIST_LIVE_MAX_STREAMS must be a positive number, got %q", v)
			}
			c.LiveMaxStreams = n
		}
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
