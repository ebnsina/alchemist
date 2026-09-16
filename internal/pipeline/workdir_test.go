package pipeline

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPruneWorkDirKeepsWhatCouldStillBeRunning(t *testing.T) {
	dir := t.TempDir()

	live := filepath.Join(dir, "asset-running")
	if err := os.MkdirAll(filepath.Join(live, "chunks"), 0o750); err != nil {
		t.Fatal(err)
	}
	abandoned := filepath.Join(dir, "asset-abandoned")
	if err := os.MkdirAll(filepath.Join(abandoned, "chunks"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(abandoned, "mezzanine.mp4"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-WorkDirTTL - time.Minute)
	if err := os.Chtimes(abandoned, old, old); err != nil {
		t.Fatal(err)
	}

	n, err := PruneWorkDir(dir, WorkDirTTL)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("removed %d directories, want 1", n)
	}
	if _, err := os.Stat(abandoned); !os.IsNotExist(err) {
		t.Fatal("an abandoned directory older than a job can live was kept")
	}
	if _, err := os.Stat(live); err != nil {
		t.Fatalf("a directory young enough to belong to a running job was removed: %v", err)
	}
}
