// Package live is broadcast ingest: streams, sessions, publish authorisation for the
// ingest server, and the worker that encodes and publishes a broadcast.
//
// A stream is an asset from the moment it is armed, so playback URLs, the signed
// prefix, the origin and the player need nothing live-specific. See docs/06-live.md.
//
// The module owns live_streams and live_sessions and writes nothing else. The asset
// a broadcast is watched at belongs to the video side, and the ladder is shared
// reference data, so both arrive as interfaces an adapter binds in cmd/.
package live

import (
	"context"
	"io"

	"github.com/go-chi/chi/v5"

	"github.com/ebnsina/alchemist/internal/platform/db"
	"github.com/ebnsina/alchemist/internal/platform/media"
)

// Assets is the slice of the asset table a broadcast needs. Declared here rather
// than written directly because live does not own assets: the day this module leaves,
// the adapter behind it becomes an HTTP call and nothing in here changes.
type Assets interface {
	CreateForBroadcast(ctx context.Context, tenantID string) (assetID string, err error)
	MarkLive(ctx context.Context, tenantID, assetID string) error
	MarkEnded(ctx context.Context, tenantID, assetID string) error
	MarkFailed(ctx context.Context, tenantID, assetID, code string) error
}

// Ladder resolves the rungs a broadcast's asset was armed with. ladder_profiles is
// shared reference data owned by the account surface, so live reads it through here.
type Ladder interface {
	Rungs(ctx context.Context, tenantID, assetID string) ([]media.Rung, error)
}

// ObjectStore is the one operation live needs of object storage: segments and the
// playlist go out as whole objects, never byte ranges.
type ObjectStore interface {
	Put(ctx context.Context, key string, body io.Reader, contentType string) error
}

// Queue enqueues the session job. An interface so the module carries no job-queue
// driver with it when it is extracted; cmd/ registers the worker that runs it.
type Queue interface {
	EnqueueSession(ctx context.Context, sessionID, tenantID string) error
}

// Events publishes the broadcast lifecycle to whoever is listening. Nil simply does
// not emit, the same way a deployment with no webhooks configured does not.
type Events interface {
	Emit(ctx context.Context, tenantID, event string, data map[string]any) error
}

// Module is the live control surface. It holds a database handle because it owns
// tables; everything it does not own arrives as an interface.
type Module struct {
	db         *db.DB
	ingestHost string
	assets     Assets
	queue      Queue
}

func New(database *db.DB, ingestHost string, assets Assets, queue Queue) *Module {
	return &Module{db: database, ingestHost: ingestHost, assets: assets, queue: queue}
}

// Routes mounts the customer-facing live surface. Mount it inside the authenticated
// group: these endpoints read the tenant the API key resolved to.
func (m *Module) Routes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(m.requireLive)
		r.Post("/live-streams", m.createStream)
		r.Get("/live-streams", m.listStreams)
		r.Get("/live-streams/{id}", m.getStream)
		r.Post("/live-streams/{id}/start", m.startStream)
		r.Post("/live-streams/{id}/key", m.replaceKey)
		r.Delete("/live-streams/{id}", m.deleteStream)
	})
}

// IngestRoutes is mounted separately because the ingest server holds no API key: it
// must sit outside authentication and be bound to a private interface.
func (m *Module) IngestRoutes(r chi.Router) {
	r.Post("/internal/live/authorize", m.authorizeIngest)
}
