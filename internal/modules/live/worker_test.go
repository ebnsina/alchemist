package live

import (
	"strings"
	"testing"

	"github.com/ebnsina/alchemist/internal/platform/media"
)

// What a live broadcast publishes is decided entirely by this parse. Miss the
// EXT-X-MAP line and every player gets a playlist whose init segment 404s -- which
// looks like a corrupt stream, not a missing file. Treat a tag as a segment and the
// worker uploads a file that does not exist and stalls behind it.
func TestPlaylistFilesTakesInitAndSegmentsOnly(t *testing.T) {
	playlist := strings.Join([]string{
		"#EXTM3U",
		"#EXT-X-VERSION:7",
		"#EXT-X-TARGETDURATION:2",
		"#EXT-X-PLAYLIST-TYPE:EVENT",
		`#EXT-X-MAP:URI="init.mp4"`,
		"#EXTINF:2.000000,",
		"0.m4s",
		"#EXTINF:2.000000,",
		"1.m4s",
		"",
	}, "\n")

	got := PlaylistFiles([]byte(playlist))
	want := []string{"init.mp4", "0.m4s", "1.m4s"}

	if len(got) != len(want) {
		t.Fatalf("PlaylistFiles = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("PlaylistFiles[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// A playlist ffmpeg has opened but not yet written a segment into must publish
// nothing, or the first viewer gets an empty stream instead of waiting.
func TestPlaylistFilesEmptyBeforeFirstSegment(t *testing.T) {
	if got := PlaylistFiles([]byte("#EXTM3U\n#EXT-X-VERSION:7\n")); len(got) != 0 {
		t.Errorf("PlaylistFiles = %v, want nothing", got)
	}
}

// The cheapest rung, and only the eager ones: a lazy rung has no business being
// encoded in realtime, and picking the tallest would spend the whole encode box on
// one broadcast.
func TestLiveRungPicksLowestEager(t *testing.T) {
	rungs := []media.Rung{
		{Height: 720, Codec: "h264"},
		{Height: 240, Codec: "h264"},
		{Height: 144, Codec: "h264", Lazy: true},
		{Height: 360, Codec: "h264"},
	}
	got, ok := liveRung(rungs)
	if !ok || got.Height != 240 {
		t.Errorf("liveRung = %d (ok=%v), want 240", got.Height, ok)
	}

	if _, ok := liveRung([]media.Rung{{Height: 144, Lazy: true}}); ok {
		t.Error("a profile with no eager rung must not yield one to encode")
	}
}
