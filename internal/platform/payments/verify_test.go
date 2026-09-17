package payments

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"
)

// A forged "paid" is the whole reason these signatures exist. Everything below is the
// attack, not the happy path.

func stripeHeader(body []byte, secret string, at time.Time) string {
	ts := fmt.Sprintf("%d", at.Unix())
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + "."))
	mac.Write(body)
	return "t=" + ts + ",v1=" + hex.EncodeToString(mac.Sum(nil))
}

func TestStripeSignature(t *testing.T) {
	const secret = "whsec_test_secret_at_least_long_enough"
	body := []byte(`{"type":"payment_intent.succeeded","data":{"object":{"id":"pi_1"}}}`)
	now := time.Now()

	if !verifyStripeSignature(body, stripeHeader(body, secret, now), secret, now) {
		t.Fatal("a genuine signature was rejected")
	}

	// The body is signed, so a single edited byte must fail -- this is the forgery.
	tampered := []byte(`{"type":"payment_intent.succeeded","data":{"object":{"id":"pi_2"}}}`)
	if verifyStripeSignature(tampered, stripeHeader(body, secret, now), secret, now) {
		t.Error("an edited body verified against the original signature")
	}
	if verifyStripeSignature(body, stripeHeader(body, "whsec_someone_elses_secret_value", now), secret, now) {
		t.Error("a signature from another secret verified")
	}
	// Without a tolerance a captured signature is replayable forever.
	if verifyStripeSignature(body, stripeHeader(body, secret, now.Add(-time.Hour)), secret, now) {
		t.Error("an hour-old signature verified")
	}
	if verifyStripeSignature(body, stripeHeader(body, secret, now.Add(time.Hour)), secret, now) {
		t.Error("a signature from the future verified")
	}
	for _, h := range []string{"", "t=123", "v1=abc", "garbage", "t=notanumber,v1=abc"} {
		if verifyStripeSignature(body, h, secret, now) {
			t.Errorf("malformed header %q verified", h)
		}
	}
	// No secret configured must fail closed, never open.
	if verifyStripeSignature(body, stripeHeader(body, secret, now), "", now) {
		t.Error("an unconfigured webhook secret verified anything")
	}

	// Stripe sends two v1 signatures while a secret is being rolled; the old one must
	// still work or every webhook fails for the length of the roll.
	h := stripeHeader(body, secret, now)
	rolled := h + ",v1=" + strings.Repeat("0", 64)
	if !verifyStripeSignature(body, rolled, secret, now) {
		t.Error("a header carrying a second signature during a roll was rejected")
	}
}

func sslSign(form url.Values, storePass string) string {
	fields := strings.Split(form.Get("verify_key"), ",")
	sort.Strings(fields)
	var parts []string
	for _, f := range fields {
		if f = strings.TrimSpace(f); f != "" {
			parts = append(parts, f+"="+form.Get(f))
		}
	}
	sum := md5.Sum([]byte(storePass))
	parts = append(parts, "store_passwd="+hex.EncodeToString(sum[:]))
	sort.Strings(parts)
	got := md5.Sum([]byte(strings.Join(parts, "&")))
	return hex.EncodeToString(got[:])
}

func TestSSLCommerzIPNHash(t *testing.T) {
	const pass = "qwerty"
	form := url.Values{
		"tran_id":      {"1f3c9a2e-0000-0000-0000-000000000001"},
		"val_id":       {"1711231900331kHP17lnrr9T8Gt"},
		"amount":       {"1560.00"},
		"currency":     {"BDT"},
		"status":       {"VALID"},
		"store_id":     {"testbox"},
		"bank_tran_id": {"1711231900331S0R8atkhAZksmM"},
		"verify_key":   {"amount,bank_tran_id,currency,status,store_id,tran_id,val_id"},
	}
	form.Set("verify_sign", sslSign(form, pass))

	if !verifySSLCommerzHash(form, pass) {
		t.Fatal("a genuine IPN was rejected")
	}

	// Raising the amount after signing is the forgery worth testing.
	tampered := cloneForm(form)
	tampered.Set("amount", "1.00")
	if verifySSLCommerzHash(tampered, pass) {
		t.Error("an IPN with an edited amount verified")
	}
	// Claiming success on a failed payment is the other one.
	tampered = cloneForm(form)
	tampered.Set("status", "FAILED")
	if verifySSLCommerzHash(tampered, pass) {
		t.Error("an IPN with an edited status verified")
	}
	if verifySSLCommerzHash(form, "not-the-store-password") {
		t.Error("an IPN verified against the wrong store password")
	}

	missing := cloneForm(form)
	missing.Del("verify_sign")
	if verifySSLCommerzHash(missing, pass) {
		t.Error("an unsigned IPN verified")
	}
	if verifySSLCommerzHash(form, "") {
		t.Error("an unconfigured store password verified anything")
	}
}

func cloneForm(in url.Values) url.Values {
	out := url.Values{}
	for k, v := range in {
		out[k] = append([]string(nil), v...)
	}
	return out
}

// Amounts cross the gateway boundary as decimal strings. Routing them through a float
// is how 1560.00 paisa becomes 1559.
func TestMinorUnitsSurviveTheGateway(t *testing.T) {
	for _, c := range []struct {
		minor int64
		text  string
	}{
		{156000, "1560.00"},
		{1, "0.01"},
		{99, "0.99"},
		{100, "1.00"},
		{0, "0.00"},
		{123456789, "1234567.89"},
	} {
		if got := formatMinor(c.minor, 100); got != c.text {
			t.Errorf("formatMinor(%d) = %q, want %q", c.minor, got, c.text)
		}
		back, err := parseMinor(c.text, 100)
		if err != nil || back != c.minor {
			t.Errorf("parseMinor(%q) = %d, %v; want %d", c.text, back, err, c.minor)
		}
	}
	// A gateway that answers without decimals must not lose a hundred times the money.
	if got, err := parseMinor("1560", 100); err != nil || got != 156000 {
		t.Errorf(`parseMinor("1560") = %d, %v; want 156000`, got, err)
	}
}
