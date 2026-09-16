package pipeline

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ebnsina/alchemist/internal/modules/live"
	"github.com/ebnsina/alchemist/internal/platform/media"
)

// Live to VOD, with no re-ingest.
//
// Every live segment came out of an encoder on the GOP grid CLAUDE.md requires of
// every rendition, so an init segment followed by its media segments in playlist order
// is already one valid fragmented MP4. Appending those bytes is lossless and costs no
// decode -- from there it is the ordinary transcode path, which is why a recording ends
// up with the same ladder, the same byte-range CMAF and the same content key as any
// upload.
//
// One file per encoder connection, joined with the concat demuxer. A reconnect restarts
// the encoder's timestamps at zero and ffmpeg marks it with EXT-X-DISCONTINUITY;
// appending across that boundary makes a file whose timeline goes backwards, and
// everything downstream quietly believes the shorter duration. A ten-second test
// broadcast with one reconnect came back as a six-second mezzanine, with no error
// anywhere. The concat demuxer is built for this and recovers the whole thing.
//
// The playlist decides the order, never a directory listing: ffmpeg names segments
// 1.m4s, 2.m4s, 10.m4s, and sorting those as strings puts 10 before 2 and produces a
// recording that plays out of order with no error anywhere.
func (w *TranscodeWorker) pullRecording(ctx context.Context, a TranscodeArgs, dst string) error {
	prefix := live.Prefix(a.TenantID, a.AssetID)

	body, err := w.Store.Get(ctx, prefix+"/"+live.PlaylistName)
	if err != nil {
		return fmt.Errorf("read recording playlist: %w", err)
	}
	playlist, err := io.ReadAll(body)
	body.Close()
	if err != nil {
		return err
	}

	runs := live.PlaylistRuns(playlist)
	if len(runs) == 0 {
		return fmt.Errorf("recording %s has no segments", a.AssetID)
	}

	dir := filepath.Dir(dst)
	parts := make([]string, 0, len(runs))
	for i, names := range runs {
		part := filepath.Join(dir, fmt.Sprintf("recording-%03d.mp4", i))
		if err := w.pullRun(ctx, prefix, names, part); err != nil {
			return err
		}
		parts = append(parts, part)
	}

	// The common case is one connection, and concatenating a single file through
	// ffmpeg only to copy it is work for nothing.
	if len(parts) == 1 {
		return os.Rename(parts[0], dst)
	}
	return media.Stitch(ctx, parts, dst)
}

// pullRun assembles one encoder connection: its init segment, then its media segments
// in playlist order.
func (w *TranscodeWorker) pullRun(ctx context.Context, prefix string, names []string, dst string) error {
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, name := range names {
		seg, err := w.Store.Get(ctx, prefix+"/"+name)
		if err != nil {
			return fmt.Errorf("read recording segment %s: %w", name, err)
		}
		_, err = io.Copy(f, seg)
		seg.Close()
		if err != nil {
			return err
		}
	}
	return f.Sync()
}
