package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

// A logo is small enough to go through the API. The presigned dance exists for source
// video, where a gigabyte through a request handler would tie up the control plane.
const maxLogoBytes = 1 << 20

// SVG is deliberately absent. It is a document, not an image: served from our own
// origin it can carry script, and a logo is not worth that.
var logoTypes = map[string]string{
	"image/png":  "png",
	"image/jpeg": "jpg",
	"image/webp": "webp",
}

func (s *Server) getBranding(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)

	var name string
	var logoKey, updated *string
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(r.Context(), `select name from tenants`).Scan(&name); err != nil {
			return err
		}
		err := tx.QueryRow(r.Context(),
			`select logo_key, to_char(updated_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
			   from tenant_branding`).Scan(&logoKey, &updated)
		if err == pgx.ErrNoRows {
			return nil
		}
		return err
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}

	resp := map[string]any{"name": name, "logo_url": nil, "logo_updated_at": updated}
	if logoKey != nil {
		// Cache-busted by the timestamp: the key never changes, so a replaced logo
		// would otherwise keep serving from every cache that already has it.
		resp["logo_url"] = fmt.Sprintf("/brand/%s/logo?v=%s", tenantID, *updated)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) putBrandingLogo(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)
	userID, _ := r.Context().Value(userKey).(string)
	if !s.canAdminister(w, r, tenantID, userID) {
		return
	}

	ext, ok := logoTypes[r.Header.Get("Content-Type")]
	if !ok {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_image",
			"Send a PNG, JPEG or WebP.")
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxLogoBytes))
	if err != nil {
		writeErrFor(w, r, http.StatusRequestEntityTooLarge, "image_too_large",
			"That image is over 1 MB. Send a smaller one.")
		return
	}
	if len(body) == 0 {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_image", "That file was empty.")
		return
	}
	// The declared type is the client's claim; the magic bytes are the fact. A PNG
	// header on a Content-Type of image/webp means one of the two is lying.
	if !sniffMatches(body, ext) {
		writeErrFor(w, r, http.StatusBadRequest, "invalid_image",
			"That file is not the image type it claims to be.")
		return
	}

	key := fmt.Sprintf("brand/%s/logo.%s", tenantID, ext)
	if err := s.store.Put(r.Context(), key, bytes.NewReader(body), r.Header.Get("Content-Type")); err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"We couldn't save that logo. Please try again.")
		return
	}

	var updated string
	err = s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`insert into tenant_branding (tenant_id, logo_key, logo_type)
			 values ($1, $2, $3)
			 on conflict (tenant_id) do update
			   set logo_key = excluded.logo_key, logo_type = excluded.logo_type,
			       updated_at = now()
			 returning to_char(updated_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')`,
			tenantID, key, r.Header.Get("Content-Type")).Scan(&updated)
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"We couldn't save that logo. Please try again.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"logo_url":        fmt.Sprintf("/brand/%s/logo?v=%s", tenantID, updated),
		"logo_updated_at": updated,
	})
}

func (s *Server) deleteBrandingLogo(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(tenantKey).(string)
	userID, _ := r.Context().Value(userKey).(string)
	if !s.canAdminister(w, r, tenantID, userID) {
		return
	}

	var key *string
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		err := tx.QueryRow(r.Context(),
			`update tenant_branding set logo_key = null, logo_type = null, updated_at = now()
			  where tenant_id = $1 returning logo_key`, tenantID).Scan(&key)
		if err == pgx.ErrNoRows {
			return nil
		}
		return err
	})
	if err != nil {
		writeErrFor(w, r, http.StatusInternalServerError, "internal_error",
			"Something went wrong on our side.")
		return
	}
	// The row is cleared first. A failed delete leaves an orphan object, which costs
	// a few kilobytes; the reverse leaves a row pointing at nothing.
	if key != nil {
		_ = s.store.Delete(r.Context(), *key)
	}
	w.WriteHeader(http.StatusNoContent)
}

// serveBrandLogo is public on purpose: a logo is shown to viewers who have no account
// and no key. Nothing else under the tenant prefix is reachable through it.
func (s *Server) serveBrandLogo(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenant")

	// Scoped to the tenant in the path, not queried off the bare pool: RLS is forced,
	// so an unscoped read returns zero rows rather than erroring and every logo 404s.
	// Scoping to the requested tenant is safe here because that id *is* the request.
	var key, ctype string
	err := s.db.AsTenant(r.Context(), tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(r.Context(),
			`select logo_key, logo_type from tenant_branding where logo_key is not null`).
			Scan(&key, &ctype)
	})
	if err != nil {
		http.NotFound(w, r)
		return
	}
	obj, err := s.store.Get(r.Context(), key)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer obj.Close()

	w.Header().Set("Content-Type", ctype)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = io.Copy(w, obj)
}

func sniffMatches(body []byte, ext string) bool {
	switch ext {
	case "png":
		return bytes.HasPrefix(body, []byte("\x89PNG\r\n\x1a\n"))
	case "jpg":
		return bytes.HasPrefix(body, []byte{0xFF, 0xD8, 0xFF})
	case "webp":
		return len(body) > 12 && bytes.Equal(body[0:4], []byte("RIFF")) &&
			bytes.Equal(body[8:12], []byte("WEBP"))
	}
	return false
}
