package media

import "testing"

// A lazy rung the source cannot fill must not be deferred. Unfiltered, a 360p upload
// on the bd-mobile profile pends 480p and 720p rows: the asset never leaves
// partially_ready, the mezzanine is retained for its whole life, and first playback
// queues a JIT job that upscales. Nothing errors, which is why it needs a test.
func TestLazyRungsDropUpscales(t *testing.T) {
	ladder := []Rung{
		{Height: 144, Codec: "h264"},
		{Height: 240, Codec: "h264"},
		{Height: 360, Codec: "h264"},
		{Height: 480, Codec: "h264", Lazy: true},
		{Height: 720, Codec: "h264", Lazy: true},
	}

	if got := LazyRungs(Applicable(ladder, 360)); len(got) != 0 {
		t.Fatalf("360p source deferred %d rungs above its height: %+v", len(got), got)
	}
	got := LazyRungs(Applicable(ladder, 1080))
	if len(got) != 2 || got[0].Height != 480 || got[1].Height != 720 {
		t.Fatalf("1080p source should defer 480p and 720p, got %+v", got)
	}
}

// Chunks are 6-24s, so a fixed 12s window averages over a boundary instead of
// straddling it -- which hides the exact oscillation MaxAdjDiff exists to catch.
func TestPerChunkScoresFollowRealBoundaries(t *testing.T) {
	// 90 frames at 30fps: 3 seconds, split 2s / 1s.
	scores := make([]float64, 90)
	for i := range scores {
		scores[i] = 90
		if i >= 60 {
			scores[i] = 70 // the second chunk is visibly worse
		}
	}
	chunks := []Chunk{{Index: 0, StartSec: 0, EndSec: 2}, {Index: 1, StartSec: 2, EndSec: 3}}

	got := perChunkScores(scores, chunks)
	if len(got) != 2 {
		t.Fatalf("want one score per chunk, got %v", got)
	}
	if got[0] != 90 || got[1] != 70 {
		t.Fatalf("scores did not follow chunk boundaries: %v", got)
	}
	if one := perChunkScores(scores, nil); len(one) != 1 {
		t.Fatalf("no chunk plan should give one window, got %v", one)
	}
}
