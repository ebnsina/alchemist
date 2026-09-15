package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

// Two views for looking at one asset closely: the chunk map, and what the pipeline
// did to it. Both are the customer's own data, and both are read through the
// tenant-scoped path so another tenant's id simply matches nothing.

type chunkRow struct {
	Rendition string  `json:"rendition"`
	Index     int     `json:"index"`
	StartSec  float64 `json:"start_sec"`
	EndSec    float64 `json:"end_sec"`
	Done      bool    `json:"done"`
}

func (s *Server) listChunks(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)
	assetID := chi.URLParam(r, "id")

	out := []chunkRow{}
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(),
			`select r.height::text || 'p ' || r.codec, c.idx, c.start_sec, c.end_sec,
			        c.completed_at is not null
			   from chunks c
			   join renditions r on r.id = c.rendition_id
			  where r.asset_id = (select coalesce(deduplicated_from, id)
			                        from assets where id = $1)
			  order by r.height desc, c.idx`, assetID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var c chunkRow
			if err := rows.Scan(&c.Rendition, &c.Index, &c.StartSec, &c.EndSec, &c.Done); err != nil {
				return err
			}
			out = append(out, c)
		}
		return rows.Err()
	})
	if err != nil {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request", "That asset id is not valid.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"chunks": out})
}

type activityRow struct {
	Step       string  `json:"step"`
	State      string  `json:"state"`
	Attempt    int     `json:"attempt"`
	Failures   int     `json:"failures"`
	QueuedAt   string  `json:"queued_at"`
	StartedAt  *string `json:"started_at"`
	FinishedAt *string `json:"finished_at"`
}

// listActivity reads the job history for one asset.
//
// Deliberately not the error text. These rows come from the queue, where a failure
// carries whatever ffmpeg or the storage client said — paths, hostnames, internal
// arguments. What a customer can act on is the stable error_code on the asset, which
// getAsset already returns; the rest is ours to read in the logs.
func (s *Server) listActivity(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)
	assetID := chi.URLParam(r, "id")

	// Ownership first, through RLS. river_job is not tenant-scoped, so nothing below
	// may run until the asset is known to belong to this tenant.
	var exists bool
	if err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`select exists (select 1 from assets where id = $1)`, assetID).Scan(&exists)
	}); err != nil {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request", "That asset id is not valid.")
		return
	}
	if !exists {
		writeErrFor(w, r, http.StatusNotFound, "asset_not_found", "We couldn't find that video.")
		return
	}

	rows, err := s.db.Pool().Query(r.Context(),
		`select kind, state::text, attempt, coalesce(array_length(errors, 1), 0),
		        created_at, attempted_at, finalized_at
		   from river_job
		  where args->>'asset_id' = $1
		  order by created_at`, assetID)
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	defer rows.Close()

	out := []activityRow{}
	for rows.Next() {
		var a activityRow
		var queued time.Time
		var started, finished *time.Time
		if err := rows.Scan(&a.Step, &a.State, &a.Attempt, &a.Failures,
			&queued, &started, &finished); err != nil {
			writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
				"Something went wrong on our side.")
			return
		}
		a.QueuedAt = queued.UTC().Format(time.RFC3339)
		if started != nil {
			v := started.UTC().Format(time.RFC3339)
			a.StartedAt = &v
		}
		if finished != nil {
			v := finished.UTC().Format(time.RFC3339)
			a.FinishedAt = &v
		}
		out = append(out, a)
	}
	writeJSON(w, http.StatusOK, map[string]any{"activity": out})
}
