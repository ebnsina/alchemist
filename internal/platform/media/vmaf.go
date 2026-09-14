package media

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
)

type VMAFReport struct {
	Mean       float64   `json:"mean"`
	Min        float64   `json:"min"`
	PerChunk   []float64 `json:"per_chunk"`
	MaxAdjDiff float64   `json:"max_adjacent_chunk_diff"`
}

type vmafLog struct {
	Frames []struct {
		FrameNum int `json:"frameNum"`
		Metrics  struct {
			VMAF float64 `json:"vmaf"`
		} `json:"metrics"`
	} `json:"frames"`
}

// ScoreVMAF compares a rendition against the mezzanine it came from.
//
// The per-chunk breakdown is the point: chunked encoding's characteristic failure is
// quality oscillating on a chunk-length cycle, which is glaring on a TV and invisible
// on a laptop. A large MaxAdjDiff means rate control is drifting between chunks.
func ScoreVMAF(ctx context.Context, distorted, reference string, workDir string) (*VMAFReport, error) {
	ref, err := Inspect(ctx, reference)
	if err != nil {
		return nil, err
	}

	logPath := filepath.Join(workDir, "vmaf.json")
	// libvmaf takes [distorted][reference]; the distorted stream is upscaled to the
	// reference resolution so rungs below source height can be scored at all.
	filter := fmt.Sprintf(
		"[0:v]scale=%d:%d:flags=bicubic,setpts=PTS-STARTPTS[dist];"+
			"[1:v]setpts=PTS-STARTPTS[ref];[dist][ref]libvmaf=log_fmt=json:log_path=%s",
		ref.Width, ref.Height, logPath)

	out, err := exec.CommandContext(ctx, "ffmpeg",
		"-hide_banner", "-loglevel", "error",
		"-i", distorted, "-i", reference,
		"-lavfi", filter, "-f", "null", "-").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("vmaf: %s", truncate(string(out)))
	}

	raw, err := os.ReadFile(logPath)
	if err != nil {
		return nil, fmt.Errorf("vmaf: read log: %w", err)
	}
	var parsed vmafLog
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("vmaf: parse log: %w", err)
	}
	if len(parsed.Frames) == 0 {
		return nil, fmt.Errorf("vmaf: no frames scored")
	}

	framesPerChunk := TargetChunkSeconds * MezzanineFrameRate
	rep := &VMAFReport{Min: math.MaxFloat64}
	var total float64
	var chunkSum float64
	var chunkN int

	for i, f := range parsed.Frames {
		v := f.Metrics.VMAF
		total += v
		rep.Min = math.Min(rep.Min, v)
		chunkSum += v
		chunkN++
		if (i+1)%framesPerChunk == 0 || i == len(parsed.Frames)-1 {
			rep.PerChunk = append(rep.PerChunk, chunkSum/float64(chunkN))
			chunkSum, chunkN = 0, 0
		}
	}
	rep.Mean = total / float64(len(parsed.Frames))

	for i := 1; i < len(rep.PerChunk); i++ {
		rep.MaxAdjDiff = math.Max(rep.MaxAdjDiff, math.Abs(rep.PerChunk[i]-rep.PerChunk[i-1]))
	}
	return rep, nil
}
