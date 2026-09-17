package media

import (
	"context"
	"fmt"
	"math"
	"os/exec"
	"strconv"
)

// GOPSeconds is the keyframe interval, shared by the mezzanine and every rendition.
// Identical GOP boundaries across the ladder are what make ABR switching clean.
const GOPSeconds = 2

// DefaultFrameRate is what a mezzanine is built at when the source rate is unusable,
// and what live encodes at, where there is no source to read ahead of.
const DefaultFrameRate = 30

// mezzanineRates are the rates a mezzanine may be built at. Integers only, so a GOP
// is a whole number of frames and every chunk boundary lands on one.
var mezzanineRates = []int{24, 25, 30}

// MezzanineRate picks the rate to normalise a source to.
//
// Forcing everything to 30 was wrong in both directions: 24fps film gained six
// invented frames a second, which reads as judder on a pan, and 60fps lost half its
// frames while still being billed and reported as if it had them. The cinema and PAL
// rates are kept as they are, the high rates halve back onto a rate in the set, and
// anything below the set is left alone rather than being inflated -- a 15fps screen
// recording doubled to 30 costs bitrate for frames that carry nothing.
//
// High frame rate is not offered. It is bitrate the budget Android fleet this is built
// for cannot spend, and no tenant has asked; the halving is where that opt-in would go.
func MezzanineRate(srcFPS float64) int {
	if srcFPS <= 0 || srcFPS > 240 {
		return DefaultFrameRate
	}

	rate := int(math.Round(srcFPS))
	// 48, 50, 60, 120 land back on 24, 25, 30, 60 -- and 60 halves again.
	for rate > 30 {
		rate /= 2
	}
	if rate < 24 {
		return max(rate, 1) // already cheap; inflating it would only cost bitrate
	}

	nearest := DefaultFrameRate
	best := math.MaxFloat64
	for _, r := range mezzanineRates {
		if d := math.Abs(float64(rate - r)); d < best {
			nearest, best = r, d
		}
	}
	return nearest
}

// BuildMezzanine normalizes an arbitrary source into a predictable intermediate:
// constant frame rate, closed GOPs on a fixed grid, rotation baked in, timestamps
// regenerated. Chunk workers seek into this, never into the customer's file.
//
// fps is chosen by MezzanineRate from the source, not fixed: everything downstream --
// the GOP grid, the chunk plan, the params hash -- is expressed in whole frames, so
// the rate has to be decided once, here, and carried.
func BuildMezzanine(ctx context.Context, src, dst string, fps int) error {
	if fps <= 0 {
		fps = DefaultFrameRate
	}
	keyint := strconv.Itoa(GOPSeconds * fps)

	args := []string{
		"-y", "-hide_banner", "-loglevel", "error",
		"-fflags", "+genpts", "-i", src,
		"-map", "0:v:0",
		"-c:v", "libx264", "-preset", "veryfast", "-crf", "17",
		"-pix_fmt", "yuv420p",
		"-r", strconv.Itoa(fps),
		"-g", keyint, "-keyint_min", keyint,
		"-sc_threshold", "0", // keyframes on the grid only, never scene-driven
		"-force_key_frames", fmt.Sprintf("expr:gte(t,n_forced*%d)", GOPSeconds),
		"-vf", "scale=trunc(iw/2)*2:trunc(ih/2)*2",
		"-movflags", "+faststart",
	}
	// Audio is optional; a silent source stays silent rather than failing.
	args = append(args, "-map", "0:a:0?", "-c:a", "aac", "-b:a", "128k", "-ac", "2", "-ar", "48000")
	args = append(args, dst)

	if out, err := exec.CommandContext(ctx, "ffmpeg", args...).CombinedOutput(); err != nil {
		return fmt.Errorf("%w: mezzanine: %s", ErrEncodeFailed, truncate(string(out)))
	}
	return nil
}

func truncate(s string) string {
	const max = 400
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}

// Conformant reports whether a source is already everything a mezzanine has to be, so
// it can be remuxed rather than re-encoded.
//
// Checked rather than trusted. The caller knows where the file came from, but "live
// encodes on the same grid" is a fact two files away that could drift, and the cost of
// being wrong is a chunk plan that does not land on keyframes -- which surfaces as
// stitch failures, not as an error here.
//
// The GOP itself is not probed: reading every keyframe position means walking the whole
// index, which is most of what remuxing was meant to save. Live builds the grid from
// media.GOPSeconds by construction, and VerifyStitch catches it downstream if it ever
// stops being true.
func Conformant(p *Probe, fps int) bool {
	// Near enough is not enough: rounding 29.97 to 30 and copying leaves the chunk
	// plan assuming 30 frames a second where there are 29.97, so every boundary drifts
	// a little further off the keyframes. Live writes exactly -r fps, so a genuine
	// recording sits on the integer and a tolerance this tight costs it nothing.
	return p != nil && p.VideoCodec == "h264" &&
		math.Abs(p.FrameRate-float64(fps)) < 0.01 &&
		p.Width > 0 && p.Height > 0
}

// RemuxMezzanine rewrites the container and copies the streams.
//
// A broadcast's recording is already H.264 on the two-second grid at a constant rate,
// because live encodes it that way. Re-encoding it to CRF 17 produced a mezzanine
// roughly twice the size of the file it came from, with a generation of loss, and then
// built the whole ladder from that degraded copy. Measured: 0.04s against 0.61s, and
// 1.5 MB against 2.9 MB, on a clip where the input was 1.5 MB.
func RemuxMezzanine(ctx context.Context, src, dst string) error {
	args := []string{
		"-y", "-hide_banner", "-loglevel", "error",
		// The concatenated segments carry the timestamps of several encoder
		// connections, so the container is rebuilt with a clean timeline.
		"-fflags", "+genpts", "-i", src,
		"-map", "0:v:0", "-map", "0:a:0?",
		"-c", "copy",
		"-movflags", "+faststart",
		dst,
	}
	if out, err := exec.CommandContext(ctx, "ffmpeg", args...).CombinedOutput(); err != nil {
		return fmt.Errorf("%w: remux mezzanine: %s", ErrEncodeFailed, truncate(string(out)))
	}
	return nil
}
