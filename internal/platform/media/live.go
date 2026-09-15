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
func LiveCommand(ctx context.Context, input string, r Rung, outDir string) *exec.Cmd {
	keyint := strconv.Itoa(LiveSegmentSeconds * MezzanineFrameRate)

	args := []string{"-hide_banner", "-loglevel", "error"}
	// TCP, because a dropped UDP packet on the private hop between the ingest server
	// and here would corrupt a segment for every viewer at once.
	args = append(args, "-rtsp_transport", "tcp")
	args = append(args,
		"-i", input,
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

// LivePullURL is the private address the transcoder reads a published stream from.
// The path is the stream id, which is what the ingest server authorized against the
// stream key -- see internal/api/live_auth.go.
func LivePullURL(base, streamID string) string {
	return base + "/" + streamID
}

// Ingest ports as configured in deploy/live/mediamtx.yml. Change them there and here
// together, or the URL handed to a customer points at a port nothing is listening on.
const (
	LiveRTMPPort = 1935
	LiveSRTPort  = 8890
)

// LiveKeyPlaceholder marks where the customer pastes their own stream key.
const LiveKeyPlaceholder = "YOUR_STREAM_KEY"

// LivePublishURL is what the customer points OBS at.
//
// The stream key travels as the password and the stream id as the path, because the
// ingest server asks the API about exactly that pair before accepting a publisher.
// One port serves every stream: the ingest server dispatches on the path, which is
// the thing ffmpeg's own listener could never do.
//
// The key is a placeholder because only its hash is stored -- it was shown once, at
// creation, and we cannot put it back into a URL later even for its owner.
func LivePublishURL(protocol, host, streamID string) string {
	if protocol == "srt" {
		return fmt.Sprintf("srt://%s:%d?streamid=publish:%s:publisher:%s",
			host, LiveSRTPort, streamID, LiveKeyPlaceholder)
	}
	return fmt.Sprintf("rtmp://%s:%d/%s?user=publisher&pass=%s",
		host, LiveRTMPPort, streamID, LiveKeyPlaceholder)
}
