package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/ebnsina/alchemist/internal/pipeline"
	"github.com/ebnsina/alchemist/internal/platform/media"
)

// Studio. An edit reads one asset and produces another; it never changes the one it
// came from, because somebody may already have handed out a link to it.

type createEditRequest struct {
	AssetID string     `json:"asset_id"`
	Ops     media.Edit `json:"ops"`
}

func (s *Server) createEdit(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	var req createEditRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&req); err != nil {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request",
			"Send a JSON body with an asset_id and what to change.")
		return
	}
	if req.AssetID == "" {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request", "Say which video to edit.")
		return
	}

	// Validated here so an impossible edit is refused while somebody is looking at
	// the screen, rather than failing in a worker minutes later.
	if _, err := req.Ops.Args("in", "out", 0); err != nil {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_edit",
			"That edit does not describe a video we can make.")
		return
	}

	ops, err := json.Marshal(req.Ops)
	if err != nil {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request", "Those settings did not come through.")
		return
	}

	var ready bool
	var editID string
	err = s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		// Only from a video that has a mezzanine. Before that there is nothing to cut.
		if err := tx.QueryRow(r.Context(),
			`select mezzanine_key is not null and mezzanine_key <> ''
			   from assets where id = $1`, req.AssetID).Scan(&ready); err != nil {
			return err
		}
		if !ready {
			return nil
		}
		return tx.QueryRow(r.Context(),
			`insert into edits (tenant_id, source_asset_id, ops)
			 values ($1, $2, $3) returning id::text`, tenantID, req.AssetID, ops).Scan(&editID)
	})
	switch {
	case err == pgx.ErrNoRows:
		writeErrFor(w, r, http.StatusNotFound, "asset_not_found", "We could not find that video.")
		return
	case err != nil:
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"We couldn't start that edit.")
		return
	case !ready:
		writeErrFor(w, r, http.StatusConflict, "not_ready",
			"That video is still being processed. Editing opens once it is ready.")
		return
	}

	if _, err := s.river.Insert(r.Context(),
		pipeline.EditArgs{EditID: editID, TenantID: tenantID}, nil); err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"We saved the edit but couldn't queue it.")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": editID, "state": "queued"})
}

func (s *Server) listEdits(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)
	assetID := r.URL.Query().Get("asset_id")

	type item struct {
		ID            string     `json:"id"`
		SourceAssetID string     `json:"source_asset_id"`
		OutputAssetID *string    `json:"output_asset_id"`
		State         string     `json:"state"`
		ErrorCode     *string    `json:"error_code"`
		Ops           media.Edit `json:"ops"`
		CreatedAt     string     `json:"created_at"`
	}
	items := []item{}
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(),
			`select id::text, source_asset_id::text, output_asset_id::text, state,
			        error_code, ops, to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
			   from edits
			  where ($1 = '' or source_asset_id = $1::uuid)
			  order by created_at desc limit 200`, assetID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var it item
			var raw []byte
			if err := rows.Scan(&it.ID, &it.SourceAssetID, &it.OutputAssetID, &it.State,
				&it.ErrorCode, &raw, &it.CreatedAt); err != nil {
				return err
			}
			_ = json.Unmarshal(raw, &it.Ops)
			items = append(items, it)
		}
		return rows.Err()
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"edits": items})
}

func (s *Server) deleteEdit(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)
	id := chi.URLParam(r, "id")

	// Only the record goes. The video it produced is an asset like any other and is
	// deleted through the asset, not by discarding the edit that made it.
	var n int64
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		ct, e := tx.Exec(r.Context(), `delete from edits where id = $1`, id)
		if e != nil {
			return e
		}
		n = ct.RowsAffected()
		return nil
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	if n == 0 {
		writeErrFor(w, r, http.StatusNotFound, "not_found", "We could not find that edit.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
