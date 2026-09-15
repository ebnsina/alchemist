package pipeline

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"

	"github.com/ebnsina/alchemist/internal/platform/db"
	"github.com/ebnsina/alchemist/internal/platform/media"
	"github.com/ebnsina/alchemist/internal/platform/storage"
)

// QueueLive is sized separately because a live job holds its slot for the whole
// broadcast. Sharing encode_cpu would let two football matches drain every worker and
// leave the VOD queue untouched for three hours.
const QueueLive = "live"

// LiveTimeout bounds a single broadcast, including the wait for the encoder to
// connect at all. An armed stream nobody publishes to releases its port here.
const LiveTimeout = 6 * time.Hour

// livePollInterval is how often the output directory is swept for new segments. Half
// a segment, so a segment is never more than one interval behind the viewer.
const livePollInterval = time.Second

// LiveWaitForEncoder is how long an armed stream waits for somebody to publish before
// giving up. Generous, because a class that starts late is normal and the cost of
// waiting is one idle worker slot.
const LiveWaitForEncoder = 30 * time.Minute

// liveRetryInterval is how often the transcoder re-checks whether a publisher has
// arrived. Two seconds is under one segment, so nothing is missed at the start.
const liveRetryInterval = 2 * time.Second

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
	DB    *db.DB
	Store *storage.Store
	River *river.Client[pgx.Tx]
	// PullBase is the ingest server's private address, e.g. rtsp://127.0.0.1:8554.
	PullBase string
	WorkDir  string
}

func (w *LiveWorker) Timeout(*river.Job[LiveArgs]) time.Duration { return LiveTimeout }

// Work runs one broadcast: bind the port, wait for the encoder, encode and segment,
// and publish each segment as it lands.
//
// ponytail: one rung, no ABR, and the recording is left as live segments rather than
// converted to a VOD asset. The ceiling is that a viewer gets a single bitrate and
// the segments stay small objects until swept. Both upgrades are additive -- N rungs
// is N outputs in the same ffmpeg command on the same GOP grid, and the recording
// concatenates losslessly into a mezzanine that the existing Transcode/Package path
// already knows how to handle. See docs/06-live.md.
func (w *LiveWorker) Work(ctx context.Context, job *river.Job[LiveArgs]) error {
	a := job.Args

	var assetID, streamID, protocol string
	var ladderRaw []byte
	err := w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`select s.asset_id::text, s.stream_id::text, l.protocol, p.rungs
			   from live_sessions s
			   join live_streams l on l.id = s.stream_id
			   join assets a on a.id = s.asset_id
			   join ladder_profiles p on p.name = a.ladder_profile
			  where s.id = $1`, a.SessionID).
			Scan(&assetID, &streamID, &protocol, &ladderRaw)
	})
	if err != nil {
		return fmt.Errorf("load live session %s: %w", a.SessionID, err)
	}

	rungs, err := media.ParseLadder(ladderRaw)
	if err != nil {
		return w.failLive(ctx, a, assetID, streamID, "invalid_ladder_profile")
	}
	rung, ok := liveRung(rungs)
	if !ok {
		return w.failLive(ctx, a, assetID, streamID, "invalid_ladder_profile")
	}

	dir := filepath.Join(w.WorkDir, "live-"+a.SessionID)
	defer os.RemoveAll(dir)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}

	// The ingest server holds the socket, so there is nothing to pull until a
	// publisher actually arrives. ffmpeg exits immediately against an empty path, so
	// the wait is a retry loop rather than a listening socket.
	pull := media.LivePullURL(w.PullBase, streamID)
	prefix := fmt.Sprintf("live/%s/%s", a.TenantID, assetID)
	published := map[string]bool{}
	waitUntil := time.Now().Add(LiveWaitForEncoder)

	for {
		ended, err := w.runIngest(ctx, a, pull, rung, dir, prefix, published, assetID, streamID)
		if err != nil {
			return w.failLive(ctx, a, assetID, streamID, "ingest_failed")
		}
		if ended {
			return w.endLive(ctx, a, assetID, streamID)
		}
		// Nothing published yet: nobody has connected. Keep waiting until the stream
		// is abandoned rather than failing a broadcast that starts five minutes late.
		if time.Now().After(waitUntil) {
			return w.failLive(ctx, a, assetID, streamID, "no_encoder")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(liveRetryInterval):
		}
	}
}

