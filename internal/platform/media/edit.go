package media

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Edit describes what to do to one video. Every field is optional: an edit with
// nothing set is a copy, which is a legitimate thing to ask for and must not error.
//
// Coordinates are fractions of the source, not pixels. A crop box drawn on a preview
// at whatever size the browser happened to render it means nothing to ffmpeg, and
// converting on the client would put the arithmetic in the one place that cannot be
// trusted about the source's real dimensions.
type Edit struct {
	// StartSec and EndSec trim. EndSec of 0 means "to the end".
	StartSec float64 `json:"start_sec"`
	EndSec   float64 `json:"end_sec"`

	// Crop is a box in fractions of the source, 0-1. Zero width or height means no
	// crop rather than a zero-pixel video.
	CropX float64 `json:"crop_x"`
	CropY float64 `json:"crop_y"`
	CropW float64 `json:"crop_w"`
	CropH float64 `json:"crop_h"`

	// Aspect reframes to a shape: "9:16" for reels, "1:1", "16:9". Applied after the
	// crop, so a focus area can be chosen and then fitted.
	Aspect string `json:"aspect"`

	// Width scales the result. Height follows from the aspect, so nothing is ever
	// stretched. Zero leaves the size alone.
	Width int `json:"width"`

	// Mute drops the audio track.
	Mute bool `json:"mute"`

	// Overlays are drawn on top, in order. Text and images are the same mechanism:
	// both become a PNG that ffmpeg composites, so nothing here depends on an ffmpeg
	// built with libfreetype.
	Overlays []Overlay `json:"overlays"`
}

// Corner is where an overlay sits. Named corners rather than raw coordinates because
// "bottom right, a bit in from the edge" is what a watermark actually is, and it has
// to keep meaning that when the frame is reshaped from wide to upright.
type Corner string

const (
	TopLeft     Corner = "top-left"
	TopRight    Corner = "top-right"
	BottomLeft  Corner = "bottom-left"
	BottomRight Corner = "bottom-right"
	Centre      Corner = "centre"
	// Free is a place picked by dragging: X and Y say where, the corners do not.
	Free Corner = "free"
)

// KnownOverlaySources is the closed set of things an overlay may be. It is a set, not
// a path, so nothing a customer sends can name an object we would then go and read.
var KnownOverlaySources = map[string]bool{"logo": true}

type Overlay struct {
	// Kind is "text" or "image". A logo and a watermark are both images; what makes
	// one a watermark is opacity and a corner, not a different code path.
	Kind string `json:"kind"`

	Text   string `json:"text,omitempty"`
	Colour string `json:"colour,omitempty"`
	Shadow bool   `json:"shadow,omitempty"`

	// Source names WHAT to put on, never WHERE it is. "logo" means this account's own
	// logo and the worker resolves it. A storage key from the client would be a way
	// to name another tenant's object and have us fetch it.
	Source string `json:"source,omitempty"`

	At Corner `json:"at"`
	// X and Y are the overlay's top-left as fractions of the frame. Read only when
	// At is "free" — a corner keeps meaning the corner when the frame is reshaped,
	// which a dragged coordinate cannot.
	X float64 `json:"x,omitempty"`
	Y float64 `json:"y,omitempty"`
	// Scale is a fraction of the frame: the text's cap height, or the image's width.
	Scale float64 `json:"scale"`
	// Margin is the inset from the edge, as a fraction of the frame.
	Margin  float64 `json:"margin"`
	Opacity float64 `json:"opacity"`

	// Empty means the whole clip.
	StartSec float64 `json:"start_sec,omitempty"`
	EndSec   float64 `json:"end_sec,omitempty"`
}

