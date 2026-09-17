// Package payments collects money for an invoice.
//
// Two gateways, and they are not the same shape, so the interface does not pretend
// they are. Stripe can charge a saved card with nobody present. SSLCommerz's published
// API is hosted checkout: the customer opens a link and pays. Both can be asked for a
// checkout URL and both report back over a webhook; only one can be told to charge.
package payments

import (
	"context"
	"errors"
	"net/http"
)

// ErrNoSavedMethod means there is nothing to charge without the customer, so the
// invoice has to be collected by sending them a link instead.
var ErrNoSavedMethod = errors.New("no saved payment method")

// ErrUnverified is a webhook that did not authenticate. It is never a 500: an attacker
// posting a forged "paid" is exactly what the signature is there to stop, and a 500
// invites the gateway to retry a request that will never be accepted.
var ErrUnverified = errors.New("webhook signature did not verify")

type Invoice struct {
	ID          string
	TenantID    string
	AmountMinor int64
	Currency    string
	Description string
}

type Customer struct {
	Email string
	Name  string
	Phone string
	// Ref is the gateway's own id for this customer where it keeps one.
	Ref string
}

// Attempt is what a gateway did when asked for money.
type Attempt struct {
	// Ref identifies this attempt at the gateway. It is the idempotency key on the
	// way back: a webhook delivered twice must credit an invoice once.
	Ref string
	// URL is set when the customer has to go and pay; empty when it was charged.
	URL       string
	Succeeded bool
	Error     string
}

// Event is what a verified webhook says happened.
type Event struct {
	// Ref matches Attempt.Ref, which is how the payment finds its invoice.
	Ref         string
	InvoiceID   string
	Succeeded   bool
	Failed      bool
	AmountMinor int64
	Currency    string
	Error       string
	// Ignored is a well-formed, genuinely-from-the-gateway event about something this
	// system does not act on. Answered 200 so the gateway stops redelivering it.
	Ignored bool
}

type Gateway interface {
	Name() string
	// Checkout returns somewhere the customer can pay this invoice.
	Checkout(ctx context.Context, inv Invoice, cus Customer, returnTo string) (Attempt, error)
	// Verify authenticates an incoming callback and says what it reports.
	Verify(ctx context.Context, r *http.Request) (Event, error)
}

// AutoCharger is a gateway that can take money with the customer absent. Stripe
// implements it; SSLCommerz does not, and an invoice on a gateway that does not is
// collected by emailing a link rather than by silently failing every month.
type AutoCharger interface {
	Charge(ctx context.Context, inv Invoice, methodRef, customerRef string) (Attempt, error)
}
