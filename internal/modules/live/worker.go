package live

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ebnsina/alchemist/internal/platform/db"
	"github.com/ebnsina/alchemist/internal/platform/media"
)

// Timeout bounds a single broadcast, including the wait for the encoder to connect
// at all. An armed stream nobody publishes to releases its slot here.
const Timeout = 6 * time.Hour

// pollInterval is how often the output directory is swept for new segments. Half a
// segment, so a segment is never more than one interval behind the viewer.
const pollInterval = time.Second

// WaitForEncoder is how long an armed stream waits for somebody to publish before
// giving up. Generous, because a class that starts late is normal and the cost of
// waiting is one idle worker slot.
const WaitForEncoder = 30 * time.Minute

// retryInterval is how often the transcoder re-checks whether a publisher has
// arrived. Two seconds is under one segment, so nothing is missed at the start.
const retryInterval = 2 * time.Second

// ErrBroadcastFailed marks an outcome already recorded against the session. A
// broadcast cannot be retried -- by the time it failed the encoder has gone -- so
// whatever runs Run must not queue it again.
var ErrBroadcastFailed = errors.New("broadcast failed")

// Worker runs one broadcast. It holds a database handle for the tables live owns and
// interfaces for everything else.
type Worker struct {
	DB     *db.DB
	Store  ObjectStore
	Assets Assets
	Ladder Ladder
	Events Events
	Queue  Queue
	// PullBase is the ingest server's private address, e.g. rtsp://127.0.0.1:8554.
	PullBase string
	WorkDir  string
}

// Run runs one broadcast: wait for the encoder, encode and segment, and publish each
// segment as it lands.
//
// ponytail: one rung, no ABR. The ceiling is that a viewer gets a single bitrate
// while the broadcast is on; the upgrade is additive -- N rungs is N outputs in the
// same ffmpeg command on the same GOP grid. See docs/06-live.md.
func (w *Worker) Run(ctx context.Context, sessionID, tenantID string) error {
	a := session{ID: sessionID, TenantID: tenantID}

	var assetID, streamID string
	err := w.DB.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`select s.asset_id::text, s.stream_id::text
			   from live_sessions s
			   join live_streams l on l.id = s.stream_id
			  where s.id = $1`, sessionID).
			Scan(&assetID, &streamID)
	})
	if err != nil {
		return fmt.Errorf("load live session %s: %w", sessionID, err)
	}

	rungs, err := w.Ladder.Rungs(ctx, tenantID, assetID)
	if err != nil {
		return w.failLive(ctx, a, assetID, streamID, "invalid_ladder_profile")
	}
	rung, ok := liveRung(rungs)
	if !ok {
		return w.failLive(ctx, a, assetID, streamID, "invalid_ladder_profile")
	}

	dir := filepath.Join(w.WorkDir, "live-"+sessionID)
	defer os.RemoveAll(dir)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}

	// The ingest server holds the socket, so there is nothing to pull until a
	// publisher actually arrives. ffmpeg exits immediately against an empty path, so
	// the wait is a retry loop rather than a listening socket.
	pull := pullURL(w.PullBase, streamID)
	prefix := Prefix(tenantID, assetID)
	published := map[string]bool{}
	waitUntil := time.Now().Add(WaitForEncoder)

	for {
		ended, err := w.runIngest(ctx, a, pull, rung, dir, prefix, published, assetID, streamID)
		if err != nil {
			return w.failLive(ctx, a, assetID, streamID, "ingest_failed")
		}
		if ended {
			return w.endLive(ctx, a, assetID, streamID)
		}
		// Stopped before anything was sent. There is no recording to keep, so this is
		// an ending rather than a broadcast: a stable code, and the segments -- there
		// are none -- are swept like any other failure.
		if w.stopRequested(ctx, a) {
			return w.failLive(ctx, a, assetID, streamID, "stopped")
		}
		// Nothing published yet: nobody has connected. Keep waiting until the stream
		// is abandoned rather than failing a broadcast that starts five minutes late.
		if time.Now().After(waitUntil) {
			return w.failLive(ctx, a, assetID, streamID, "no_encoder")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryInterval):
		}
	}
}

