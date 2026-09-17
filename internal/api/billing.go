package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/ebnsina/alchemist/internal/pipeline"
	"github.com/ebnsina/alchemist/internal/platform/httpx"
	"github.com/ebnsina/alchemist/internal/platform/payments"
)

type invoiceLine struct {
	Kind        string `json:"kind"`
	Unit        string `json:"unit"`
	Quantity    string `json:"quantity"`
	UnitAmount  string `json:"unit_amount"`
	Per         string `json:"per"`
	AmountMinor int64  `json:"amount_minor"`
}

type invoiceResponse struct {
	ID          string        `json:"id"`
	PeriodStart string        `json:"period_start"`
	PeriodEnd   string        `json:"period_end"`
	Currency    string        `json:"currency"`
	TotalMinor  int64         `json:"total_minor"`
	Status      string        `json:"status"`
	Attempts    int           `json:"attempts"`
	LastError   *string       `json:"last_error,omitempty"`
	IssuedAt    string        `json:"issued_at"`
	PaidAt      *string       `json:"paid_at,omitempty"`
	Lines       []invoiceLine `json:"lines"`
}

// getBilling is the whole billing picture in one call: what the account owes, what it
// is on, and what happens next. A customer chasing a suspension should not have to
// assemble that from three endpoints.
func (s *Server) getBilling(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	var status, currency, card string
	var email *string
	var outstanding int64
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(r.Context(),
			`select t.billing_status, t.billing_email, coalesce(t.rate_card, 'bd-standard'),
			        c.currency
			   from tenants t
			   join rate_cards c on c.name = coalesce(t.rate_card, 'bd-standard')`).
			Scan(&status, &email, &card, &currency); err != nil {
			return err
		}
		return tx.QueryRow(r.Context(),
			`select coalesce(sum(total_minor), 0) from invoices
			  where status in ('open','failed')`).Scan(&outstanding)
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}

	out := map[string]any{
		"status":            status,
		"currency":          currency,
		"rate_card":         card,
		"gateway":           pipeline.GatewayFor(currency),
		"outstanding_minor": outstanding,
		"billing_email":     email,
		// Said rather than implied: a suspended account's videos keep playing, and
		// somebody deciding whether to panic needs to know that.
		"playback_continues": true,
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) listInvoices(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	page, ok := httpx.ParseList(w, r, []string{"issued_at", "total_minor", "status"}, "issued_at")
	if !ok {
		return
	}

	out := []invoiceResponse{}
	var total int
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(),
			`select id::text, period_start, period_end, currency, total_minor, status,
			        attempts, last_error, issued_at, paid_at
			   from invoices order by `+page.OrderBy()+` limit $1 offset $2`,
			page.Limit, page.Offset)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var v invoiceResponse
			var ps, pe, issued time.Time
			var paid *time.Time
			if err := rows.Scan(&v.ID, &ps, &pe, &v.Currency, &v.TotalMinor, &v.Status,
				&v.Attempts, &v.LastError, &issued, &paid); err != nil {
				return err
			}
			v.PeriodStart, v.PeriodEnd = ps.Format("2006-01-02"), pe.Format("2006-01-02")
			v.IssuedAt = issued.UTC().Format(time.RFC3339)
			if paid != nil {
				p := paid.UTC().Format(time.RFC3339)
				v.PaidAt = &p
			}
			v.Lines = []invoiceLine{}
			out = append(out, v)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		if err := tx.QueryRow(r.Context(), `select count(*) from invoices`).Scan(&total); err != nil {
			return err
		}
		// Lines in one pass rather than per invoice: a year of bills is twelve rows
		// and thirty-six lines, not thirteen round trips.
		lines, err := tx.Query(r.Context(),
			`select invoice_id::text, kind, unit, quantity::text, unit_amount::text,
			        per::text, amount_minor from invoice_lines`)
		if err != nil {
			return err
		}
		defer lines.Close()
		byInvoice := map[string][]invoiceLine{}
		for lines.Next() {
			var id string
			var l invoiceLine
			if err := lines.Scan(&id, &l.Kind, &l.Unit, &l.Quantity, &l.UnitAmount,
				&l.Per, &l.AmountMinor); err != nil {
				return err
			}
			byInvoice[id] = append(byInvoice[id], l)
		}
		if err := lines.Err(); err != nil {
			return err
		}
		for i := range out {
			if ls, ok := byInvoice[out[i].ID]; ok {
				out[i].Lines = ls
			}
		}
		return nil
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"invoices": out, "total": total})
}

