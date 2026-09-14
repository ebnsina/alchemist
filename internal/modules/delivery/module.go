// Package delivery is the playback origin: manifests, media byte ranges, content
// keys, and edge authorization.
//
// It is deliberately the thinnest module. It serves every byte the platform
// delivers, has almost no business logic, and wants to run close to storage and the
// edge -- an entirely different scaling profile from the control plane. It is
// therefore the most likely candidate for extraction into its own service, and the
// seams here are shaped for that: it depends on interfaces it declares itself, never
// on another module.
package delivery

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/go-chi/chi/v5"

	"github.com/ebnsina/alchemist/internal/platform/signing"
)

// ObjectStore is the slice of object storage this module needs. Declared here rather
// than imported wholesale so extraction does not drag the rest of the platform along.
type ObjectStore interface {
	GetPassthrough(ctx context.Context, key, rangeHeader string) (*Object, error)
}

// ContentKeys resolves an asset's decryption key. The transcoder writes keys through
// the same interface, so neither module has to know how the other stores them.
type ContentKeys interface {
	Get(ctx context.Context, tenantID, assetID string) ([]byte, error)
	Put(ctx context.Context, tenantID, assetID string, keyID, key []byte) error
}

// PlaybackObserver is told when an asset's master playlist is served, so deferred
// renditions can be generated on demand.
//
// An interface rather than a queue handle: delivery must not know that a job queue
// exists, or it stops being separable. A standalone origin passes nil and simply does
// not trigger generation.
type PlaybackObserver interface {
	OnPlaybackStarted(ctx context.Context, tenantID, assetID string)
}

// AssetResolver maps a requested asset to the storage prefix that actually holds its
// media.
//
// They differ when an asset was deduplicated: the media is stored once under the
// first asset that carried that content, and every later asset with the same bytes
// points at it. Without this the duplicate reports ready and every byte range 404s.
type AssetResolver interface {
	StoragePrefix(ctx context.Context, tenantID, assetID string) (string, error)
}

// Module is the delivery plane. Constructing it needs no database handle: storage
// and key lookup arrive as interfaces, which is what makes it separable.
type Module struct {
	store    ObjectStore
	keys     ContentKeys
	signer   *signing.Keyring
	observer PlaybackObserver
	resolver AssetResolver
}

func New(store ObjectStore, keys ContentKeys, signer *signing.Keyring) *Module {
	return &Module{store: store, keys: keys, signer: signer}
}

// WithResolver enables deduplicated assets to resolve to the media they share.
// Without it, storage paths are taken literally from the request.
func (m *Module) WithResolver(r AssetResolver) *Module {
	m.resolver = r
	return m
}

// prefix is where this asset's media actually lives.
func (m *Module) prefix(ctx context.Context, tenantID, assetID string) string {
	if m.resolver != nil {
		if p, err := m.resolver.StoragePrefix(ctx, tenantID, assetID); err == nil && p != "" {
			return p
		}
	}
	return fmt.Sprintf("cmaf/%s/%s", tenantID, assetID)
}

// WithObserver attaches on-demand rendition generation. Optional by design.
func (m *Module) WithObserver(o PlaybackObserver) *Module {
	m.observer = o
	return m
}

// Routes mounts the public playback surface. A standalone origin service mounts
// exactly this and nothing else.
func (m *Module) Routes(r chi.Router) {
	// Methods are registered explicitly. chi's Handle binds every method, which would
	// mean DELETE and PUT on a media URL reaching the read handler and being answered
	// with content -- harmless here, but it misleads caches and proxies and invites a
	// later handler to act on a verb nobody intended to expose.
	//
	// The key route is declared first so "key" is not matched as a filename.
	r.Get("/playback/{tenant}/{asset}/key", m.serveContentKey)
	r.Head("/playback/{tenant}/{asset}/key", m.serveContentKey)
	r.Options("/playback/{tenant}/{asset}/key", m.serveContentKey)

	r.Get("/playback/{tenant}/{asset}/{file}", m.servePlayback)
	r.Head("/playback/{tenant}/{asset}/{file}", m.servePlayback)
	r.Options("/playback/{tenant}/{asset}/{file}", m.servePlayback)

	r.Get("/internal/verify-playback", m.verifyPlayback)
}

// SignPlayback mints the URLs the control plane hands to customers. It lives here
// because the verification rules live here; keeping them together is what stops the
// two drifting apart.
func (m *Module) SignPlayback(prefix string, expUnix int64) (kid, sig string) {
	return m.signer.Sign(prefix, expUnix)
}

// VerifyPlayback exposes signature checking to other modules that authorize viewer
// traffic, such as the QoE beacon endpoint, without duplicating the rules.
func (m *Module) VerifyPlayback(prefix, kid, sig, exp string) bool {
	return m.verify(prefix, kid, sig, exp)
}

func (m *Module) verify(prefix, kid, sig, exp string) bool {
	return m.signer.Verify(prefix, kid, sig, exp)
}

// Object mirrors the storage response fields the origin passes through verbatim so
// that range and conditional requests behave correctly.
type Object struct {
	Body          io.ReadCloser
	ETag          string
	ContentLength int64
	ContentRange  string
}

// Errors a store may return that the origin maps to specific status codes.
var (
	ErrNotFound            = errors.New("object not found")
	ErrRangeNotSatisfiable = errors.New("range not satisfiable")
)