// session is the broadcast being run, carried so the helpers read as they did when
// they took the job arguments.
type session struct {
	ID       string
	TenantID string
}

// runIngest reads one connection to completion. It reports ended=true when video was
// seen, which is a broadcast that finished; ended=false means nobody was publishing
// and the caller should wait and try again.
func (w *Worker) runIngest(ctx context.Context, a session, pull string, rung media.Rung,
	dir, prefix string, published map[string]bool, assetID, streamID string) (bool, error) {
	// Its own context so a stop can end ffmpeg without cancelling the publish of the
	// segments it already wrote -- those go out on the parent, after it exits.
	runCtx, stopFFmpeg := context.WithCancel(ctx)
	defer stopFFmpeg()

	cmd := segmentCommand(runCtx, pull, rung, dir)
	if err := cmd.Start(); err != nil {
		return false, err
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			// Publish whatever the last tick missed before declaring the broadcast
			// over, or the final segments are lost even though they were encoded.
			_, _ = w.publishSegments(ctx, dir, prefix, published)
			// And the playlist unconditionally: the ENDLIST ffmpeg appends on its way
			// out is how a viewer learns the broadcast is over, and it can arrive with
			// no new segment beside it, which publishSegments would skip.
			if body, err := os.ReadFile(filepath.Join(dir, PlaylistName)); err == nil && len(published) > 0 {
				_ = w.Store.Put(ctx, prefix+"/"+PlaylistName,
					bytes.NewReader(body), "application/vnd.apple.mpegurl")
			}
			// An encoder that hangs up is how a broadcast normally ends, so an exit
			// after we saw video is success. An exit with nothing published is simply
			// nobody there yet.
			return len(published) > 0, nil
		case <-ticker.C:
			added, err := w.publishSegments(ctx, dir, prefix, published)
			if err == nil && added > 0 {
				w.markSeen(ctx, a, assetID, streamID, len(published) == added)
			}
			// Asked to stop. Ending ffmpeg here rather than returning straight away
			// is what makes a stopped broadcast identical to one whose encoder hung
			// up: the exit runs the same tail publish, and the recording converts.
			if w.stopRequested(ctx, a) {
				stopFFmpeg()
			}
		}
	}
}

// publishSegments uploads every file the playlist already names, then the playlist.
func (w *Worker) publishSegments(ctx context.Context, dir, prefix string,
	published map[string]bool) (int, error) {
	body, err := os.ReadFile(filepath.Join(dir, PlaylistName))
	if err != nil {
		return 0, err
	}

	added := 0
	for _, name := range PlaylistFiles(body) {
		if published[name] {
			continue
		}
		f, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			continue // named in the playlist but not on disk yet; next tick
		}
		err = w.Store.Put(ctx, prefix+"/"+name, f, "video/mp4")
		f.Close()
		if err != nil {
			return added, err
		}
		published[name] = true
		added++
	}
	if added == 0 {
		return 0, nil
	}
	// Playlist last, always: a playlist naming a segment that is not yet stored is a
	// 404 in the middle of a broadcast.
	return added, w.Store.Put(ctx, prefix+"/"+PlaylistName,
		bytes.NewReader(body), "application/vnd.apple.mpegurl")
}

// PlaylistFiles lists the media files a playlist references: the init segment named
// by EXT-X-MAP and every segment line.
//
// Driven by the playlist rather than by a directory listing because ffmpeg writes a
// segment first and names it only once it is closed. Publishing whatever is in the
// directory would push a half-written segment and stall the player on it.
func PlaylistFiles(body []byte) []string {
	var out []string
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if name, ok := strings.CutPrefix(line, `#EXT-X-MAP:URI="`); ok {
			line, _, _ = strings.Cut(name, `"`)
		} else if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	return out
}

// stopRequested reports whether the customer has asked for this broadcast to end. A
// read failure answers no: dropping a live class because one query timed out is worse
// than noticing the stop a second later.
func (w *Worker) stopRequested(ctx context.Context, a session) bool {
	var asked bool
	err := w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`select stop_requested_at is not null from live_sessions where id = $1`,
			a.ID).Scan(&asked)
	})
	return err == nil && asked
}

