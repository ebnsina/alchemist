package live

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// The reaper, and the sweep of the live prefix.
//
// Everything that decides whether a broadcast's segments still matter lives here,
// because it is the same decision in four disguises: a broadcast that ended, one that
// failed, one whose encoder vanished, and one nobody ever published to. Splitting it
// across the live worker and the transcoder is how a leak survives -- the failing
// paths are precisely the ones nobody runs twice.

// LiveGrace is how long a broadcast may go without a segment before it counts as
// over. Two minutes rather than the 60s in docs/06-live.md: a segment upload stalled
// behind a slow object store is not a dead encoder, and ending a match early is worse
// than reclaiming its segments a minute late.
const LiveGrace = 2 * time.Minute

// WaitGrace covers the armed-but-silent case. It is WaitForEncoder plus slack, so the
// worker always gives up on its own first and this only catches a worker that died.
const WaitGrace = WaitForEncoder + 5*time.Minute

// RetryWindow bounds how long a finished broadcast keeps being handed back to the
// VOD path.
//
// ponytail: after a day of failed conversions the segments are left alone and a human
// has to look. Deleting them destroys the only copy of the recording and retrying
// forever hides the failure; the upgrade is an alert on live_sessions where swept_at
// is null and ended_at is old.
const RetryWindow = 24 * time.Hour

// ReapInterval is how often Reap runs for each tenant that has live.
const ReapInterval = time.Minute

// StopGrace is how long a stop may go unanswered before the reaper does it.
//
// The worker reads the flag on its next poll, within a second, so half a minute means
// no worker is reading it -- the job is still queued behind a full box, or the worker
// died. Without this backstop a stop on a queued broadcast does nothing at all until
// the session times out half an hour later, which is exactly the stuck stream the
// customer pressed the button about.
const StopGrace = 30 * time.Second

// OrphanGrace is how long an asset may sit in a live state with no session before it
// counts as one whose arming never finished. Generous, because the session lands a
// moment after the asset and a race here would fail a broadcast that is about to
// start.
const OrphanGrace = 5 * time.Minute

// Reap moves abandoned broadcasts to a terminal state and reclaims the segments of
// every broadcast that no longer needs them. Scoped to one tenant so live_sessions is
// read under ordinary RLS; finding the tenants is the caller's problem.
func (w *Worker) Reap(ctx context.Context, tenantID string) error {
	type row struct {
		sessionID, assetID, streamID, state string
		stopped                             bool
	}
	var rows []row

	if err := w.DB.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		r, err := tx.Query(ctx,
			`select id::text, asset_id::text, stream_id::text, state,
			        stop_requested_at is not null
			   from live_sessions
			  where (state = 'live'
			         and coalesce(last_seen_at, started_at, created_at)
			             < now() - make_interval(secs => $1))
			     or (state = 'waiting' and created_at < now() - make_interval(secs => $2))
			     or (state in ('waiting', 'live')
			         and stop_requested_at < now() - make_interval(secs => $4))
			     or (state = 'failed' and swept_at is null)
			     or (state = 'ended' and swept_at is null
			         and ended_at > now() - make_interval(secs => $3))
			  limit 100`,
			LiveGrace.Seconds(), WaitGrace.Seconds(), RetryWindow.Seconds(),
			StopGrace.Seconds())
		if err != nil {
			return err
		}
		defer r.Close()
		for r.Next() {
			var s row
			if err := r.Scan(&s.sessionID, &s.assetID, &s.streamID, &s.state,
				&s.stopped); err != nil {
				return err
			}
			rows = append(rows, s)
		}
		return r.Err()
	}); err != nil {
		return err
	}

	a := session{TenantID: tenantID}
	for _, s := range rows {
		a.ID = s.sessionID
		if err := w.reapOne(ctx, a, s.assetID, s.streamID, s.state, s.stopped); err != nil {
			return err
		}
	}
	return w.reapOrphans(ctx, tenantID)
}

// reapOrphans ends broadcast assets whose session never landed.
//
// The asset and its session are created in two steps, because live does not own
// assets, and the second step can fail or the process can die between them. Every
// other reaper pass walks sessions, so an asset with none was reachable by nothing:
// it stayed ON AIR forever, with no error and no row to find it by.
func (w *Worker) reapOrphans(ctx context.Context, tenantID string) error {
	ids, err := w.Assets.Orphans(ctx, tenantID, OrphanGrace)
	if err != nil || len(ids) == 0 {
		return err
	}

	// One query rather than one per asset: every broadcast still on air is a candidate
	// here for its whole run, and only the sessionless ones are orphans.
	attached := map[string]bool{}
	if err := w.DB.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx,
			`select distinct asset_id::text from live_sessions where asset_id = any($1::uuid[])`,
			ids)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return err
			}
			attached[id] = true
		}
		return rows.Err()
	}); err != nil {
		return err
	}

	for _, assetID := range ids {
		if attached[assetID] {
			continue
		}
		if err := w.Assets.MarkFailed(ctx, tenantID, assetID, "start_failed"); err != nil {
			return err
		}
	}
	return nil
}

func (w *Worker) reapOne(ctx context.Context, a session, assetID, streamID, state string,
	stopped bool) error {
	switch state {
	case "live":
		// Segments exist, so the broadcast happened: it is converted like any other
		// rather than discarded because the encoder failed to say goodbye.
		return w.endLive(ctx, a, assetID, streamID)
	case "waiting":
		// Nobody ever published. Why says which code: a customer who pressed stop is
		// told that, not that their encoder never turned up.
		code := "no_encoder"
		if stopped {
			code = "stopped"
		}
		_ = w.failLive(ctx, a, assetID, streamID, code)
		return w.sweep(ctx, a, assetID)
	case "failed":
		return w.sweep(ctx, a, assetID)
	}

	// Ended. live_ended means the segments are still the only copy of the recording;
	// any other state means the cmaf/ objects are in place and these are dead weight.
	assetState, err := w.Assets.State(ctx, a.TenantID, assetID)
	if err != nil {
		return err
	}
	if assetState == "live_ended" {
		return w.Queue.ConvertRecording(ctx, a.TenantID, assetID)
	}
	return w.sweep(ctx, a, assetID)
}

// sweep deletes the broadcast's segments and records that it happened. The mark is
// written only after the delete is queued, so a crash in between costs one extra pass
// and never a silently skipped prefix.
func (w *Worker) sweep(ctx context.Context, a session, assetID string) error {
	// The trailing slash is load-bearing: without it a prefix delete could walk into
	// a sibling asset whose id merely starts with this one's.
	if err := w.Queue.ReclaimPrefix(ctx, Prefix(a.TenantID, assetID)+"/"); err != nil {
		return err
	}
	return w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`update live_sessions set swept_at = now() where id = $1`, a.ID)
		return err
	})
}
