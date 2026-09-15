package pipeline

import (
	"context"
	"errors"
	"time"

	"github.com/riverqueue/river"

	"github.com/ebnsina/alchemist/internal/modules/live"
	"github.com/ebnsina/alchemist/internal/platform/db"
)

// The River binding for the live module. The broadcast itself lives in
// internal/modules/live, which holds no queue: this is the ten lines a deployment
// writes to schedule it, and the only thing a split would rewrite.

// QueueLive is sized separately because a live job holds its slot for the whole
// broadcast. Sharing encode_cpu would let two football matches drain every worker and
// leave the VOD queue untouched for three hours.
const QueueLive = "live"

type LiveArgs struct {
	SessionID string `json:"session_id"`
	TenantID  string `json:"tenant_id"`
}

func (LiveArgs) Kind() string { return "live_session" }

// MaxAttempts is 1 on purpose. A broadcast cannot be retried: by the time the job
// failed the encoder has gone and the moment has passed. Retrying would only rebind
// the port and wait out the timeout again.
func (LiveArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: QueueLive, MaxAttempts: 1}
}

type LiveWorker struct {
	river.WorkerDefaults[LiveArgs]
	Live *live.Worker
}

func (w *LiveWorker) Timeout(*river.Job[LiveArgs]) time.Duration { return live.Timeout }

func (w *LiveWorker) Work(ctx context.Context, job *river.Job[LiveArgs]) error {
	err := w.Live.Run(ctx, job.Args.SessionID, job.Args.TenantID)
	// The outcome is already recorded against the session, so a retry would only
	// rebind and wait out the timeout again.
	if errors.Is(err, live.ErrBroadcastFailed) {
		return river.JobCancel(err)
	}
	return err
}

// LiveReapArgs sweeps abandoned broadcasts and the segments of finished ones.
//
// Without it a worker that dies mid-broadcast leaves an asset live forever, its
// segments never converted and never deleted -- silently, which is the expensive kind
// of failure. See docs/06-live.md.
type LiveReapArgs struct{}

func (LiveReapArgs) Kind() string { return "live_reap" }

func (LiveReapArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: QueueIO, MaxAttempts: 2}
}

type LiveReapWorker struct {
	river.WorkerDefaults[LiveReapArgs]
	DB   *db.DB
	Live *live.Worker
}

func (w *LiveReapWorker) Timeout(*river.Job[LiveReapArgs]) time.Duration { return 5 * time.Minute }

// Work reaps one tenant at a time so the module's own queries run under ordinary RLS.
//
// ponytail: one pass per live tenant per minute. The ceiling is the number of tenants
// with live enabled, which is a sales number long before it is a query cost; the
// upgrade is for live_tenants() to return only tenants with unswept sessions.
func (w *LiveReapWorker) Work(ctx context.Context, _ *river.Job[LiveReapArgs]) error {
	var tenants []string
	// Cross-tenant, so it goes through the definer function: RLS is forced on
	// tenant_limits and a plain select returns zero rows without erroring.
	rows, err := w.DB.Pool().Query(ctx, `select tenant_id::text from live_tenants()`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}
		tenants = append(tenants, id)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, tenantID := range tenants {
		if err := w.Live.Reap(ctx, tenantID); err != nil {
			return err
		}
	}
	return nil
}
