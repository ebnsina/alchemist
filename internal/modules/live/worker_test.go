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

// The best rung one realtime core can carry, and only the eager ones: a lazy rung has
// no business being encoded in realtime, and the tallest would spend the whole encode
// box on one broadcast. Taking the cheapest instead put every class out at 144p.
func TestLiveRungPicksBestUnderTheCap(t *testing.T) {
	rungs := []media.Rung{
		{Height: 720, Codec: "h264"},
		{Height: 240, Codec: "h264"},
		{Height: 144, Codec: "h264", Lazy: true},
		{Height: 360, Codec: "h264"},
	}
	got, ok := liveRung(rungs)
	if !ok || got.Height != LiveMaxHeight {
		t.Errorf("liveRung = %d (ok=%v), want %d", got.Height, ok, LiveMaxHeight)
	}

	// Better the smallest thing the profile has than no broadcast at all.
	tall, ok := liveRung([]media.Rung{{Height: 1080}, {Height: 720}})
	if !ok || tall.Height != 720 {
		t.Errorf("all rungs above the cap: got %d (ok=%v), want 720", tall.Height, ok)
	}

	if _, ok := liveRung([]media.Rung{{Height: 144, Lazy: true}}); ok {
		t.Error("a profile with no eager rung must not yield one to encode")
	}
}

// A dropped uplink and a presenter closing OBS are the same event from here. ffmpeg
// writes ENDLIST on either, and a player that has seen it does not come back when the
// segments resume -- so it is stripped until the broadcast is genuinely over.
func TestOpenPlaylistDropsEndList(t *testing.T) {
	body := []byte("#EXTM3U\n#EXT-X-PLAYLIST-TYPE:EVENT\n#EXTINF:2.0,\n0.m4s\n#EXT-X-ENDLIST\n")
	got := string(openPlaylist(body))
	if strings.Contains(got, "#EXT-X-ENDLIST") {
		t.Fatalf("playlist still ends the stream:\n%s", got)
	}
	if !strings.Contains(got, "0.m4s") {
		t.Fatalf("segments were lost:\n%s", got)
	}
}

// Appending the bytes of two encoder connections makes a file whose timeline runs
// backwards, and everything downstream believes the shorter duration: a ten-second
// test broadcast with one reconnect came back as a six-second mezzanine, silently.
func TestPlaylistRunsSplitOnDiscontinuity(t *testing.T) {
	// The shape ffmpeg 9.0.1 actually writes with -hls_flags append_list.
	body := []byte(`#EXTM3U
#EXT-X-VERSION:7
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-PLAYLIST-TYPE:EVENT
#EXT-X-MAP:URI="init.mp4"
#EXTINF:2.000000,
0.m4s
#EXTINF:2.000000,
1.m4s
#EXT-X-DISCONTINUITY
#EXTINF:2.000000,
2.m4s
#EXT-X-ENDLIST
`)
	runs := PlaylistRuns(body)
	if len(runs) != 2 {
		t.Fatalf("want one group per connection, got %v", runs)
	}
	want := [][]string{{"init.mp4", "0.m4s", "1.m4s"}, {"init.mp4", "2.m4s"}}
	for i := range want {
		if strings.Join(runs[i], ",") != strings.Join(want[i], ",") {
			t.Errorf("run %d = %v, want %v", i, runs[i], want[i])
		}
	}

	// One connection is the ordinary case and must stay one group.
	one := PlaylistRuns([]byte("#EXT-X-MAP:URI=\"init.mp4\"\n#EXTINF:2.0,\n0.m4s\n"))
	if len(one) != 1 || len(one[0]) != 2 {
		t.Errorf("a single connection = %v, want one group of two", one)
	}

	// A playlist opened but never written into has no run to convert.
	if got := PlaylistRuns([]byte("#EXTM3U\n#EXT-X-VERSION:7\n")); len(got) != 0 {
		t.Errorf("PlaylistRuns = %v, want nothing", got)
	}
}