// markSeen records the health signal, and on the first segment flips the stream live.
func (w *Worker) markSeen(ctx context.Context, a session, assetID, streamID string, first bool) {
	_ = w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`update live_sessions set last_seen_at = now(),
			        state = 'live', started_at = coalesce(started_at, now())
			  where id = $1`, a.ID)
		return err
	})
	if !first {
		return
	}
	_ = w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`update live_streams set state = 'live', updated_at = now() where id = $1`,
			streamID)
		return err
	})
	_ = w.Assets.MarkLive(ctx, a.TenantID, assetID)
	w.emitLive(ctx, a.TenantID, "live.started", assetID, streamID, "")
}

// endLive closes the broadcast and hands the recording to the VOD path.
//
// The asset goes to live_ended, not ready: the segments are still the only copy of
// the recording and the storage prefix switches on that state, so playback keeps
// reading live/ until the cmaf/ objects exist. The reaper re-issues the conversion if
// this enqueue is lost, so a dropped job is a delay and not a lost recording.
func (w *Worker) endLive(ctx context.Context, a session, assetID, streamID string) error {
	// A fresh context: the job context may already be past its deadline, which is
	// exactly when recording the outcome matters most.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
	defer cancel()

	if err := w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			`update live_sessions set state = 'ended', ended_at = now() where id = $1`,
			a.ID); err != nil {
			return err
		}
		// The port is released here and nowhere else, so a finished broadcast never
		// holds one.
		_, err := tx.Exec(ctx,
			`update live_streams set state = 'ended', ingest_port = null,
			        updated_at = now() where id = $1`, streamID)
		return err
	}); err != nil {
		return err
	}
	if err := w.Assets.MarkEnded(ctx, a.TenantID, assetID); err != nil {
		return err
	}
	w.emitLive(ctx, a.TenantID, "live.ended", assetID, streamID, "")
	return w.Queue.ConvertRecording(ctx, a.TenantID, assetID)
}

// failLive records a stable code and stops. No ffmpeg text crosses this boundary.
func (w *Worker) failLive(ctx context.Context, a session, assetID, streamID, code string) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
	defer cancel()
	_ = w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			`update live_sessions set state = 'failed', error_code = $2, ended_at = now()
			  where id = $1`, a.ID, code); err != nil {
			return err
		}
		_, err := tx.Exec(ctx,
			`update live_streams set state = 'ended', ingest_port = null,
			        updated_at = now() where id = $1`, streamID)
		return err
	})
	_ = w.Assets.MarkFailed(ctx, a.TenantID, assetID, code)
	w.emitLive(ctx, a.TenantID, "live.failed", assetID, streamID, code)
	// Segments are left for the reaper rather than deleted here: the failure path is
	// exactly where a second storage call is most likely to fail too.
	return fmt.Errorf("%w: live session %s: %s", ErrBroadcastFailed, a.ID, code)
}

// emitLive carries the stable code on a failure, because "it never went on air" and
// "it stopped mid-match" need different things from the customer.
func (w *Worker) emitLive(ctx context.Context, tenantID, event, assetID, streamID, code string) {
	if w.Events == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	data := map[string]any{"asset_id": assetID, "stream_id": streamID}
	if code != "" {
		data["error_code"] = code
	}
	_ = w.Events.Emit(ctx, tenantID, event, data)
}

// liveRung picks the lowest eager rung in the tenant's profile.
//
// One rung, and the cheapest one, because realtime encoding is roughly a core per
// rung and the whole BD phase-1 encode budget is about three cores: a full live
// ladder on CPU consumes it. The ladder comes back with the GPU pool.
func liveRung(rungs []media.Rung) (media.Rung, bool) {
	var best media.Rung
	found := false
	for _, r := range media.Eager(rungs) {
		if !found || r.Height < best.Height {
			best, found = r, true
		}
	}
	return best, found
}
