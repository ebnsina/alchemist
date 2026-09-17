package delivery

import (
	"net/http"
	"net/url"
	"strings"
)

// Origin locking: a token may name the one web origin it plays on.
//
// This is the control every commercial platform ships and this one did not. Playback
// answered Access-Control-Allow-Origin: * and checked no Referer, so a link lifted out
// of a customer's page played inside any page on any site for the life of the token --
// long enough to build a competing site around somebody else's course.
//
// The origin travels inside the signature rather than being looked up per tenant,
// because the edge decides this. njs verifies in the worker with no database contact,
// and a cached slice never reaches the origin at all, so a rule the edge cannot read
// is a rule that stops applying the moment the cache is warm.

// originAllowed reports whether a request may be served under a token locked to want.
//
// An empty want is an unlocked token and allows anything, which is what every existing
// link is. Otherwise the browser has to say where it is: no Origin and no Referer means
// refused, because "send no header" is otherwise the whole bypass.
//
// The consequence, stated rather than discovered: a locked token cannot be played by
// a native app, curl, or anything else that sends neither header. A customer with an
// iOS or Android app leaves the lock off, and their protection is the signature and
// its expiry, as before.
func originAllowed(r *http.Request, want string) bool {
	if want == "" {
		return true
	}
	if o := r.Header.Get("Origin"); o != "" {
		return strings.EqualFold(o, want)
	}
	// Safari omits Origin on media subrequests that are not CORS, so Referer is the
	// fallback rather than a nicety. Only its scheme and host are compared: the path
	// is the customer's page, which is none of our business.
	if ref := r.Header.Get("Referer"); ref != "" {
		u, err := url.Parse(ref)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return false
		}
		return strings.EqualFold(u.Scheme+"://"+u.Host, want)
	}
	return false
}

// NormalizeOrigin reduces a customer-supplied address to the exact string a browser
// puts in an Origin header: scheme and host, lowercased, no path, no trailing slash.
// Stored and signed in this form so the comparison is never doing normalisation.
func NormalizeOrigin(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return "", false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", false
	}
	return strings.ToLower(u.Scheme + "://" + u.Host), true
}

// corsOrigin is what to answer Access-Control-Allow-Origin with.
//
// A locked token echoes its own origin and nothing else; an unlocked one keeps the
// wildcard, because a player embedded anywhere is the product for everybody who has
// not asked for a lock. Vary: Origin goes with it or a cache hands one site's answer
// to another.
func corsOrigin(locked string) string {
	if locked == "" {
		return "*"
	}
	return locked
}
