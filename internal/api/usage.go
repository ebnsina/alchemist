package api

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
)

type usageLine struct {
	Kind     string  `json:"kind"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
}

type usageResponse struct {
	From  string      `json:"from"`
	To    string      `json:"to"`
	Lines []usageLine `json:"lines"`
}

// getUsage returns billable aggregates for a period. Events are written alongside
// the work that produces them, because usage cannot be reconstructed after the fact:
// by the time anyone asks, the evidence is gone.
func (s *Server) getUsage(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	// Default to the current calendar month, which is what a bill is drawn against.
	now := time.Now().UTC()
	from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)

	if v := r.URL.Query().Get("from"); v != "" {
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			writeErrFor(w, r, http.StatusBadRequest, "invalid_date",
				"Use an RFC 3339 date, for example 2026-09-01T00:00:00Z.")
			return
		}
		from = parsed
	}
	if v := r.URL.Query().Get("to"); v != "" {
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			writeErrFor(w, r, http.StatusBadRequest, "invalid_date",
				"Use an RFC 3339 date, for example 2026-10-01T00:00:00Z.")
			return
		}
		to = parsed
	}
	if !to.After(from) {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_range",
			"The end of the period must come after the start.")
		return
	}

	resp := usageResponse{
		From:  from.Format(time.RFC3339),
		To:    to.Format(time.RFC3339),
		Lines: []usageLine{},
	}
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(),
			`select kind, sum(quantity)::float8, unit
			   from usage_events
			  where occurred_at >= $1 and occurred_at < $2
			  group by kind, unit
			  order by kind`, from, to)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var l usageLine
			if err := rows.Scan(&l.Kind, &l.Quantity, &l.Unit); err != nil {
				return err
			}
			resp.Lines = append(resp.Lines, l)
		}
		return rows.Err()
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
