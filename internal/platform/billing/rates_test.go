package billing

import (
	"math/big"
	"testing"
)

func rat(s string) *big.Rat {
	v, ok := new(big.Rat).SetString(s)
	if !ok {
		panic(s)
	}
	return v
}

// The published BD card, against a month of real-shaped usage. These numbers are the
// product's price list: if this test changes, somebody's bill changed.
func TestChargeAgainstThePublishedCard(t *testing.T) {
	rates, err := ParseRates([]byte(`[
	  {"kind":"ingest",  "unit":"seconds",  "per":60,         "amount":2.00},
	  {"kind":"storage", "unit":"gb_hours", "per":720,        "amount":1.20},
	  {"kind":"egress",  "unit":"bytes",    "per":1000000000, "amount":1.20}
	]`))
	if err != nil {
		t.Fatal(err)
	}

	lines, total, err := Charge([]Usage{
		// 10 hours of video in: 600 minutes x ৳2 = ৳1200 = 120000 paisa.
		{Kind: "ingest", Unit: "seconds", Quantity: rat("36000")},
		// 50 GB held for a 30-day month: 50 x 720 gb_hours / 720 x ৳1.2 = ৳60.
		{Kind: "storage", Unit: "gb_hours", Quantity: rat("36000")},
		// 250 GB watched x ৳1.2 = ৳300.
		{Kind: "egress", Unit: "bytes", Quantity: rat("250000000000")},
		// Metered, and this card does not charge for it. Not an error.
		{Kind: "peak_viewers", Unit: "viewers", Quantity: rat("410")},
	}, rates, "BDT")
	if err != nil {
		t.Fatal(err)
	}

	if len(lines) != 3 {
		t.Fatalf("want a line per charged kind, got %d", len(lines))
	}
	want := map[string]int64{"ingest": 120000, "storage": 6000, "egress": 30000}
	for _, l := range lines {
		if want[l.Kind] != l.AmountMinor {
			t.Errorf("%s = %d paisa, want %d", l.Kind, l.AmountMinor, want[l.Kind])
		}
	}
	if total != 156000 {
		t.Fatalf("total = %d paisa (৳%.2f), want 156000 (৳1560.00)", total, float64(total)/100)
	}
}

// 0.016 has no exact float64. Parsing the card through a float and billing a year of
// traffic against it is how a bill drifts from the price list by real money.
func TestRatesAreExactNotFloats(t *testing.T) {
	rates, err := ParseRates([]byte(`[{"kind":"egress","unit":"bytes","per":1000000000,"amount":0.016}]`))
	if err != nil {
		t.Fatal(err)
	}
	if rates[0].Amount != "0.016" {
		t.Fatalf("amount came back as %q, want the literal 0.016", rates[0].Amount)
	}

	// A petabyte at $0.016/GB is exactly $16,000.00 -- 1600000 cents, no remainder.
	_, total, err := Charge([]Usage{
		{Kind: "egress", Unit: "bytes", Quantity: rat("1000000000000000")},
	}, rates, "USD")
	if err != nil {
		t.Fatal(err)
	}
	if total != 1600000 {
		t.Fatalf("total = %d cents, want 1600000", total)
	}
}

// Rounding once at the line, not per event: a thousand small egress rows each rounding
// up would bill for traffic nobody served.
func TestRoundingIsHalfUpAndHappensOnce(t *testing.T) {
	rates, _ := ParseRates([]byte(`[{"kind":"egress","unit":"bytes","per":1000000000,"amount":1.20}]`))

	for _, c := range []struct {
		bytes string
		want  int64
	}{
		{"1000000000", 120}, // exactly 1 GB
		{"4166666", 0},      // rounds down to nothing
		{"4166667", 1},      // the halfway point rounds up
		{"0", 0},
	} {
		_, got, err := Charge([]Usage{{Kind: "egress", Unit: "bytes", Quantity: rat(c.bytes)}}, rates, "BDT")
		if err != nil {
			t.Fatal(err)
		}
		if got != c.want {
			t.Errorf("%s bytes = %d paisa, want %d", c.bytes, got, c.want)
		}
	}
}

// Nothing used is nothing owed, and an invoice for zero must not be sent to a gateway.
func TestNoUsageIsNoCharge(t *testing.T) {
	rates, _ := ParseRates([]byte(`[{"kind":"ingest","unit":"seconds","per":60,"amount":2}]`))
	lines, total, err := Charge([]Usage{{Kind: "ingest", Unit: "seconds", Quantity: rat("0")}}, rates, "BDT")
	if err != nil || total != 0 || len(lines) != 0 {
		t.Fatalf("lines=%v total=%d err=%v; want no lines and no charge", lines, total, err)
	}
}

func TestBadCardIsRefusedNotGuessed(t *testing.T) {
	if _, err := ParseRates([]byte(`[]`)); err == nil {
		t.Error("an empty card must not price anything")
	}
	if _, err := ParseRates([]byte(`[{"unit":"seconds"}]`)); err == nil {
		t.Error("a rate with no kind must be refused")
	}
	rates := []Rate{{Kind: "ingest", Unit: "seconds", Per: "0", Amount: "2"}}
	if _, _, err := Charge([]Usage{{Kind: "ingest", Quantity: rat("60")}}, rates, "BDT"); err == nil {
		t.Error("dividing by a per of zero must error, not panic or bill nonsense")
	}
}
