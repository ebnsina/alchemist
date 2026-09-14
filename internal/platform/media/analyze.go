package media

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
)

var ptsTimeRE = regexp.MustCompile(`pts_time:([0-9.]+)`)

// SceneThreshold is ffmpeg's scene score above which a cut is assumed.
//
// Measured, not guessed: hard cuts between flat-coloured frames score as low as
// 0.04, so the commonly cited 0.3 silently misses real cuts. 0.1 errs toward
// over-detection, which is the safe direction — see SceneChanges.
const SceneThreshold = 0.1

// SceneChanges returns timestamps where the picture changes substantially.
//
// Cutting chunks on a scene change is free: the encoder was going to spend a
// keyframe there anyway, so the forced I-frame at a chunk boundary costs no extra
// bitrate. Cutting mid-scene pays for a keyframe that earns nothing.
//
// This is an optimisation with a safe fallback, never a correctness requirement.
// Every boundary is snapped to the GOP grid regardless, and PlanChunksAtScenes falls
// back to fixed spacing when no usable cut is found. A missed cut costs a little
// efficiency; it cannot produce broken output. Thresholds want tuning against real
// customer content, which synthetic test material cannot stand in for.
func SceneChanges(ctx context.Context, mezzanine string, threshold float64) ([]float64, error) {
	out, err := exec.CommandContext(ctx, "ffmpeg",
		"-hide_banner", "-loglevel", "info", "-i", mezzanine,
		"-filter:v", fmt.Sprintf("select='gt(scene,%.3f)',metadata=print:file=-", threshold),
		"-an", "-sn", "-f", "null", "-").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: scene detection: %s", ErrEncodeFailed, truncate(string(out)))
	}

	var times []float64
	for _, m := range ptsTimeRE.FindAllStringSubmatch(string(out), -1) {
		if t, err := strconv.ParseFloat(m[1], 64); err == nil {
			times = append(times, t)
		}
	}
	return times, nil
}

// MinChunkSeconds and MaxChunkSeconds bound the search for a scene cut. Too short
// wastes bitrate on keyframes and encoder ramp-up; too long starves parallelism and
// lengthens the tail, and one slow chunk holds up the whole stitch.
const (
	MinChunkSeconds = 6
	MaxChunkSeconds = 24
)

// PlanChunksAtScenes snaps chunk boundaries to scene changes where one falls in the
// acceptable window, and to the keyframe grid otherwise. Every boundary lands on the
// GOP grid either way, which is what keeps stitching a copy operation.
func PlanChunksAtScenes(durationSec float64, scenes []float64) []Chunk {
	if durationSec <= 0 {
		return nil
	}

	var chunks []Chunk
	start := 0.0
	for start < durationSec-0.001 {
		ideal := start + TargetChunkSeconds
		if ideal >= durationSec {
			chunks = append(chunks, Chunk{Index: len(chunks), StartSec: start, EndSec: durationSec})
			break
		}

		best := snapToGrid(ideal)
		bestDist := math.MaxFloat64
		for _, s := range scenes {
			if s <= start+MinChunkSeconds || s > start+MaxChunkSeconds {
				continue
			}
			g := snapToGrid(s)
			if g <= start || g >= durationSec {
				continue
			}
			if d := math.Abs(s - ideal); d < bestDist {
				best, bestDist = g, d
			}
		}
		if best <= start {
			best = snapToGrid(start + TargetChunkSeconds)
		}
		if best >= durationSec {
			best = durationSec
		}
		chunks = append(chunks, Chunk{Index: len(chunks), StartSec: start, EndSec: best})
		start = best
	}
	return chunks
}

func snapToGrid(t float64) float64 {
	return math.Round(t/GOPSeconds) * GOPSeconds
}

// referenceBitrate360 is what averagely-complex content costs at 360p CRF 23. The
// measured sample is compared against it to get a complexity ratio.
const referenceBitrate360 = 700_000

// Complexity measures how expensive this particular content is to encode, by
// sampling a few seconds from across the timeline at fixed quality and seeing what
// bitrate falls out. Cost is constant regardless of source length.
//
// This is what makes a lecture cheaper than a football match. For BD edtech, screen
// recordings routinely land 40-60% under the fixed ladder at identical quality, and
// that saving is the viewer's mobile data bill.
func Complexity(ctx context.Context, mezzanine string, durationSec float64, workDir string) (float64, error) {
	const (
		samples    = 3
		sampleSecs = 4
	)
	var totalBytes int64
	var totalSecs float64

	for i := range samples {
		at := durationSec * float64(i+1) / float64(samples+1)
		if at+sampleSecs > durationSec {
			at = math.Max(0, durationSec-sampleSecs)
		}
		dst := filepath.Join(workDir, fmt.Sprintf("sample-%d.mp4", i))

		out, err := exec.CommandContext(ctx, "ffmpeg",
			"-y", "-hide_banner", "-loglevel", "error",
			"-ss", fmt.Sprintf("%.3f", at), "-i", mezzanine,
			"-t", strconv.Itoa(sampleSecs),
			"-map", "0:v:0", "-an",
			"-c:v", "libx264", "-preset", "veryfast", "-crf", "23",
			"-vf", "scale=-2:360", "-pix_fmt", "yuv420p", dst).CombinedOutput()
		if err != nil {
			return 1, fmt.Errorf("%w: complexity sample: %s", ErrEncodeFailed, truncate(string(out)))
		}
		st, err := os.Stat(dst)
		if err != nil {
			return 1, err
		}
		totalBytes += st.Size()
		totalSecs += sampleSecs
		os.Remove(dst)
	}

	if totalSecs == 0 {
		return 1, nil
	}
	measured := float64(totalBytes) * 8 / totalSecs
	return measured / referenceBitrate360, nil
}

// ApplyComplexity scales a ladder's bitrate caps to the content.
//
// Bounded deliberately: an unbounded per-title ladder produces variants a naive
// client-side quality selector cannot reason about, and the extra saving past these
// limits is small.
func ApplyComplexity(rungs []Rung, ratio float64) []Rung {
	// The floor admits the full saving screen-recorded lecture content genuinely
	// shows; the ceiling stops pathological input inflating the ladder.
	scale := math.Max(0.5, math.Min(1.4, ratio))
	out := make([]Rung, len(rungs))
	for i, r := range rungs {
		r.MaxrateBPS = int(math.Round(float64(r.MaxrateBPS) * scale))
		out[i] = r
	}
	return out
}
