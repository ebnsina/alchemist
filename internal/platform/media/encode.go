package media

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os/exec"
	"strconv"
)

// EncoderVersion changes whenever encoding behaviour changes. It is part of every
// output key, so a bump re-encodes rather than silently mixing old and new output.
const EncoderVersion = "x264-v1"

// EncodeChunk encodes one chunk at one rung.
//
// Rate control is CRF with a VBV cap, never per-chunk two-pass ABR. Two-pass would
// make each chunk independently hit a bitrate target, so an easy chunk is encoded
// needlessly well and a hard one needlessly badly, and quality visibly oscillates on
// a chunk-length cycle. CRF holds quality constant across boundaries by construction
// while maxrate/bufsize keep any chunk from blowing the client's bandwidth budget.
func EncodeChunk(ctx context.Context, mezzanine string, c Chunk, r Rung, dst string) error {
	keyint := strconv.Itoa(GOPSeconds * MezzanineFrameRate)
	bufsize := r.MaxrateBPS * 2

	args := []string{
		"-y", "-hide_banner", "-loglevel", "error",
		// Seek before -i: the mezzanine's keyframe grid makes this exact and cheap.
		"-ss", fmt.Sprintf("%.3f", c.StartSec),
		"-i", mezzanine,
		"-t", fmt.Sprintf("%.3f", c.EndSec-c.StartSec),
		"-map", "0:v:0",
		"-c:v", "libx264",
		"-profile:v", r.Profile,
		"-preset", r.Preset,
		"-crf", strconv.Itoa(r.CRF),
		"-maxrate", strconv.Itoa(r.MaxrateBPS),
		"-bufsize", strconv.Itoa(bufsize),
		"-vf", fmt.Sprintf("scale=-2:%d", r.Height),
		"-pix_fmt", "yuv420p",
		// Identical GOP structure on every rung: the precondition for clean ABR
		// switching and for concat without re-encoding.
		"-g", keyint, "-keyint_min", keyint, "-sc_threshold", "0",
		"-force_key_frames", fmt.Sprintf("expr:gte(t,n_forced*%d)", GOPSeconds),
		"-avoid_negative_ts", "make_zero",
		"-an", // audio is encoded once, separately, never per rendition
		dst,
	}

	if out, err := exec.CommandContext(ctx, "ffmpeg", args...).CombinedOutput(); err != nil {
		return fmt.Errorf("%w: chunk %d rung %dp: %s",
			ErrEncodeFailed, c.Index, r.Height, truncate(string(out)))
	}
	return nil
}

// EncodeAudio produces the single shared audio rendition. Audio is cheap enough to
// encode in one pass, and storing it once instead of muxing it into every video rung
// saves the whole ladder's worth of duplicate audio bytes.
func EncodeAudio(ctx context.Context, mezzanine, dst string) error {
	out, err := exec.CommandContext(ctx, "ffmpeg",
		"-y", "-hide_banner", "-loglevel", "error",
		"-i", mezzanine,
		"-map", "0:a:0",
		"-c:a", "aac", "-b:a", "96k", "-ac", "2", "-ar", "48000",
		"-vn", dst).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: audio: %s", ErrEncodeFailed, truncate(string(out)))
	}
	return nil
}

// ParamsHash identifies an exact encode. Output objects are keyed by it, so a retry
// overwrites deterministically and a parameter change never collides with old output.
func ParamsHash(r Rung) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s|%d|%s|%s|%d|%d|%d|%d",
		EncoderVersion, r.Height, r.Codec, r.Profile, r.CRF, r.MaxrateBPS,
		GOPSeconds, MezzanineFrameRate)))
	return hex.EncodeToString(h[:8])
}
