package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// Plan defaults, used when a tenant has no explicit limits row. Defined here rather
// than only in the schema so the meaning is visible where it is enforced.
const (
	defaultMaxConcurrentJobs = 4
	defaultMaxSourceBytes    = int64(32) << 30
)

type limits struct {
	MaxConcurrentJobs int
	MaxSourceBytes    int64
	MaxIngestHoursMo  *float64
}

var errQuotaExceeded = errors.New("quota_exceeded")

// checkIngestQuota refuses a new ingest when the tenant is already at its limit.
//
// Concurrency is the limit that matters operationally: without it one customer's bulk
// import fills the queue and every other tenant's upload waits behind it. Monthly
// hours are a billing control and only apply when set.
func (s *Server) checkIngestQuota(ctx context.Context, tenantID string) (string, error) {
	var l limits
	var inFlight int
	var usedHours float64

	err := s.db.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		l = limits{
			MaxConcurrentJobs: defaultMaxConcurrentJobs,
			MaxSourceBytes:    defaultMaxSourceBytes,
		}
		err := tx.QueryRow(ctx,
			`select max_concurrent_jobs, max_source_bytes, max_ingest_hours_mo
			   from tenant_limits`).
			Scan(&l.MaxConcurrentJobs, &l.MaxSourceBytes, &l.MaxIngestHoursMo)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		if err := tx.QueryRow(ctx,
			`select count(*) from assets
			  where state in ('uploaded','probing','mezzanine','analyzing',
			                  'encoding','packaging')`).Scan(&inFlight); err != nil {
			return err
		}

		if l.MaxIngestHoursMo != nil {
			return tx.QueryRow(ctx,
				`select coalesce(sum(quantity),0)::float8 / 3600 from usage_events
				  where kind = 'ingest'
				    and occurred_at >= date_trunc('month', now())`).Scan(&usedHours)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	if inFlight >= l.MaxConcurrentJobs {
		return fmt.Sprintf(
			"You have %d videos processing, which is your limit. Try again once they finish.",
			inFlight), errQuotaExceeded
	}
	if l.MaxIngestHoursMo != nil && usedHours >= *l.MaxIngestHoursMo {
		return fmt.Sprintf(
			"You have used %.1f of your %.1f hours this month.",
			usedHours, *l.MaxIngestHoursMo), errQuotaExceeded
	}
	return "", nil
}
