package pipeline

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"

	"github.com/ebnsina/alchemist/internal/platform/billing"
	"github.com/ebnsina/alchemist/internal/platform/db"
	"github.com/ebnsina/alchemist/internal/platform/payments"
)

// Monthly invoicing, and collecting what it produces.
//
// Two jobs rather than one, because they fail for different reasons and retry on
// different clocks: issuing is arithmetic over rows we already hold and either works or
// is a bug, while collecting talks to a card network and fails for reasons that are
// nobody's fault and often clear up on their own.

// MaxChargeAttempts is how many times an invoice is presented before the account is
// suspended. Three over three days: a card that declines three times is not a blip.
const MaxChargeAttempts = 3

// InvoiceArgs closes one month and issues every invoice for it.
type InvoiceArgs struct {
	// PeriodStart is the first day of the month being billed. Explicit rather than
	// derived from now(), so a missed run can be issued for the month it belongs to.
	PeriodStart string `json:"period_start"`
}

func (InvoiceArgs) Kind() string { return "issue_invoices" }

func (InvoiceArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: QueueIO, MaxAttempts: 3}
}

type InvoiceWorker struct {
	river.WorkerDefaults[InvoiceArgs]
	DB    *db.DB
	River *river.Client[pgx.Tx]
}

func (w *InvoiceWorker) Timeout(*river.Job[InvoiceArgs]) time.Duration { return 15 * time.Minute }

func (w *InvoiceWorker) Work(ctx context.Context, job *river.Job[InvoiceArgs]) error {
	start, err := time.Parse("2006-01-02", job.Args.PeriodStart)
	if err != nil {
		return river.JobCancel(fmt.Errorf("period_start %q is not a date", job.Args.PeriodStart))
	}
	end := start.AddDate(0, 1, 0)

	type billable struct{ tenantID, card, currency string }
	var tenants []billable
	// Cross-tenant, before any tenant is in scope, so it goes through the definer
	// function. It also excludes tenants already invoiced for this period, which is
	// what makes a retried job safe rather than a second bill.
	rows, err := w.DB.Pool().Query(ctx,
		`select tenant_id::text, rate_card, currency from billable_tenants($1)`, start)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var b billable
		if err := rows.Scan(&b.tenantID, &b.card, &b.currency); err != nil {
			return err
		}
		tenants = append(tenants, b)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, b := range tenants {
		if err := w.issueOne(ctx, b.tenantID, b.card, b.currency, start, end); err != nil {
			// One tenant's bad rate card must not stop everybody else being invoiced.
			w.recordBillingError(ctx, b.tenantID, err)
		}
	}
	return nil
}

// issueOne prices one tenant's month and writes the invoice.
func (w *InvoiceWorker) issueOne(ctx context.Context, tenantID, card, currency string,
	start, end time.Time) error {
	var rawRates []byte
	if err := w.DB.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `select rates from rate_cards where name = $1`, card).Scan(&rawRates)
	}); err != nil {
		return fmt.Errorf("read rate card %s: %w", card, err)
	}
	rates, err := billing.ParseRates(rawRates)
	if err != nil {
		return err
	}

	var usage []billing.Usage
	if err := w.DB.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		// Summed in the database as numeric, and read back as a string, so a month of
		// egress bytes never passes through a float on its way to being money.
		rows, err := tx.Query(ctx,
			`select kind, unit, sum(quantity)::text
			   from usage_events
			  where occurred_at >= $1 and occurred_at < $2
			  group by kind, unit
			  order by kind`, start, end)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var u billing.Usage
			var qty string
			if err := rows.Scan(&u.Kind, &u.Unit, &qty); err != nil {
				return err
			}
			q, ok := new(big.Rat).SetString(qty)
			if !ok {
				return fmt.Errorf("usage %s came back as %q", u.Kind, qty)
			}
			u.Quantity = q
			usage = append(usage, u)
		}
		return rows.Err()
	}); err != nil {
		return err
	}

	lines, total, err := billing.Charge(usage, rates, currency)
	if err != nil {
		return err
	}
	// Nothing used is nothing owed. An invoice for zero is still written, because a
	// customer who used nothing should see a zero bill rather than silence, but it is
	// marked paid rather than sent to a gateway that would refuse it.
	status := "open"
	if total == 0 {
		status = "paid"
	}

	return w.DB.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		var invoiceID string
		if err := tx.QueryRow(ctx,
			`insert into invoices (tenant_id, period_start, period_end, currency,
			        total_minor, status, paid_at)
			 values ($1,$2,$3,$4,$5,$6, case when $6 = 'paid' then now() end)
			 on conflict (tenant_id, period_start) do nothing
			 returning id::text`,
			tenantID, start, end.AddDate(0, 0, -1), currency, total, status).Scan(&invoiceID); err != nil {
			// Already invoiced by a concurrent run. Not an error: the unique key did
			// its job, which is the whole reason it is there.
			return nil
		}
		for _, l := range lines {
			if _, err := tx.Exec(ctx,
				`insert into invoice_lines (invoice_id, tenant_id, kind, unit, quantity,
				        unit_amount, per, amount_minor)
				 values ($1,$2,$3,$4,$5::numeric,$6::numeric,$7::numeric,$8)`,
				invoiceID, tenantID, l.Kind, l.Unit,
				l.Quantity.FloatString(4), l.UnitAmount.FloatString(6),
				l.Per.FloatString(4), l.AmountMinor); err != nil {
				return err
			}
		}
		return nil
	})
}

