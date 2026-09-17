package media

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"

	"golang.org/x/sync/errgroup"
)

type TranscodeResult struct {
	Probe *Probe
	// FrameRate is what the mezzanine and every rendition were built at, which is not
	// the source's rate whenever MezzanineRate snapped it. It is the rate delivered,
	// so it is the one worth reporting.
	FrameRate int
	// MezzProbe is what was actually encoded. It differs from Probe whenever the
	// source carries a rotation matrix: ffmpeg applies it building the mezzanine, so
	// a 1920x1080 phone clip tagged rotate:90 becomes a 1080x1920 intermediate and
	// every dimension derived from Probe is transposed.
	MezzProbe  *Probe
	Rungs      []Rung
	Chunks     int
	Scenes     int
	Complexity float64
	OutDir     string
	Manifest   *Result
	Thumbs     *ThumbnailResult       `json:",omitempty"`
	VMAF       map[string]*VMAFReport `json:",omitempty"`
}

// Options controls the analysis and quality stages.
type Options struct {
	// PerTitle scales bitrate caps to measured content complexity.
	PerTitle bool
	// SceneAware snaps chunk boundaries to scene changes.
	SceneAware bool
	// Encrypt, when set, protects the output with the given content key.
	Encrypt *Encryption
	// SourceConformant says the caller believes this source was produced to the
	// mezzanine's own rules -- which a broadcast recording is, by construction. It is
	// a hint and not a promise: Conformant checks the source before acting on it.
	SourceConformant bool
	// VMAFSample scores this fraction of renditions against the mezzanine. Sampling
	// rather than scoring everything: the point is catching a regression, not a
	// per-asset report, and VMAF costs more than the encode it is checking.
	VMAFSample float64

	// OnPlan is called once, with the ladder and the chunk plan, before any chunk is
	// encoded. OnChunkDone is called after each chunk lands, from several goroutines
	// at once, so an implementation has to be safe to call concurrently.
	//
	// These exist so progress can be recorded where somebody can see it. Without
	// them the only observable states are "started" and "finished", which on an hour
	// of video is an hour of nothing.
	OnPlan      func(rungs []Rung, chunks []Chunk, fps int) error
	OnChunkDone func(r Rung, c Chunk)
}

func DefaultOptions() Options {
	return Options{PerTitle: true, SceneAware: true, VMAFSample: 0.02}
}

