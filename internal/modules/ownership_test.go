package modules_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// A module owns its tables and nobody else writes them. That is the rule that
// decides whether extracting a module is a copy or an archaeology project, and
// import analysis cannot see it: a module reaching into another's tables is a join
// that becomes a network call later.
//
// Add a module here when you add one. An unlisted module fails rather than passing
// silently, because silence is how the list goes stale.
var moduleTables = map[string][]string{
	"delivery": {}, // stateless by design
	"live":     {"live_streams", "live_sessions"},
}

// Shared reference data belongs to the platform, is read by everyone, and is written
// only through the account surface.
var sharedTables = []string{"tenants", "tenant_limits", "ladder_profiles", "api_keys"}

// Only literals that look like SQL are inspected, and only the table position of the
// four clauses that name one. A false positive is better than no enforcement.
var (
	looksLikeSQL = regexp.MustCompile(`(?i)\b(select|insert|update|delete)\b`)
	tableRef     = regexp.MustCompile(`(?i)\b(from|into|update|join)\s+([a-z_][a-z0-9_]*)`)
)

func TestModulesOnlyNameTablesTheyOwn(t *testing.T) {
	entries, err := filepath.Glob("*")
	if err != nil {
		t.Fatal(err)
	}

	for _, dir := range entries {
		info, err := filepath.Glob(filepath.Join(dir, "*.go"))
		if err != nil || len(info) == 0 {
			continue
		}
		module := filepath.Base(dir)
		owned, listed := moduleTables[module]
		if !listed {
			t.Errorf("module %q is not in moduleTables.\n"+
				"List the tables it owns, or {} if it owns none, so a later table "+
				"added to it cannot slip in unnoticed.", module)
			continue
		}
		allowed := map[string]bool{}
		for _, name := range append(append([]string{}, owned...), sharedTables...) {
			allowed[name] = true
		}

		err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
				return err
			}
			fset := token.NewFileSet()
			f, perr := parser.ParseFile(fset, path, nil, 0)
			if perr != nil {
				return perr
			}
			ast.Inspect(f, func(n ast.Node) bool {
				lit, ok := n.(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return true
				}
				sql, uerr := strconv.Unquote(lit.Value)
				if uerr != nil || !looksLikeSQL.MatchString(sql) {
					return true
				}
				for _, m := range tableRef.FindAllStringSubmatch(sql, -1) {
					table := strings.ToLower(m[2])
					if !allowed[table] {
						t.Errorf("%s names table %q, which module %q does not own.\n"+
							"Declare an interface for what it needs and wire an "+
							"adapter in internal/adapters, or add the table to "+
							"moduleTables if ownership really moved.",
							path, table, module)
					}
				}
				return true
			})
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
