package media

import (
	"context"
	"image/color"
	"os/exec"
	"path/filepath"
	"testing"
)

// Text is rendered here rather than by ffmpeg's drawtext, so the drawing has to be
// checked on its own: a zero-sized or blank PNG would composite as nothing at all and
// look like the overlay was ignored.
func TestRenderTextProducesAVisibleImage(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "t.png")

	w, h, err := RenderText(TextImage{
		Text: "Nile Academy", PixelSize: 48, Colour: color.RGBA{255, 255, 255, 255}, Shadow: true,
	}, dst)
	if err != nil {
		t.Fatal(err)
	}
	if w < 100 || h < 30 {
		t.Fatalf("text rendered %dx%d, too small to be the words asked for", w, h)
	}

	// Longer text must be wider, or the measurement is not being used.
	w2, _, err := RenderText(TextImage{
		Text: "Nile Academy, Dhaka", PixelSize: 48, Colour: color.RGBA{255, 255, 255, 255},
	}, filepath.Join(dir, "t2.png"))
	if err != nil {
		t.Fatal(err)
	}
	if w2 <= w {
		t.Errorf("longer text was not wider: %d then %d", w, w2)
	}
}

func TestEmptyTextIsRefused(t *testing.T) {
	if _, _, err := RenderText(TextImage{Text: "   ", PixelSize: 40}, filepath.Join(t.TempDir(), "x.png")); err == nil {
		t.Fatal("blank text was accepted")
	}
}

func TestParseHexColour(t *testing.T) {
	for _, tc := range []struct {
		in   string
		ok   bool
		want color.RGBA
	}{
		{"#ffffff", true, color.RGBA{255, 255, 255, 255}},
		{"c9f24d", true, color.RGBA{201, 242, 77, 255}},
		{"#fff", true, color.RGBA{255, 255, 255, 255}},
		{"lime", false, color.RGBA{}},
		{"#ff", false, color.RGBA{}},
		{"", false, color.RGBA{}},
	} {
		got, err := ParseHexColour(tc.in)
		if tc.ok && (err != nil || got != tc.want) {
			t.Errorf("%q: got %v %v, want %v", tc.in, got, err, tc.want)
		}
		if !tc.ok && err == nil {
			t.Errorf("%q was accepted, want refused", tc.in)
		}
	}
}

func TestOverlayValidation(t *testing.T) {
	for _, tc := range []struct {
		name string
		o    Overlay
		ok   bool
	}{
		{"a caption", Overlay{Kind: "text", Text: "hello", At: BottomLeft}, true},
		{"a watermark", Overlay{Kind: "image", Source: "logo", At: BottomRight, Opacity: 0.6}, true},
		{"text with nothing in it", Overlay{Kind: "text", Text: " "}, false},
		{"image with no image", Overlay{Kind: "image"}, false},
		// A storage key from the client would be a way to name another tenant's
		// object and have the worker fetch it. Only known names are resolvable.
		{"an image naming a storage key", Overlay{Kind: "image", Source: "src/other-tenant/asset/original"}, false},
		{"an image naming someone else's logo", Overlay{Kind: "image", Source: "brand/someone-else/logo.png"}, false},
		{"neither", Overlay{Kind: "sticker", Text: "x"}, false},
		{"a corner that is not one", Overlay{Kind: "text", Text: "x", At: "middle-ish"}, false},
		{"dragged somewhere on the frame", Overlay{Kind: "text", Text: "x", At: Free, X: 0.3, Y: 0.8}, true},
		{"dragged off the frame", Overlay{Kind: "text", Text: "x", At: Free, X: 1.4, Y: 0}, false},
		{"opacity past 1", Overlay{Kind: "text", Text: "x", Opacity: 4}, false},
		{"off-frame scale", Overlay{Kind: "text", Text: "x", Scale: 2}, false},
		{"ends before it starts", Overlay{Kind: "text", Text: "x", StartSec: 5, EndSec: 2}, false},
	} {
		err := tc.o.validate()
		if tc.ok != (err == nil) {
			t.Errorf("%s: err=%v, want ok=%v", tc.name, err, tc.ok)
		}
	}
}

// The graph is a string handed to another program. Asserting on the string proves
// nothing — this renders a real video with a caption and a watermark on it.
func TestRenderWithOverlaysProducesAVideo(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not on PATH")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "src.mp4")
	if out, err := exec.Command("ffmpeg", "-y", "-hide_banner", "-loglevel", "error",
		"-f", "lavfi", "-i", "testsrc=size=640x360:rate=30:duration=4",
		"-c:v", "libx264", "-pix_fmt", "yuv420p", src).CombinedOutput(); err != nil {
		t.Skipf("fixture: %s", out)
	}

	caption := filepath.Join(dir, "caption.png")
	if _, _, err := RenderText(TextImage{
		Text: "Week one", PixelSize: 40, Colour: color.RGBA{255, 255, 255, 255}, Shadow: true,
	}, caption); err != nil {
		t.Fatal(err)
	}
	mark := filepath.Join(dir, "mark.png")
	if _, _, err := RenderText(TextImage{
		Text: "ALCHEMIST", PixelSize: 28, Colour: color.RGBA{201, 242, 77, 255},
	}, mark); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name string
		e    Edit
		ovl  []PreparedOverlay
		w, h int
	}{
		{
			name: "a caption on an untouched frame",
			e:    Edit{},
			ovl:  []PreparedOverlay{{Overlay{Kind: "text", At: BottomLeft, Scale: 0.4}, caption}},
			w:    640, h: 360,
		},
		{
			name: "caption and watermark together",
			e:    Edit{},
			ovl: []PreparedOverlay{
				{Overlay{Kind: "text", At: TopLeft, Scale: 0.35}, caption},
				{Overlay{Kind: "image", At: BottomRight, Scale: 0.2, Opacity: 0.5}, mark},
			},
			w: 640, h: 360,
		},
		{
			name: "a watermark survives a reel reframe",
			e:    Edit{Aspect: "9:16", Width: 360},
			ovl:  []PreparedOverlay{{Overlay{Kind: "image", At: BottomRight, Scale: 0.3, Opacity: 0.7}, mark}},
			w:    360, h: 640,
		},
		{
			name: "a caption dragged to a spot on the frame",
			e:    Edit{},
			ovl:  []PreparedOverlay{{Overlay{Kind: "text", At: Free, X: 0.31, Y: 0.72, Scale: 0.4}, caption}},
			w:    640, h: 360,
		},
		{
			name: "an overlay for part of the clip only",
			e:    Edit{},
			ovl:  []PreparedOverlay{{Overlay{Kind: "text", At: Centre, Scale: 0.5, StartSec: 1, EndSec: 2}, caption}},
			w:    640, h: 360,
		},
	} {
		dst := filepath.Join(dir, tc.name+".mp4")
		if err := RenderWith(context.Background(), tc.e, src, dst, 4, tc.ovl); err != nil {
			t.Errorf("%s: %v", tc.name, err)
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
	}
}
