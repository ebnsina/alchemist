package pipeline

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
)

// Caption is one subtitle track: a sidecar WebVTT the customer uploaded, referenced
// from the manifests rather than burned into the picture. Burning in would mean
// re-encoding the whole ladder per language and leaving the viewer no way to turn it
// off; a sidecar rewrites two small playlists and evicts nothing from any edge cache.
type Caption struct {
	Language string
	Label    string
}

// CaptionPrefix is the object-name stem for a track, shared by the API that uploads
// it and the manifests that point at it. One place, because a mismatch is a 404 that
// only shows up as "subtitles do not work".
func CaptionKey(prefix, language string) string {
	return fmt.Sprintf("%s/subs_%s.vtt", prefix, language)
}

func captionPlaylistName(language string) string { return "subs_" + language + ".m3u8" }
func captionVTTName(language string) string      { return "subs_" + language + ".vtt" }

// captionsFor reads an asset's tracks, ordered so the menu a viewer sees is stable
// between reloads rather than following whatever order the rows came back in.
func (w *TranscodeWorker) captionsFor(ctx context.Context, tenantID, assetID string) ([]Caption, error) {
	var out []Caption
	err := w.DB.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx,
			`select language, label from captions
			  where asset_id = $1 order by language`, assetID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var c Caption
			if err := rows.Scan(&c.Language, &c.Label); err != nil {
				return err
			}
			out = append(out, c)
		}
		return rows.Err()
	})
	return out, err
}

// captionPlaylist is the media playlist HLS insists on for a subtitle track: a player
// will not read a .vtt the master points at directly, it wants a playlist that names
// it. The whole track is one segment, which is what a VOD subtitle file is.
func captionPlaylist(language string, durationSec float64) []byte {
	if durationSec <= 0 {
		durationSec = 1
	}
	var b bytes.Buffer
	fmt.Fprintf(&b, "#EXTM3U\n#EXT-X-VERSION:6\n#EXT-X-PLAYLIST-TYPE:VOD\n"+
		"#EXT-X-TARGETDURATION:%d\n#EXTINF:%.3f,\n%s\n#EXT-X-ENDLIST\n",
		int(math.Ceil(durationSec)), durationSec, captionVTTName(language))
	return b.Bytes()
}

// writeCaptionPlaylists puts one media playlist next to each track. Rewritten on every
// republish rather than at upload: the duration is the asset's, and a track uploaded
// before the encode finished would otherwise carry a playlist claiming a duration of
// nothing.
func (w *TranscodeWorker) writeCaptionPlaylists(ctx context.Context, prefix string,
	captions []Caption, durationSec float64) error {
	for _, c := range captions {
		if err := w.Store.Put(ctx, prefix+"/"+captionPlaylistName(c.Language),
			bytes.NewReader(captionPlaylist(c.Language, durationSec)),
			"application/vnd.apple.mpegurl"); err != nil {
			return fmt.Errorf("publish %s subtitles: %w", c.Language, err)
		}
	}
	return nil
}

// hlsCaptionMedia is the EXT-X-MEDIA block naming every track, and the attribute each
// variant needs to point at the group. Returned together because a group declared and
// never referenced is silently ignored by players -- the tracks exist and no menu
// shows them.
func hlsCaptionMedia(captions []Caption) (media, attr string) {
	if len(captions) == 0 {
		return "", ""
	}
	var b strings.Builder
	for i, c := range captions {
		// Nothing is DEFAULT: subtitles a viewer did not ask for, switched on for
		// them, is the behaviour people turn off and never turn back on.
		fmt.Fprintf(&b, "#EXT-X-MEDIA:TYPE=SUBTITLES,GROUP-ID=\"subs\",NAME=%q,"+
			"LANGUAGE=%q,DEFAULT=NO,AUTOSELECT=%s,URI=%q\n",
			c.Label, c.Language, boolYes(i == 0), captionPlaylistName(c.Language))
	}
	b.WriteString("\n")
	return b.String(), `,SUBTITLES="subs"`
}

func boolYes(v bool) string {
	if v {
		return "YES"
	}
	return "NO"
}

// dashCaptionSets are the text AdaptationSets, one per track, inserted before the
// Period closes. Kept out of the video AdaptationSet's representation placeholder on
// purpose: a text track is not a rendition of the video and a player looking for
// subtitles never reads that set.
func dashCaptionSets(captions []Caption) string {
	var b strings.Builder
	for i, c := range captions {
		fmt.Fprintf(&b, `    <AdaptationSet id="%d" contentType="text" mimeType="text/vtt" lang="%s">
      <Role schemeIdUri="urn:mpeg:dash:role:2011" value="subtitle"/>
      <Representation id="subs-%s" bandwidth="1000">
        <BaseURL>%s</BaseURL>
      </Representation>
    </AdaptationSet>
`, 900+i, xmlAttr(c.Language), xmlAttr(c.Language), xmlAttr(captionVTTName(c.Language)))
	}
	return b.String()
}

// xmlAttr keeps a language tag or filename from breaking the manifest. Language tags
// are validated at the API boundary, so this is the second line rather than the first.
func xmlAttr(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(s)
}

// RepublishArgs rewrites an asset's manifests from whatever is currently stored.
//
// It exists so uploading or removing a subtitle track can refresh the manifests
// without the API reaching into the transcode worker's storage and key handling. The
// media is untouched either way: only master.m3u8, the MPD and the small subtitle
// playlists change.
type RepublishArgs struct {
	AssetID  string `json:"asset_id"`
	TenantID string `json:"tenant_id"`
}

func (RepublishArgs) Kind() string { return "republish_manifests" }

func (RepublishArgs) InsertOpts() river.InsertOpts {
	// On the IO queue, not encode: it moves two small files and must not wait behind
	// an hour of transcoding when a customer has just added subtitles.
	return river.InsertOpts{Queue: QueueIO, MaxAttempts: 3}
}

type RepublishWorker struct {
	river.WorkerDefaults[RepublishArgs]
	*TranscodeWorker
}

// Declared because the name is promoted from both embedded types at the same depth,
// which drops it from the method set and leaves this short of river.Worker.
func (w *RepublishWorker) Timeout(*river.Job[RepublishArgs]) time.Duration {
	return 5 * time.Minute
}

func (w *RepublishWorker) Work(ctx context.Context, job *river.Job[RepublishArgs]) error {
	a := job.Args
	// An asset with nothing published yet has no manifest to refresh, and the encode
	// that is coming will write one including whatever tracks exist by then.
	var published bool
	if err := w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`select exists (select 1 from renditions
			                 where asset_id = $1 and state = 'ready')`,
			a.AssetID).Scan(&published)
	}); err != nil {
		return err
	}
	if !published {
		return nil
	}
	return w.republish(ctx, a.TenantID, a.AssetID)
}
