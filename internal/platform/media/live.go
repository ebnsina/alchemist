package media

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
)

// Live ingest and segmentation.
//
// This is a second entry point beside Package() rather than a call into it, because
// Package() shells out once over complete files and returns when the process exits.
// A live packager never exits, cannot read a file that is still being written, and
// cannot write byte ranges into an object store that has no append. Calling it once
// per segment would restart the timeline every two seconds and produce output no
// player can join up. See docs/06-live.md.
//
// shaka-packager stays the VOD packager. Live segments with ffmpeg because neither
// binary can emit LL-HLS partial segments today, so there is nothing shaka would buy
// here that ffmpeg does not already do in one process.

// LiveSegmentSeconds is the live segment length. It is GOPSeconds so every segment
// opens on a keyframe, which is what lets the recording concatenate losslessly into
// a mezzanine afterwards instead of being re-encoded.
const LiveSegmentSeconds = GOPSeconds

// LivePlaylist is the name of the playlist written into the output directory. It is
// master.m3u8 because a single-rendition media playlist is a valid top-level
// playlist, and because the asset's existing playback URL already points there.
const LivePlaylist = "master.m3u8"

// LiveCommand builds the one process that terminates ingest, encodes and segments.
//
// Rate control matches the VOD path (CRF with a VBV cap, never per-chunk ABR) so the
// recording looks like everything else in the library; only the preset drops to
// veryfast, because realtime is a hard constraint and a soft frame beats a late one.
func LiveCommand(ctx context.Context, protocol string, port int, r Rung, outDir string) *exec.Cmd {
	keyint := strconv.Itoa(LiveSegmentSeconds * MezzanineFrameRate)

	args := []string{"-hide_banner", "-loglevel", "error"}
	// RTMP listening is an input flag; SRT carries it in the URL.
	if protocol == "rtmp" {
		args = append(args, "-listen", "1")
	}
	args = append(args,
		"-i", LiveIngestURL(protocol, port),
		"-c:v", "libx264",
		"-profile:v", r.Profile,
		"-preset", "veryfast",
		"-crf", strconv.Itoa(r.CRF),
		"-maxrate", strconv.Itoa(r.MaxrateBPS),
		"-bufsize", strconv.Itoa(r.MaxrateBPS*2),
		"-vf", fmt.Sprintf("scale=-2:%d", r.Height),
		"-pix_fmt", "yuv420p",
		"-g", keyint, "-keyint_min", keyint, "-sc_threshold", "0",
		"-c:a", "aac", "-b:a", "96k", "-ac", "2", "-ar", "48000",
		"-f", "hls",
		"-hls_time", strconv.Itoa(LiveSegmentSeconds),
		"-hls_segment_type", "fmp4",
		// EVENT, not LIVE: the playlist only grows, so a viewer can seek back to the
		// start of the broadcast without a separate DVR mechanism.
		"-hls_playlist_type", "event",
		"-hls_list_size", "0",
		"-hls_fmp4_init_filename", "init.mp4",
		"-hls_segment_filename", filepath.Join(outDir, "%d.m4s"),
		filepath.Join(outDir, LivePlaylist),
	)
	return exec.CommandContext(ctx, "ffmpeg", args...)
}

// LiveIngestURL is the address the encoder publishes to, in ffmpeg's listener form.
//
// One port per stream: ffmpeg accepts a single connection and cannot dispatch on
// SRT's streamid, so streams cannot share a port.
//
// ponytail: port-per-stream, ceiling is the size of the configured range. The upgrade
// is one SRT listener that reads streamid at handshake and hands the socket on -- the
// stream key already travels there, so nothing about auth changes.
func LiveIngestURL(protocol string, port int) string {
	if protocol == "srt" {
		return fmt.Sprintf("srt://0.0.0.0:%d?mode=listener", port)
	}
	return fmt.Sprintf("rtmp://0.0.0.0:%d/live", port)
}

// LivePublishURL is what the customer points OBS at.
func LivePublishURL(protocol, host string, port int, streamKey string) string {
	if protocol == "srt" {
		return fmt.Sprintf("srt://%s:%d?mode=caller&streamid=%s", host, port, streamKey)
	}
	return fmt.Sprintf("rtmp://%s:%d/live/%s", host, port, streamKey)
}
