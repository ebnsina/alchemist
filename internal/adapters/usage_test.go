package adapters

import "testing"

// Draining twice must not bill the same bytes twice, and a failed write puts them
// back: both are money, and both are silent when they go wrong.
func TestUsageDrainAccumulatesThenResets(t *testing.T) {
	var e Usage
	e.RecordEgress("tenant-a", 100)
	e.RecordEgress("tenant-a", 50)
	e.RecordEgress("tenant-b", 7)

	got, _ := e.drain()
	if got["tenant-a"] != 150 || got["tenant-b"] != 7 {
		t.Errorf("bytes not accumulated per tenant: %v", got)
	}
	if again, _ := e.drain(); len(again) != 0 {
		t.Error("a second drain returned bytes that were already billed")
	}

	e.RecordEgress("tenant-a", 9)
	if got, _ := e.drain(); got["tenant-a"] != 9 {
		t.Error("counting did not resume after a drain")
	}
}

// The peak is a count of distinct viewers in one interval, not of requests. A player
// re-reads the playlist every few seconds, so counting requests would report one
// person watching a two-hour lecture as thousands of viewers and bill for them.
func TestUsageViewersCountedOncePerInterval(t *testing.T) {
	var e Usage
	for i := 0; i < 5; i++ {
		e.RecordViewer("tenant-a", "asset-1", "viewer-one")
	}
	e.RecordViewer("tenant-a", "asset-1", "viewer-two")
	e.RecordViewer("tenant-a", "asset-2", "viewer-one")

	_, peaks := e.drain()
	if got := peaks[assetKey{"tenant-a", "asset-1"}]; got != 2 {
		t.Errorf("asset-1 had 2 distinct viewers, counted %d", got)
	}
	// The same person watching two assets is present on both.
	if got := peaks[assetKey{"tenant-a", "asset-2"}]; got != 1 {
		t.Errorf("asset-2 had 1 viewer, counted %d", got)
	}

	// The interval resets, or a viewer who left would be counted forever.
	if _, again := e.drain(); len(again) != 0 {
		t.Error("viewers carried over into the next interval")
	}
}