// runIngest reads one connection to completion. It reports ended=true when video was
// seen, which is a broadcast that finished; ended=false means nobody was publishing
// and the caller should wait and try again.
func (w *LiveWorker) runIngest(ctx context.Context, a LiveArgs, pull string, rung media.Rung,
	dir, prefix string, published map[string]bool, assetID, streamID string) (bool, error) {
	cmd := media.LiveCommand(ctx, pull, rung, dir)
	if err := cmd.Start(); err != nil {
		return false, err
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	ticker := time.NewTicker(livePollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			// Publish whatever the last tick missed before declaring the broadcast
			// over, or the final segments are lost even though they were encoded.
			_, _ = w.publishSegments(ctx, a, dir, prefix, published)
			// An encoder that hangs up is how a broadcast normally ends, so an exit
			// after we saw video is success. An exit with nothing published is simply
			// nobody there yet.
			return len(published) > 0, nil
		case <-ticker.C:
			added, err := w.publishSegments(ctx, a, dir, prefix, published)
			if err != nil || added == 0 {
				continue
			}
			w.markSeen(ctx, a, assetID, streamID, len(published) == added)
		}
	}
}

// publishSegments uploads every file the playlist already names, then the playlist.
func (w *LiveWorker) publishSegments(ctx context.Context, a LiveArgs, dir, prefix string,
	published map[string]bool) (int, error) {
	body, err := os.ReadFile(filepath.Join(dir, media.LivePlaylist))
	if err != nil {
		return 0, err
	}

	added := 0
	for _, name := range playlistFiles(body) {
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
	return added, w.Store.Put(ctx, prefix+"/"+media.LivePlaylist,
		bytes.NewReader(body), "application/vnd.apple.mpegurl")
}

// playlistFiles lists the media files a playlist references: the init segment named
// by EXT-X-MAP and every segment line.
//
// Driven by the playlist rather than by a directory listing because ffmpeg writes a
// segment first and names it only once it is closed. Publishing whatever is in the
// directory would push a half-written segment and stall the player on it.
func playlistFiles(body []byte) []string {
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

// markSeen records the health signal, and on the first segment flips the stream live.
func (w *LiveWorker) markSeen(ctx context.Context, a LiveArgs, assetID, streamID string, first bool) {
	_ = w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`update live_sessions set last_seen_at = now(),
			        state = 'live', started_at = coalesce(started_at, now())
			  where id = $1`, a.SessionID)
		return err
	})
	if !first {
		return
	}
	_ = w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			`update live_streams set state = 'live', updated_at = now() where id = $1`,
			streamID); err != nil {
			return err
		}
		_, err := tx.Exec(ctx,
			`update assets set state = 'live', updated_at = now() where id = $1`, assetID)
		return err
	})
	w.emitLive(ctx, a.TenantID, "live.started", assetID, streamID)
}

// endLive closes the broadcast. The segments stay where they are and keep playing:
// the asset is live_ended, not ready, because the recording has not been converted.
func (w *LiveWorker) endLive(ctx context.Context, a LiveArgs, assetID, streamID string) error {
	// A fresh context: the job context may already be past its deadline, which is
	// exactly when recording the outcome matters most.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
	defer cancel()

	if err := w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			`update live_sessions set state = 'ended', ended_at = now() where id = $1`,
			a.SessionID); err != nil {
			return err
		}
		// The port is released here and nowhere else, so a finished broadcast never
		// holds one.
		if _, err := tx.Exec(ctx,
			`update live_streams set state = 'ended', ingest_port = null,
			        updated_at = now() where id = $1`, streamID); err != nil {
			return err
		}
		_, err := tx.Exec(ctx,
			`update assets set state = 'live_ended', updated_at = now() where id = $1`,
			assetID)
		return err
	}); err != nil {
		return err
	}
	w.emitLive(ctx, a.TenantID, "live.ended", assetID, streamID)
	return nil
}

// failLive records a stable code and stops. No ffmpeg text crosses this boundary.
func (w *LiveWorker) failLive(ctx context.Context, a LiveArgs, assetID, streamID, code string) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
	defer cancel()
	_ = w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			`update live_sessions set state = 'failed', error_code = $2, ended_at = now()
			  where id = $1`, a.SessionID, code); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx,
			`update live_streams set state = 'ended', ingest_port = null,
			        updated_at = now() where id = $1`, streamID); err != nil {
			return err
		}
		_, err := tx.Exec(ctx,
			`update assets set state = 'failed', error_code = $2, updated_at = now()
			  where id = $1`, assetID, code)
		return err
	})
	return river.JobCancel(fmt.Errorf("live session %s: %s", a.SessionID, code))
}

func (w *LiveWorker) emitLive(ctx context.Context, tenantID, event, assetID, streamID string) {
	if w.River == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	_ = Emit(ctx, w.DB, w.River, tenantID, event,
		map[string]any{"asset_id": assetID, "stream_id": streamID})
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
