package pipeline

import (
	"errors"
	"os"
	"path/filepath"
	"time"
)

// WorkDirTTL is how long an entry in the working directory can still belong to a
// running job. Every job that creates one is bounded -- transcode and JIT by
// JobTimeout, a broadcast by live.Timeout, an edit by thirty minutes -- so the
// longest a directory can be in use is JobTimeout, and an hour past that means it
// belongs to nothing.
const WorkDirTTL = JobTimeout + time.Hour

// PruneWorkDir removes working directories left behind by jobs that died.
//
// `defer os.RemoveAll` cleans up a job that returns. It does not run on an OOM kill,
// a SIGKILL, or a deploy in the middle of a six-hour encode, and nothing else ever
// looked at this directory -- so every crash leaked a source, a mezzanine and a full
// chunk set until the disk filled and every later job failed on write.
//
// By age, not by emptying the directory at startup: a live broadcast or an encode on
// another process sharing the box must not have its files deleted underneath it. A
// directory's modification time is when it was created, which is the age wanted here.
func PruneWorkDir(dir string, ttl time.Duration) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	cutoff := time.Now().Add(-ttl)
	removed := 0
	var errs []error
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue // vanished between the listing and the stat: nothing to remove
		}
		if info.ModTime().After(cutoff) {
			continue
		}
		if err := os.RemoveAll(filepath.Join(dir, e.Name())); err != nil {
			errs = append(errs, err)
			continue
		}
		removed++
	}
	return removed, errors.Join(errs...)
}
