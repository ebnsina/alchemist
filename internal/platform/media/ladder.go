package media

import (
	"encoding/json"
	"fmt"
)

// Rung is one entry in a ladder profile. Profiles live in the database, so adding a
// market or changing bitrates is a row change, not a deploy.
type Rung struct {
	Height     int    `json:"height"`
	Codec      string `json:"codec"`
	Profile    string `json:"profile"`
	Preset     string `json:"preset"`
	CRF        int    `json:"crf"`
	MaxrateBPS int    `json:"maxrate_bps"`
	// Lazy rungs are generated on first playback instead of at ingest.
	//
	// This is the single largest cost decision in the system. Pre-encoding a full
	// ladder for everything grows storage ~1.8 PB/month at 100k ingest hours and
	// never stops; a library's cold majority is never watched. Storing a cheap low
	// ladder always and generating the expensive rungs on demand cuts stored bytes
	// by roughly 86%. See docs/03-cost-model.md.
	//
	// Absent means eager, so a profile written before this existed still behaves as
	// it did.
	Lazy bool `json:"lazy"`
}

// Eager returns the rungs encoded at ingest.
func Eager(rungs []Rung) []Rung {
	var out []Rung
	for _, r := range rungs {
		if !r.Lazy {
			out = append(out, r)
		}
	}
	return out
}

// LazyRungs returns the rungs deferred until something asks for them.
func LazyRungs(rungs []Rung) []Rung {
	var out []Rung
	for _, r := range rungs {
		if r.Lazy {
			out = append(out, r)
		}
	}
	return out
}

// Find locates a rung by height and codec.
func Find(rungs []Rung, height int, codec string) (Rung, bool) {
	for _, r := range rungs {
		if r.Height == height && r.Codec == codec {
			return r, true
		}
	}
	return Rung{}, false
}

func ParseLadder(raw []byte) ([]Rung, error) {
	var rungs []Rung
	if err := json.Unmarshal(raw, &rungs); err != nil {
		return nil, fmt.Errorf("parse ladder: %w", err)
	}
	if len(rungs) == 0 {
		return nil, fmt.Errorf("ladder profile has no rungs")
	}
	return rungs, nil
}

// Applicable drops rungs that would upscale the source. Upscaling spends the
// viewer's data on pixels the source never had. The lowest rung is always kept so
// poor connections still have something to play.
func Applicable(rungs []Rung, srcHeight int) []Rung {
	var out []Rung
	for _, r := range rungs {
		if r.Height <= srcHeight {
			out = append(out, r)
		}
	}
	if len(out) == 0 && len(rungs) > 0 {
		out = append(out, rungs[0])
	}
	return out
}
