package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/ebnsina/alchemist/internal/pipeline"
)

// MaxCaptionBytes bounds one subtitle track. Three hours of dialogue is a few hundred
// kilobytes; anything past this is a file that is not a transcript.
const MaxCaptionBytes = 2 << 20

// language is BCP-47 as the manifests carry it: "bn", "en", "bn-BD". Restricted
// because it is written into an HLS attribute and a DASH attribute, and into an
// object key -- three places where an arbitrary string is someone else's bug.
var language = regexp.MustCompile(`^[a-z]{2,3}(-[A-Za-z0-9]{2,8})*$`)

type captionResponse struct {
	Language  string `json:"language"`
	Label     string `json:"label"`
	Bytes     *int64 `json:"bytes"`
	UpdatedAt string `json:"updated_at"`
}

func (s *Server) listCaptions(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)
	assetID := chi.URLParam(r, "id")

	out := []captionResponse{}
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		// Through deduplicated_from: a duplicate plays the canonical asset's media and
		// its manifests, so it shows the canonical asset's tracks or it shows none.
		rows, err := tx.Query(r.Context(),
			`select c.language, c.label, c.bytes, c.updated_at
			   from captions c
			  where c.asset_id = (select coalesce(deduplicated_from, id)
			                        from assets where id = $1)
			  order by c.language`, assetID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var c captionResponse
			var updated time.Time
			if err := rows.Scan(&c.Language, &c.Label, &c.Bytes, &updated); err != nil {
				return err
			}
			c.UpdatedAt = updated.Format(time.RFC3339)
			out = append(out, c)
		}
		return rows.Err()
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"captions": out})
}

// putCaption stores one subtitle track and refreshes the manifests.
//
// The body is the WebVTT file itself rather than JSON: it is a file the customer
// exported from something else, and base64 in a JSON envelope only gives them a step
// to get wrong. The label rides in a query parameter.
func (s *Server) putCaption(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)
	assetID := chi.URLParam(r, "id")
	lang := strings.TrimSpace(chi.URLParam(r, "lang"))

	if !language.MatchString(lang) {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_language",
			"Use a language code like bn, en or bn-BD.")
		return
	}
	label := strings.TrimSpace(r.URL.Query().Get("label"))
	if label == "" {
		label = lang
	}
	if len(label) > 60 {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_label",
			"Keep the name under 60 characters — it has to fit a player's menu.")
		return
	}

	// One byte past the limit is enough to know it is over, and stops a huge body
	// being read into memory to find that out.
	body, err := io.ReadAll(io.LimitReader(r.Body, MaxCaptionBytes+1))
	if err != nil {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_request",
			"We could not read that file.")
		return
	}
	if len(body) > MaxCaptionBytes {
		writeErrFor(w, r, http.StatusRequestEntityTooLarge, "caption_too_large",
			"That subtitle file is over 2 MB. A transcript of a long lecture is far smaller.")
		return
	}
	// Checked here because a player does not report a malformed track: it shows no
	// subtitles and no error, and the customer is told the feature is broken.
	if !strings.HasPrefix(strings.TrimLeft(string(body), "\uFEFF \t\r\n"), "WEBVTT") {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_captions",
			"That is not a WebVTT file. It has to start with the word WEBVTT.")
		return
	}

	// The canonical asset owns the media and the manifests, so a track uploaded
	// against a duplicate has to land there or nothing will ever read it.
	var canonical string
	if err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`select coalesce(deduplicated_from, id)::text from assets where id = $1`,
			assetID).Scan(&canonical)
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeErrFor(w, r, http.StatusNotFound, "asset_not_found", "We couldn't find that video.")
			return
		}
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}

	key := pipeline.CaptionKey(fmt.Sprintf("cmaf/%s/%s", tenantID, canonical), lang)
	if err := s.store.Put(r.Context(), key, strings.NewReader(string(body)), "text/vtt"); err != nil {
		writeErrFor(w, r, http.StatusServiceUnavailable, "storage_unavailable",
			"We couldn't store that just now. Please try again shortly.")
		return
	}

	if err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(r.Context(),
			`insert into captions (asset_id, tenant_id, language, label, object_key, bytes)
			 values ($1,$2,$3,$4,$5,$6)
			 on conflict (asset_id, language) do update set
			   label = excluded.label, object_key = excluded.object_key,
			   bytes = excluded.bytes, updated_at = now()`,
			canonical, tenantID, lang, label, key, int64(len(body)))
		return err
	}); err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}

	s.refreshManifests(r, tenantID, canonical)
	writeJSON(w, http.StatusOK, map[string]any{
		"language": lang, "label": label, "bytes": len(body),
	})
}

func (s *Server) deleteCaption(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)
	assetID := chi.URLParam(r, "id")
	lang := chi.URLParam(r, "lang")

	var key, canonical string
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`delete from captions
			  where asset_id = (select coalesce(deduplicated_from, id)
			                      from assets where id = $1)
			    and language = $2
			 returning object_key, asset_id::text`, assetID, lang).Scan(&key, &canonical)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		writeErrFor(w, r, http.StatusNotFound, "caption_not_found",
			"There is no subtitle track in that language on this video.")
		return
	}
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}

	// The manifests stop naming it first. An object deleted while a manifest still
	// points at it is a 404 a player reports as a broken video, not a missing track.
	s.refreshManifests(r, tenantID, canonical)
	if _, err := s.river.Insert(r.Context(), pipeline.ReclaimArgs{Keys: []string{key}}, nil); err != nil {
		// The row is gone and the manifests no longer name it, so the track is off the
		// video either way. What is left is an object to sweep.
		writeJSON(w, http.StatusNoContent, nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// refreshManifests queues the rewrite rather than doing it inline: it needs the
// worker's storage and content-key handling, and the customer does not wait on it.
func (s *Server) refreshManifests(r *http.Request, tenantID, assetID string) {
	_, _ = s.river.Insert(r.Context(),
		pipeline.RepublishArgs{AssetID: assetID, TenantID: tenantID}, nil)
}
