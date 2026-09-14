package media

import (
	"math"
	"testing"
)

// Guards the pumping detector itself: bucketing must survive a partial final chunk.
func TestPlanChunksCoversDuration(t *testing.T) {
	for _, dur := range []float64{1, 11.9, 12, 12.1, 40, 3600.5} {
		chunks := PlanChunks(dur)
		if len(chunks) == 0 {
			t.Fatalf("duration %.1f produced no chunks", dur)
		}
		if got := chunks[len(chunks)-1].EndSec; math.Abs(got-dur) > 0.001 {
			t.Errorf("duration %.1f: last chunk ends at %.3f, want %.3f", dur, got, dur)
		}
		for i, c := range chunks {
			if c.EndSec <= c.StartSec {
				t.Errorf("duration %.1f chunk %d is empty: %+v", dur, i, c)
			}
			if i > 0 && c.StartSec != chunks[i-1].EndSec {
				t.Errorf("duration %.1f: gap between chunk %d and %d", dur, i-1, i)
			}
		}
	}
}

func TestApplicableNeverUpscales(t *testing.T) {
	ladder := []Rung{{Height: 240}, {Height: 360}, {Height: 720}, {Height: 1080}}
	got := Applicable(ladder, 480)
	if len(got) != 2 || got[1].Height != 360 {
		t.Errorf("480p source got rungs %+v, want 240p and 360p only", got)
	}
	// A source below every rung still gets the lowest one rather than nothing.
	if got := Applicable(ladder, 100); len(got) != 1 || got[0].Height != 240 {
		t.Errorf("tiny source got %+v, want the single lowest rung", got)
	}
}
