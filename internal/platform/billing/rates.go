// Package billing turns metered usage into an amount of money.
//
// Money is integer minor units everywhere it is stored or charged -- paisa, cents.
// Rates are not: they are fractions of a paisa per megabyte, so they stay decimal and
// only the line total rounds. A float anywhere in this path is how a cent goes missing
// and never comes back.
package billing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
)

// Rate is one line a customer can be charged for. `Per` is how many units the amount
// buys -- 60 seconds of ingest, a gigabyte of egress -- so a rate reads the way it is
// quoted rather than as an unreadable per-byte fraction.
type Rate struct {
	Kind   string `json:"kind"`
	Unit   string `json:"unit"`
	Per    string `json:"per"`
	Amount string `json:"amount"`
}

// rateJSON accepts the numbers as JSON numbers and keeps them exact by carrying the
// literal text, never a float64. encoding/json would otherwise parse 0.016 into the
// nearest binary double and bill the difference forever.
type rateJSON struct {
	Kind   string      `json:"kind"`
	Unit   string      `json:"unit"`
	Per    json.Number `json:"per"`
	Amount json.Number `json:"amount"`
}

func ParseRates(raw []byte) ([]Rate, error) {
	var in []rateJSON
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	if err := d.Decode(&in); err != nil {
		return nil, fmt.Errorf("parse rate card: %w", err)
	}
	if len(in) == 0 {
		return nil, fmt.Errorf("rate card has no rates")
	}
	out := make([]Rate, len(in))
	for i, r := range in {
		if r.Kind == "" || r.Unit == "" {
			return nil, fmt.Errorf("rate %d has no kind or unit", i)
		}
		out[i] = Rate{Kind: r.Kind, Unit: r.Unit, Per: r.Per.String(), Amount: r.Amount.String()}
	}
	return out, nil
}

// Line is one charge on an invoice, with the arithmetic that produced it kept beside
// the answer. A customer asking "why is this ৳412" gets the quantity, the rate and the
// unit rather than a number to take on trust.
type Line struct {
	Kind        string
	Unit        string
	Quantity    *big.Rat
	UnitAmount  *big.Rat
	Per         *big.Rat
	AmountMinor int64
}

// Usage is what was metered for one kind over the period, in that kind's own unit.
type Usage struct {
	Kind     string
	Unit     string
	Quantity *big.Rat
}

// MinorUnits is how many minor units make one of the currency. Both currencies here
// are hundredths; the constant exists so a zero-decimal currency is a table entry
// rather than a rewrite.
func MinorUnits(currency string) int64 {
	switch currency {
	case "JPY", "KRW":
		return 1
	default:
		return 100
	}
}

// Charge prices usage against a rate card.
//
// A kind with no matching rate is not billed and not an error: a rate card that omits
// peak_viewers is a card that does not charge for it, which is a pricing decision, not
// a broken row.
//
// Rounding is half-up at the line, once. Rounding per-event would let a thousand small
// egress rows each round up and bill for traffic nobody served.
func Charge(usage []Usage, rates []Rate, currency string) ([]Line, int64, error) {
	minor := big.NewRat(MinorUnits(currency), 1)
	var lines []Line
	var total int64

	for _, u := range usage {
		r, ok := findRate(rates, u.Kind)
		if !ok {
			continue
		}
		amount, ok := new(big.Rat).SetString(r.Amount)
		if !ok {
			return nil, 0, fmt.Errorf("rate %s has an unreadable amount %q", r.Kind, r.Amount)
		}
		per, ok := new(big.Rat).SetString(r.Per)
		if !ok || per.Sign() == 0 {
			return nil, 0, fmt.Errorf("rate %s has an unusable per %q", r.Kind, r.Per)
		}
		if u.Quantity == nil || u.Quantity.Sign() <= 0 {
			continue
		}

		// quantity / per * amount * minorUnits, all exact until the final round.
		v := new(big.Rat).Quo(u.Quantity, per)
		v.Mul(v, amount)
		v.Mul(v, minor)

		cents := roundHalfUp(v)
		lines = append(lines, Line{
			Kind: u.Kind, Unit: u.Unit, Quantity: u.Quantity,
			UnitAmount: amount, Per: per, AmountMinor: cents,
		})
		total += cents
	}
	return lines, total, nil
}

func findRate(rates []Rate, kind string) (Rate, bool) {
	for _, r := range rates {
		if r.Kind == kind {
			return r, true
		}
	}
	return Rate{}, false
}

// roundHalfUp rounds a non-negative rational to the nearest integer, halves upward.
// big.Rat has no rounding of its own, and doing it via float64 would reintroduce
// exactly the imprecision the rationals are here to avoid.
func roundHalfUp(v *big.Rat) int64 {
	num, den := v.Num(), v.Denom()
	// (2*num + den) / (2*den), floored, is the half-up round for a non-negative value.
	twice := new(big.Int).Mul(num, big.NewInt(2))
	twice.Add(twice, den)
	twiceDen := new(big.Int).Mul(den, big.NewInt(2))
	q := new(big.Int).Div(twice, twiceDen)
	return q.Int64()
}
