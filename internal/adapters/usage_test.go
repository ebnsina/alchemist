package adapters

import (
	"testing"
	"time"
)

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

// The device cap is the answer to one login shared with a class, so the rules that
// matter are: a device already watching keeps watching, a new one past the cap is
// refused rather than the incumbent being kicked, and a slot comes back when someone
// stops.
func TestDeviceCap(t *testing.T) {
	var e Usage
	k := bindingKey{"tenant-a", "student-1"}
	now := time.Now()

	if !e.allowDevice(k, "phone", 2, now) || !e.allowDevice(k, "laptop", 2, now) {
		t.Fatal("the first two devices were refused")
	}
	// A device already counted is not a new stream: a playlist re-read every few
	// seconds must not consume the cap over and over.
	if !e.allowDevice(k, "phone", 2, now.Add(10*time.Second)) {
		t.Error("a device already watching was refused on its next poll")
	}
	if e.allowDevice(k, "friend", 2, now.Add(11*time.Second)) {
		t.Error("a third device was admitted past a cap of two")
	}
	// The incumbents are untouched: refusing the newcomer is the whole point.
	if !e.allowDevice(k, "phone", 2, now.Add(12*time.Second)) {
		t.Error("an existing device lost its slot to the one that was refused")
	}
	// Another viewer has their own allowance.
	if !e.allowDevice(bindingKey{"tenant-a", "student-2"}, "friend", 2, now) {
		t.Error("one viewer's cap applied to a different viewer")
	}
	// Stop watching and the slot returns, or a closed tab would lock a student out
	// until the process restarted.
	if !e.allowDevice(k, "friend", 2, now.Add(deviceWindow+time.Minute)) {
		t.Error("a slot never came back after both devices went quiet")
	}
}
