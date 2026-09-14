package media

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Stitch concatenates encoded chunks with no re-encode. This is exact and fast, and
// it is only valid because every chunk starts on a keyframe and shares encoder
// settings, which PlanChunks and EncodeChunk guarantee.
func Stitch(ctx context.Context, chunkPaths []string, dst string) error {
	if len(chunkPaths) == 0 {
		return fmt.Errorf("%w: no chunks to stitch", ErrStitchFailed)
	}

	listPath := filepath.Join(filepath.Dir(dst), "concat-"+filepath.Base(dst)+".txt")
	var b strings.Builder
	for _, p := range chunkPaths {
		abs, err := filepath.Abs(p)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrStitchFailed, err)
		}
		// Single quotes are escaped per the concat demuxer's own rules.
		fmt.Fprintf(&b, "file '%s'\n", strings.ReplaceAll(abs, "'", `'\''`))
	}
	if err := os.WriteFile(listPath, []byte(b.String()), 0o600); err != nil {
		return fmt.Errorf("%w: %v", ErrStitchFailed, err)
	}
	defer os.Remove(listPath)

	out, err := exec.CommandContext(ctx, "ffmpeg",
		"-y", "-hide_banner", "-loglevel", "error",
		"-f", "concat", "-safe", "0", "-i", listPath,
		"-c", "copy", "-movflags", "+faststart", dst).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrStitchFailed, truncate(string(out)))
	}
	return nil
}

// VerifyStitch catches a silent concat failure, which produces a file that plays for
// a few seconds and then stops. Without this the first report comes from a customer.
func VerifyStitch(ctx context.Context, path string, wantSec float64) error {
	p, err := Inspect(ctx, path)
	if err != nil {
		return fmt.Errorf("%w: stitched output is unreadable", ErrStitchFailed)
	}
	const toleranceSec = 0.5
	if diff := p.DurationSec - wantSec; diff > toleranceSec || diff < -toleranceSec {
		return fmt.Errorf("%w: stitched duration %.2fs, expected %.2fs",
			ErrStitchFailed, p.DurationSec, wantSec)
	}
	return nil
}
