package api

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// The embed: an iframe a customer drops into their page.
//
// It is authorised by the same playback signature as the media, not by an API key,
// because the caller is a viewer's browser. The customer's server mints the signed
// URL it already gets from /v1/assets/{id} and puts it in the iframe src; this page
// hands that straight to the player.
//
// Mounted only when a player bundle is configured. Without one the page could only
// ever render an empty box, and a route that always fails is worse than no route.
func (s *Server) embedEnabled() bool { return s.playerURL != "" }

// embedPage is deliberately thin. Everything about playing video lives in the player
// bundle, which is versioned and cached separately -- a transcoder deploy must never
// change the player running inside thousands of already-embedded iframes.
var embedPage = template.Must(template.New("embed").Parse(`<!doctype html>
<html lang="en"><head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>{{.Title}}</title>
<style>html,body{margin:0;height:100%;background:#0C0C0E}#p{height:100%}</style>
</head><body>
<div id="p"></div>
<script type="module">
import { AlchemistPlayer } from {{.Bundle}};
new AlchemistPlayer(document.getElementById('p'), {
  src: {{.Src}},
  poster: {{.Poster}},
  lang: 'auto',
  dataSaver: 'auto',
});
</script>
</body></html>`))

func (s *Server) serveEmbed(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenant")
	assetID := chi.URLParam(r, "asset")

	prefix := fmt.Sprintf("/playback/%s/%s", tenantID, assetID)
	q := r.URL.Query()
	if !s.delivery.VerifyPlayback(prefix, q.Get("kid"), q.Get("sig"), q.Get("exp"),
		q.Get("vid"), q.Get("wm")) {
		writeErrFor(w, r, http.StatusForbidden, "playback_not_authorized",
			"This playback link has expired or is not valid.")
		return
	}

	// Whatever the link was signed for travels on, so a bound link stays bound and the
	// watermark cannot be dropped by embedding it.
	file := "master.m3u8"
	if strings.EqualFold(q.Get("format"), "dash") {
		file = "manifest.mpd"
	}
	// format chose the manifest and has no business on every segment request after it.
	q.Del("format")
	signed := q.Encode()

	// Framing is the entire point, so nothing here may set X-Frame-Options. The page
	// carries no customer data beyond the URLs already in the iframe src.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=60")
	w.Header().Set("Referrer-Policy", "no-referrer")
	_ = embedPage.Execute(w, map[string]any{
		"Title":  "Video",
		"Bundle": s.playerURL,
		"Src":    fmt.Sprintf("%s/%s?%s", prefix, file, signed),
		"Poster": fmt.Sprintf("%s/poster.jpg?%s", prefix, signed),
	})
}