func (o Overlay) validate() error {
	if o.Kind != "text" && o.Kind != "image" {
		return fmt.Errorf("%w: an overlay is text or an image", ErrBadEdit)
	}
	if o.Kind == "text" && strings.TrimSpace(o.Text) == "" {
		return fmt.Errorf("%w: the text is empty", ErrBadEdit)
	}
	if o.Kind == "image" && !KnownOverlaySources[o.Source] {
		return fmt.Errorf("%w: %q is not an image we can put on a video", ErrBadEdit, o.Source)
	}
	switch o.At {
	case TopLeft, TopRight, BottomLeft, BottomRight, Centre, "":
	case Free:
		if o.X < 0 || o.X > 1 || o.Y < 0 || o.Y > 1 {
			return fmt.Errorf("%w: that position is off the frame", ErrBadEdit)
		}
	default:
		return fmt.Errorf("%w: %q is not a place on the frame", ErrBadEdit, o.At)
	}
	if o.Scale < 0 || o.Scale > 1 {
		return fmt.Errorf("%w: that size is off the frame", ErrBadEdit)
	}
	if o.Margin < 0 || o.Margin > 0.5 {
		return fmt.Errorf("%w: that margin is off the frame", ErrBadEdit)
	}
	if o.Opacity < 0 || o.Opacity > 1 {
		return fmt.Errorf("%w: opacity runs 0 to 1", ErrBadEdit)
	}
	if o.EndSec > 0 && o.EndSec <= o.StartSec {
		return fmt.Errorf("%w: that overlay ends before it starts", ErrBadEdit)
	}
	return nil
}

// position turns a corner into an ffmpeg overlay expression. W/H are the frame, w/h
// the overlay, which is what the filter already gives us.
func (o Overlay) position() (x, y string) {
	m := o.Margin
	if m == 0 {
		m = 0.04
	}
	inset := f(m)
	switch o.At {
	case Free:
		// Clamped in the expression too, so an overlay wider than expected stays on.
		return "min(W*" + f(o.X) + "\\,W-w)", "min(H*" + f(o.Y) + "\\,H-h)"
	case TopLeft:
		return "W*" + inset, "H*" + inset
	case TopRight:
		return "W-w-W*" + inset, "H*" + inset
	case BottomLeft:
		return "W*" + inset, "H-h-H*" + inset
	case Centre:
		return "(W-w)/2", "(H-h)/2"
	default: // bottom right is where a watermark goes unless told otherwise
		return "W-w-W*" + inset, "H-h-H*" + inset
	}
}

// ErrBadEdit is a request that cannot be rendered, as opposed to a render that fails.
var ErrBadEdit = fmt.Errorf("that edit does not describe a video")

func (e Edit) validate(durationSec float64) error {
	if e.StartSec < 0 || e.EndSec < 0 {
		return fmt.Errorf("%w: negative times", ErrBadEdit)
	}
	if e.EndSec > 0 && e.EndSec <= e.StartSec {
		return fmt.Errorf("%w: it would end before it starts", ErrBadEdit)
	}
	if durationSec > 0 && e.StartSec >= durationSec {
		return fmt.Errorf("%w: it starts after the video ends", ErrBadEdit)
	}
	for _, v := range []float64{e.CropX, e.CropY, e.CropW, e.CropH} {
		if v < 0 || v > 1 {
			return fmt.Errorf("%w: the crop box is outside the video", ErrBadEdit)
		}
	}
	if e.CropX+e.CropW > 1.0001 || e.CropY+e.CropH > 1.0001 {
		return fmt.Errorf("%w: the crop box runs off the edge", ErrBadEdit)
	}
	if e.Aspect != "" {
		if _, _, err := parseAspect(e.Aspect); err != nil {
			return err
		}
	}
	if e.Width < 0 || e.Width > 7680 {
		return fmt.Errorf("%w: that width is not one we make", ErrBadEdit)
	}
	// A frame covered in overlays is a mistake, not a design, and each one costs a
	// filter pass.
	if len(e.Overlays) > 8 {
		return fmt.Errorf("%w: that is more than eight things on one frame", ErrBadEdit)
	}
	for _, o := range e.Overlays {
		if err := o.validate(); err != nil {
			return err
		}
	}
	return nil
}

func parseAspect(s string) (w, h int, err error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("%w: %q is not an aspect like 9:16", ErrBadEdit, s)
	}
	w, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	h, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil || w <= 0 || h <= 0 {
		return 0, 0, fmt.Errorf("%w: %q is not an aspect like 9:16", ErrBadEdit, s)
	}
	return w, h, nil
}

