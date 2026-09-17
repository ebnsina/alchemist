package payments

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// SSLCommerz. Hosted checkout only, deliberately.
//
// Its published v4 API is initiate-then-redirect: there is no merchant-initiated
// charge against a stored card in it, so this does not implement AutoCharger and a BDT
// invoice is collected by sending the customer a link. Recurring against a saved card
// needs a separate agreement with SSLCommerz; when there is one, this is where it goes.
type SSLCommerz struct {
	StoreID   string
	StorePass string
	// Sandbox switches both the checkout and validation hosts. Their sandbox is a
	// different domain, not a flag, so one bool decides both or they disagree.
	Sandbox bool
	HTTP    *http.Client
	BaseURL string
}

func NewSSLCommerz(storeID, storePass string, sandbox bool) *SSLCommerz {
	base := "https://securepay.sslcommerz.com"
	if sandbox {
		base = "https://sandbox.sslcommerz.com"
	}
	return &SSLCommerz{
		StoreID: storeID, StorePass: storePass, Sandbox: sandbox,
		HTTP: &http.Client{Timeout: 30 * time.Second}, BaseURL: base,
	}
}

func (g *SSLCommerz) Name() string { return "sslcommerz" }

// Checkout asks for a payment page and returns where to send the customer.
//
// tran_id is ours and is the invoice id, so the IPN that comes back names the invoice
// without a lookup table in between.
func (g *SSLCommerz) Checkout(ctx context.Context, inv Invoice, cus Customer, returnTo string) (Attempt, error) {
	// Their API takes an amount in whole currency units, not minor ones. Formatting it
	// from the integer keeps the paisa exact instead of routing it through a float.
	amount := formatMinor(inv.AmountMinor, MinorUnitsFor(inv.Currency))

	form := url.Values{
		"store_id":         {g.StoreID},
		"store_passwd":     {g.StorePass},
		"total_amount":     {amount},
		"currency":         {inv.Currency},
		"tran_id":          {inv.ID},
		"success_url":      {returnTo + "?paid=1"},
		"fail_url":         {returnTo + "?paid=0"},
		"cancel_url":       {returnTo},
		"ipn_url":          {returnTo + "/ipn"},
		"product_name":     {inv.Description},
		"product_profile":  {"non-physical-goods"},
		"product_category": {"service"},
		"shipping_method":  {"NO"},
		"cus_name":         {orDefault(cus.Name, "Customer")},
		"cus_email":        {orDefault(cus.Email, "billing@example.com")},
		"cus_add1":         {"N/A"},
		"cus_city":         {"Dhaka"},
		"cus_postcode":     {"1000"},
		"cus_country":      {"Bangladesh"},
		"cus_phone":        {orDefault(cus.Phone, "01700000000")},
		// Carried through the gateway and handed back on the IPN untouched.
		"value_a": {inv.TenantID},
	}

	var out struct {
		Status         string `json:"status"`
		FailedReason   string `json:"failedreason"`
		SessionKey     string `json:"sessionkey"`
		GatewayPageURL string `json:"GatewayPageURL"`
	}
	if err := g.post(ctx, "/gwprocess/v4/api.php", form, &out); err != nil {
		return Attempt{}, err
	}
	if !strings.EqualFold(out.Status, "SUCCESS") || out.GatewayPageURL == "" {
		return Attempt{}, fmt.Errorf("sslcommerz refused the session: %s %s",
			out.Status, out.FailedReason)
	}
	return Attempt{Ref: out.SessionKey, URL: out.GatewayPageURL}, nil
}

