package pipeline

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"

	"github.com/ebnsina/alchemist/internal/platform/db"
	"github.com/ebnsina/alchemist/internal/platform/storage"
)

// SweepInterval is how often failed original-deletions are retried.
const SweepInterval = time.Hour

type SweepArgs struct{}

func (SweepArgs) Kind() string { return "retention_sweep" }

func (SweepArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: QueueIO, MaxAttempts: 2}
}

// SweepWorker retries originals whose deletion failed.
//
// Storage errors are usually transient, but nothing retried them, so one blip meant
// paying to store that file indefinitely. Recording the error without ever acting on
// it is the worst of both: the cost is invisible and nothing fixes it.
type SweepWorker struct {
	river.WorkerDefaults[SweepArgs]
	DB    *db.DB
	Store *storage.Store
	// BatchSize bounds one pass so a large backlog does not monopolise the queue.
	BatchSize int
}

func (w *SweepWorker) Timeout(*river.Job[SweepArgs]) time.Duration { return 5 * time.Minute }

func (w *SweepWorker) Work(ctx context.Context, _ *river.Job[SweepArgs]) error {
	batch := w.BatchSize
	if batch <= 0 {
		batch = 100
	}

	type pending struct {
		assetID, tenantID, sourceKey string
	}
	var items []pending

	// Runs across tenants before any is in scope, hence the definer function.
	rows, err := w.DB.Pool().Query(ctx,
		`select asset_id::text, tenant_id::text, source_key
		   from assets_pending_cleanup($1)`, batch)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var p pending
		if err := rows.Scan(&p.assetID, &p.tenantID, &p.sourceKey); err != nil {
			return err
		}
		items = append(items, p)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, p := range items {
		if err := w.Store.Delete(ctx, p.sourceKey); err != nil {
			continue // still failing; the next sweep tries again
		}
		_ = w.DB.AsTenant(ctx, p.tenantID, func(tx pgx.Tx) error {
			_, e := tx.Exec(ctx,
				`update assets
				    set source_deleted_at = now(), source_key = null,
				        last_retention_error = null
				  where id = $1`, p.assetID)
			return e
		})
	}
	return nil
}
