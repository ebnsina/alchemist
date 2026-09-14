package pipeline

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5"

	"github.com/ebnsina/alchemist/internal/platform/media"
)

// buildRendition encodes one rung from the mezzanine and packages it on its own.
//
// Packaged standalone rather than alongside the existing renditions: re-running the
// packager over everything would rewrite every .cmfv, changing ETags and evicting the
// whole asset from every edge cache -- a large egress bill to add one rung.
func (w *TranscodeWorker) buildRendition(ctx context.Context, a JITArgs, mezz, dir string, rung media.Rung) error {
	probe, err := media.Inspect(ctx, mezz)
	if err != nil {
		return err
	}

	full := filepath.Join(dir, fmt.Sprintf("%dp.mp4", rung.Height))
	chunks := media.PlanChunks(probe.DurationSec)
	paths := make([]string, len(chunks))
	for i, c := range chunks {
		out := filepath.Join(dir, fmt.Sprintf("chunk-%05d.mp4", c.Index))
		if err := media.EncodeChunk(ctx, mezz, c, rung, out); err != nil {
			return err
		}
		paths[i] = out
	}
	if err := media.Stitch(ctx, paths, full); err != nil {
		return err
	}
	if err := media.VerifyStitch(ctx, full, probe.DurationSec); err != nil {
		return err
	}

	outDir := filepath.Join(dir, "out")
	if err := os.MkdirAll(outDir, 0o750); err != nil {
		return err
	}

	name := fmt.Sprintf("%dp", rung.Height)
	enc, err := w.encryptionFor(ctx, a.TenantID, a.AssetID)
	if err != nil {
		return err
	}
	if _, err := media.Package(ctx, []media.Input{{
		Path: full, Kind: media.StreamVideo, Name: name, Height: rung.Height,
	}}, outDir, enc); err != nil {
		return err
	}

	prefix := fmt.Sprintf("cmaf/%s/%s", a.TenantID, a.AssetID)
	for _, f := range []string{name + ".cmfv", name + ".m3u8"} {
		src := filepath.Join(outDir, f)
		fh, err := os.Open(src)
		if err != nil {
			return err
		}
		err = w.Store.Put(ctx, prefix+"/"+f, fh, contentType(f))
		fh.Close()
		if err != nil {
			return err
		}
	}

	width := rung.Height * probe.Width / probe.Height
	if width%2 != 0 {
		width++
	}
	return w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`update renditions
			    set state = 'ready', object_key = $4, width = $5,
			        codec_string = $6, avg_bandwidth_bps = $7, lazy = false
			  where asset_id = $1 and height = $2 and codec = $3`,
			a.AssetID, rung.Height, rung.Codec,
			fmt.Sprintf("%s/%dp.cmfv", prefix, rung.Height),
			width, codecString(rung), rung.MaxrateBPS)
		return err
	})
}

// codecString is the RFC 6381 identifier a player needs before it will consider a
// variant. Derived from the H.264 profile and a level high enough for the rung.
func codecString(r media.Rung) string {
	profileIDC := map[string]string{"baseline": "42", "main": "4d", "high": "64"}[r.Profile]
	if profileIDC == "" {
		profileIDC = "4d"
	}
	level := "1e" // 3.0
	switch {
	case r.Height > 1080:
		level = "33" // 5.1
	case r.Height > 720:
		level = "28" // 4.0
	case r.Height > 480:
		level = "1f" // 3.1
	}
	return fmt.Sprintf("avc1.%s40%s", profileIDC, level)
}

// republish rewrites the HLS master playlist from what is currently ready.
//
// ponytail: HLS only. The DASH manifest still lists the rungs present at ingest and
// is refreshed only by a full re-package. BD playback is HLS on both Android and iOS,
// so this is the path that matters here; composing the MPD needs each
// Representation's indexRange, which would have to be captured at package time.
func (w *TranscodeWorker) republish(ctx context.Context, tenantID, assetID, dir string) error {
	type variant struct {
		height, width, bandwidth int
		codec                    string
	}
	var variants []variant
	hasAudio := false

	if err := w.DB.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx,
			`select height, coalesce(width,0), coalesce(avg_bandwidth_bps, bitrate_bps),
			        coalesce(codec_string,'avc1.4d401e')
			   from renditions
			  where asset_id = $1 and state = 'ready'
			  order by height`, assetID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var v variant
			if err := rows.Scan(&v.height, &v.width, &v.bandwidth, &v.codec); err != nil {
				return err
			}
			variants = append(variants, v)
		}
		return rows.Err()
	}); err != nil {
		return err
	}
	if len(variants) == 0 {
		return fmt.Errorf("no ready renditions for %s", assetID)
	}

	prefix := fmt.Sprintf("cmaf/%s/%s", tenantID, assetID)
	if _, err := w.Store.GetPassthrough(ctx, prefix+"/audio.m3u8", ""); err == nil {
		hasAudio = true
	}

	sort.Slice(variants, func(i, j int) bool { return variants[i].height < variants[j].height })

	var b bytes.Buffer
	b.WriteString("#EXTM3U\n#EXT-X-VERSION:6\n#EXT-X-INDEPENDENT-SEGMENTS\n\n")
	audioAttr := ""
	if hasAudio {
		b.WriteString(`#EXT-X-MEDIA:TYPE=AUDIO,URI="audio.m3u8",GROUP-ID="audio",` +
			`NAME="AUDIO",DEFAULT=YES,AUTOSELECT=YES,CHANNELS="2"` + "\n\n")
		audioAttr = `,AUDIO="audio"`
	}
	for _, v := range variants {
		codecs := v.codec
		if hasAudio {
			codecs += ",mp4a.40.2"
		}
		fmt.Fprintf(&b,
			"#EXT-X-STREAM-INF:BANDWIDTH=%d,CODECS=\"%s\",RESOLUTION=%dx%d%s\n%dp.m3u8\n",
			v.bandwidth, codecs, v.width, v.height, audioAttr, v.height)
	}

	if err := w.Store.Put(ctx, prefix+"/master.m3u8", bytes.NewReader(b.Bytes()),
		"application/vnd.apple.mpegurl"); err != nil {
		return err
	}

	// Everything the profile asked for now exists, so the asset is fully ready.
	return w.DB.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`update assets set state = case
			     when exists (select 1 from renditions
			                   where asset_id = $1 and state <> 'ready')
			     then 'partially_ready'::asset_state else 'ready'::asset_state end,
			    updated_at = now()
			  where id = $1 and state <> 'failed'`, assetID)
		return err
	})
}

// encryptionFor returns the asset's existing content key.
//
// Reused, never regenerated: the player fetches one key for the asset and uses it for
// every rendition. A fresh key here would produce a rung nothing can decrypt, and the
// failure would look like a corrupt file rather than a key mismatch.
func (w *TranscodeWorker) encryptionFor(ctx context.Context, tenantID, assetID string) (*media.Encryption, error) {
	var keyID, wrapped, nonce []byte
	err := w.DB.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`select key_id, wrapped_key, nonce from content_keys where asset_id = $1`,
			assetID).Scan(&keyID, &wrapped, &nonce)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // asset was published unencrypted
	}
	if err != nil {
		return nil, err
	}

	key, err := w.Keys.Unwrap(wrapped, nonce, assetID)
	if err != nil {
		return nil, fmt.Errorf("unwrap content key: %w", err)
	}
	return &media.Encryption{
		KeyID: keyID, Key: key, KeyURI: "key", ClearLeadSeconds: 0,
	}, nil
}
