package media

import (
	"fmt"
	"math"
	"os"
	"strings"
)

// writeSpriteVTT indexes each tile by time range and pixel position, which is the
// format players expect for scrubbing previews.
func writeSpriteVTT(path string, tiles int, interval, durationSec float64, tileH int) error {
	var b strings.Builder
	b.WriteString("WEBVTT\n\n")
	for i := range tiles {
		start := float64(i) * interval
		end := math.Min(start+interval, durationSec)
		if start >= durationSec {
			break
		}
		x := (i % spriteColumns) * spriteTileWidth
		y := (i / spriteColumns) * tileH
		fmt.Fprintf(&b, "%s --> %s\nsprite.jpg#xywh=%d,%d,%d,%d\n\n",
			vttTime(start), vttTime(end), x, y, spriteTileWidth, tileH)
	}
	return os.WriteFile(path, []byte(b.String()), 0o600)
}

func vttTime(sec float64) string {
	h := int(sec) / 3600
	m := (int(sec) % 3600) / 60
	s := sec - float64(h*3600+m*60)
	return fmt.Sprintf("%02d:%02d:%06.3f", h, m, s)
}
