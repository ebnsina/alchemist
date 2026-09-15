package adapters

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ebnsina/alchemist/internal/platform/db"
)

// Usage turns what the origin served into billable numbers: bytes, and the peak
// number of viewers watching one asset at once.
//
// Counted in memory and folded into one row per tenant per UTC day. A usage_events
// row per byte-range request would be millions a day for one popular video and would
// bill the identical number, while putting a write on the path of every seek.
//
// ponytail: the counter is per process, so an ungraceful kill loses at most one flush
// interval of egress. Persisting it would cost the write this exists to avoid; move
// the counter into the database only if that interval ever matters.
type Usage struct {
	DB      *db.DB
	mu      sync.Mutex
	pending map[string]int64
	// Distinct viewers seen on each asset since the last flush. The set is emptied
	// every interval, so its size is the number watching during that interval —
	// which is what "concurrent" means for billing.
	viewers map[assetKey]map[string]struct{}
	// Devices each bound viewer is streaming from, and when each was last seen.
	// Rolling rather than drained, because a cap that resets every flush would let
	// a shared login back in once a minute.
	devices map[bindingKey]map[string]time.Time
	caps    map[string]deviceCap
}

type assetKey struct{ tenant, asset string }

type bindingKey struct{ tenant, viewer string }

type deviceCap struct {
	n     int
	until time.Time
}

// deviceWindow is how long after its last playlist read a device still counts as
// watching. A player re-reads the playlist every segment duration, so this is
// several missed polls -- long enough that a tunnel or a lift does not free a slot,
// short enough that closing a tab does within a minute or two.
const deviceWindow = 2 * time.Minute

// AllowViewer reports whether one more device may stream for this viewer.
//
// The newcomer is refused, never the incumbent: kicking the oldest session lets a
// shared password boot the student who paid out of their own lecture, repeatedly and
// invisibly, which is a worse product than telling the twenty-first friend no.
//
// ponytail: the count is per process, so a restart forgives everyone and two origins
// count separately. Move it into Postgres or Redis when several actually run at once.
func (e *Usage) AllowViewer(ctx context.Context, tenantID, viewer, device string) bool {
	n := e.deviceCap(ctx, tenantID)
	if n <= 0 {
		return true
	}
	return e.allowDevice(bindingKey{tenantID, viewer}, device, n, time.Now())
}

// allowDevice is the decision itself, with the clock and the cap passed in so it can
// be tested without a database.
func (e *Usage) allowDevice(k bindingKey, device string, limit int, now time.Time) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.devices == nil {
		e.devices = map[bindingKey]map[string]time.Time{}
	}
	seen := e.devices[k]
	if seen == nil {
		seen = map[string]time.Time{}
		e.devices[k] = seen
	}
	if _, known := seen[device]; !known {
		for d, at := range seen {
			if now.Sub(at) > deviceWindow {
				delete(seen, d)
			}
		}
		if len(seen) >= limit {
			return false
		}
	}
	seen[device] = now
	return true
}

// deviceCap reads the tenant's cap, cached for a minute: this is on the playback
// path, and a query per playlist read would put the database back in front of every
// viewer. A read that fails caches nothing and allows -- a database blip must not
// stop playback.
func (e *Usage) deviceCap(ctx context.Context, tenantID string) int {
	e.mu.Lock()
	c, ok := e.caps[tenantID]
	e.mu.Unlock()
	if ok && time.Now().Before(c.until) {
		return c.n
	}

	var n int
	err := e.DB.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, `select max_viewer_devices from tenant_limits`).Scan(&n)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil // no limits row means plan defaults, which is no cap
		}
		return err
	})
	if err != nil {
		return 0
	}
	e.mu.Lock()
	if e.caps == nil {
		e.caps = map[string]deviceCap{}
	}
	e.caps[tenantID] = deviceCap{n: n, until: time.Now().Add(time.Minute)}
	e.mu.Unlock()
	return n
}

// RecordViewer notes one viewer on one asset. Cheap on purpose: this runs on every
// playlist read, which for a live stream is once per viewer per segment duration.
func (e *Usage) RecordViewer(tenantID, assetID, viewer string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.viewers == nil {
		e.viewers = map[assetKey]map[string]struct{}{}
	}
	k := assetKey{tenantID, assetID}
	if e.viewers[k] == nil {
		e.viewers[k] = map[string]struct{}{}
	}
	e.viewers[k][viewer] = struct{}{}
}

func (e *Usage) RecordEgress(tenantID string, bytes int64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.pending == nil {
		e.pending = map[string]int64{}
	}
	e.pending[tenantID] += bytes
}

// Flush writes what has accumulated, and once more on the way down so a rolling
// deploy does not silently drop the last interval.
func (e *Usage) Flush(ctx context.Context, every time.Duration) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			final, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
			defer cancel()
			e.flush(final)
			return
		case <-t.C:
			e.flush(ctx)
		}
	}
}

// drain takes what has accumulated and resets the counter in one step, so bytes are
// never billed twice and never counted into a batch that is already being written.
func (e *Usage) drain() (map[string]int64, map[assetKey]int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	batch := e.pending
	e.pending = nil

	peaks := make(map[assetKey]int, len(e.viewers))
	for k, set := range e.viewers {
		peaks[k] = len(set)
	}
	e.viewers = nil

	// Sweep viewers nobody is watching as any more, or the map grows for the life
	// of the process.
	now := time.Now()
	for k, seen := range e.devices {
		for d, at := range seen {
			if now.Sub(at) > deviceWindow {
				delete(seen, d)
			}
		}
		if len(seen) == 0 {
			delete(e.devices, k)
		}
	}
	return batch, peaks
}

func (e *Usage) flush(ctx context.Context) {
	bytes, peaks := e.drain()
	for k, n := range peaks {
		// greatest, not sum: the peak is the highest interval of the day, and a day
		// with many quiet intervals must not add up to a busy one.
		_ = e.DB.AsTenant(ctx, k.tenant, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx,
				`insert into usage_events (tenant_id, asset_id, kind, quantity, unit)
				 values ($1, $2, 'peak_viewers', $3, 'viewers')
				 on conflict (tenant_id, asset_id, kind, ((occurred_at at time zone 'UTC')::date))
				   where asset_id is not null
				 do update set quantity = greatest(usage_events.quantity, excluded.quantity)`,
				k.tenant, k.asset, n)
			return err
		})
	}
	for tenantID, n := range bytes {
		err := e.DB.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx,
				`insert into usage_events (tenant_id, kind, quantity, unit)
				 values ($1, 'egress', $2, 'bytes')
				 on conflict (tenant_id, kind, ((occurred_at at time zone 'UTC')::date))
				   where asset_id is null
				 do update set quantity = usage_events.quantity + excluded.quantity`,
				tenantID, n)
			return err
		})
		if err != nil {
			// Put the bytes back rather than lose them; the next tick tries again.
			e.RecordEgress(tenantID, n)
		}
	}
}
