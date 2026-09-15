package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/ebnsina/alchemist/internal/pipeline"
	"github.com/ebnsina/alchemist/internal/platform/providers"
)

// listProviders tells the dashboard which services can be migrated from, and what
// each one needs beyond a key. Hosts that expose no downloadable source are absent
// rather than listed and broken.
func (s *Server) listProviders(w http.ResponseWriter, r *http.Request) {
	type field struct {
		Key   string `json:"key"`
		Label string `json:"label"`
		Hint  string `json:"hint"`
	}
	type secret struct {
		Label string `json:"label"`
		Hint  string `json:"hint"`
		Kind  string `json:"kind"`
	}
	type item struct {
		Name   string  `json:"name"`
		Label  string  `json:"label"`
		Secret secret  `json:"secret"`
		Config []field `json:"config"`
	}
	out := []item{}
	for _, p := range providers.All() {
		fields := []field{}
		for _, f := range p.NeedsConfig() {
			fields = append(fields, field{Key: f.Key, Label: f.Label, Hint: f.Hint})
		}
		sp := p.SecretSpec()
		out = append(out, item{
			Name:   p.Name(),
			Label:  p.Label(),
			Secret: secret{Label: sp.Label, Hint: sp.Hint, Kind: sp.Kind},
			Config: fields,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"providers": out})
}

type createMigrationRequest struct {
	Provider string            `json:"provider"`
	Secret   string            `json:"secret"`
	Config   map[string]string `json:"config"`
}

func (s *Server) createMigration(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	var req createMigrationRequest
	// A pasted list of links is the body here, not a one-line key, so the cap is
	// generous — roughly twenty thousand URLs.
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<20)).Decode(&req); err != nil {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request",
			"Send a JSON body with a provider and a key.")
		return
	}
	p, ok := providers.Get(req.Provider)
	if !ok {
		writeErrFor(w, r, http.StatusBadRequest, "unknown_provider",
			"We cannot migrate from that service. Some hosts expose no downloadable original at all.")
		return
	}
	if strings.TrimSpace(req.Secret) == "" {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request", "Include the provider's API key.")
		return
	}
	if req.Config == nil {
		req.Config = map[string]string{}
	}
	for _, f := range p.NeedsConfig() {
		if strings.TrimSpace(req.Config[f.Key]) == "" {
			writeErrFor(w, r, http.StatusBadRequest, "invalid_request",
				"That service also needs its "+strings.ToLower(f.Label)+".")
			return
		}
	}

	cfg, err := json.Marshal(req.Config)
	if err != nil {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request", "Those settings did not come through.")
		return
	}

	// Same two-step as a bucket source: the row first, because wrapping is bound to
	// its id, then the secret, and the row is removed if wrapping fails so a
	// placeholder is never left behind.
	var sourceID string
	err = s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`insert into migration_sources (tenant_id, provider, config, wrapped_secret, secret_nonce)
			 values ($1, $2, $3, '\x00', '\x00') returning id::text`,
			tenantID, req.Provider, cfg).Scan(&sourceID)
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"We couldn't start that migration.")
		return
	}

	wrapped, nonce, err := s.keys.Wrap([]byte(req.Secret), sourceID)
	if err == nil {
		err = s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
			_, e := tx.Exec(r.Context(),
				`update migration_sources set wrapped_secret = $2, secret_nonce = $3
				  where id = $1`, sourceID, wrapped, nonce)
			return e
		})
	}
	if err != nil {
		_ = s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
			_, e := tx.Exec(r.Context(), `delete from migration_sources where id = $1`, sourceID)
			return e
		})
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"We couldn't start that migration.")
		return
	}

	if _, err := s.river.Insert(r.Context(),
		pipeline.MigrateArgs{SourceID: sourceID, TenantID: tenantID}, nil); err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"We saved it but couldn't start the first pass.")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"id": sourceID, "provider": req.Provider, "state": "previewing",
	})
}