// Verify authenticates an IPN and then asks SSLCommerz what actually happened.
//
// Both steps, not one. The hash proves the POST came from them and was not edited; it
// does not prove the payment succeeded, because the fields it covers are supplied in
// the same request. The validation API is the authoritative answer and is the only
// thing an invoice is marked paid on.
func (g *SSLCommerz) Verify(ctx context.Context, r *http.Request) (Event, error) {
	if err := r.ParseForm(); err != nil {
		return Event{}, ErrUnverified
	}
	if !verifySSLCommerzHash(r.PostForm, g.StorePass) {
		return Event{}, ErrUnverified
	}

	valID := r.PostForm.Get("val_id")
	invoiceID := r.PostForm.Get("tran_id")
	if valID == "" || invoiceID == "" {
		return Event{}, ErrUnverified
	}

	var v struct {
		Status   string `json:"status"`
		TranID   string `json:"tran_id"`
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
		BankTran string `json:"bank_tran_id"`
	}
	q := url.Values{
		"val_id": {valID}, "store_id": {g.StoreID}, "store_passwd": {g.StorePass},
		"v": {"1"}, "format": {"json"},
	}
	if err := g.get(ctx, "/validator/api/validationserverAPI.php", q, &v); err != nil {
		return Event{}, err
	}

	// Their id for the transaction, not ours: the IPN and the redirect both describe
	// one payment, and crediting it twice is what the unique key on it prevents.
	ref := v.BankTran
	if ref == "" {
		ref = valID
	}
	ev := Event{Ref: ref, InvoiceID: invoiceID, Currency: strings.ToUpper(v.Currency)}
	if minor, err := parseMinor(v.Amount, MinorUnitsFor(ev.Currency)); err == nil {
		ev.AmountMinor = minor
	}

	// VALIDATED is a card payment that settled; VALID is the same for other methods.
	switch strings.ToUpper(v.Status) {
	case "VALID", "VALIDATED":
		ev.Succeeded = true
	case "FAILED", "UNATTEMPTED", "EXPIRED", "CANCELLED":
		ev.Failed = true
		ev.Error = "payment " + strings.ToLower(v.Status)
	default:
		ev.Ignored = true
	}
	return ev, nil
}

// verifySSLCommerzHash reproduces the verify_sign on an IPN.
//
// verify_key names the fields that were signed, in the order they must be joined after
// sorting; the store password's md5 is appended as the secret. Reading the field list
// off the request rather than hardcoding it is how this survives them adding a field.
func verifySSLCommerzHash(form url.Values, storePass string) bool {
	sign := form.Get("verify_sign")
	keys := form.Get("verify_key")
	if sign == "" || keys == "" || storePass == "" {
		return false
	}

	fields := strings.Split(keys, ",")
	sort.Strings(fields)
	parts := make([]string, 0, len(fields)+1)
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		parts = append(parts, f+"="+form.Get(f))
	}
	sum := md5.Sum([]byte(storePass))
	parts = append(parts, "store_passwd="+hex.EncodeToString(sum[:]))
	sort.Strings(parts)

	got := md5.Sum([]byte(strings.Join(parts, "&")))
	return hmac.Equal([]byte(strings.ToLower(sign)), []byte(hex.EncodeToString(got[:])))
}

// MinorUnitsFor mirrors billing.MinorUnits without importing it: payments must not
// depend on pricing, and the number is the currency's, not either package's.
func MinorUnitsFor(currency string) int64 {
	switch strings.ToUpper(currency) {
	case "JPY", "KRW":
		return 1
	default:
		return 100
	}
}

// formatMinor renders minor units as the decimal string a gateway expects, by integer
// arithmetic. Dividing by 100.0 first would hand them a float's idea of the amount.
func formatMinor(minor, units int64) string {
	if units <= 1 {
		return strconv.FormatInt(minor, 10)
	}
	whole := minor / units
	frac := minor % units
	if frac < 0 {
		frac = -frac
	}
	return fmt.Sprintf("%d.%02d", whole, frac)
}

func parseMinor(s string, units int64) (int64, error) {
	whole, frac, _ := strings.Cut(strings.TrimSpace(s), ".")
	w, err := strconv.ParseInt(whole, 10, 64)
	if err != nil {
		return 0, err
	}
	frac = (frac + "00")[:2]
	f, err := strconv.ParseInt(frac, 10, 64)
	if err != nil {
		return 0, err
	}
	if units <= 1 {
		return w, nil
	}
	return w*units + f, nil
}

func orDefault(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func (g *SSLCommerz) post(ctx context.Context, path string, form url.Values, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		g.BaseURL+path, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return g.do(req, out)
}

func (g *SSLCommerz) get(ctx context.Context, path string, q url.Values, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		g.BaseURL+path+"?"+q.Encode(), nil)
	if err != nil {
		return err
	}
	return g.do(req, out)
}

func (g *SSLCommerz) do(req *http.Request, out any) error {
	res, err := g.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("sslcommerz %s: %w", req.URL.Path, err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("sslcommerz %s: unreadable response (%d)", req.URL.Path, res.StatusCode)
	}
	return nil
}
