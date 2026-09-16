package api

import "testing"

// Reclaiming media that another asset still plays is the expensive mistake here: the
// duplicate keeps reporting ready while every byte range 404s.
func TestReclaimForSkipsSharedMedia(t *testing.T) {
	src, mezz := "src/t/a/original", "mez/t/a/mezzanine.mp4"

	shared := reclaimFor("cmaf/t/a", &src, &mezz, false)
	if len(shared.Prefixes) != 0 {
		t.Errorf("shared media was queued for deletion: %v", shared.Prefixes)
	}
	if len(shared.Keys) != 1 || shared.Keys[0] != src {
		t.Errorf("the asset's own source should still go: %v", shared.Keys)
	}

	sole := reclaimFor("cmaf/t/a", &src, &mezz, true)
	if len(sole.Keys) != 2 || len(sole.Prefixes) != 1 || sole.Prefixes[0] != "cmaf/t/a/" {
		t.Errorf("the last asset holding the media should reclaim all of it: %+v", sole)
	}
}

// A URL import with no title takes the file name off the URL. Getting this wrong
// gives a library migrated in bulk several hundred rows called "/" or ".mp4".
func TestTitleFromPath(t *testing.T) {
	for in, want := range map[string]string{
		"/videos/Lecture%2001.mp4":    "Lecture 01",
		"/videos/lecture.final.mov":   "lecture.final",
		"lecture.mp4":                 "lecture",
		"/videos/lecture":             "lecture",
		"/videos/":                    "videos",
		"/":                           "",
		"":                            "",
		"/a/b/%E0%A6%AA%E0%A6%A1.mp4": "পড",
	} {
		if got := titleFromPath(in); got != want {
			t.Errorf("titleFromPath(%q) = %q, want %q", in, got, want)
		}
	}
}
