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
