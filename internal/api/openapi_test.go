package api

import (
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ebnsina/alchemist/internal/modules/live"
)

// An API spec that drifts from the router is worse than no spec: customers build
// against documented behaviour that does not exist, or miss endpoints that do.
//
// This walks the real router and the checked-in spec and requires them to agree in
// both directions. It compares paths and methods only -- schemas and descriptions
// still need review.
func TestOpenAPIMatchesRouter(t *testing.T) {
	raw, err := os.ReadFile("../../api/openapi.yaml")
	if err != nil {
		t.Fatalf("reading spec: %v", err)
	}
	documented := parseSpecRoutes(string(raw))

	// A server with nil dependencies is enough: only route registration is inspected.
	// The accounts and live surfaces are mounted deliberately — without an origin or
	// an ingest host they are absent from the router, and the spec drift they can
	// carry would go unnoticed.
	routes := map[string]bool{}
	srv := &Server{webOrigins: []string{"https://example.test"},
		live:        live.New(nil, "ingest.example", nil, nil),
		playerURL:   "https://cdn.example/player/v1/alchemist-player.js",
		authLimiter: newAuthLimiter(10, time.Minute)}
	router, ok := srv.Routes().(chi.Routes)
	if !ok {
		t.Fatal("router does not expose chi.Routes")
	}
	err = chi.Walk(router, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		route = strings.TrimSuffix(route, "/")
		if route == "" {
			return nil
		}
		routes[method+" "+normalise(route)] = true
		return nil
	})
	if err != nil {
		t.Fatalf("walking router: %v", err)
	}

	for r := range routes {
		if !documented[r] {
			t.Errorf("route %s is served but missing from api/openapi.yaml.\n"+
				"Document it in the same change that added it.", r)
		}
	}
	for d := range documented {
		if !routes[d] {
			t.Errorf("api/openapi.yaml documents %s, which the router does not serve.\n"+
				"Remove it, or restore the route.", d)
		}
	}
}

// chi writes wildcards as {name}; the spec uses the same form, so only trailing
// slashes and chi's regex constraints need normalising.
var chiParam = regexp.MustCompile(`\{([a-zA-Z0-9_]+)(:[^}]+)?\}`)

func normalise(route string) string {
	return chiParam.ReplaceAllString(route, "{$1}")
}

var (
	specPath   = regexp.MustCompile(`^  (/[^:]*):\s*$`)
	specMethod = regexp.MustCompile(`^    (get|post|put|patch|delete|head|options):\s*$`)
)

func parseSpecRoutes(spec string) map[string]bool {
	out := map[string]bool{}
	var current string
	inPaths := false
	for _, line := range strings.Split(spec, "\n") {
		if strings.HasPrefix(line, "paths:") {
			inPaths = true
			continue
		}
		if inPaths && len(line) > 0 && line[0] != ' ' {
			inPaths = false
		}
		if !inPaths {
			continue
		}
		if m := specPath.FindStringSubmatch(line); m != nil {
			current = strings.TrimSuffix(m[1], "/")
			continue
		}
		if m := specMethod.FindStringSubmatch(line); m != nil && current != "" {
			out[strings.ToUpper(m[1])+" "+current] = true
		}
	}
	return out
}
