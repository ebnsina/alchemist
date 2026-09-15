package adapters

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ebnsina/alchemist/internal/platform/db"
)

// Egress turns bytes served into billable usage.
//
// Counted in memory and folded into one row per tenant per UTC day. A usage_events
// row per byte-range request would be millions a day for one popular video and would
// bill the identical number, while putting a write on the path of every seek.
//
// ponytail: the counter is per process, so an ungraceful kill loses at most one flush
// interval of egress. Persisting it would cost the write this exists to avoid; move
// the counter into the database only if that interval ever matters.
type Egress struct {
	DB      *db.DB
	mu      sync.Mutex
	pending map[string]int64
}

func (e *Egress) RecordEgress(tenantID string, bytes int64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.pending == nil {
		e.pending = map[string]int64{}
	}
	e.pending[tenantID] += bytes
}

// Flush writes what has accumulated, and once more on the way down so a rolling
// deploy does not silently drop the last interval.
func (e *Egress) Flush(ctx context.Context, every time.Duration) {
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
func (e *Egress) drain() map[string]int64 {
	e.mu.Lock()
	defer e.mu.Unlock()
	batch := e.pending
	e.pending = nil
	return batch
}

func (e *Egress) flush(ctx context.Context) {
	for tenantID, n := range e.drain() {
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