// payInvoice hands back somewhere to pay. It is how a BDT customer settles a bill at
// all, and how a card customer replaces a card that keeps declining.
func (s *Server) payInvoice(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)
	invoiceID := chi.URLParam(r, "id")

	var currency, status string
	var totalMinor int64
	var email *string
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(r.Context(),
			`select currency, status, total_minor from invoices where id = $1`,
			invoiceID).Scan(&currency, &status, &totalMinor); err != nil {
			return err
		}
		return tx.QueryRow(r.Context(), `select billing_email from tenants`).Scan(&email)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		writeErrFor(w, r, http.StatusNotFound, "invoice_not_found", "We couldn't find that invoice.")
		return
	}
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	if status == "paid" || totalMinor <= 0 {
		writeErrFor(w, r, http.StatusConflict, "invoice_not_payable",
			"That invoice has nothing left to pay.")
		return
	}

	gw, ok := s.gateways[pipeline.GatewayFor(currency)]
	if !ok {
		writeErrFor(w, r, http.StatusServiceUnavailable, "billing_not_configured",
			"Paying online isn't switched on yet. Talk to us and we'll settle it directly.")
		return
	}

	att, err := gw.Checkout(r.Context(),
		payments.Invoice{ID: invoiceID, TenantID: tenantID, AmountMinor: totalMinor,
			Currency: currency, Description: "Alchemist usage"},
		payments.Customer{Email: strValue(email)}, s.billingURL)
	if err != nil {
		writeErrFor(w, r, http.StatusServiceUnavailable, "gateway_unavailable",
			"The payment page could not be opened just now. Please try again shortly.")
		return
	}

	_ = s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		_, e := tx.Exec(r.Context(),
			`update invoices set gateway = $2, gateway_ref = $3 where id = $1`,
			invoiceID, gw.Name(), att.Ref)
		return e
	})
	writeJSON(w, http.StatusOK, map[string]any{"pay_url": att.URL})
}

// gatewayWebhook takes a callback from Stripe or SSLCommerz.
//
// Unauthenticated by necessity -- a gateway holds no API key -- so the signature is the
// only thing standing between a stranger and a paid invoice. It is checked first, and a
// failure is 403 rather than 500: a forged "paid" is exactly what this is for, and a
// 500 invites the gateway to redeliver a request that will never be accepted.
func (s *Server) gatewayWebhook(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "gateway")
	gw, ok := s.gateways[name]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ev, err := gw.Verify(r.Context(), r)
	if errors.Is(err, payments.ErrUnverified) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if err != nil {
		// A real failure talking to the gateway's own validation API. 500 is right
		// here: it should be redelivered.
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Genuine, and about something this does not act on. 200, or it is redelivered
	// for days.
	if ev.Ignored {
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := s.applyPayment(r.Context(), name, ev); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// applyPayment credits an invoice from a verified event, once.
//
// The gateway's own reference is a unique key on payments, so an IPN and a redirect
// describing the same money, or a webhook delivered twice, settle the invoice once.
func (s *Server) applyPayment(ctx context.Context, gateway string, ev payments.Event) error {
	var invoiceID, tenantID, currency, status string
	var totalMinor int64
	// No tenant in scope yet, hence the definer function.
	if err := s.db.Pool().QueryRow(ctx,
		`select invoice_id::text, tenant_id::text, currency, total_minor, status
		   from resolve_invoice($1, $2, $3)`,
		ev.InvoiceID, gateway, ev.Ref).
		Scan(&invoiceID, &tenantID, &currency, &totalMinor, &status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Verified, but about an invoice this deployment does not have. Accepted
			// so it stops being redelivered; nothing is credited.
			return nil
		}
		return err
	}

	// An event that says less was paid than is owed does not settle the invoice.
	// Trusting the event's own amount is how a one-taka payment clears a 1560 bill.
	if ev.Succeeded && ev.AmountMinor > 0 && ev.AmountMinor < totalMinor {
		ev.Succeeded = false
		ev.Failed = true
		ev.Error = fmt.Sprintf("paid %d of %d %s", ev.AmountMinor, totalMinor, currency)
	}

	return s.db.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		payStatus := "failed"
		if ev.Succeeded {
			payStatus = "succeeded"
		}
		if _, err := tx.Exec(ctx,
			`insert into payments (invoice_id, tenant_id, gateway, external_id,
			        amount_minor, currency, status, error)
			 values ($1,$2,$3,$4,$5,$6,$7,nullif($8,''))
			 on conflict (gateway, external_id) do update set
			   status = excluded.status, error = excluded.error`,
			invoiceID, tenantID, gateway, ev.Ref, ev.AmountMinor, currency,
			payStatus, ev.Error); err != nil {
			return err
		}
		if !ev.Succeeded {
			return nil
		}
		if _, err := tx.Exec(ctx,
			`update invoices set status = 'paid', paid_at = now(), last_error = null
			  where id = $1 and status <> 'paid'`, invoiceID); err != nil {
			return err
		}
		_, err := tx.Exec(ctx,
			`update tenants set billing_status = 'active'
			  where billing_status <> 'active'
			    and not exists (select 1 from invoices
			                     where status in ('open','failed') and total_minor > 0)`)
		return err
	})
}

func strValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
