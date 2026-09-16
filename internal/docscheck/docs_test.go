package docscheck_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Every path a doc mentions in backticks, plus every relative markdown link.
var (
	backtickPath = regexp.MustCompile("`((?:internal|cmd|deploy|docs|test|web|player|src|scripts)/[A-Za-z0-9_./-]+|[A-Za-z0-9_-]+\\.(?:go|sql|js|conf|service|sh|py|yml))`")
	mdLink       = regexp.MustCompile(`\]\(([^)h][^)]*)\)`)
)

// Docs rot silently. A path renamed during a refactor leaves prose that still reads
// plausibly and sends the next person somewhere that no longer exists -- which is
// worse than no documentation, because it is confidently wrong.
//
// This makes the rot fail the build instead. It only checks paths, which is the part
// that can be verified mechanically; prose still needs judgement.
func TestDocsReferenceRealPaths(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}

	var docs []string
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "bin", ".local", "node_modules":
				return filepath.SkipDir
			}
			return nil
		}
		// llms.txt is markdown despite the extension, and it is the file an agent
		// integrating against this reads first, so a dead link there is the most
		// expensive kind.
		// A changelog records what a release did, so a "Removed" entry naming a file
		// that no longer exists is correct rather than rotten.
		if d.Name() == "CHANGELOG.md" {
			return nil
		}
		if strings.HasSuffix(path, ".md") || d.Name() == "llms.txt" {
			docs = append(docs, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) == 0 {
		t.Fatal("no markdown files found; the walk root is wrong")
	}

	for _, doc := range docs {
		body, err := os.ReadFile(doc)
		if err != nil {
			t.Fatal(err)
		}
		rel, _ := filepath.Rel(root, doc)
		dir := filepath.Dir(doc)

		check := func(ref string) {
			ref = strings.TrimSpace(ref)
			if ref == "" || strings.HasPrefix(ref, "#") {
				return
			}
			ref = strings.SplitN(ref, "#", 2)[0]
			if ref == "" {
				return
			}
			// Resolve against the document first, then the repo root, since docs
			// reference their siblings and repo-relative paths interchangeably.
			if _, err := os.Stat(filepath.Join(dir, ref)); err == nil {
				return
			}
			if _, err := os.Stat(filepath.Join(root, ref)); err == nil {
				return
			}
			t.Errorf("%s references %q, which does not exist.\n"+
				"Update the doc in the same change that moved or renamed the file.", rel, ref)
		}

		for _, m := range backtickPath.FindAllStringSubmatch(string(body), -1) {
			check(m[1])
		}
		for _, m := range mdLink.FindAllStringSubmatch(string(body), -1) {
			check(m[1])
		}
	}
}
