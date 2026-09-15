package pipeline

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"

	"github.com/ebnsina/alchemist/internal/platform/db"
)

// StorageInterval is how often stored bytes become a billable quantity. Storage is a
// rate, not an event, so it has to be sampled; a day is the shortest period anyone
// bills on and the cheapest correct thing.
const StorageInterval = 24 * time.Hour

type StorageArgs struct{}

func (StorageArgs) Kind() string { return "storage_usage" }

func (StorageArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: QueueIO, MaxAttempts: 2}
}

// StorageWorker records GB-hours per tenant from what is actually stored.
type StorageWorker struct {
	river.WorkerDefaults[StorageArgs]
	DB *db.DB
}

func (w *StorageWorker) Timeout(*river.Job[StorageArgs]) time.Duration { return 5 * time.Minute }

func (w *StorageWorker) Work(ctx context.Context, _ *river.Job[StorageArgs]) error {
	type stored struct {
		tenantID string
		bytes    float64
	}
	var totals []stored

	// Cross-tenant, hence the definer function: under forced RLS with no tenant in
	// scope a plain select returns zero rows and reports no error at all.
	rows, err := w.DB.Pool().Query(ctx,
		`select tenant_id::text, bytes::float8 from tenant_stored_bytes()`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var s stored
		if err := rows.Scan(&s.tenantID, &s.bytes); err != nil {
			return err
		}
		totals = append(totals, s)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, s := range totals {
		gbHours := s.bytes / 1e9 * StorageInterval.Hours()
		// Replaced rather than added: this is a snapshot of a day, so a restart that
		// runs it twice must not bill the day twice.
		if err := w.DB.AsTenant(ctx, s.tenantID, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx,
				`insert into usage_events (tenant_id, kind, quantity, unit)
				 values ($1, 'storage', $2, 'gb_hours')
				 on conflict (tenant_id, kind, ((occurred_at at time zone 'UTC')::date))
				   where asset_id is null
				 do update set quantity = excluded.quantity`,
				s.tenantID, gbHours)
			return err
		}); err != nil {
			return err
		}
	}
	return nil
}
