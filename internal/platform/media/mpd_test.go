package media

import (
	"strings"
	"testing"
)

// Both fixtures are real shaka-packager output. They differ in a way that is easy to
// miss and silently breaks manifest composition: with several renditions the packager
// keeps width/height on each Representation, but with exactly one it hoists them onto
// the AdaptationSet. Packaging a single deferred rung produces the second shape, so a
// parser written against only the first finds nothing and drops the rendition from
// the manifest without erroring.
func TestVideoRepresentationsBothShapes(t *testing.T) {
	multi, err := VideoRepresentations("testdata/multi_rendition.mpd")
	if err != nil {
		t.Fatalf("multi-rendition: %v", err)
	}
	if len(multi) < 2 {
		t.Fatalf("multi-rendition found %d representations, want several", len(multi))
	}
	for h, rep := range multi {
		if h == 0 {
			t.Error("a representation was keyed by height 0")
		}
		if !strings.Contains(rep, "<BaseURL>") {
			t.Errorf("height %d: representation lost its BaseURL", h)
		}
	}

	single, err := VideoRepresentations("testdata/single_rendition.mpd")
	if err != nil {
		t.Fatalf("single-rendition: %v", err)
	}
	if len(single) != 1 {
		t.Fatalf("single-rendition found %d representations, want 1", len(single))
	}
	if _, ok := single[480]; !ok {
		t.Errorf("single-rendition keyed by %v, want height 480 from the AdaptationSet", keys(single))
	}
}

func TestMPDSkeletonAndCompose(t *testing.T) {
	skeleton, err := MPDSkeleton("testdata/multi_rendition.mpd")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(skeleton, RepresentationPlaceholder) {
		t.Fatal("skeleton has no placeholder to inject into")
	}
	// Only the video representations are stripped. The audio set keeps its own, which
	// is why this checks inside the video AdaptationSet rather than the whole document.
	videoSet := videoSetRE.FindString(skeleton)
	if videoSet == "" {
		t.Fatal("skeleton lost its video AdaptationSet")
	}
	if strings.Contains(videoSet, "<Representation ") {
		t.Error("video representations survived into the skeleton; they would be duplicated")
	}
	// The audio adaptation set must survive, or playback loses its audio track.
	if !strings.Contains(skeleton, `contentType="audio"`) {
		t.Error("skeleton dropped the audio AdaptationSet")
	}

	reps, err := VideoRepresentations("testdata/multi_rendition.mpd")
	if err != nil {
		t.Fatal(err)
	}
	var all []string
	for _, r := range reps {
		all = append(all, r)
	}
	out := ComposeMPD(skeleton, all)

	if strings.Contains(out, RepresentationPlaceholder) {
		t.Error("placeholder survived composition")
	}
	composedVideo := videoSetRE.FindString(out)
	if got := strings.Count(composedVideo, "<Representation "); got != len(all) {
		t.Errorf("composed video set has %d representations, want %d", got, len(all))
	}
	if !strings.Contains(out, `contentType="audio"`) {
		t.Error("composition dropped the audio AdaptationSet")
	}
}

func TestMissingFile(t *testing.T) {
	if _, err := VideoRepresentations("testdata/does-not-exist.mpd"); err == nil {
		t.Error("a missing manifest did not error")
	}
}

func keys(m map[int]string) []int {
	out := make([]int, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
