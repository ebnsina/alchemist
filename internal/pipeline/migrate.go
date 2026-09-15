package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"

	"github.com/ebnsina/alchemist/internal/platform/db"
	"github.com/ebnsina/alchemist/internal/platform/keys"
	"github.com/ebnsina/alchemist/internal/platform/providers"
)

type MigrateArgs struct {
	SourceID string `json:"source_id"`
	TenantID string `json:"tenant_id"`
	// Page only exists to vary the uniqueness hash. Without it the follow-up a run
	// queues for the next page is deduplicated against that same run, still Running,
	// and every migration of more than one page stops silently after the first.
	Page int `json:"page"`
}

func (MigrateArgs) Kind() string { return "migrate_source" }

// One migration of a given source at a time: two passes would fight over the cursor
// and double-import. Same explicit state list as bucket sync, and for the same
// reason — River's default includes Completed, which would throttle this to one page
// a day.
func (MigrateArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       QueueIO,
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

type MigrateWorker struct {
	river.WorkerDefaults[MigrateArgs]
	DB    *db.DB
	Keys  *keys.Wrapper
	River *river.Client[pgx.Tx]
}

func (w *MigrateWorker) Timeout(*river.Job[MigrateArgs]) time.Duration {
	return 10 * time.Minute
}

type migrationSource struct {
	Provider                   string
	Config                     map[string]string
	WrappedSecret, SecretNonce []byte
	Cursor                     string
	State                      string
}

// Work takes one page of the far side's library per run and queues the next, so a
// library of any size makes steady progress instead of running into the timeout.
func (w *MigrateWorker) Work(ctx context.Context, job *river.Job[MigrateArgs]) error {
	a := job.Args

	var src migrationSource
	var rawConfig []byte
	err := w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`select provider, config, wrapped_secret, secret_nonce, cursor, state
			   from migration_sources where id = $1`, a.SourceID).
			Scan(&src.Provider, &rawConfig, &src.WrappedSecret, &src.SecretNonce,
				&src.Cursor, &src.State)
	})
	if err != nil {
		return fmt.Errorf("load migration source: %w", err)
	}
	// Paused, finished or abandoned between pages: stop without touching the row.
	if src.State != "previewing" && src.State != "scanning" {
		return nil
	}
	previewing := src.State == "previewing"
	if err := json.Unmarshal(rawConfig, &src.Config); err != nil {
		return fmt.Errorf("migration config: %w", err)
	}

	p, ok := providers.Get(src.Provider)
	if !ok {
		return w.fail(ctx, a, "We no longer support migrating from that service.")
	}
	secret, err := w.Keys.Unwrap(src.WrappedSecret, src.SecretNonce, a.SourceID)
	if err != nil {
		return fmt.Errorf("unwrap migration secret: %w", err)
	}
	creds := providers.Creds{Secret: string(secret), Config: src.Config}

	videos, next, err := p.List(ctx, creds, src.Cursor)
	if err != nil {
		if errors.Is(err, providers.ErrBadCredentials) {
			// Nothing retryable about a rejected key, so it stops here rather than
			// burning attempts and leaving the customer watching a spinner.
			return w.fail(ctx, a, "That key was rejected. Check it and connect again.")
		}
		return fmt.Errorf("list %s: %w", src.Provider, err)
	}

	imported := 0
	for _, v := range videos {
		// While previewing we only write down what is there. No asset, no transcode,
		// nothing fetched from the far side until a person has said yes.
		if previewing {
			if err := w.record(ctx, a, v, "pending", "", nil); err != nil {
				return err
			}
			continue
		}
		done, err := w.importOne(ctx, a, p, creds, v)
		if err != nil {
			return err
		}
		if done {
			imported++
		}
	}

	state := src.State
	if next == "" && !previewing {
		state = "done"
	}
	if err := w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		_, e := tx.Exec(ctx,
			`update migration_sources
			    set cursor = $2, state = $3, discovered = discovered + $4,
			        imported = imported + $5, preview_done = preview_done or $6,
			        last_error = null, updated_at = now()
			  where id = $1 and state = $7`,
			a.SourceID, next, state, len(videos), imported,
			previewing && next == "", src.State)
		return e
	}); err != nil {
		return fmt.Errorf("advance migration: %w", err)
	}

	if next == "" {
		return nil
	}
	// A pause landing mid-page must not be undone by the follow-up we are about to
	// queue, so the state is read again rather than trusted from the top of the run.
	var state2 string
	if err := w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `select state from migration_sources where id = $1`,
			a.SourceID).Scan(&state2)
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("re-read migration state: %w", err)
	}
	if state2 != src.State {
		return nil
	}
	// More to do: queue it now rather than waiting, so a library finishes promptly.
	if _, err := w.River.Insert(ctx, MigrateArgs{
		SourceID: a.SourceID, TenantID: a.TenantID, Page: a.Page + 1,
	}, nil); err != nil {
		return fmt.Errorf("queue next migration page: %w", err)
	}
	return nil
}