// FilterGraph builds the -vf chain. Order matters and is fixed: crop the area wanted,
// reframe that area to the target shape, then scale. Doing it the other way round
// scales pixels that are about to be thrown away.
func (e Edit) FilterGraph() string {
	var parts []string

	if e.CropW > 0 && e.CropH > 0 {
		// Expressed against ffmpeg's own iw/ih so the real source size decides the
		// pixels, and rounded to even numbers because H.264 chroma requires it.
		parts = append(parts, fmt.Sprintf(
			"crop=trunc(iw*%s/2)*2:trunc(ih*%s/2)*2:trunc(iw*%s):trunc(ih*%s)",
			f(e.CropW), f(e.CropH), f(e.CropX), f(e.CropY)))
	}

	if e.Aspect != "" {
		aw, ah, err := parseAspect(e.Aspect)
		if err == nil {
			// Cover, then centre-crop: a reel filled with black bars is not a reel.
			// force_original_aspect_ratio=increase scales the short side up so the
			// frame is covered, and the crop takes the middle back out.
			parts = append(parts,
				fmt.Sprintf("scale=w=if(gte(iw/ih\\,%d/%d)\\,-2\\,iw):h=if(gte(iw/ih\\,%d/%d)\\,ih\\,-2)", aw, ah, aw, ah),
				fmt.Sprintf("crop=trunc(min(iw\\,ih*%d/%d)/2)*2:trunc(min(ih\\,iw*%d/%d)/2)*2", aw, ah, ah, aw))
		}
	}

	if e.Width > 0 {
		w := e.Width - e.Width%2
		// With a target shape the height is computed, not left to ffmpeg's -2. The
		// reframe crop rounds to even pixels, so the frame reaching this scale is a
		// hair off the exact ratio and -2 inherits the drift — 9:16 at 360 wide came
		// out 360x642 rather than 360x640.
		if aw, ah, err := parseAspect(e.Aspect); e.Aspect != "" && err == nil {
			h := int(float64(w) * float64(ah) / float64(aw))
			parts = append(parts, fmt.Sprintf("scale=%d:%d", w, h-h%2))
		} else {
			// Height follows the width so nothing is stretched, and stays even.
			parts = append(parts, fmt.Sprintf("scale=%d:-2", w))
		}
	}

	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, ",")
}

// f formats a fraction without scientific notation, which ffmpeg will not parse.
func f(v float64) string { return strconv.FormatFloat(v, 'f', 6, 64) }

// Args builds the whole command. The mezzanine is the input: it is already a clean,
// constant-frame-rate master, so an edit never re-reads the customer's original.
func (e Edit) Args(src, dst string, durationSec float64) ([]string, error) {
	if err := e.validate(durationSec); err != nil {
		return nil, err
	}

	args := []string{"-y", "-hide_banner", "-loglevel", "error"}
	// Seeking before -i is the fast one. Accuracy is still frame-exact because the
	// mezzanine has keyframes on a fixed two-second grid and we re-encode anyway.
	if e.StartSec > 0 {
		args = append(args, "-ss", f(e.StartSec))
	}
	args = append(args, "-i", src)
	if e.EndSec > 0 {
		args = append(args, "-t", f(e.EndSec-e.StartSec))
	}

	args = append(args, "-map", "0:v:0")
	if graph := e.FilterGraph(); graph != "" {
		args = append(args, "-vf", graph)
	}

	// No -r: the input is a mezzanine, already constant at whatever rate its source
	// was normalised to. Forcing 30 here threw that away before the pipeline that
	// re-ingests this file ever saw it, so a 24fps edit came back with judder.
	keyint := strconv.Itoa(GOPSeconds * DefaultFrameRate)
	args = append(args,
		"-c:v", "libx264", "-preset", "veryfast", "-crf", "17",
		"-pix_fmt", "yuv420p",
		"-g", keyint, "-keyint_min", keyint,
		"-sc_threshold", "0",
		"-movflags", "+faststart")

	if e.Mute {
		args = append(args, "-an")
	} else {
		args = append(args, "-map", "0:a:0?", "-c:a", "aac", "-b:a", "128k", "-ac", "2", "-ar", "48000")
	}
	return append(args, dst), nil
}

// PreparedOverlay is an overlay whose picture already exists on disk — rendered text
// or a fetched image. The worker does that part, because it needs storage and this
// package must not.
type PreparedOverlay struct {
	Overlay
	Path string
}

