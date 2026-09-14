package media

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
)

// GOPSeconds is the keyframe interval, shared by the mezzanine and every rendition.
// Identical GOP boundaries across the ladder are what make ABR switching clean.
const GOPSeconds = 2

// MezzanineFrameRate is forced so chunk boundaries land on exact frame counts.
const MezzanineFrameRate = 30

// BuildMezzanine normalizes an arbitrary source into a predictable intermediate:
// constant frame rate, closed GOPs on a fixed grid, rotation baked in, timestamps
// regenerated. Chunk workers seek into this, never into the customer's file.
func BuildMezzanine(ctx context.Context, src, dst string) error {
	keyint := strconv.Itoa(GOPSeconds * MezzanineFrameRate)

	args := []string{
		"-y", "-hide_banner", "-loglevel", "error",
		"-fflags", "+genpts", "-i", src,
		"-map", "0:v:0",
		"-c:v", "libx264", "-preset", "veryfast", "-crf", "17",
		"-pix_fmt", "yuv420p",
		"-r", strconv.Itoa(MezzanineFrameRate),
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
