package pipeline

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"

	"github.com/ebnsina/alchemist/internal/platform/media"
)

// JITArgs generates one deferred rendition and republishes the manifests.
type JITArgs struct {
	AssetID  string `json:"asset_id"`
	TenantID string `json:"tenant_id"`
	Height   int    `json:"height"`
	Codec    string `json:"codec"`
}

func (JITArgs) Kind() string { return "jit_rendition" }

// Highest priority: a viewer is waiting. Unique per (asset, rung) so a burst of
// requests for the same video produces one encode, not one per viewer -- which is
// exactly what happens when something is shared and suddenly popular.
func (JITArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       QueueEncode,
		Priority:    1,
		MaxAttempts: 3,
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
			ByState: []rivertype.JobState{
				rivertype.JobStateAvailable,
				rivertype.JobStatePending,
				rivertype.JobStateRunning,
				rivertype.JobStateRetryable,
				rivertype.JobStateScheduled,
			},
		},
	}
}

type JITWorker struct {
	river.WorkerDefaults[JITArgs]
	*TranscodeWorker
}

func (w *JITWorker) Timeout(*river.Job[JITArgs]) time.Duration { return JobTimeout }

func (w *JITWorker) Work(ctx context.Context, job *river.Job[JITArgs]) error {
	a := job.Args

	var ladderRaw []byte
	var mezzKey *string
	var complexity *float64
	err := w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`select p.rungs, a.mezzanine_key, a.complexity
			   from assets a join ladder_profiles p on p.name = a.ladder_profile
			  where a.id = $1`, a.AssetID).Scan(&ladderRaw, &mezzKey, &complexity)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("load asset: %w", err)
	}
	if mezzKey == nil || *mezzKey == "" {
		// Without a mezzanine the rendition cannot be produced. Regenerating one from
		// the original is a separate, much more expensive path.
		return river.JobCancel(fmt.Errorf("no mezzanine retained for %s", a.AssetID))
	}

	rungs, err := media.ParseLadder(ladderRaw)
	if err != nil {
		return err
	}
	// Reuse the complexity measured at ingest so this rung is tuned to the same
	// content as the rest of the ladder. Re-measuring would cost another sampling
	// pass and could land on a slightly different answer, leaving the ladder
	// internally inconsistent.
	if complexity != nil {
		rungs = media.ApplyComplexity(rungs, *complexity)
	}

	rung, ok := media.Find(rungs, a.Height, a.Codec)
	if !ok {
		return river.JobCancel(fmt.Errorf("rung %d/%s is not in the profile", a.Height, a.Codec))
	}

	dir := filepath.Join(w.WorkDir, a.AssetID+"-jit")
	defer os.RemoveAll(dir)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}

	mezz := filepath.Join(dir, "mezzanine.mp4")
	if err := w.download(ctx, *mezzKey, mezz); err != nil {
		return fmt.Errorf("fetch mezzanine: %w", err)
	}

	if err := w.buildRendition(ctx, a, mezz, dir, rung); err != nil {
		return err
	}
	return w.republish(ctx, a.TenantID, a.AssetID, dir)
}
