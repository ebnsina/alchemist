package pipeline

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ebnsina/alchemist/internal/platform/media"
)

// A wrong codec string is worse than a missing one: players use it to decide whether
// a variant is playable at all, so an over-stated level makes a phone skip a rung it
// could have played, and an under-stated one makes it try a rung it cannot.
func TestCodecString(t *testing.T) {
	cases := []struct {
		profile string
		height  int
		want    string
	}{
		{"baseline", 144, "avc1.42401e"}, // baseline, level 3.0
		{"main", 240, "avc1.4d401e"},
		{"main", 360, "avc1.4d401e"},
		{"main", 480, "avc1.4d401e"},  // 480 is not above 480
		{"main", 720, "avc1.4d401f"},  // level 3.1
		{"high", 1080, "avc1.644028"}, // level 4.0
		{"high", 2160, "avc1.644033"}, // level 5.1
		// An unknown profile must fall back to main rather than emit something no
		// player recognises.
		{"nonsense", 360, "avc1.4d401e"},
		{"", 360, "avc1.4d401e"},
	}
	for _, tc := range cases {
		got := codecString(media.Rung{Profile: tc.profile, Height: tc.height})
		if got != tc.want {
			t.Errorf("codecString(%s %dp) = %s, want %s", tc.profile, tc.height, got, tc.want)
		}
	}
}

// Manifests served with the wrong type are refused by players, and media served as
// octet-stream defeats range caching in some proxies.
func TestContentType(t *testing.T) {
	cases := map[string]string{
		"master.m3u8":   "application/vnd.apple.mpegurl",
		"720p.m3u8":     "application/vnd.apple.mpegurl",
		"manifest.mpd":  "application/dash+xml",
		"720p.cmfv":     "video/mp4",
		"audio.cmfa":    "video/mp4",
		"mezzanine.mp4": "video/mp4",
	}
	for name, want := range cases {
		if got := contentType(name); got != want {
			t.Errorf("contentType(%q) = %q, want %q", name, got, want)
		}
	}
	// Anything unrecognised must still get a type, never an empty header.
	if got := contentType("strange.xyz"); got == "" {
		t.Error("an unknown extension produced an empty content type")
	}
}

// Error codes are the API contract; customers branch on them. A media failure that
// maps to the wrong code sends the customer looking in the wrong place.
func TestErrorCodeMapping(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{media.ErrNoVideoStream, "no_video_stream"},
		{media.ErrUnreadableSource, "unreadable_source"},
		{media.ErrStitchFailed, "stitch_failed"},
		{media.ErrPackageFailed, "package_failed"},
		{media.ErrEncodeFailed, "encode_failed"},
		{errors.New("something unexpected"), "encode_failed"},
		// Wrapping must not lose the classification.
		{fmt.Errorf("chunk 3: %w", media.ErrStitchFailed), "stitch_failed"},
		{fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", media.ErrNoVideoStream)), "no_video_stream"},
	}
	for _, tc := range cases {
		if got := errorCode(tc.err); got != tc.want {
			t.Errorf("errorCode(%v) = %q, want %q", tc.err, got, tc.want)
		}
	}
}

// The extension filter is what stops a customer's PDFs and images being transcoded,
// which costs them money and produces nothing.
func TestVideoExtensionFilter(t *testing.T) {
	accept := []string{
		"lecture.mp4", "clip.MOV", "recording.mkv", "old.avi",
		"stream.ts", "broadcast.mxf", "web.webm", "phone.m4v",
		"lectures/week1/part2.MP4",
	}
	for _, k := range accept {
		if !videoExtensions[strings.ToLower(path.Ext(k))] {
			t.Errorf("%q was rejected but is a video", k)
		}
	}

	reject := []string{
		"notes.pdf", "slide.png", "archive.zip", "readme.txt",
		"subtitles.srt", "audio.mp3", "data.json", "noextension",
		// A directory marker, not a file.
		"lectures/",
	}
	for _, k := range reject {
		if videoExtensions[strings.ToLower(path.Ext(k))] {
			t.Errorf("%q was accepted but is not a video", k)
		}
	}
}

func TestDeref(t *testing.T) {
	if got := deref(nil); got != "" {
		t.Errorf("deref(nil) = %q, want empty", got)
	}
	v := "value"
	if got := deref(&v); got != "value" {
		t.Errorf("deref(&v) = %q, want %q", got, v)
	}
}

// The hash decides whether an upload is re-encoded or linked to existing renditions,
// so it must depend on content and nothing else.
func TestHashFile(t *testing.T) {
	dir := t.TempDir()
	write := func(name string, content []byte) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, content, 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}

	body := bytes.Repeat([]byte("alchemist"), 5000)
	a := write("a.mp4", body)
	b := write("b.mp4", body) // same content, different name
	c := write("c.mp4", append(bytes.Clone(body), 'x'))

	ha, err := hashFile(a)
	if err != nil {
		t.Fatal(err)
	}
	hb, err := hashFile(b)
	if err != nil {
		t.Fatal(err)
	}
	hc, err := hashFile(c)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(ha, hb) {
		t.Error("identical content under different names produced different hashes; it would be re-encoded")
	}
	if bytes.Equal(ha, hc) {
		t.Error("differing content produced the same hash; the wrong renditions would be served")
	}
	if len(ha) != 32 {
		t.Errorf("hash is %d bytes, want 32", len(ha))
	}

	if _, err := hashFile(filepath.Join(dir, "missing.mp4")); err == nil {
		t.Error("hashing a missing file did not error")
	}
}