func (w *InvoiceWorker) recordBillingError(ctx context.Context, tenantID string, cause error) {
	_ = w.DB.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`insert into invoices (tenant_id, period_start, period_end, currency,
			        total_minor, status, last_error)
			 values ($1, current_date, current_date, 'BDT', 0, 'failed', $2)
			 on conflict (tenant_id, period_start) do update set last_error = excluded.last_error`,
			tenantID, cause.Error())
		return err
	})
}

// CollectArgs presents every open invoice for payment.
type CollectArgs struct{}

func (CollectArgs) Kind() string { return "collect_invoices" }

func (CollectArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: QueueIO, MaxAttempts: 2}
}

type CollectWorker struct {
	river.WorkerDefaults[CollectArgs]
	DB *db.DB
	// Gateways by name. A tenant's currency picks one; an account on a gateway that
	// is not configured is left alone rather than marked failed, because an operator
	// forgetting a key is not a customer's unpaid bill.
	Gateways map[string]payments.Gateway
	// BillingURL is where a customer goes to pay, which is what a checkout link
	// returns to.
	BillingURL string
}

func (w *CollectWorker) Timeout(*river.Job[CollectArgs]) time.Duration { return 15 * time.Minute }

func (w *CollectWorker) Work(ctx context.Context, _ *river.Job[CollectArgs]) error {
	type due struct {
		invoiceID, tenantID, currency string
		totalMinor                    int64
		attempts                      int
	}
	var invoices []due
	rows, err := w.DB.Pool().Query(ctx,
		`select invoice_id::text, tenant_id::text, currency, total_minor, attempts
		   from collectable_invoices($1)`, MaxChargeAttempts)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var d due
		if err := rows.Scan(&d.invoiceID, &d.tenantID, &d.currency, &d.totalMinor, &d.attempts); err != nil {
			return err
		}
		invoices = append(invoices, d)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, d := range invoices {
		if err := w.collectOne(ctx, d.invoiceID, d.tenantID, d.currency, d.totalMinor); err != nil {
			w.markAttempt(ctx, d.tenantID, d.invoiceID, err.Error())
		}
	}
	return nil
}

// GatewayFor picks the gateway by currency. One per account: holding a balance in one
// currency and charging in another is an accounting problem nobody asked for.
func GatewayFor(currency string) string {
	if currency == "BDT" {
		return "sslcommerz"
	}
	return "stripe"
}

