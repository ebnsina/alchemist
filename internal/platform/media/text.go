package media

import (
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"strings"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// Text is drawn here and composited by ffmpeg's overlay filter, rather than by
// ffmpeg's drawtext. drawtext needs an ffmpeg built with libfreetype, which is not a
// safe thing to assume of whatever build a machine happens to have — the one this was
// written against does not have it. Rendering here also means overlays use the
// product's own typeface instead of whatever font the host happens to ship.
//
//go:embed fonts/archivo.ttf
var archivoTTF []byte

var (
	faceOnce sync.Once
	parsed   *opentype.Font
	parseErr error
)

func loadFont() (*opentype.Font, error) {
	faceOnce.Do(func() {
		parsed, parseErr = opentype.Parse(archivoTTF)
	})
	return parsed, parseErr
}

// TextImage describes a line of text to be drawn onto a transparent canvas.
type TextImage struct {
	Text      string
	PixelSize float64
	Colour    color.RGBA
	// Shadow keeps text readable over a bright frame without a heavy box behind it.
	Shadow bool
}

// RenderText writes the text to a transparent PNG at dst and returns its size. The
// canvas is the text's own bounds plus a little padding, so positioning it later is
// a matter of placing a rectangle rather than guessing at baselines.
func RenderText(t TextImage, dst string) (w, h int, err error) {
	if strings.TrimSpace(t.Text) == "" {
		return 0, 0, fmt.Errorf("%w: the text is empty", ErrBadEdit)
	}
	f, err := loadFont()
	if err != nil {
		return 0, 0, fmt.Errorf("load font: %w", err)
	}
	if t.PixelSize < 8 {
		t.PixelSize = 8
	}

	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size: t.PixelSize, DPI: 72, Hinting: font.HintingFull,
	})
	if err != nil {
		return 0, 0, fmt.Errorf("make face: %w", err)
	}
	defer face.Close()

	d := &font.Drawer{Face: face}
	advance := d.MeasureString(t.Text)
	metrics := face.Metrics()

	pad := int(t.PixelSize * 0.25)
	w = advance.Ceil() + pad*2
	h = (metrics.Ascent + metrics.Descent).Ceil() + pad*2

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), image.Transparent, image.Point{}, draw.Src)

	baseline := fixed.P(pad, pad+metrics.Ascent.Ceil())
	if t.Shadow {
		// One offset pass in translucent black. Cheap, and it is the difference
		// between readable and invisible over a pale frame.
		d.Dst, d.Src = img, image.NewUniform(color.RGBA{0, 0, 0, 140})
		d.Dot = fixed.P(baseline.X.Ceil()+2, baseline.Y.Ceil()+2)
		d.DrawString(t.Text)
	}
	d.Dst, d.Src = img, image.NewUniform(t.Colour)
	d.Dot = baseline
	d.DrawString(t.Text)

	out, err := os.Create(dst)
	if err != nil {
		return 0, 0, err
	}
	defer out.Close()
	if err := png.Encode(out, img); err != nil {
		return 0, 0, err
	}
	return w, h, nil
}

// ParseHexColour reads #rgb or #rrggbb. Anything else is refused rather than guessed
// at, because a silently-black caption on a dark video looks like a bug in the render.
func ParseHexColour(s string) (color.RGBA, error) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "#")
	var r, g, b uint8
	switch len(s) {
	case 3:
		if _, err := fmt.Sscanf(strings.ToLower(s), "%1x%1x%1x", &r, &g, &b); err != nil {
			return color.RGBA{}, fmt.Errorf("%w: %q is not a colour", ErrBadEdit, s)
		}
		r, g, b = r*17, g*17, b*17
	case 6:
		if _, err := fmt.Sscanf(strings.ToLower(s), "%02x%02x%02x", &r, &g, &b); err != nil {
			return color.RGBA{}, fmt.Errorf("%w: %q is not a colour", ErrBadEdit, s)
		}
	default:
		return color.RGBA{}, fmt.Errorf("%w: %q is not a colour", ErrBadEdit, s)
	}
	return color.RGBA{R: r, G: g, B: b, A: 255}, nil
}
