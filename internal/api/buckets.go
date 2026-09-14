package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/ebnsina/alchemist/internal/pipeline"
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

func (s *Server) listBucketSources(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	type item struct {
		ID           string  `json:"id"`
		Bucket       string  `json:"bucket"`
		Prefix       string  `json:"prefix"`
		Active       bool    `json:"active"`
		LastSyncedAt *string `json:"last_synced_at"`
		LastError    *string `json:"last_error"`
		Imported     int     `json:"imported_objects"`
	}
	items := []item{}
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(),
			`select s.id::text, s.bucket, s.prefix, s.active,
			        to_char(s.last_synced_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), s.last_error,
			        (select count(*) from bucket_objects o where o.source_id = s.id)
			   from bucket_sources s order by s.created_at desc`)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var it item
			if err := rows.Scan(&it.ID, &it.Bucket, &it.Prefix, &it.Active,
				&it.LastSyncedAt, &it.LastError, &it.Imported); err != nil {
				return err
			}
			items = append(items, it)
		}
		return rows.Err()
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"bucket_sources": items})
}