// importOne records the video and queues a transcode, unless this remote id has been
// seen before. Dedupe is on the remote id, so re-running a migration is safe.
func (w *MigrateWorker) importOne(
	ctx context.Context, a MigrateArgs, p providers.Provider,
	creds providers.Creds, v providers.Video,
) (bool, error) {
	var seen bool
	if err := w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`select exists (select 1 from migration_items
			                 where source_id = $1 and remote_id = $2
			                   and state <> 'pending')`,
			a.SourceID, v.ID).Scan(&seen)
	}); err != nil {
		return false, fmt.Errorf("check migration item: %w", err)
	}
	if seen {
		return false, nil
	}

	link, err := p.DownloadURL(ctx, creds, v.ID)
	if err != nil {
		// A video the provider will not hand over is a fact about that video. It is
		// recorded and the migration carries on; failing the run would strand every
		// video behind it.
		reason := "The provider would not give us a file for this one."
		if errors.Is(err, providers.ErrBadCredentials) {
			return false, err
		}
		return false, w.record(ctx, a, v, "skipped", reason, nil)
	}

	// Dedupe across the whole account, not just this migration. Somebody who adds one
	// video to their list and re-pastes it should get one video, not a second copy of
	// everything — and the item points at the asset they already have.
	var existing *string
	if err := w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`select id::text from assets
			  where source_url = $1 and state <> 'failed'
			  order by created_at limit 1`, link).Scan(&existing)
	}); err != nil && err != pgx.ErrNoRows {
		return false, fmt.Errorf("look for an existing asset: %w", err)
	}
	if existing != nil {
		return false, w.record(ctx, a, v, "skipped", "You already have this one.", existing)
	}

	var assetID string
	if err := w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`insert into assets (tenant_id, state, ladder_profile, source_url)
			 select $1, 'uploaded', ladder_profile, $2 from tenants where id = $1
			 returning id::text`, a.TenantID, link).Scan(&assetID)
	}); err != nil {
		return false, fmt.Errorf("create migrated asset: %w", err)
	}

	if err := w.record(ctx, a, v, "imported", "", &assetID); err != nil {
		return false, err
	}
	if _, err := w.River.Insert(ctx,
		TranscodeArgs{AssetID: assetID, TenantID: a.TenantID}, nil); err != nil {
		return false, fmt.Errorf("queue migrated transcode: %w", err)
	}
	return true, nil
}

func (w *MigrateWorker) record(
	ctx context.Context, a MigrateArgs, v providers.Video,
	state, reason string, assetID *string,
) error {
	return w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`insert into migration_items (source_id, remote_id, title, asset_id, state, reason)
			 values ($1, $2, $3, $4, $5, nullif($6, ''))
			 on conflict (source_id, remote_id) do update
			    set state = excluded.state, reason = excluded.reason,
			        asset_id = excluded.asset_id`,
			a.SourceID, v.ID, v.Title, assetID, state, reason)
		return err
	})
}

// fail stops the migration with something the customer can read. Returning nil is
// deliberate: the job is finished, the problem is theirs to fix, and a retry would
// only fail the same way.
func (w *MigrateWorker) fail(ctx context.Context, a MigrateArgs, msg string) error {
	return w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`update migration_sources set state = 'failed', last_error = $2, updated_at = now()
			  where id = $1`, a.SourceID, msg)
		return err
	})
}
