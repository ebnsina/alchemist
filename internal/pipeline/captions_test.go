package pipeline

import (
	"strings"
	"testing"
)

// A subtitle group declared and never referenced by a variant is silently ignored:
// the tracks are in the manifest and no menu ever shows them.
func TestHLSCaptionMediaIsReferencedByVariants(t *testing.T) {
	media, attr := hlsCaptionMedia([]Caption{
		{Language: "bn", Label: "বাংলা"},
		{Language: "en", Label: "English"},
	})
	if attr != `,SUBTITLES="subs"` {
		t.Fatalf("variants would not point at the group: %q", attr)
	}
	for _, want := range []string{
		`TYPE=SUBTITLES`, `GROUP-ID="subs"`, `LANGUAGE="bn"`, `LANGUAGE="en"`,
		`URI="subs_bn.m3u8"`, `URI="subs_en.m3u8"`,
	} {
		if !strings.Contains(media, want) {
			t.Errorf("missing %s in:\n%s", want, media)
		}
	}
	// Subtitles switched on for a viewer who did not ask are turned off once and
	// never turned back on.
	if strings.Contains(media, "DEFAULT=YES") {
		t.Error("a track is marked default")
	}

	if m, a := hlsCaptionMedia(nil); m != "" || a != "" {
		t.Errorf("no tracks should add nothing, got %q and %q", m, a)
	}
}

// HLS will not read a .vtt the master points at directly; it wants a playlist naming
// it. A missing or empty one shows as "subtitles unavailable" with nothing logged.
func TestCaptionPlaylistNamesItsTrack(t *testing.T) {
	got := string(captionPlaylist("bn", 3725.4))
	for _, want := range []string{"#EXTM3U", "#EXT-X-TARGETDURATION:3726", "subs_bn.vtt", "#EXT-X-ENDLIST"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %s in:\n%s", want, got)
		}
	}
	// A track uploaded before the encode finished has no duration to work from.
	if p := string(captionPlaylist("en", 0)); !strings.Contains(p, "#EXT-X-TARGETDURATION:1") {
		t.Errorf("a zero duration must not produce a zero target duration:\n%s", p)
	}
}

func TestDashCaptionSetsAreWellFormed(t *testing.T) {
	got := dashCaptionSets([]Caption{{Language: "bn", Label: "বাংলা"}})
	for _, want := range []string{
		`contentType="text"`, `mimeType="text/vtt"`, `lang="bn"`,
		`value="subtitle"`, `<BaseURL>subs_bn.vtt</BaseURL>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %s in:\n%s", want, got)
		}
	}
	if dashCaptionSets(nil) != "" {
		t.Error("no tracks should add no AdaptationSet")
	}
}
