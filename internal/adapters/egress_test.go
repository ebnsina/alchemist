package adapters

import "testing"

// Draining twice must not bill the same bytes twice, and a failed write puts them
// back: both are money, and both are silent when they go wrong.
func TestEgressDrainAccumulatesThenResets(t *testing.T) {
	var e Egress
	e.RecordEgress("tenant-a", 100)
	e.RecordEgress("tenant-a", 50)
	e.RecordEgress("tenant-b", 7)

	got := e.drain()
	if got["tenant-a"] != 150 || got["tenant-b"] != 7 {
		t.Errorf("bytes not accumulated per tenant: %v", got)
	}
	if len(e.drain()) != 0 {
		t.Error("a second drain returned bytes that were already billed")
	}

	e.RecordEgress("tenant-a", 9)
	if e.drain()["tenant-a"] != 9 {
		t.Error("counting did not resume after a drain")
	}
}