func (s *Server) listMigrations(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	type item struct {
		ID          string  `json:"id"`
		Provider    string  `json:"provider"`
		State       string  `json:"state"`
		PreviewDone bool    `json:"preview_done"`
		Total       int     `json:"total"`
		Handled     int     `json:"handled"`
		Imported    int     `json:"imported"`
		Skipped     int     `json:"skipped"`
		LastError   *string `json:"last_error"`
		CreatedAt   string  `json:"created_at"`
	}
	items := []item{}
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(),
			`select m.id::text, m.provider, m.state, m.preview_done,
			        count(i.*),
			        count(i.*) filter (where i.state in ('imported', 'skipped')),
			        count(i.*) filter (where i.state = 'imported'),
			        count(i.*) filter (where i.state = 'skipped'),
			        m.last_error,
			        to_char(m.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
			   from migration_sources m
			   left join migration_items i on i.source_id = m.id
			  group by m.id order by m.created_at desc`)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var it item
			if err := rows.Scan(&it.ID, &it.Provider, &it.State, &it.PreviewDone,
				&it.Total, &it.Handled, &it.Imported, &it.Skipped,
				&it.LastError, &it.CreatedAt); err != nil {
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
	writeJSON(w, http.StatusOK, map[string]any{"migrations": items})
}

// listMigrationItems is what makes a partial migration explainable: which videos came
// across, which did not, and why not.
func (s *Server) listMigrationItems(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)
	id := chi.URLParam(r, "id")

	type item struct {
		RemoteID string  `json:"remote_id"`
		Title    string  `json:"title"`
		AssetID  *string `json:"asset_id"`
		State    string  `json:"state"`
		Reason   *string `json:"reason"`
	}
	items := []item{}
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(),
			`select i.remote_id, coalesce(i.title, ''), i.asset_id::text, i.state, i.reason
			   from migration_items i
			   join migration_sources m on m.id = i.source_id
			  where i.source_id = $1
			  order by i.created_at desc limit 500`, id)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var it item
			if err := rows.Scan(&it.RemoteID, &it.Title, &it.AssetID, &it.State, &it.Reason); err != nil {
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
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

// setMigrationState is the one place the allowed moves live, so confirm, pause and
// resume cannot disagree about what is legal.
func (s *Server) setMigrationState(w http.ResponseWriter, r *http.Request, from []string, to string) {
	tenantID, _ := r.Context().Value(tenantKey).(string)
	id := chi.URLParam(r, "id")

	var exists bool
	var moved bool
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(r.Context(),
			`select exists (select 1 from migration_sources where id = $1)`, id).
			Scan(&exists); err != nil || !exists {
			return err
		}
		// Confirming resets the paging so the whole library is walked again, this
		// time importing; the per-item dedupe keeps that from doubling anything.
		return tx.QueryRow(r.Context(),
			`update migration_sources
			    set state = $2,
			        cursor = case when $2 = 'scanning' and state = 'previewing' then '' else cursor end,
			        last_error = null, updated_at = now()
			  where id = $1 and state = any($3)
			    and (state <> 'previewing' or preview_done)
			 returning true`, id, to, from).Scan(&moved)
	})
	if err != nil && err != pgx.ErrNoRows {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	if !exists {
		writeErrFor(w, r, http.StatusNotFound, "not_found", "We could not find that migration.")
		return
	}
	if !moved {
		writeErrFor(w, r, http.StatusConflict, "invalid_state",
			"That migration is not in a state where this can be done.")
		return
	}
	if to == "scanning" {
		if _, err := s.river.Insert(r.Context(),
			pipeline.MigrateArgs{SourceID: id, TenantID: tenantID}, nil); err != nil {
			writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
				"We saved it but couldn't get it moving again.")
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "state": to})
}

// confirmMigration is the consent gate: until this is called nothing has been
// imported, only listed.
func (s *Server) confirmMigration(w http.ResponseWriter, r *http.Request) {
	s.setMigrationState(w, r, []string{"previewing"}, "scanning")
}

func (s *Server) pauseMigration(w http.ResponseWriter, r *http.Request) {
	s.setMigrationState(w, r, []string{"scanning"}, "paused")
}

func (s *Server) resumeMigration(w http.ResponseWriter, r *http.Request) {
	s.setMigrationState(w, r, []string{"paused", "failed"}, "scanning")
}

// deleteMigration forgets the migration and its key. Videos already brought across
// are assets now and stay; nothing on the far side is touched.
func (s *Server) deleteMigration(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)
	id := chi.URLParam(r, "id")

	var gone bool
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		e := tx.QueryRow(r.Context(),
			`delete from migration_sources where id = $1 returning true`, id).Scan(&gone)
		if e == pgx.ErrNoRows {
			return nil
		}
		return e
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	if !gone {
		writeErrFor(w, r, http.StatusNotFound, "not_found", "We could not find that migration.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
