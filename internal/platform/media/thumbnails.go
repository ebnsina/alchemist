package media

import (
	"context"
	"fmt"
	"math"
	"os/exec"
	"path/filepath"
)

// ThumbnailResult describes generated preview imagery.
type ThumbnailResult struct {
	Poster     string
	Sprite     string
	SpriteVTT  string
	Columns    int
	TileWidth  int
	TileHeight int
	IntervalS  float64
}

const (
	spriteTileWidth = 160
	spriteColumns   = 10
	// Roughly this many tiles regardless of duration, so a 3-hour lecture does not
	// produce a sprite sheet too large to download on mobile.
	spriteTargetTiles = 100
)

// Thumbnails produces a poster frame and a scrubbing sprite sheet with its WebVTT
// index. One sprite image beats hundreds of separate thumbnail requests, which
// matters most on the high-latency mobile connections this is built for.
func Thumbnails(ctx context.Context, mezzanine string, durationSec float64, height int, outDir string) (*ThumbnailResult, error) {
	tileH := int(math.Round(float64(spriteTileWidth) * float64(height) / float64(heightToWidth(height))))
	if tileH%2 != 0 {
		tileH++
	}

	poster := filepath.Join(outDir, "poster.jpg")
	// Seek to 10% rather than frame zero: many videos open on black or a slate.
	at := math.Max(0, durationSec*0.1)
	if out, err := exec.CommandContext(ctx, "ffmpeg",
		"-y", "-hide_banner", "-loglevel", "error",
		"-ss", fmt.Sprintf("%.3f", at), "-i", mezzanine,
		"-frames:v", "1", "-q:v", "3", "-vf", "scale=-2:720", poster).CombinedOutput(); err != nil {
		return nil, fmt.Errorf("%w: poster: %s", ErrEncodeFailed, truncate(string(out)))
	}

	interval := math.Max(1, durationSec/float64(spriteTargetTiles))
	tiles := int(math.Ceil(durationSec / interval))
	rows := int(math.Ceil(float64(tiles) / float64(spriteColumns)))

	sprite := filepath.Join(outDir, "sprite.jpg")
	if out, err := exec.CommandContext(ctx, "ffmpeg",
		"-y", "-hide_banner", "-loglevel", "error", "-i", mezzanine,
		"-vf", fmt.Sprintf("fps=1/%.4f,scale=%d:%d,tile=%dx%d",
			interval, spriteTileWidth, tileH, spriteColumns, rows),
		"-frames:v", "1", "-q:v", "4", sprite).CombinedOutput(); err != nil {
		return nil, fmt.Errorf("%w: sprite: %s", ErrEncodeFailed, truncate(string(out)))
	}

	vtt := filepath.Join(outDir, "sprite.vtt")
	if err := writeSpriteVTT(vtt, tiles, interval, durationSec, tileH); err != nil {
		return nil, err
	}

	return &ThumbnailResult{
		Poster: poster, Sprite: sprite, SpriteVTT: vtt,
		Columns: spriteColumns, TileWidth: spriteTileWidth,
		TileHeight: tileH, IntervalS: interval,
	}, nil
}

func heightToWidth(h int) int {
	// Tiles are generated from 16:9 mezzanine output.
	return h * 16 / 9
}
