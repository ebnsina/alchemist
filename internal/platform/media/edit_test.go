package media

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Validation is the difference between "we cannot render that" and a worker burning
// three attempts on ffmpeg errors nobody can read.
func TestEditValidationRejectsWhatCannotBeRendered(t *testing.T) {
	for _, tc := range []struct {
		name string
		e    Edit
		ok   bool
	}{
		{"nothing set is a copy, not an error", Edit{}, true},
		{"a plain trim", Edit{StartSec: 2, EndSec: 10}, true},
		{"ends before it starts", Edit{StartSec: 10, EndSec: 2}, false},
		{"ends exactly when it starts", Edit{StartSec: 5, EndSec: 5}, false},
		{"open-ended trim", Edit{StartSec: 5}, true},
		{"negative start", Edit{StartSec: -1}, false},
		{"starts after the video ends", Edit{StartSec: 999}, false},
		{"a crop inside the frame", Edit{CropX: 0.1, CropY: 0.1, CropW: 0.5, CropH: 0.5}, true},
		{"a crop running off the right", Edit{CropX: 0.8, CropW: 0.5, CropH: 0.5}, false},
		{"a crop running off the bottom", Edit{CropY: 0.8, CropW: 0.5, CropH: 0.5}, false},
		{"a crop outside 0-1", Edit{CropW: 2}, false},
		{"a known aspect", Edit{Aspect: "9:16"}, true},
		{"nonsense aspect", Edit{Aspect: "portrait"}, false},
		{"zero aspect", Edit{Aspect: "0:16"}, false},
		{"a sane width", Edit{Width: 1080}, true},
		{"wider than 8K", Edit{Width: 99999}, false},
	} {
		err := tc.e.validate(60)
		if tc.ok && err != nil {
			t.Errorf("%s: rejected with %v, want accepted", tc.name, err)
		}
		if !tc.ok && err == nil {
			t.Errorf("%s: accepted, want rejected", tc.name)
		}
		if !tc.ok && err != nil && !errors.Is(err, ErrBadEdit) {
			t.Errorf("%s: error is not ErrBadEdit, so a caller cannot tell it apart", tc.name)
		}
	}
}

// An empty edit must produce no -vf at all. An empty filter string passed to ffmpeg
// is a parse error, not a no-op.
func TestEmptyEditHasNoFilter(t *testing.T) {
	if got := (Edit{}).FilterGraph(); got != "" {
		t.Fatalf("FilterGraph() = %q, want empty", got)
	}
	args, err := (Edit{}).Args("in.mp4", "out.mp4", 60)
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range args {
		if a == "-vf" {
			t.Fatal("-vf passed for an edit with no filters")
		}
	}
}

// Order is the whole point: cropping after scaling throws away pixels that were just
// paid for, and reframing before cropping reframes the wrong area.
func TestFilterOrderIsCropThenReframeThenScale(t *testing.T) {
	g := Edit{CropW: 0.5, CropH: 0.5, Aspect: "9:16", Width: 1080}.FilterGraph()
	crop := strings.Index(g, "crop=trunc(iw*")
	scale := strings.LastIndex(g, "scale=1080")
	if crop == -1 || scale == -1 || crop > scale {
		t.Fatalf("expected crop before the final scale, got %q", g)
	}
}

func TestMuteDropsAudioMapping(t *testing.T) {
	args, err := Edit{Mute: true}.Args("in.mp4", "out.mp4", 60)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-an") || strings.Contains(joined, "0:a:0?") {
		t.Fatalf("mute did not drop audio: %s", joined)
	}
}

// The graphs are strings handed to another program. Asserting on the string proves
// nothing about whether ffmpeg accepts it, so this renders a real file.
func TestRenderProducesTheRequestedShape(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not on PATH")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "src.mp4")

	// Six seconds of 640x360 test pattern with a tone, so trim and audio are real.
	mk := exec.Command("ffmpeg", "-y", "-hide_banner", "-loglevel", "error",
		"-f", "lavfi", "-i", "testsrc=size=640x360:rate=30:duration=6",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=6",
		"-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "aac", "-shortest", src)
	if out, err := mk.CombinedOutput(); err != nil {
		t.Skipf("could not build a fixture: %s", out)
	}

	for _, tc := range []struct {
		name   string
		e      Edit
		w, h   int
		maxDur float64
	}{
		{"untouched", Edit{}, 640, 360, 6.5},
		{"trimmed to two seconds", Edit{StartSec: 1, EndSec: 3}, 640, 360, 2.5},
		{"cropped to the middle", Edit{CropX: 0.25, CropY: 0.25, CropW: 0.5, CropH: 0.5}, 320, 180, 6.5},
		{"a reel, 9:16", Edit{Aspect: "9:16", Width: 360}, 360, 640, 6.5},
		{"square", Edit{Aspect: "1:1", Width: 300}, 300, 300, 6.5},
		{"resized only", Edit{Width: 320}, 320, 180, 6.5},
	} {
		dst := filepath.Join(dir, strings.ReplaceAll(tc.name, " ", "_")+".mp4")
		if err := Render(context.Background(), tc.e, src, dst, 6); err != nil {
			t.Errorf("%s: %v", tc.name, err)
			continue
		}
		st, err := os.Stat(dst)
		if err != nil || st.Size() == 0 {
			t.Errorf("%s: produced no file", tc.name)
			continue
		}
		p, err := Inspect(context.Background(), dst)
		if err != nil {
			t.Errorf("%s: probe: %v", tc.name, err)
			continue
		}
		if p.Width != tc.w || p.Height != tc.h {
			t.Errorf("%s: %dx%d, want %dx%d", tc.name, p.Width, p.Height, tc.w, tc.h)
		}
		if p.DurationSec > tc.maxDur {
			t.Errorf("%s: %.2fs, want under %.2fs", tc.name, p.DurationSec, tc.maxDur)
		}
	}
}
