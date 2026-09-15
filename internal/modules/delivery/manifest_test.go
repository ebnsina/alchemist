package delivery

import (
	"strings"
	"testing"
)

func TestSignManifestHLS(t *testing.T) {
	in := []byte("#EXTM3U\n#EXT-X-MAP:URI=\"720p.cmfv\",BYTERANGE=\"872@0\"\n" +
		"#EXT-X-BYTERANGE:408467@1144\n720p.cmfv\n")
	got := string(signManifest(in, "exp=1&sig=abc", "720p.m3u8"))

	if !strings.Contains(got, `URI="720p.cmfv?exp=1&sig=abc"`) {
		t.Errorf("EXT-X-MAP URI was not signed:\n%s", got)
	}
	if !strings.Contains(got, "\n720p.cmfv?exp=1&sig=abc") {
		t.Errorf("bare segment URI was not signed:\n%s", got)
	}
	if strings.Contains(got, "#EXT-X-BYTERANGE:408467@1144?") {
		t.Errorf("a tag line was wrongly treated as a URI:\n%s", got)
	}
}

// A raw & breaks XML parsing, so every DASH player would reject the manifest.
func TestSignManifestDASHEscapesAmpersand(t *testing.T) {
	in := []byte("<MPD><BaseURL>720p.cmfv</BaseURL></MPD>")
	got := string(signManifest(in, "exp=1&sig=abc", "manifest.mpd"))

	if !strings.Contains(got, "<BaseURL>720p.cmfv?exp=1&amp;sig=abc</BaseURL>") {
		t.Errorf("ampersand not XML-escaped:\n%s", got)
	}
	if strings.Contains(got, "sig=abc</BaseURL>") && strings.Contains(got, "1&sig") {
		t.Errorf("raw ampersand survived:\n%s", got)
	}
}

func TestSignManifestLeavesAbsoluteURLsAlone(t *testing.T) {
	in := []byte("#EXTM3U\nhttps://cdn.example/seg.cmfv\n")
	if got := string(signManifest(in, "exp=1&sig=abc", "a.m3u8")); strings.Contains(got, "?exp=") {
		t.Errorf("absolute URL was rewritten:\n%s", got)
	}
}

// The cue payload carries an #xywh fragment. A query appended after it is read as
// part of the fragment, so the tile request arrives unsigned and 403s.
func TestSignManifestVTTKeepsQueryBeforeFragment(t *testing.T) {
	in := []byte("WEBVTT\n\n00:00:00.000 --> 00:00:05.000\nsprite.jpg#xywh=160,0,160,90\n")
	got := string(signManifest(in, "exp=1&sig=abc", "sprite.vtt"))

	if !strings.Contains(got, "sprite.jpg?exp=1&sig=abc#xywh=160,0,160,90") {
		t.Errorf("cue image was not signed before its fragment:\n%s", got)
	}
	if strings.Contains(got, "WEBVTT?") || strings.Contains(got, "00:00:05.000?") {
		t.Errorf("a header or timing line was treated as a URI:\n%s", got)
	}
}
