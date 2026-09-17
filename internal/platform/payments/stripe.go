package payments

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Stripe, over plain HTTPS rather than the SDK. Four endpoints do not justify a
// dependency that ships the whole API surface, and the one security-critical piece --
// webhook signature verification -- is twenty lines either way.
type Stripe struct {
	SecretKey     string
	WebhookSecret string
	HTTP          *http.Client
	// BaseURL is overridden by tests. Nothing else should set it.
	BaseURL string
}

func NewStripe(secretKey, webhookSecret string) *Stripe {
	return &Stripe{
		SecretKey:     secretKey,
		WebhookSecret: webhookSecret,
		HTTP:          &http.Client{Timeout: 30 * time.Second},
		BaseURL:       "https://api.stripe.com",
	}
}

func (s *Stripe) Name() string { return "stripe" }

// Charge takes the money with the customer absent, which is what a monthly invoice is.
//
// off_session says so to the card network: it changes how an issuer treats the
// transaction and is what keeps a recurring charge from being declined as a suspicious
// cardholder-absent one. confirm=true attempts it now rather than leaving it pending.
func (s *Stripe) Charge(ctx context.Context, inv Invoice, methodRef, customerRef string) (Attempt, error) {
	if methodRef == "" {
		return Attempt{}, ErrNoSavedMethod
	}
	form := url.Values{
		"amount":         {strconv.FormatInt(inv.AmountMinor, 10)},
		"currency":       {strings.ToLower(inv.Currency)},
		"payment_method": {methodRef},
		"confirm":        {"true"},
		"off_session":    {"true"},
		"description":    {inv.Description},
		// The invoice id, so a webhook arriving without our reference still finds it.
		"metadata[invoice_id]": {inv.ID},
		"metadata[tenant_id]":  {inv.TenantID},
	}
	if customerRef != "" {
		form.Set("customer", customerRef)
	}

	var out struct {
		ID           string                   `json:"id"`
		Status       string                   `json:"status"`
		LastError    struct{ Message string } `json:"last_payment_error"`
		ErrorWrapper struct {
			Message string `json:"message"`
			Code    string `json:"code"`
		} `json:"error"`
	}
	// An idempotency key per invoice attempt: a retried job after a timeout must not
	// charge a customer twice for one invoice.
	key := fmt.Sprintf("alchemist-invoice-%s", inv.ID)
	if err := s.post(ctx, "/v1/payment_intents", form, key, &out); err != nil {
		return Attempt{}, err
	}
	if out.ErrorWrapper.Message != "" {
		return Attempt{Ref: out.ID, Error: out.ErrorWrapper.Message}, nil
	}
	att := Attempt{Ref: out.ID, Succeeded: out.Status == "succeeded"}
	if !att.Succeeded {
		att.Error = out.LastError.Message
		if att.Error == "" {
			att.Error = "card was not charged: " + out.Status
		}
	}
	return att, nil
}

// Checkout is the fallback when there is no saved card: a hosted page that both
// collects this invoice and saves the card, so next month can be charged without
// asking again.
func (s *Stripe) Checkout(ctx context.Context, inv Invoice, cus Customer, returnTo string) (Attempt, error) {
	form := url.Values{
		"mode":                                   {"payment"},
		"success_url":                            {returnTo + "?paid=1"},
		"cancel_url":                             {returnTo},
		"line_items[0][price_data][currency]":    {strings.ToLower(inv.Currency)},
		"line_items[0][price_data][unit_amount]": {strconv.FormatInt(inv.AmountMinor, 10)},
		"line_items[0][price_data][product_data][name]": {inv.Description},
		"line_items[0][quantity]":                       {"1"},
		"payment_intent_data[setup_future_usage]":       {"off_session"},
		"payment_intent_data[metadata][invoice_id]":     {inv.ID},
		"payment_intent_data[metadata][tenant_id]":      {inv.TenantID},
		"metadata[invoice_id]":                          {inv.ID},
	}
	if cus.Email != "" {
		form.Set("customer_email", cus.Email)
	}
	if cus.Ref != "" {
		form.Del("customer_email")
		form.Set("customer", cus.Ref)
	}

	var out struct {
		ID           string                   `json:"id"`
		URL          string                   `json:"url"`
		ErrorWrapper struct{ Message string } `json:"error"`
	}
	if err := s.post(ctx, "/v1/checkout/sessions", form, "", &out); err != nil {
		return Attempt{}, err
	}
	if out.ErrorWrapper.Message != "" {
		return Attempt{}, fmt.Errorf("stripe checkout: %s", out.ErrorWrapper.Message)
	}
	return Attempt{Ref: out.ID, URL: out.URL}, nil
}

