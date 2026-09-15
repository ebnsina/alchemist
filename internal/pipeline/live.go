package pipeline

import (
	"context"
	"errors"
	"time"

	"github.com/riverqueue/river"

	"github.com/ebnsina/alchemist/internal/modules/live"
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
