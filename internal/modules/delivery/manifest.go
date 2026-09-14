package delivery

import (
	"regexp"
	"strings"
)

var (
	// URI="segment.m3u8" inside EXT-X-MAP, EXT-X-MEDIA and friends.
	hlsAttrURI = regexp.MustCompile(`URI="([^"]+)"`)
	dashBase   = regexp.MustCompile(`<BaseURL>([^<]+)</BaseURL>`)
)

// signManifest propagates the playback query string onto every relative URI a
// manifest references. Without this a player authorizes the master playlist and then
// gets 403 on every child playlist and segment, because those are fetched as bare
// relative paths carrying no signature.
func signManifest(body []byte, query, file string) []byte {
	if query == "" {
		return body
	}
	if strings.HasSuffix(file, ".mpd") {
		// A raw & is not well-formed XML, so the query must be entity-escaped or
		// every DASH player fails to parse the manifest at all.
		xmlQuery := strings.ReplaceAll(query, "&", "&amp;")
		return dashBase.ReplaceAllFunc(body, func(m []byte) []byte {
			sub := dashBase.FindSubmatch(m)
			return []byte("<BaseURL>" + appendQuery(string(sub[1]), xmlQuery) + "</BaseURL>")
		})
	}

	lines := strings.Split(string(body), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == "":
			continue
		case strings.HasPrefix(trimmed, "#"):
			lines[i] = hlsAttrURI.ReplaceAllStringFunc(line, func(m string) string {
				sub := hlsAttrURI.FindStringSubmatch(m)
				return `URI="` + appendQuery(sub[1], query) + `"`
			})
		default:
			lines[i] = appendQuery(trimmed, query)
		}
	}
	return []byte(strings.Join(lines, "\n"))
}

// appendQuery leaves absolute URLs alone; only our own relative paths are signed.
func appendQuery(uri, query string) string {
	if strings.Contains(uri, "://") || strings.Contains(uri, "?") {
		return uri
	}
	return uri + "?" + query
}
