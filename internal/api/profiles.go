package api

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5"
)

type ladderRung struct {
	Height     int    `json:"height"`
	Codec      string `json:"codec"`
	MaxrateBps int    `json:"maxrate_bps"`
	Lazy       bool   `json:"lazy"`
}

type ladderProfile struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Rungs       []ladderRung `json:"rungs"`
	Current     bool         `json:"current"`
}

// listProfiles exposes the encoding presets. Until now the only way to know which
// existed was to read the seed, and the only way to be on one other than bd-mobile
// was for an operator to update the row by hand.
func (s *Server) listProfiles(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	profiles := []ladderProfile{}
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		var current string
		if err := tx.QueryRow(r.Context(), `select ladder_profile from tenants`).Scan(&current); err != nil {
			return err
		}
		rows, err := tx.Query(r.Context(),
			`select name, description, rungs from ladder_profiles order by name`)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var p ladderProfile
			var raw []byte
			if err := rows.Scan(&p.Name, &p.Description, &raw); err != nil {
				return err
			}
			if err := json.Unmarshal(raw, &p.Rungs); err != nil {
				return err
			}
			p.Current = p.Name == current
			profiles = append(profiles, p)
		}
		return rows.Err()
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"profiles": profiles})
}

type setProfileRequest struct {
	Profile string `json:"profile"`
}

type playbackSettings struct {
	EncryptPlayback bool `json:"encrypt_playback"`
}

// getPlayback and setPlayback expose the one delivery setting a customer can change.
// Encryption is off by default: without a licence server it is not DRM — the key is
// served from the same signed URL as the segments — and it makes playback impossible
// outside Safari, which is most viewers.
func (s *Server) getPlayback(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	var out playbackSettings
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(), `select encrypt_playback from tenants`).
			Scan(&out.EncryptPlayback)
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) setPlayback(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)
	userID, _ := r.Context().Value(userKey).(string)
	if !s.canAdminister(w, r, tenantID, userID) {
		return
	}

	var req playbackSettings
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&req); err != nil {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request",
			"Send a JSON body with encrypt_playback.")
		return
	}
	// Videos already published keep whatever they were made with. Changing this does
	// not silently re-package a library, which would change every ETag and evict the
	// lot from every edge cache.
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		_, e := tx.Exec(r.Context(), `update tenants set encrypt_playback = $1`, req.EncryptPlayback)
		return e
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	writeJSON(w, http.StatusOK, req)
}

// setProfile changes which ladder new uploads are built against. Existing assets keep
// the profile they were encoded with — re-encoding a library because a setting moved
// would be a surprise that costs real money.
func (s *Server) setProfile(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)
	userID, _ := r.Context().Value(userKey).(string)
	if !s.canAdminister(w, r, tenantID, userID) {
		return
	}

	var req setProfileRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&req); err != nil {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request",
			"Send a JSON body with a profile name.")
		return
	}

	var unknown bool
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		var exists bool
		if err := tx.QueryRow(r.Context(),
			`select exists (select 1 from ladder_profiles where name = $1)`, req.Profile).
			Scan(&exists); err != nil {
			return err
		}
		if !exists {
			unknown = true
			return nil
		}
		_, err := tx.Exec(r.Context(), `update tenants set ladder_profile = $1`, req.Profile)
		return err
	})
	switch {
	case err != nil:
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
	case unknown:
		writeErrFor(w, r, http.StatusBadRequest, "unknown_profile",
			"We do not have an encoding preset by that name.")
	default:
		writeJSON(w, http.StatusOK, map[string]string{"profile": req.Profile})
	}
}
