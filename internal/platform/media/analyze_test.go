package media

import (
	"math"
	"testing"
)

// Scene-aware boundaries must preserve every invariant the plain grid planner has:
// no gaps, no overlaps, full coverage, and every cut on the GOP grid. A boundary off
// the grid silently breaks stitching, which is a copy operation only because chunks
// begin on keyframes.
func TestPlanChunksAtScenesKeepsInvariants(t *testing.T) {
	cases := []struct {
		name     string
		duration float64
		scenes   []float64
	}{
		{"no scenes", 120, nil},
		{"dense scenes", 120, []float64{3, 7, 9, 14, 19, 23, 31, 44, 55, 61, 70, 88, 99, 110}},
		{"sparse scenes", 300, []float64{47, 155, 230}},
		{"scenes outside window", 100, []float64{1, 2, 99.5}},
		{"short source", 5, []float64{2}},
		{"exact multiple", 24, []float64{12}},
		{"odd tail", 37.5, []float64{11, 26}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			chunks := PlanChunksAtScenes(tc.duration, tc.scenes)
			if len(chunks) == 0 {
				t.Fatal("no chunks produced")
			}
			if chunks[0].StartSec != 0 {
				t.Errorf("first chunk starts at %.3f, want 0", chunks[0].StartSec)
			}
			last := chunks[len(chunks)-1].EndSec
			if math.Abs(last-tc.duration) > 0.001 {
				t.Errorf("coverage ends at %.3f, want %.3f", last, tc.duration)
			}
			for i, c := range chunks {
				if c.EndSec <= c.StartSec {
					t.Errorf("chunk %d is empty: %+v", i, c)
				}
				if c.Index != i {
					t.Errorf("chunk %d has index %d", i, c.Index)
				}
				if i > 0 && c.StartSec != chunks[i-1].EndSec {
					t.Errorf("gap or overlap between chunk %d and %d", i-1, i)
				}
				// Interior boundaries must sit on the keyframe grid; the final one
				// is the true duration and need not.
				if i < len(chunks)-1 {
					if r := math.Mod(c.EndSec, GOPSeconds); math.Abs(r) > 0.001 {
						t.Errorf("chunk %d ends at %.3f, off the %ds GOP grid",
							i, c.EndSec, GOPSeconds)
					}
				}
			}
		})
	}
}

func TestApplyComplexityIsBounded(t *testing.T) {
	ladder := []Rung{{Height: 360, MaxrateBPS: 700_000}}
	// Trivially simple content cannot shrink the ladder without limit.
	if got := ApplyComplexity(ladder, 0.01)[0].MaxrateBPS; got != 350_000 {
		t.Errorf("floor not applied: got %d, want 350000", got)
	}
	// Nor can pathological content inflate it without limit.
	if got := ApplyComplexity(ladder, 99)[0].MaxrateBPS; got != 980_000 {
		t.Errorf("ceiling not applied: got %d, want 980000", got)
	}
	// A lecture at 60% of reference complexity scales the cap to match.
	if got := ApplyComplexity(ladder, 0.6)[0].MaxrateBPS; got != 420_000 {
		t.Errorf("scaling wrong: got %d, want 420000", got)
	}
}