// RenderWith applies the edit including overlays. The filter graph is built as a
// chain: the base video is cropped and scaled first, then each overlay is scaled and
// composited onto the result in turn.
func RenderWith(ctx context.Context, e Edit, src, dst string, durationSec float64, ovl []PreparedOverlay) error {
	if err := e.validate(durationSec); err != nil {
		return err
	}
	if len(ovl) == 0 {
		return Render(ctx, e, src, dst, durationSec)
	}

	args := []string{"-y", "-hide_banner", "-loglevel", "error"}
	if e.StartSec > 0 {
		args = append(args, "-ss", f(e.StartSec))
	}
	args = append(args, "-i", src)
	if e.EndSec > 0 {
		args = append(args, "-t", f(e.EndSec-e.StartSec))
	}
	for _, o := range ovl {
		args = append(args, "-i", o.Path)
	}

	var chain []string
	base := "[0:v]"
	if g := e.FilterGraph(); g != "" {
		chain = append(chain, base+g+"[base]")
		base = "[base]"
	}

	for i, o := range ovl {
		scale := o.Scale
		if scale <= 0 {
			scale = 0.18
		}
		op := o.Opacity
		if op <= 0 {
			op = 1
		}

		// Opacity first, on the overlay's own stream.
		chain = append(chain, fmt.Sprintf(
			"[%d:v]format=rgba,colorchannelmixer=aa=%s[o%d]", i+1, f(op), i))

		// scale2ref takes two inputs and returns two: the scaled overlay and the
		// reference passed through. `main_w` is the frame — `iw` would be the overlay
		// itself, which is how a "fraction of the frame" ends up a fraction of the
		// wrong thing.
		chain = append(chain, fmt.Sprintf(
			"[o%d]%sscale2ref=w=main_w*%s:h=-1[ov%d][bg%d]", i, base, f(scale), i, i))

		x, y := o.position()
		enable := ""
		if o.StartSec > 0 || o.EndSec > 0 {
			end := o.EndSec
			if end <= 0 {
				end = durationSec + 1
			}
			enable = fmt.Sprintf(":enable='between(t,%s,%s)'", f(o.StartSec), f(end))
		}
		out := fmt.Sprintf("[v%d]", i)
		if i == len(ovl)-1 {
			out = "[vout]"
		}
		chain = append(chain, fmt.Sprintf(
			"[bg%d][ov%d]overlay=%s:%s%s%s", i, i, x, y, enable, out))
		base = out
	}

	args = append(args,
		"-filter_complex", strings.Join(chain, ";"),
		"-map", "[vout]")

	// No -r: the input is a mezzanine, already constant at whatever rate its source
	// was normalised to. Forcing 30 here threw that away before the pipeline that
	// re-ingests this file ever saw it, so a 24fps edit came back with judder.
	keyint := strconv.Itoa(GOPSeconds * DefaultFrameRate)
	args = append(args,
		"-c:v", "libx264", "-preset", "veryfast", "-crf", "17",
		"-pix_fmt", "yuv420p",
		"-g", keyint, "-keyint_min", keyint,
		"-sc_threshold", "0",
		"-movflags", "+faststart")
	if e.Mute {
		args = append(args, "-an")
	} else {
		args = append(args, "-map", "0:a:0?", "-c:a", "aac", "-b:a", "128k", "-ac", "2", "-ar", "48000")
	}
	args = append(args, dst)

	if out, err := exec.CommandContext(ctx, "ffmpeg", args...).CombinedOutput(); err != nil {
		return fmt.Errorf("%w: overlay: %s", ErrEncodeFailed, truncate(string(out)))
	}
	return nil
}

// Render applies the edit. The output is a normal video file, handed to the ordinary
// ingest path afterwards — an edit produces a new asset, it does not mutate one.
func Render(ctx context.Context, e Edit, src, dst string, durationSec float64) error {
	args, err := e.Args(src, dst, durationSec)
	if err != nil {
		return err
	}
	if out, err := exec.CommandContext(ctx, "ffmpeg", args...).CombinedOutput(); err != nil {
		return fmt.Errorf("%w: edit: %s", ErrEncodeFailed, truncate(string(out)))
	}
	return nil
}