// Transcode runs the whole chain locally: mezzanine, chunk plan, parallel chunk
// encode, stitch, package. The distributed path runs the identical primitives as
// separate jobs; this is the same code with a local worker pool instead of a queue.
func Transcode(ctx context.Context, src, workDir string, rungs []Rung, opts Options) (*TranscodeResult, error) {
	if err := os.MkdirAll(workDir, 0o750); err != nil {
		return nil, err
	}

	probe, err := Inspect(ctx, src)
	if err != nil {
		return nil, err
	}

	// Chosen once, here, and carried: the GOP grid, the chunk plan, the params hash and
	// the VMAF frame-to-time mapping are all expressed in whole frames of this rate.
	fps := MezzanineRate(probe.FrameRate)

	mezz := filepath.Join(workDir, "mezzanine.mp4")
	if !done(mezz) {
		build := func() error { return BuildMezzanine(ctx, src, mezz+tmpSuffix, fps) }
		// A recording is already H.264 on the grid at a constant rate, so re-encoding
		// it only makes a bigger, worse copy to build the ladder from.
		if opts.SourceConformant && Conformant(probe, fps) {
			build = func() error { return RemuxMezzanine(ctx, src, mezz+tmpSuffix) }
		}
		if err := build(); err != nil {
			return nil, err
		}
		if err := os.Rename(mezz+tmpSuffix, mezz); err != nil {
			return nil, err
		}
	}
	// Chunk against the mezzanine's own duration, not the source's: normalizing to
	// CFR can shift the total by a frame or two.
	mezzProbe, err := Inspect(ctx, mezz)
	if err != nil {
		return nil, err
	}

	// Against the mezzanine, not the source: rotation is already applied here.
	rungs = Applicable(rungs, mezzProbe.Height)

	complexity := 1.0
	if opts.PerTitle {
		complexity, err = Complexity(ctx, mezz, mezzProbe.DurationSec, workDir)
		if err != nil {
			return nil, err
		}
		rungs = ApplyComplexity(rungs, complexity)
	}

	var scenes []float64
	if opts.SceneAware {
		if scenes, err = SceneChanges(ctx, mezz, SceneThreshold); err != nil {
			return nil, err
		}
	}

	var chunks []Chunk
	if len(scenes) > 0 {
		chunks = PlanChunksAtScenes(mezzProbe.DurationSec, scenes)
	} else {
		chunks = PlanChunks(mezzProbe.DurationSec)
	}
	if len(chunks) == 0 {
		return nil, fmt.Errorf("%w: source too short to chunk", ErrUnreadableSource)
	}

	if opts.OnPlan != nil {
		if err := opts.OnPlan(rungs, chunks, fps); err != nil {
			return nil, err
		}
	}

	chunkDir := filepath.Join(workDir, "chunks")
	if err := os.MkdirAll(chunkDir, 0o750); err != nil {
		return nil, err
	}

	// One task per (chunk, rendition) — the same unit the distributed queue uses.
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(runtime.NumCPU())
	paths := make([][]string, len(rungs))
	for ri, r := range rungs {
		paths[ri] = make([]string, len(chunks))
		for ci, c := range chunks {
			ri, r, ci, c := ri, r, ci, c
			// Content-addressed: a retry overwrites deterministically and a
			// parameter change can never collide with older output.
			out := filepath.Join(chunkDir,
				fmt.Sprintf("%dp-%s-%05d.mp4", r.Height, ParamsHash(r, fps), c.Index))
			paths[ri][ci] = out
			g.Go(func() error {
				if !done(out) {
					if err := EncodeChunk(gctx, mezz, c, r, fps, out+tmpSuffix); err != nil {
						return err
					}
					if err := os.Rename(out+tmpSuffix, out); err != nil {
						return err
					}
				}
				if opts.OnChunkDone != nil {
					opts.OnChunkDone(r, c)
				}
				return nil
			})
		}
	}
	audio := filepath.Join(workDir, "audio.mp4")
	if probe.HasAudio && !done(audio) {
		g.Go(func() error {
			if err := EncodeAudio(gctx, mezz, audio+tmpSuffix); err != nil {
				return err
			}
			return os.Rename(audio+tmpSuffix, audio)
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	outDir := filepath.Join(workDir, "out")
	if err := os.MkdirAll(outDir, 0o750); err != nil {
		return nil, err
	}

	var inputs []Input
	for ri, r := range rungs {
		full := filepath.Join(workDir, fmt.Sprintf("%dp.mp4", r.Height))
		if err := Stitch(ctx, paths[ri], full); err != nil {
			return nil, err
		}
		if err := VerifyStitch(ctx, full, mezzProbe.DurationSec); err != nil {
			return nil, err
		}
		inputs = append(inputs, Input{
			Path: full, Kind: StreamVideo,
			Name: fmt.Sprintf("%dp", r.Height), Height: r.Height,
		})
	}
	if probe.HasAudio {
		inputs = append(inputs, Input{Path: audio, Kind: StreamAudio, Name: "audio"})
	}

	manifest, err := Package(ctx, inputs, outDir, opts.Encrypt)
	if err != nil {
		return nil, err
	}

	// Thumbnails come from the mezzanine, so they are unaffected by which rungs
	// exist or by encryption.
	thumbs, err := Thumbnails(ctx, mezz, mezzProbe.DurationSec,
		mezzProbe.Width, mezzProbe.Height, outDir)
	if err != nil {
		return nil, err
	}

	res := &TranscodeResult{
		Probe: probe, MezzProbe: mezzProbe, FrameRate: fps, Rungs: rungs, Chunks: len(chunks), Scenes: len(scenes),
		Complexity: complexity, OutDir: outDir, Manifest: manifest, Thumbs: thumbs,
	}

	// Quality gate. A failure here is not fatal to the asset: the renditions are
	// already correct, and the score is evidence for alerting, not a delivery block.
	//
	// Genuinely sampled, because VMAF costs more than the encode it checks -- it is
	// the slowest stage in this pipeline by a wide margin. The point is detecting a
	// regression across the fleet, which a small sample does as well as scoring
	// everything, at a fraction of the compute.
	if len(rungs) > 0 && rand.Float64() < opts.VMAFSample {
		top := inputs[len(rungs)-1]
		if rep, err := ScoreVMAF(ctx, top.Path, mezz, workDir, chunks, fps); err == nil {
			res.VMAF = map[string]*VMAFReport{top.Name: rep}
		}
	}
	return res, nil
}

// tmpSuffix marks output that is still being written.
//
// Every expensive stage writes to a temporary name and renames on success, so the
// final name existing means the work behind it finished. Without the rename a process
// killed mid-write leaves a truncated file that the next attempt would accept and
// stitch, producing a video that plays for a few seconds and stops.
const tmpSuffix = ".partial"

// done reports whether a stage's output is already there from an earlier attempt.
// Retries re-run the whole chain, and re-encoding two hours of chunks because the
// upload at the end timed out is the difference between a ten-minute retry and a
// six-hour one.
func done(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.Size() > 0
}
