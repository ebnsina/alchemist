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

	got := perChunkScores(scores, chunks, 30)
	if len(got) != 2 {
		t.Fatalf("want one score per chunk, got %v", got)
	}
	if got[0] != 90 || got[1] != 70 {
		t.Fatalf("scores did not follow chunk boundaries: %v", got)
	}
	if one := perChunkScores(scores, nil, 30); len(one) != 1 {
		t.Fatalf("no chunk plan should give one window, got %v", one)
	}
}

// Forcing everything to 30 invented six frames a second on film and threw half of a
// 60fps source away while still reporting it as 60.
func TestMezzanineRate(t *testing.T) {
	for _, c := range []struct {
		src  float64
		want int
	}{
		{23.976, 24}, {24, 24}, {25, 25}, {29.97, 30}, {30, 30},
		{48, 24}, {50, 25}, {59.94, 30}, {60, 30}, {120, 30},
		{15, 15},           // already cheap: inflating it only costs bitrate
		{0, 30}, {900, 30}, // unreadable rate falls back rather than failing
	} {
		if got := MezzanineRate(c.src); got != c.want {
			t.Errorf("MezzanineRate(%v) = %d, want %d", c.src, got, c.want)
		}
	}
}

// Two assets normalised to different rates must not share a params hash: the hash is
// what says "this is the same encode", and the GOP length differs between them.
func TestParamsHashSeparatesFrameRates(t *testing.T) {
	r := Rung{Height: 720, Codec: "h264", Profile: "main", CRF: 24, MaxrateBPS: 2_200_000}
	if ParamsHash(r, 24) == ParamsHash(r, 30) {
		t.Fatal("24fps and 30fps encodes of the same rung hash the same")
	}
}

// A recording is already what a mezzanine has to be. Re-encoding it made a copy
// roughly twice the size with a generation of loss, and the ladder was then built from
// that -- so the check that decides has to be right in both directions.
func TestConformantDecidesRemux(t *testing.T) {
	grid := &Probe{VideoCodec: "h264", FrameRate: 30, Width: 640, Height: 360}
	if !Conformant(grid, 30) {
		t.Error("a recording on the grid should be remuxed")
	}
	// 29.97 is not 30. Copying it leaves a chunk plan that drifts off the keyframes,
	// which surfaces as a stitch failure rather than as an error here.
	if Conformant(&Probe{VideoCodec: "h264", FrameRate: 29.97, Width: 640, Height: 360}, 30) {
		t.Error("a near-miss frame rate must be re-encoded, not copied")
	}
	if Conformant(&Probe{VideoCodec: "hevc", FrameRate: 30, Width: 640, Height: 360}, 30) {
		t.Error("another codec must be re-encoded")
	}
	if Conformant(&Probe{VideoCodec: "h264", FrameRate: 30}, 30) {
		t.Error("a probe with no dimensions must not be trusted")
	}
	if Conformant(nil, 30) {
		t.Error("no probe at all must not be trusted")
	}
}
