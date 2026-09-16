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
// chunks are the real boundaries the encode used, which is what MaxAdjDiff has to be
// measured across. A fixed window cannot stand in for them: PlanChunksAtScenes emits
// anything from 6 to 24 seconds, so a 12-second window averages over the boundary
// instead of straddling it and the oscillation this exists to catch is smoothed away.
func ScoreVMAF(ctx context.Context, distorted, reference string, workDir string,
	chunks []Chunk, fps int) (*VMAFReport, error) {
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

	scores := make([]float64, len(parsed.Frames))
	rep := &VMAFReport{Min: math.MaxFloat64}
	var total float64
	for i, f := range parsed.Frames {
		scores[i] = f.Metrics.VMAF
		total += scores[i]
		rep.Min = math.Min(rep.Min, scores[i])
	}
	rep.Mean = total / float64(len(scores))
	rep.PerChunk = perChunkScores(scores, chunks, fps)

	for i := 1; i < len(rep.PerChunk); i++ {
		rep.MaxAdjDiff = math.Max(rep.MaxAdjDiff, math.Abs(rep.PerChunk[i]-rep.PerChunk[i-1]))
	}
	return rep, nil
}

// perChunkScores averages frame scores within each real chunk.
//
// Frames are numbered against the mezzanine's constant rate, so a frame index converts
// to a timestamp exactly. With no chunk plan the whole run is one window,
// which is honest rather than wrong: MaxAdjDiff is then zero because there is no
// boundary to measure across.
func perChunkScores(scores []float64, chunks []Chunk, fps int) []float64 {
	if fps <= 0 {
		fps = DefaultFrameRate
	}
	var out []float64
	var sum float64
	var n int
	at := 0

	for i, v := range scores {
		t := float64(i) / float64(fps)
		for at < len(chunks)-1 && t >= chunks[at].EndSec {
			if n > 0 {
				out = append(out, sum/float64(n))
			}
			sum, n = 0, 0
			at++
		}
		sum += v
		n++
	}
	if n > 0 {
		out = append(out, sum/float64(n))
	}
	return out
}
