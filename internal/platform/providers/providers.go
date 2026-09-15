// Package providers reads a customer's library out of another video host.
//
// Only hosts that will hand back a source file belong here. Platforms built around
// locked-down delivery deliberately expose no downloadable original through their
// API — that is the product, not an oversight — so they cannot be migrated
// automatically and are not listed. The csv adapter is the way across from those.
package providers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Video is the little we need about something on the far side: enough to record it,
// deduplicate it, and show a person what is being moved.
type Video struct {
	ID       string
	Title    string
	Duration float64
}

// ErrNoDownload means the provider has this video but will not hand over a file for
// it. Common and expected — a plan that excludes downloads, or an encode that has not
// finished — so it marks one item skipped rather than failing the whole migration.
var ErrNoDownload = errors.New("provider exposes no downloadable file for this video")

type Provider interface {
	// Name is the stable key stored on the row.
	Name() string
	// Label names the service for a person.
	Label() string
	// NeedsConfig lists the non-secret settings this provider requires, so the form
	// can ask for exactly those and nothing more.
	NeedsConfig() []ConfigField
	// SecretSpec describes what to ask for. Most providers want an API key in a
	// password field; a pasted CSV is neither secret nor one line, and a form that
	// cannot tell the difference asks for the wrong thing.
	SecretSpec() Secret
	// List returns one page. The cursor is opaque; an empty one returned means done.
	List(ctx context.Context, c Creds, cursor string) ([]Video, string, error)
	// DownloadURL resolves a URL the fetcher can pull the source from. These are
	// short-lived on every provider, so it is resolved per video at import time
	// rather than stored.
	DownloadURL(ctx context.Context, c Creds, id string) (string, error)
}

// Kind is "key" for a credential or "text" for something pasted in bulk.
type Secret struct {
	Label string `json:"label"`
	Hint  string `json:"hint"`
	Kind  string `json:"kind"`
}

type ConfigField struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Hint  string `json:"hint"`
}

// Creds carries the secret and whatever non-secret settings the adapter declared.
type Creds struct {
	Secret string
	Config map[string]string
}

var registry = map[string]Provider{
	"vimeo": Vimeo{},
	"bunny": Bunny{},
	"csv":   CSV{},
}

func Get(name string) (Provider, bool) {
	p, ok := registry[name]
	return p, ok
}

func All() []Provider {
	// Stable order, so the form does not reshuffle between requests.
	return []Provider{registry["vimeo"], registry["bunny"], registry["csv"]}
}

// client is shared by every adapter: a short timeout, because a provider that is slow
// to answer should stall one page rather than a worker.
var client = &http.Client{Timeout: 30 * time.Second}

func do(ctx context.Context, method, url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
		res.Body.Close()
		return nil, ErrBadCredentials
	}
	if res.StatusCode >= 400 {
		res.Body.Close()
		return nil, fmt.Errorf("%s %s: %s", method, url, res.Status)
	}
	return res, nil
}

// ErrBadCredentials is the one failure the customer can act on, so it is separated
// from every other transport error.
var ErrBadCredentials = errors.New("the provider rejected those credentials")
