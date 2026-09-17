package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ebnsina/alchemist/internal/modules/delivery"
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
	// PlaybackOrigins are the web origins a playback link may be locked to. Empty
	// means unrestricted, which is what an account with a native app needs: the lock
	// is read from a browser's Origin or Referer and an app sends neither.
	PlaybackOrigins []string `json:"playback_origins"`
	// PlaybackTTLSeconds is how long a minted link lasts, and the ceiling for a
	// per-request ?ttl=. It has to outlast the longest single viewing: the signature
	// covers every segment, so a link that expires mid-lecture stops playback there.
	PlaybackTTLSeconds int `json:"playback_ttl_seconds"`
	MaxViewerDevices   int `json:"max_viewer_devices"`
}

// getPlayback and setPlayback expose the delivery settings a customer can change.
//
// Encryption protects the bytes at rest: a lifted bucket or backup is useless without
// the key. It is not DRM and does not stop a viewer who is entitled to watch from
// keeping a copy. max_viewer_devices is the one that answers password sharing, and it
// applies only to links minted with a viewer id.
func (s *Server) getPlayback(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	out := playbackSettings{PlaybackOrigins: []string{}}
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`select t.encrypt_playback, t.playback_origins, t.playback_ttl_seconds,
			        coalesce(l.max_viewer_devices, 0)
			   from tenants t left join tenant_limits l on l.tenant_id = t.id`).
			Scan(&out.EncryptPlayback, &out.PlaybackOrigins, &out.PlaybackTTLSeconds,
				&out.MaxViewerDevices)
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
			"Send a JSON body with encrypt_playback and max_viewer_devices.")
		return
	}
	// Bounded: a cap of a thousand devices is a typo, not a policy, and it would read
	// as protection while being none.
	if req.MaxViewerDevices < 0 || req.MaxViewerDevices > 20 {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request",
			"max_viewer_devices is between 0 (no limit) and 20.")
		return
	}
	// Normalised to exactly what a browser puts in an Origin header, so the comparison
	// at the edge is never doing normalisation -- njs would have to agree with it, and
	// two implementations of "is this the same site" is how a lock becomes a coin flip.
	origins := []string{}
	seen := map[string]bool{}
	for _, raw := range req.PlaybackOrigins {
		o, ok := delivery.NormalizeOrigin(raw)
		if !ok {
			writeErrFor(w, r, http.StatusBadRequest, "invalid_origin",
				"A site looks like https://app.example.com — scheme and domain, no path.")
			return
		}
		if seen[o] {
			continue
		}
		seen[o] = true
		origins = append(origins, o)
	}
	// Bounded: a list this long is a paste accident, and every entry is a string the
	// token may be locked to.
	if len(origins) > 20 {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_origin",
			"Twenty sites is as many as one account can list.")
		return
	}

	ttl := time.Duration(req.PlaybackTTLSeconds) * time.Second
	if ttl < MinPlaybackTTL || ttl > MaxPlaybackTTL {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_ttl",
			fmt.Sprintf("A playback link lasts between %d seconds and %d.",
				int(MinPlaybackTTL.Seconds()), int(MaxPlaybackTTL.Seconds())))
		return
	}

	// Videos already published keep whatever they were made with. Changing this does
	// not silently re-package a library, which would change every ETag and evict the
	// lot from every edge cache.
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		if _, e := tx.Exec(r.Context(),
			`update tenants set encrypt_playback = $1, playback_origins = $2,
			        playback_ttl_seconds = $3`,
			req.EncryptPlayback, origins, req.PlaybackTTLSeconds); e != nil {
			return e
		}
		// Upsert: a tenant on plan defaults has no limits row, and setting a cap must
		// not require one to have been created first.
		_, e := tx.Exec(r.Context(),
			`insert into tenant_limits (tenant_id, max_viewer_devices) values ($1, $2)
			 on conflict (tenant_id) do update
			   set max_viewer_devices = excluded.max_viewer_devices, updated_at = now()`,
			tenantID, req.MaxViewerDevices)
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
