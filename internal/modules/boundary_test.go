package modules_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const modulePrefix = "github.com/ebnsina/alchemist/internal/modules/"

// Modules may depend on the platform packages listed here. Anything else would have
// to be carried along when the module is extracted, so it needs a deliberate decision
// rather than an accidental import.
var platformAllowed = map[string]bool{
	"httpx":   true, // response conventions, no state
	"signing": true, // playback signatures, pure
	// A module that owns tables carries a database handle when it is extracted, the
	// same way any service carries its driver. platform/db is a thin pgx wrapper and
	// its AsTenant is what enforces RLS.
	"db": true,
	// Ladder types and the encode primitives a realtime rung is built from. Pure
	// value types and argv construction, no state and no connections.
	"media": true,
}

// A module must not import another module. Cross-module needs go through an interface
// the consumer declares and main wires up, which is what keeps extraction mechanical:
// copy the module plus its adapters, and nothing else comes with it.
//
// Without this test the boundary erodes within weeks -- one "just this once" import
// turns a copy-out into a refactor, quietly and with nothing failing.
func TestModulesDoNotImportEachOther(t *testing.T) {
	root := "."
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		module := e.Name()
		err := filepath.WalkDir(filepath.Join(root, module), walkGoFiles(t, module))
		if err != nil {
			t.Fatal(err)
		}
	}
}

func walkGoFiles(t *testing.T, module string) fs.WalkDirFunc {
	t.Helper()
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		fset := token.NewFileSet()
		f, perr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if perr != nil {
			return perr
		}
		for _, imp := range f.Imports {
			p := strings.Trim(imp.Path.Value, `"`)

			if strings.HasPrefix(p, modulePrefix) {
				other := strings.SplitN(strings.TrimPrefix(p, modulePrefix), "/", 2)[0]
				if other != module {
					t.Errorf("%s imports module %q.\n"+
						"Modules must not depend on each other. Declare an interface in "+
						"%s describing what it needs and wire the implementation in "+
						"cmd/ via internal/adapters.", path, other, module)
				}
			}

			const platformPrefix = "github.com/ebnsina/alchemist/internal/platform/"
			if strings.HasPrefix(p, platformPrefix) {
				pkg := strings.SplitN(strings.TrimPrefix(p, platformPrefix), "/", 2)[0]
				if !platformAllowed[pkg] {
					t.Errorf("%s imports platform/%s.\n"+
						"Modules depend on interfaces they declare, not on concrete "+
						"infrastructure, so that extracting the module does not drag "+
						"the platform with it. Add an adapter in internal/adapters "+
						"instead, or add %q to platformAllowed if it is genuinely "+
						"dependency-free.", path, pkg, pkg)
				}
			}

			if strings.HasPrefix(p, "github.com/ebnsina/alchemist/internal/adapters") {
				t.Errorf("%s imports internal/adapters.\n"+
					"Adapters bind modules to the platform and must only be imported "+
					"by cmd/. A module importing them inverts the dependency.", path)
			}
		}
		return nil
	}
}
