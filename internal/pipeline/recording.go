package pipeline

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/ebnsina/alchemist/internal/modules/live"
)

// Live to VOD, with no re-ingest.
//
// Every live segment came out of one encoder process on the GOP grid CLAUDE.md
// requires of every rendition, so the init segment followed by the media segments in
// playlist order is already one valid fragmented MP4. Concatenating the bytes is
// therefore lossless and costs no decode -- from there it is the ordinary transcode
// path, which is why a recording ends up with the same ladder, the same byte-range
// CMAF and the same content key as any upload.
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

	names := live.PlaylistFiles(playlist)
	if len(names) == 0 {
		return fmt.Errorf("recording %s has no segments", a.AssetID)
	}

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
