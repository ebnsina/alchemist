package pipeline

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"

	"github.com/ebnsina/alchemist/internal/modules/live"
	"github.com/ebnsina/alchemist/internal/platform/db"
)

// ReconcileInterval is how often every active bucket source is re-listed.
//
// This exists because S3 event notifications get dropped, and plenty of
// S3-compatible stores do not emit them at all. Events give low latency; this gives
// the guarantee. A sync built on events alone quietly misses files, and the customer
// discovers it rather than you.
const ReconcileInterval = 15 * time.Minute

type ReconcileArgs struct{}

func (ReconcileArgs) Kind() string { return "bucket_reconcile" }

func (ReconcileArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: QueueIO, MaxAttempts: 2}
}

type ReconcileWorker struct {
	river.WorkerDefaults[ReconcileArgs]
	DB    *db.DB
	River *river.Client[pgx.Tx]
}

func (w *ReconcileWorker) Timeout(*river.Job[ReconcileArgs]) time.Duration {
	return 2 * time.Minute
}

// Work queues a sync for every active source. It reads across tenants, so it runs
// outside the per-tenant RLS path; each queued job then scopes itself to its tenant.
func (w *ReconcileWorker) Work(ctx context.Context, _ *river.Job[ReconcileArgs]) error {
	type src struct{ id, tenant string }
	var sources []src

	// A plain select here returns nothing: RLS is forced on bucket_sources and no
	// tenant is in scope during a cross-tenant sweep. The function runs as the owner
	// and returns only ids, never credentials or customer data.
	rows, err := w.DB.Pool().Query(ctx,
		`select source_id::text, tenant_id::text from active_bucket_sources()`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var s src
		if err := rows.Scan(&s.id, &s.tenant); err != nil {
			return err
		}
		sources = append(sources, s)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, s := range sources {
		// Unique-by-args means a source already syncing is skipped rather than queued
		// twice, so this is safe to run on a short interval.
		if _, err := w.River.Insert(ctx,
			BucketSyncArgs{SourceID: s.id, TenantID: s.tenant}, nil); err != nil {
			return err
		}
	}
	return nil
}

// PeriodicJobs are registered on the River client at startup.
func PeriodicJobs() []*river.PeriodicJob {
	return []*river.PeriodicJob{
		river.NewPeriodicJob(
			river.PeriodicInterval(ReconcileInterval),
			func() (river.JobArgs, *river.InsertOpts) { return ReconcileArgs{}, nil },
			&river.PeriodicJobOpts{RunOnStart: true},
		),
		river.NewPeriodicJob(
			river.PeriodicInterval(SweepInterval),
			func() (river.JobArgs, *river.InsertOpts) { return SweepArgs{}, nil },
			&river.PeriodicJobOpts{RunOnStart: true},
		),
		river.NewPeriodicJob(
			river.PeriodicInterval(live.ReapInterval),
			func() (river.JobArgs, *river.InsertOpts) { return LiveReapArgs{}, nil },
			&river.PeriodicJobOpts{RunOnStart: true},
		),
		// Invoicing checks the day rather than being scheduled monthly, because River
		// has interval jobs and not a calendar. Issuing is keyed on (tenant, period)
		// and the definer function skips anyone already invoiced, so running it hourly
		// on the wrong day costs one cheap query and can never bill twice.
		river.NewPeriodicJob(
			river.PeriodicInterval(InvoiceCheckInterval),
			func() (river.JobArgs, *river.InsertOpts) {
				return InvoiceArgs{PeriodStart: lastClosedMonth(time.Now())}, nil
			},
			&river.PeriodicJobOpts{RunOnStart: true},
		),
		// Collection is separate and slower: a declined card is not worth retrying
		// every hour, and three attempts a day apart is the dunning schedule.
		river.NewPeriodicJob(
			river.PeriodicInterval(CollectInterval),
			func() (river.JobArgs, *river.InsertOpts) { return CollectArgs{}, nil },
			&river.PeriodicJobOpts{RunOnStart: false},
		),
		// Safe to run on start because the day's row is replaced, not added to.
		river.NewPeriodicJob(
			river.PeriodicInterval(StorageInterval),
			func() (river.JobArgs, *river.InsertOpts) { return StorageArgs{}, nil },
			&river.PeriodicJobOpts{RunOnStart: true},
		),
	}
}

// InvoiceCheckInterval and CollectInterval pace the two billing jobs.
//
// Issuing is idempotent per (tenant, period) so checking often is harmless and means a
// box that was down on the first of the month still invoices when it comes back.
// Collecting is not idempotent in the same way -- every run is a real attempt against a
// card network -- so it runs daily and the attempt counter does the rest.
const (
	InvoiceCheckInterval = time.Hour
	CollectInterval      = 24 * time.Hour
)

// lastClosedMonth is the month to bill: the one before the month we are in. Billing
// the current month would invoice usage that is still accruing.
func lastClosedMonth(now time.Time) string {
	first := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	return first.AddDate(0, -1, 0).Format("2006-01-02")
}
