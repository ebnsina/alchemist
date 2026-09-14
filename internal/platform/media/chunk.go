package media

import "math"

// TargetChunkSeconds balances parallelism against compression efficiency and the
// per-chunk encoder ramp-up. Must be a multiple of GOPSeconds so every chunk starts
// on a keyframe and can be stitched without re-encoding.
const TargetChunkSeconds = 12

type Chunk struct {
	Index    int
	StartSec float64
	EndSec   float64
}

// PlanChunks cuts the timeline on the mezzanine's known keyframe grid. Scene-aware
// boundaries are a phase-2 refinement; the grid is already correct, just not optimal.
func PlanChunks(durationSec float64) []Chunk {
	if durationSec <= 0 {
		return nil
	}
	size := float64(TargetChunkSeconds)
	n := int(math.Ceil(durationSec / size))
	chunks := make([]Chunk, 0, n)
	for i := range n {
		start := float64(i) * size
		end := math.Min(start+size, durationSec)
		if end-start < 0.001 {
			continue
		}
		chunks = append(chunks, Chunk{Index: i, StartSec: start, EndSec: end})
	}
	return chunks
}
