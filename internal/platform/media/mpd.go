package media

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	representationRE = regexp.MustCompile(`(?s)<Representation\b[^>]*>.*?</Representation>`)
	heightAttrRE     = regexp.MustCompile(`height="(\d+)"`)
	videoSetRE       = regexp.MustCompile(`(?s)<AdaptationSet\b[^>]*contentType="video".*?</AdaptationSet>`)
	adaptationOpenRE = regexp.MustCompile(`<AdaptationSet\b[^>]*>`)
)

// VideoRepresentations extracts each video <Representation> from a packaged MPD,
// keyed by height.
//
// Captured at package time because it cannot be recovered later: the published media
// is encrypted, and the packager cannot re-read it to regenerate a manifest.
func VideoRepresentations(mpdPath string) (map[int]string, error) {
	raw, err := os.ReadFile(mpdPath)
	if err != nil {
		return nil, err
	}
	videoSet := videoSetRE.Find(raw)
	if videoSet == nil {
		return nil, fmt.Errorf("no video AdaptationSet in %s", mpdPath)
	}

	// With several renditions the packager keeps width/height on each Representation.
	// With exactly one it hoists them onto the AdaptationSet instead, so a
	// single-rendition manifest -- which is what packaging one deferred rung produces
	// -- has no height to key on unless the enclosing element is consulted.
	setHeight := 0
	if open := adaptationOpenRE.Find(videoSet); open != nil {
		if m := heightAttrRE.FindSubmatch(open); m != nil {
			fmt.Sscanf(string(m[1]), "%d", &setHeight)
		}
	}

	out := map[int]string{}
	for _, rep := range representationRE.FindAllString(string(videoSet), -1) {
		h := setHeight
		if m := heightAttrRE.FindStringSubmatch(rep); m != nil {
			fmt.Sscanf(m[1], "%d", &h)
		}
		if h == 0 {
			continue
		}
		out[h] = rep
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no representations found in %s", mpdPath)
	}
	return out, nil
}

// MPDSkeleton returns the manifest with its video representations replaced by a
// placeholder, so they can be substituted as renditions are added.
func MPDSkeleton(mpdPath string) (string, error) {
	raw, err := os.ReadFile(mpdPath)
	if err != nil {
		return "", err
	}
	videoSet := videoSetRE.Find(raw)
	if videoSet == nil {
		return "", fmt.Errorf("no video AdaptationSet in %s", mpdPath)
	}
	stripped := representationRE.ReplaceAllString(string(videoSet), "")
	// Collapse the blank lines the removal leaves behind.
	stripped = regexp.MustCompile(`\n\s*\n`).ReplaceAllString(stripped, "\n")
	stripped = strings.Replace(stripped, "</AdaptationSet>",
		RepresentationPlaceholder+"\n    </AdaptationSet>", 1)
	return strings.Replace(string(raw), string(videoSet), stripped, 1), nil
}

// RepresentationPlaceholder marks where video representations are injected.
const RepresentationPlaceholder = "<!--ALCHEMIST_VIDEO_REPRESENTATIONS-->"

// ComposeMPD substitutes representations into a skeleton, in the given order.
func ComposeMPD(skeleton string, representations []string) string {
	return strings.Replace(skeleton, RepresentationPlaceholder,
		strings.Join(representations, "\n"), 1)
}
