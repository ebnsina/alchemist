// Command alchemist-transcode runs the transcode chain on a local file. It exists to
// exercise and debug the media library without a database, queue, or object store.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/ebnsina/alchemist/internal/platform/media"
)

func main() {
	workDir := flag.String("work", "", "working directory for intermediates and output")
	ladder := flag.String("ladder", "", "ladder profile JSON (defaults to a 360p/720p pair)")
	flag.Parse()

	if flag.NArg() != 1 || *workDir == "" {
		fmt.Fprintln(os.Stderr, "usage: alchemist-transcode -work DIR [-ladder JSON] SOURCE")
		os.Exit(2)
	}

	raw := []byte(`[
	  {"height":360,"codec":"h264","profile":"main","preset":"veryfast","crf":26,"maxrate_bps":700000},
	  {"height":720,"codec":"h264","profile":"main","preset":"veryfast","crf":24,"maxrate_bps":2200000}
	]`)
	if *ladder != "" {
		raw = []byte(*ladder)
	}
	rungs, err := media.ParseLadder(raw)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ladder:", err)
		os.Exit(1)
	}

	res, err := media.Transcode(context.Background(), flag.Arg(0), *workDir, rungs, media.DefaultOptions())
	if err != nil {
		fmt.Fprintln(os.Stderr, "transcode:", err)
		os.Exit(1)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(res)
}