// stripeTolerance bounds how old a signed webhook may be. Without it a signature
// captured once is replayable forever.
const stripeTolerance = 5 * time.Minute

func (s *Stripe) Verify(ctx context.Context, r *http.Request) (Event, error) {
	body, err := io.ReadAll(http.MaxBytesReader(nil, r.Body, 1<<20))
	if err != nil {
		return Event{}, ErrUnverified
	}
	if !verifyStripeSignature(body, r.Header.Get("Stripe-Signature"), s.WebhookSecret, time.Now()) {
		return Event{}, ErrUnverified
	}

	var ev struct {
		Type string `json:"type"`
		Data struct {
			Object struct {
				ID       string                   `json:"id"`
				Amount   int64                    `json:"amount"`
				Currency string                   `json:"currency"`
				Metadata map[string]string        `json:"metadata"`
				LastErr  struct{ Message string } `json:"last_payment_error"`
				// Checkout sessions carry the intent rather than being one.
				PaymentIntent string `json:"payment_intent"`
			} `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &ev); err != nil {
		return Event{}, ErrUnverified
	}

	o := ev.Data.Object
	ref := o.ID
	if o.PaymentIntent != "" {
		ref = o.PaymentIntent
	}
	out := Event{
		Ref: ref, InvoiceID: o.Metadata["invoice_id"],
		AmountMinor: o.Amount, Currency: strings.ToUpper(o.Currency),
	}
	switch ev.Type {
	case "payment_intent.succeeded", "checkout.session.completed":
		out.Succeeded = true
	case "payment_intent.payment_failed":
		out.Failed = true
		out.Error = o.LastErr.Message
	default:
		// Genuinely from Stripe, and about something this does not act on. Answered
		// 200 rather than refused, or Stripe redelivers it for days.
		out.Ignored = true
	}
	return out, nil
}

// verifyStripeSignature checks the t=/v1= header Stripe sends.
//
// The signed payload is the timestamp, a dot, and the exact bytes of the body -- which
// is why the body is read raw and never re-marshalled. Every v1 signature is tried
// because Stripe sends two during a webhook secret roll.
func verifyStripeSignature(body []byte, header, secret string, now time.Time) bool {
	if header == "" || secret == "" {
		return false
	}
	var ts string
	var sigs []string
	for _, part := range strings.Split(header, ",") {
		k, v, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		switch k {
		case "t":
			ts = v
		case "v1":
			sigs = append(sigs, v)
		}
	}
	if ts == "" || len(sigs) == 0 {
		return false
	}
	secs, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return false
	}
	age := now.Sub(time.Unix(secs, 0))
	if age > stripeTolerance || age < -stripeTolerance {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts))
	mac.Write([]byte("."))
	mac.Write(body)
	want := hex.EncodeToString(mac.Sum(nil))
	for _, got := range sigs {
		if hmac.Equal([]byte(got), []byte(want)) {
			return true
		}
	}
	return false
}

func (s *Stripe) post(ctx context.Context, path string, form url.Values, idempotencyKey string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		s.BaseURL+path, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(s.SecretKey, "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}

	res, err := s.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("stripe %s: %w", path, err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return err
	}
	// A declined card is a 402 with a body worth reading, not a transport failure, so
	// the status is not checked before decoding.
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("stripe %s: unreadable response (%d)", path, res.StatusCode)
	}
	return nil
}
