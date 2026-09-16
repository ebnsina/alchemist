package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/ebnsina/alchemist/internal/pipeline"
	"github.com/ebnsina/alchemist/internal/platform/httpx"
)

type createBucketSourceRequest struct {
	Endpoint  string `json:"endpoint"`
	Region    string `json:"region"`
	Bucket    string `json:"bucket"`
	Prefix    string `json:"prefix"`
	AccessKey string `json:"access_key_id"`
	SecretKey string `json:"secret_access_key"`
}

// createBucketSource registers a customer-owned bucket to ingest from.
//
// The secret is wrapped with the platform KEK before it is stored and is never
// returned by any endpoint afterwards.
func (s *Server) createBucketSource(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	var req createBucketSourceRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&req); err != nil {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request",
			"Send a JSON body with endpoint, bucket, access_key_id and secret_access_key.")
		return
	}
	req.Endpoint = strings.TrimSpace(req.Endpoint)
	req.Bucket = strings.TrimSpace(req.Bucket)

	u, err := url.Parse(req.Endpoint)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_url",
			"The endpoint must be an http or https address.")
		return
	}
	if req.Bucket == "" || req.AccessKey == "" || req.SecretKey == "" {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request",
			"Include the bucket name and its access keys.")
		return
	}
	if req.Region == "" {
		req.Region = "us-east-1"
	}

	var sourceID string
	err = s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`insert into bucket_sources
			   (tenant_id, endpoint, region, bucket, prefix, access_key_id,
			    wrapped_secret, secret_nonce)
			 values ($1,$2,$3,$4,$5,$6,'\x00','\x00') returning id::text`,
			tenantID, req.Endpoint, req.Region, req.Bucket, req.Prefix, req.AccessKey).
			Scan(&sourceID)
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"We couldn't save that bucket.")
		return
	}

	// Wrapping is bound to the source id, so the row has to exist first.
	wrapped, nonce, err := s.keys.Wrap([]byte(req.SecretKey), sourceID)
	if err == nil {
		err = s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
			_, e := tx.Exec(r.Context(),
				`update bucket_sources set wrapped_secret = $2, secret_nonce = $3
				  where id = $1`, sourceID, wrapped, nonce)
			return e
		})
	}
	if err != nil {
		// Never leave a row holding a placeholder secret.
		_ = s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
			_, e := tx.Exec(r.Context(), `delete from bucket_sources where id = $1`, sourceID)
			return e
		})
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"We couldn't save that bucket.")
		return
	}

	if _, err := s.river.Insert(r.Context(),
		pipeline.BucketSyncArgs{SourceID: sourceID, TenantID: tenantID}, nil); err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"We saved the bucket but couldn't start the first scan.")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id": sourceID, "bucket": req.Bucket, "prefix": req.Prefix, "active": true,
	})
}

type bucketRow struct {
	ID           string  `json:"id"`
	Bucket       string  `json:"bucket"`
	Prefix       string  `json:"prefix"`
	Active       bool    `json:"active"`
	LastSyncedAt *string `json:"last_synced_at"`
	LastError    *string `json:"last_error"`
	Imported     int     `json:"imported_objects"`
}

// bucketColumns is shared by the list and the single-source reads so all three
// report the same import count.
const bucketColumns = `s.id::text, s.bucket, s.prefix, s.active,
	to_char(s.last_synced_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), s.last_error,
	(select count(*) from bucket_objects o where o.source_id = s.id)`

func scanBucket(row pgx.Row, it *bucketRow) error {
	return row.Scan(&it.ID, &it.Bucket, &it.Prefix, &it.Active, &it.LastSyncedAt,
		&it.LastError, &it.Imported)
}

func (s *Server) listBucketSources(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	page, ok := httpx.ParseList(w, r, []string{"created_at", "bucket"}, "created_at")
	if !ok {
		return
	}
	// A bucket is connected or paused; "state" keeps the name every other list uses
	// rather than making this the one endpoint that says "active" in a filter.
	state := r.URL.Query().Get("state")
	if state != "" && state != "active" && state != "paused" {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_filter",
			"Filter by state active or paused.")
		return
	}

	const where = `from bucket_sources s
	  where ($1::text = '' or s.bucket ilike '%' || $1::text || '%'
	                       or s.prefix ilike '%' || $1::text || '%')
	    and ($2::text = '' or s.active = ($2::text = 'active'))`

	items := []bucketRow{}
	var total int
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(),
			`select `+bucketColumns+` `+where+
				` order by s.`+page.OrderBy()+` limit $3 offset $4`,
			page.Q, state, page.Limit, page.Offset)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var it bucketRow
			if err := scanBucket(rows, &it); err != nil {
				return err
			}
			items = append(items, it)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		return tx.QueryRow(r.Context(), `select count(*) `+where, page.Q, state).Scan(&total)
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"bucket_sources": items, "total": total})
}

func (s *Server) getBucketSource(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	var it bucketRow
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return scanBucket(tx.QueryRow(r.Context(),
			`select `+bucketColumns+` from bucket_sources s where s.id = $1`,
			chi.URLParam(r, "id")), &it)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		writeErrFor(w, r, http.StatusNotFound, "bucket_not_found",
			"We couldn't find that bucket.")
		return
	}
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	// Never the access keys: the secret is wrapped and is not readable here at all.
	writeJSON(w, http.StatusOK, it)
}

// patchBucketSource pauses or resumes a connection. Pausing stops the next scan;
// videos already imported are ordinary assets and are untouched.
func (s *Server) patchBucketSource(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	var req struct {
		Active *bool `json:"active"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&req); err != nil || req.Active == nil {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request",
			"Send a JSON body with \"active\": true or false.")
		return
	}

	var it bucketRow
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		id := chi.URLParam(r, "id")
		tag, err := tx.Exec(r.Context(),
			`update bucket_sources set active = $2 where id = $1`, id, *req.Active)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return pgx.ErrNoRows
		}
		return scanBucket(tx.QueryRow(r.Context(),
			`select `+bucketColumns+` from bucket_sources s where s.id = $1`, id), &it)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		writeErrFor(w, r, http.StatusNotFound, "bucket_not_found",
			"We couldn't find that bucket.")
		return
	}
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	writeJSON(w, http.StatusOK, it)
}

// deleteBucketSource disconnects a bucket for good.
//
// It refuses while a video from it is still being processed: the transcode worker
// reads the source object back through this row's credentials, so deleting it mid
// flight leaves a job that cannot fetch its own input and an asset stuck short of
// ready with no error a customer could act on.
func (s *Server) deleteBucketSource(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	var inFlight, deleted int64
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		id := chi.URLParam(r, "id")
		if err := tx.QueryRow(r.Context(),
			`select count(*) from assets
			  where bucket_source_id = $1
			    and state::text not in ('ready', 'partially_ready', 'failed')`, id).
			Scan(&inFlight); err != nil {
			return err
		}
		if inFlight > 0 {
			return nil
		}
		// Imported assets keep their media and lose only the pointer back here.
		// The object ledger goes with the row, so reconnecting re-imports the bucket.
		tag, err := tx.Exec(r.Context(), `delete from bucket_sources where id = $1`, id)
		if err != nil {
			return err
		}
		deleted = tag.RowsAffected()
		return nil
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	if inFlight > 0 {
		writeErrFor(w, r, http.StatusConflict, "bucket_in_use",
			"Videos from this bucket are still being processed. Try again once they finish.")
		return
	}
	if deleted == 0 {
		writeErrFor(w, r, http.StatusNotFound, "bucket_not_found",
			"We couldn't find that bucket.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
