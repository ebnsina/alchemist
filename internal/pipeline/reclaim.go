package pipeline

import (
	"context"
	"time"

	"github.com/riverqueue/river"

	"github.com/ebnsina/alchemist/internal/platform/storage"
)

// ReclaimArgs removes object storage nothing references any more.
//
// The keys travel in the job rather than being re-derived, because by the time this
// runs the rows that named them are gone. Deletes are idempotent, so a retry after a
// partial pass is safe.
type ReclaimArgs struct {
	Keys     []string `json:"keys,omitempty"`
	Prefixes []string `json:"prefixes,omitempty"`
}

func (ReclaimArgs) Kind() string { return "reclaim_objects" }

func (ReclaimArgs) InsertOpts() river.InsertOpts {
	// Generous retries: a storage blip here is the difference between reclaimed
	// space and paying to store an asset the customer has already deleted.
	return river.InsertOpts{Queue: QueueIO, MaxAttempts: 5}
}

type ReclaimWorker struct {
	river.WorkerDefaults[ReclaimArgs]
	Store *storage.Store
}

func (w *ReclaimWorker) Timeout(*river.Job[ReclaimArgs]) time.Duration { return 10 * time.Minute }

func (w *ReclaimWorker) Work(ctx context.Context, job *river.Job[ReclaimArgs]) error {
	for _, key := range job.Args.Keys {
		if err := w.Store.Delete(ctx, key); err != nil {
			return err
		}
	}
	for _, prefix := range job.Args.Prefixes {
		cursor := ""
		for {
			page, err := w.Store.List(ctx, prefix, cursor, 1000)
			if err != nil {
				return err
			}
			for _, obj := range page.Objects {
				if err := w.Store.Delete(ctx, obj.Key); err != nil {
					return err
				}
			}
			if !page.HasMore {
				break
			}
			cursor = page.Cursor
		}
	}
	return nil
}
