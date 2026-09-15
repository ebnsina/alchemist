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
	Probe      *Probe
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
	OnPlan      func(rungs []Rung, chunks []Chunk) error
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

	mezz := filepath.Join(workDir, "mezzanine.mp4")
	if err := BuildMezzanine(ctx, src, mezz); err != nil {
		return nil, err
	}
	// Chunk against the mezzanine's own duration, not the source's: normalizing to
	// CFR can shift the total by a frame or two.
	mezzProbe, err := Inspect(ctx, mezz)
	if err != nil {
		return nil, err
	}

	rungs = Applicable(rungs, probe.Height)

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
		if err := opts.OnPlan(rungs, chunks); err != nil {
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
				fmt.Sprintf("%dp-%s-%05d.mp4", r.Height, ParamsHash(r), c.Index))
			paths[ri][ci] = out
			g.Go(func() error {
				if err := EncodeChunk(gctx, mezz, c, r, out); err != nil {
					return err
				}
				if opts.OnChunkDone != nil {
					opts.OnChunkDone(r, c)
				}
				return nil
			})
		}
	}
	audio := filepath.Join(workDir, "audio.mp4")
	if probe.HasAudio {
		g.Go(func() error { return EncodeAudio(gctx, mezz, audio) })
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
	thumbs, err := Thumbnails(ctx, mezz, mezzProbe.DurationSec, mezzProbe.Height, outDir)
	if err != nil {
		return nil, err
	}

	res := &TranscodeResult{
		Probe: probe, Rungs: rungs, Chunks: len(chunks), Scenes: len(scenes),
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
		if rep, err := ScoreVMAF(ctx, top.Path, mezz, workDir); err == nil {
			res.VMAF = map[string]*VMAFReport{top.Name: rep}
		}
	}
	return res, nil
}
