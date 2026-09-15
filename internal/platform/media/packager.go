package media

import (
	"context"
	"encoding/hex"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
)

// PackagerBinary is shaka-packager. ffmpeg can emit HLS and DASH, but shaka is
// correct about CMAF and CENC structure, and that correctness is what lets studio
// DRM be added later as key management instead of re-packaging the library.
var PackagerBinary = "packager"

type StreamKind string

const (
	StreamVideo StreamKind = "video"
	StreamAudio StreamKind = "audio"
)

// Input is one elementary stream to package.
type Input struct {
	Path   string
	Kind   StreamKind
	Name   string // e.g. "720p" or "audio"
	Height int    // video only, for the DASH/HLS variant metadata
}

type Result struct {
	MasterPlaylist string
	MPD            string
}

// Encryption configures content protection.
//
// The scheme is cenc, not cbcs: Clear Key -- the only key system that works without
// a licence vendor -- always supports cenc, and cbcs left the stream unplayable in
// Chrome and Firefox. Both are CENC common encryption, so studio DRM later is still
// key management rather than re-packaging the library.
type Encryption struct {
	KeyID  []byte
	Key    []byte
	KeyURI string
	// ClearLeadSeconds leaves the opening seconds unencrypted so playback starts
	// while the key request is still in flight. Zero encrypts everything.
	ClearLeadSeconds int
}

// Package muxes inputs into single-file CMAF and writes HLS + DASH manifests.
//
// Each rendition becomes ONE file addressed by byte range rather than thousands of
// small segment objects. That keeps object counts ~1000x lower, lets edge caches
// fill in large sequential reads, and makes a rendition one object to delete.
//
// This is a library call, not a pipeline step: the live path will call it per
// segment with the same signature.
func Package(ctx context.Context, inputs []Input, outDir string, enc *Encryption) (*Result, error) {
	if len(inputs) == 0 {
		return nil, fmt.Errorf("%w: no inputs", ErrPackageFailed)
	}

	args := make([]string, 0, len(inputs)+6)
	for _, in := range inputs {
		ext := ".cmfv"
		descriptor := fmt.Sprintf("in=%s,stream=video,output=%s,playlist_name=%s.m3u8",
			in.Path, filepath.Join(outDir, in.Name+ext), filepath.Join(outDir, in.Name))
		if in.Kind == StreamAudio {
			ext = ".cmfa"
			descriptor = fmt.Sprintf(
				"in=%s,stream=audio,output=%s,playlist_name=%s.m3u8,hls_group_id=audio,hls_name=AUDIO",
				in.Path, filepath.Join(outDir, in.Name+ext), filepath.Join(outDir, in.Name))
		}
		args = append(args, descriptor)
	}

	master := filepath.Join(outDir, "master.m3u8")
	mpd := filepath.Join(outDir, "manifest.mpd")
	args = append(args,
		"--hls_master_playlist_output", master,
		"--mpd_output", mpd,
		"--segment_duration", fmt.Sprintf("%d", GOPSeconds),
	)

	if enc != nil {
		args = append(args,
			"--enable_raw_key_encryption",
			"--keys", fmt.Sprintf("label=:key_id=%s:key=%s",
				hex.EncodeToString(enc.KeyID), hex.EncodeToString(enc.Key)),
			"--protection_scheme", "cenc",
			"--clear_lead", strconv.Itoa(enc.ClearLeadSeconds),
		)
		if enc.KeyURI != "" {
			args = append(args, "--hls_key_uri", enc.KeyURI)
		}
	}

	if out, err := exec.CommandContext(ctx, PackagerBinary, args...).CombinedOutput(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrPackageFailed, truncate(string(out)))
	}
	return &Result{MasterPlaylist: master, MPD: mpd}, nil
}