func (w *CollectWorker) collectOne(ctx context.Context, invoiceID, tenantID, currency string,
	totalMinor int64) error {
	gw, ok := w.Gateways[GatewayFor(currency)]
	if !ok {
		return nil // not configured on this deployment; not the customer's problem
	}

	var methodRef, customerRef, email string
	_ = w.DB.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		_ = tx.QueryRow(ctx,
			`select coalesce(external_id,''), coalesce(customer_id,'')
			   from payment_methods
			  where gateway = $1 order by created_at desc limit 1`, gw.Name()).
			Scan(&methodRef, &customerRef)
		return tx.QueryRow(ctx,
			`select coalesce(billing_email, '') from tenants`).Scan(&email)
	})

	inv := payments.Invoice{
		ID: invoiceID, TenantID: tenantID, AmountMinor: totalMinor, Currency: currency,
		Description: "Alchemist usage",
	}

	// Charged without the customer where the gateway can; otherwise a link they open.
	// SSLCommerz has no merchant-initiated charge in its published API, so a BDT
	// invoice always takes the second path.
	if charger, canCharge := gw.(payments.AutoCharger); canCharge && methodRef != "" {
		att, err := charger.Charge(ctx, inv, methodRef, customerRef)
		if err != nil {
			return err
		}
		w.recordPayment(ctx, tenantID, invoiceID, gw.Name(), att, currency, totalMinor)
		if att.Succeeded {
			return w.markPaid(ctx, tenantID, invoiceID)
		}
		w.markAttempt(ctx, tenantID, invoiceID, att.Error)
		return nil
	}

	att, err := gw.Checkout(ctx, inv, payments.Customer{Email: email, Ref: customerRef},
		w.BillingURL)
	if err != nil {
		return err
	}
	// The link is recorded, not followed. The invoice stays open until a verified
	// webhook says it was paid -- a checkout session created is not money received.
	return w.DB.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		_, e := tx.Exec(ctx,
			`update invoices set gateway = $2, gateway_ref = $3 where id = $1`,
			invoiceID, gw.Name(), att.Ref)
		return e
	})
}

func (w *CollectWorker) recordPayment(ctx context.Context, tenantID, invoiceID, gateway string,
	att payments.Attempt, currency string, amount int64) {
	if att.Ref == "" {
		return
	}
	status := "failed"
	if att.Succeeded {
		status = "succeeded"
	}
	_ = w.DB.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`insert into payments (invoice_id, tenant_id, gateway, external_id,
			        amount_minor, currency, status, error)
			 values ($1,$2,$3,$4,$5,$6,$7,nullif($8,''))
			 on conflict (gateway, external_id) do update set
			   status = excluded.status, error = excluded.error`,
			invoiceID, tenantID, gateway, att.Ref, amount, currency, status, att.Error)
		return err
	})
}

func (w *CollectWorker) markPaid(ctx context.Context, tenantID, invoiceID string) error {
	return w.DB.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			`update invoices set status = 'paid', paid_at = now(), last_error = null
			  where id = $1 and status <> 'paid'`, invoiceID); err != nil {
			return err
		}
		// Back to active only when nothing else is outstanding, or paying one of two
		// overdue invoices would lift a suspension the other still justifies.
		_, err := tx.Exec(ctx,
			`update tenants set billing_status = 'active'
			  where billing_status <> 'active'
			    and not exists (select 1 from invoices
			                     where status in ('open','failed') and total_minor > 0)`)
		return err
	})
}

// markAttempt records a failure and, at the limit, suspends.
//
// Suspension refuses new ingest and nothing else. Playback keeps working: taking a
// customer's viewers offline over an unpaid invoice punishes people who are not party
// to it, and it is the one action that cannot be undone by paying.
func (w *CollectWorker) markAttempt(ctx context.Context, tenantID, invoiceID, reason string) {
	_ = w.DB.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		var attempts int
		if err := tx.QueryRow(ctx,
			`update invoices set attempts = attempts + 1, status = 'failed',
			        last_error = nullif($2,'')
			  where id = $1 returning attempts`, invoiceID, reason).Scan(&attempts); err != nil {
			return err
		}
		status := "past_due"
		if attempts >= MaxChargeAttempts {
			status = "suspended"
		}
		_, err := tx.Exec(ctx,
			`update tenants set billing_status = $1 where billing_status <> 'suspended'`, status)
		return err
	})
}
