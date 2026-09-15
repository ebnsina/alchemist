// Package pipeline turns transcoding into durable background work.
package pipeline

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"

	"github.com/ebnsina/alchemist/internal/platform/db"
	"github.com/ebnsina/alchemist/internal/platform/fetch"
	"github.com/ebnsina/alchemist/internal/platform/keys"
	"github.com/ebnsina/alchemist/internal/platform/media"
	"github.com/ebnsina/alchemist/internal/platform/metrics"
	"github.com/ebnsina/alchemist/internal/platform/storage"
)

// Queue classes are sized independently so a bulk import cannot starve playback-
// driven work. Capability tags let GPU nodes join encode_gpu without a scheduler change.
const (
	QueueIO     = "io"
	QueueEncode = "encode_cpu"
)

type TranscodeArgs struct {
	AssetID  string `json:"asset_id"`
	TenantID string `json:"tenant_id"`
}

func (TranscodeArgs) Kind() string { return "transcode" }

func (TranscodeArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: QueueEncode, MaxAttempts: 3}
}

// JobTimeout has to cover the whole chain for the longest asset a tenant can upload,
// including uploading the packaged output. River's default is far too short: a
// multi-hour lecture blows through it, and the job dies partway with output already
// written.
const JobTimeout = 6 * time.Hour

func (w *TranscodeWorker) Timeout(*river.Job[TranscodeArgs]) time.Duration {
	return JobTimeout
}

type TranscodeWorker struct {
	river.WorkerDefaults[TranscodeArgs]
	DB      *db.DB
	Store   *storage.Store
	Keys    *keys.Wrapper
	River   *river.Client[pgx.Tx]
	Metrics *metrics.Registry
	WorkDir string
}

// Work runs the full chain for one asset.
//
// ponytail: one job per asset, encoding chunks in a local pool. The ceiling is a
// single machine's cores and the mezzanine must fit on local disk, which is fine to
// roughly 10k source hours/month. Splitting into (chunk x rendition) jobs that fetch
// byte ranges from object storage is additive — the media primitives, the schema and
// the API contract all stay as they are; see docs/04-roadmap.md phase 5.
func (w *TranscodeWorker) Work(ctx context.Context, job *river.Job[TranscodeArgs]) (err error) {
	a := job.Args

	// An asset must never be left in a non-terminal state. Without this a job that
	// exhausts its retries leaves the asset stuck in "encoding" forever, with no
	// signal to the customer and nothing to retry against.
	defer func() {
		if err != nil && job.Attempt >= job.MaxAttempts {
			w.markFailed(ctx, a, "processing_failed")
		}
	}()

	var profile string
	var sourceKey, sourceURL, bucketSourceID, objectKey *string
	var ladderRaw []byte
	err = w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`select a.source_key, a.source_url, a.bucket_source_id::text,
			        a.source_object_key, a.ladder_profile, p.rungs
			   from assets a join ladder_profiles p on p.name = a.ladder_profile
			  where a.id = $1`, a.AssetID).
			Scan(&sourceKey, &sourceURL, &bucketSourceID, &objectKey, &profile, &ladderRaw)
	})
	if err != nil {
		return fmt.Errorf("load asset %s: %w", a.AssetID, err)
	}

	rungs, err := media.ParseLadder(ladderRaw)
	if err != nil {
		return w.fail(ctx, a, "invalid_ladder_profile", err)
	}

	dir := filepath.Join(w.WorkDir, a.AssetID)
	defer os.RemoveAll(dir)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}

	src := filepath.Join(dir, "source")
	switch {
	case bucketSourceID != nil && objectKey != nil:
		if err := w.pullFromBucket(ctx, a, *bucketSourceID, *objectKey, src); err != nil {
			return err
		}
	case sourceURL != nil && *sourceURL != "":
		if err := w.pull(ctx, a, *sourceURL, src); err != nil {
			return err
		}
	default:
		if err := w.download(ctx, deref(sourceKey), src); err != nil {
			return w.fail(ctx, a, "source_unreadable", err)
		}
	}

	// Hash the source before encoding. At scale, re-uploads and provider migrations
	// mean the same file arrives repeatedly; encoding it twice is pure waste.
	// Scoped to the tenant -- a cross-tenant match would leak the fact that another
	// customer holds the same file.
	sum, err := hashFile(src)
	if err != nil {
		return fmt.Errorf("hash source: %w", err)
	}
	srcBytes := fileBytes(src)
	if existing, err := w.findDuplicate(ctx, a, sum); err != nil {
		return err
	} else if existing != "" {
		return w.linkToDuplicate(ctx, a, existing, sum, srcBytes)
	}

	w.setState(ctx, a, "encoding")

	// Encode cost per source hour is the number the cost model turns on: a regression
	// here shows up as a compute bill months before anyone notices it in a profile.
	stop := w.observeEncode(a)
	defer stop()

	opts := media.DefaultOptions()

	// On by default since migration 033. cenc plus an EME Clear Key licence plays in
	// Chrome, Firefox and Edge with no vendor; what it buys is that a lifted bucket
	// decodes to nothing, not DRM. Safari and iOS cannot play it.
	var encrypt bool
	if err := w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `select encrypt_playback from tenants`).Scan(&encrypt)
	}); err != nil {
		return fmt.Errorf("read playback setting: %w", err)
	}

	if encrypt {
		keyID, key, err := keys.Generate()
		if err != nil {
			return err
		}
		// The key URI is relative so the playback signature is appended to it at
		// serve time, the same way it is for manifests and segments.
		opts.Encrypt = &media.Encryption{
			KeyID: keyID, Key: key, KeyURI: "key", ClearLeadSeconds: 0,
		}

		wrapped, nonce, err := w.Keys.Wrap(key, a.AssetID)
		if err != nil {
			return err
		}
		if err := w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx,
				`insert into content_keys (asset_id, tenant_id, key_id, wrapped_key, nonce)
				 values ($1,$2,$3,$4,$5)
				 on conflict (asset_id) do update set
				   key_id = excluded.key_id, wrapped_key = excluded.wrapped_key,
				   nonce = excluded.nonce`,
				a.AssetID, a.TenantID, keyID, wrapped, nonce)
			return err
		}); err != nil {
			return fmt.Errorf("store content key: %w", err)
		}
	}

	// Progress, written where it can be seen. The rendition rows are created up
	// front in the encoding state with their chunk plan, so a customer watching an
	// hour of video go through does not stare at one unchanging state for an hour.
	opts.OnPlan = func(rungs []media.Rung, chunks []media.Chunk) error {
		return w.recordPlan(ctx, a, rungs, chunks)
	}
	opts.OnChunkDone = func(r media.Rung, c media.Chunk) {
		// Called from several encode goroutines. A progress write that fails is not
		// worth failing the encode over — the work is the point, the counter is not.
		w.recordChunkDone(ctx, a, r, c)
	}

	// Only the eager rungs are encoded now. The rest are recorded as pending and
	// generated when a viewer first asks for them.
	lazy := media.LazyRungs(rungs)
	res, err := media.Transcode(ctx, src, dir, media.Eager(rungs), opts)
	if err != nil {
		return w.fail(ctx, a, errorCode(err), err)
	}

	w.setState(ctx, a, "packaging")
	prefix := fmt.Sprintf("cmaf/%s/%s", a.TenantID, a.AssetID)

	// Stored only when a rung is deferred, and then kept for the life of the asset:
	// studio edits render from it too, and the original is usually already deleted.
	mezzKey := ""
	mezzBytes := int64(0)
	if len(lazy) > 0 {
		mezzKey = fmt.Sprintf("mez/%s/%s/mezzanine.mp4", a.TenantID, a.AssetID)
		fh, err := os.Open(filepath.Join(dir, "mezzanine.mp4"))
		if err != nil {
			return fmt.Errorf("open mezzanine: %w", err)
		}
		err = w.Store.Put(ctx, mezzKey, fh, "video/mp4")
		fh.Close()
		if err != nil {
			return fmt.Errorf("retain mezzanine: %w", err)
		}
		mezzBytes = fileBytes(filepath.Join(dir, "mezzanine.mp4"))
	}
	if err := w.uploadDir(ctx, res.OutDir, prefix); err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	// Only after the renditions and the mezzanine are safely stored.
	// Captured now because it cannot be recovered later: the published media is
	// encrypted and the packager cannot re-read it to rebuild a manifest.
	reps, _ := media.VideoRepresentations(res.Manifest.MPD)
	skeleton, _ := media.MPDSkeleton(res.Manifest.MPD)

	w.applySourceRetention(ctx, a, deref(sourceKey), mezzKey)

	if err := w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		// partially_ready is a real, published state: playback works on the low
		// ladder while the expensive rungs do not exist yet.
		state := "ready"
		if len(lazy) > 0 {
			state = "partially_ready"
		}
		if _, err := tx.Exec(ctx,
			`update assets set state = $6::asset_state, duration_sec = $2, width = $3,
			        height = $4, frame_rate = $5, mezzanine_key = $7,
			        complexity = $8, source_sha256 = $9, dash_skeleton = $10,
			        source_bytes = nullif($11,0)::bigint,
			        mezzanine_bytes = nullif($12,0)::bigint, updated_at = now()
			  where id = $1`,
			a.AssetID, res.Probe.DurationSec, res.Probe.Width,
			res.Probe.Height, res.Probe.FrameRate, state, mezzKey,
			res.Complexity, sum, skeleton, srcBytes, mezzBytes); err != nil {
			return err
		}
		for _, r := range res.Rungs {
			width := r.Height * res.Probe.Width / res.Probe.Height
			if width%2 != 0 {
				width++
			}
			// The update repeats every column, because the row always exists by now --
			// recordPlan created it before the first chunk. Setting state alone left
			// object_key and dash_representation null on every eager rung, and a null
			// dash_representation drops that rung out of the rebuilt DASH manifest.
			if _, err := tx.Exec(ctx,
				`insert into renditions (asset_id, tenant_id, height, codec, bitrate_bps,
				        encoder_version, params_hash, state, object_key, lazy,
				        width, codec_string, avg_bandwidth_bps, dash_representation, bytes)
				 values ($1,$2,$3,$4,$5,$6,$7,'ready',$8,false,$9,$10,$5,$11,
				         nullif($12,0)::bigint)
				 on conflict (asset_id, height, codec) do update set
				   state = 'ready', bytes = excluded.bytes,
				   object_key = excluded.object_key, lazy = false,
				   width = excluded.width, codec_string = excluded.codec_string,
				   avg_bandwidth_bps = excluded.avg_bandwidth_bps,
				   dash_representation = excluded.dash_representation`,
				a.AssetID, a.TenantID, r.Height, r.Codec, r.MaxrateBPS,
				media.EncoderVersion, media.ParamsHash(r),
				fmt.Sprintf("%s/%dp.cmfv", prefix, r.Height),
				width, codecString(r), reps[r.Height],
				fileBytes(filepath.Join(res.OutDir, fmt.Sprintf("%dp.cmfv", r.Height)))); err != nil {
				return err
			}
		}
		// Pending rungs are recorded now so the API can report what is still coming,
		// and so a playback request has a row to flip rather than having to re-derive
		// the profile.
		for _, r := range lazy {
			if _, err := tx.Exec(ctx,
				`insert into renditions (asset_id, tenant_id, height, codec, bitrate_bps,
				        encoder_version, params_hash, state, lazy)
				 values ($1,$2,$3,$4,$5,$6,$7,'pending',true)
				 on conflict (asset_id, height, codec) do nothing`,
				a.AssetID, a.TenantID, r.Height, r.Codec, r.MaxrateBPS,
				media.EncoderVersion, media.ParamsHash(r)); err != nil {
				return err
			}
		}
		// Billing data cannot be backfilled, so it is emitted with the work.
		_, err := tx.Exec(ctx,
			`insert into usage_events (tenant_id, asset_id, kind, quantity, unit)
			 values ($1, $2, 'ingest', $3, 'seconds')`,
			a.TenantID, a.AssetID, res.Probe.DurationSec)
		return err
	}); err != nil {
		return err
	}

	w.emit(ctx, a, "asset.ready", map[string]any{
		"asset_id":         a.AssetID,
		"duration_seconds": res.Probe.DurationSec,
		"width":            res.Probe.Width, "height": res.Probe.Height,
	})
	return nil
}

func (w *TranscodeWorker) download(ctx context.Context, key, dst string) error {
	body, err := w.Store.Get(ctx, key)
	if err != nil {
		return err
	}
	defer body.Close()

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, body)
	return err
}

func (w *TranscodeWorker) uploadDir(ctx context.Context, dir, prefix string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		f, err := os.Open(filepath.Join(dir, e.Name()))
		if err != nil {
			return err
		}
		err = w.Store.Put(ctx, prefix+"/"+e.Name(), f, contentType(e.Name()))
		f.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func contentType(name string) string {
	switch filepath.Ext(name) {
	case ".m3u8":
		return "application/vnd.apple.mpegurl"
	case ".mpd":
		return "application/dash+xml"
	case ".cmfv", ".cmfa", ".mp4":
		return "video/mp4"
	}
	if ct := mime.TypeByExtension(filepath.Ext(name)); ct != "" {
		return ct
	}
	return "application/octet-stream"
}

// recordPlan creates the rendition rows before the work starts and lays down one
// row per chunk, so progress has somewhere to accumulate.
func (w *TranscodeWorker) recordPlan(ctx context.Context, a TranscodeArgs,
	rungs []media.Rung, chunks []media.Chunk) error {
	return w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		for _, r := range rungs {
			var renditionID string
			if err := tx.QueryRow(ctx,
				`insert into renditions (asset_id, tenant_id, height, codec, bitrate_bps,
				        encoder_version, params_hash, state, lazy, chunks_total, chunks_done)
				 values ($1,$2,$3,$4,$5,$6,$7,'encoding',false,$8,0)
				 on conflict (asset_id, height, codec) do update set
				   state = 'encoding', chunks_total = excluded.chunks_total, chunks_done = 0
				 returning id::text`,
				a.AssetID, a.TenantID, r.Height, r.Codec, r.MaxrateBPS,
				media.EncoderVersion, media.ParamsHash(r), len(chunks)).Scan(&renditionID); err != nil {
				return err
			}
			for _, c := range chunks {
				if _, err := tx.Exec(ctx,
					`insert into chunks (rendition_id, idx, tenant_id, start_sec, end_sec)
					 values ($1,$2,$3,$4,$5)
					 on conflict (rendition_id, idx) do update set
					   start_sec = excluded.start_sec, end_sec = excluded.end_sec,
					   completed_at = null`,
					renditionID, c.Index, a.TenantID, c.StartSec, c.EndSec); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// recordChunkDone marks one chunk finished and moves the rendition's counter.
func (w *TranscodeWorker) recordChunkDone(ctx context.Context, a TranscodeArgs,
	r media.Rung, c media.Chunk) {
	_ = w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`with target as (
			   select id from renditions
			    where asset_id = $1 and height = $2 and codec = $3
			 ), marked as (
			   update chunks set completed_at = now()
			    where rendition_id = (select id from target) and idx = $4
			      and completed_at is null
			   returning 1
			 )
			 update renditions set chunks_done = chunks_done + (select count(*) from marked)
			  where id = (select id from target)`,
			a.AssetID, r.Height, r.Codec, c.Index)
		return err
	})
}

func (w *TranscodeWorker) setState(ctx context.Context, a TranscodeArgs, state string) {
	_ = w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`update assets set state = $2::asset_state, updated_at = now() where id = $1`,
			a.AssetID, state)
		return err
	})
}

// fail records a stable code for the customer and stops retrying: a corrupt source
// will not become readable on the third attempt.
func (w *TranscodeWorker) fail(ctx context.Context, a TranscodeArgs, code string, cause error) error {
	w.markFailed(ctx, a, code)
	w.emit(ctx, a, "asset.failed", map[string]any{"asset_id": a.AssetID, "error_code": code})
	return river.JobCancel(fmt.Errorf("%s: %w", code, cause))
}

// markFailed uses a fresh context: the job context may already be cancelled or past
// its deadline, which is exactly when recording the failure matters most.
func (w *TranscodeWorker) markFailed(ctx context.Context, a TranscodeArgs, code string) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
	defer cancel()
	_ = w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`update assets set state = 'failed', error_code = $2, updated_at = now()
			  where id = $1 and state <> 'ready'`, a.AssetID, code)
		return err
	})
}

func errorCode(err error) string {
	switch {
	case errorsIs(err, media.ErrNoVideoStream):
		return "no_video_stream"
	case errorsIs(err, media.ErrUnreadableSource):
		return "source_unreadable"
	case errorsIs(err, media.ErrStitchFailed):
		return "stitch_failed"
	case errorsIs(err, media.ErrPackageFailed):
		return "package_failed"
	default:
		return "encode_failed"
	}
}

func errorsIs(err, target error) bool { return errors.Is(err, target) }

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// pull downloads a customer-supplied URL. Every failure here maps to a stable code,
// because "your link didn't work" needs to say why.
func (w *TranscodeWorker) pull(ctx context.Context, a TranscodeArgs, url, dst string) error {
	max, err := w.maxSourceBytes(ctx, a.TenantID)
	if err != nil {
		return fmt.Errorf("read source limit: %w", err)
	}
	if _, err := fetch.ToFile(ctx, url, dst, max, 2*time.Hour); err != nil {
		switch {
		case errors.Is(err, fetch.ErrBlockedAddress):
			return w.fail(ctx, a, "source_url_not_allowed", err)
		case errors.Is(err, fetch.ErrBadScheme):
			return w.fail(ctx, a, "source_url_not_allowed", err)
		case errors.Is(err, fetch.ErrTooLarge):
			return w.fail(ctx, a, "source_too_large", err)
		case errors.Is(err, fetch.ErrTooManyHops):
			return w.fail(ctx, a, "source_unreachable", err)
		default:
			// Transient network problems deserve the retry the queue already gives us.
			return fmt.Errorf("pull source: %w", err)
		}
	}
	return nil
}

// DefaultMaxSourceBytes bounds a pull-from-URL download when no tenant limit applies.
const DefaultMaxSourceBytes = 32 << 30

// maxSourceBytes reads the tenant's cap. It used to be a worker field that no binary
// ever set, so tenant_limits.max_source_bytes was stored, shown, and never applied.
func (w *TranscodeWorker) maxSourceBytes(ctx context.Context, tenantID string) (int64, error) {
	max := int64(DefaultMaxSourceBytes)
	err := w.DB.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, `select max_source_bytes from tenant_limits`).Scan(&max)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil // no limits row: plan default applies
		}
		return err
	})
	if max <= 0 {
		max = DefaultMaxSourceBytes
	}
	return max, err
}

func (w *TranscodeWorker) emit(ctx context.Context, a TranscodeArgs, event string, data any) {
	if w.River == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	_ = Emit(ctx, w.DB, w.River, a.TenantID, event, data)
}

// pullFromBucket reads a source object out of the customer's own bucket using their
// credentials, so the bucket never has to be public or pre-signed.
func (w *TranscodeWorker) pullFromBucket(ctx context.Context, a TranscodeArgs, sourceID, objectKey, dst string) error {
	var endpoint, region, bucket, accessKey string
	var wrapped, nonce []byte

	// Credentials are resolved before any tenant is in scope, hence the definer
	// function; RLS would otherwise return no rows here.
	err := w.DB.Pool().QueryRow(ctx,
		`select endpoint, region, bucket, access_key_id, wrapped_secret, secret_nonce
		   from bucket_source_credentials($1)`, sourceID).
		Scan(&endpoint, &region, &bucket, &accessKey, &wrapped, &nonce)
	if err != nil {
		return w.fail(ctx, a, "source_unreadable", err)
	}

	secret, err := w.Keys.Unwrap(wrapped, nonce, sourceID)
	if err != nil {
		return w.fail(ctx, a, "source_unreadable", err)
	}

	client, err := storage.NewClient(ctx, endpoint, region, bucket, accessKey, string(secret))
	if err != nil {
		return w.fail(ctx, a, "source_unreadable", err)
	}

	body, err := client.Get(ctx, objectKey)
	if err != nil {
		return w.fail(ctx, a, "source_unreadable", err)
	}
	defer body.Close()

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, body); err != nil {
		return fmt.Errorf("read %s from customer bucket: %w", objectKey, err)
	}
	return nil
}

// fileBytes is what an object costs to store, taken where the file is already on
// disk. Zero on failure, which the callers store as null: billing wrong is worse than
// billing nothing.
func fileBytes(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return fi.Size()
}

// hashFile streams the file rather than reading it whole: sources run to gigabytes.
func hashFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

// findDuplicate looks for a completed asset with identical content in this tenant.
func (w *TranscodeWorker) findDuplicate(ctx context.Context, a TranscodeArgs, sum []byte) (string, error) {
	var id string
	err := w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`select id::text from assets
			  where source_sha256 = $1 and id <> $2
			    and state in ('ready','partially_ready')
			  limit 1`, sum, a.AssetID).Scan(&id)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return id, err
}

// linkToDuplicate points the new asset at renditions that already exist.
//
// The asset stays its own row with its own id, because the customer asked for it and
// may delete it independently; only the encoding work is skipped.
func (w *TranscodeWorker) linkToDuplicate(ctx context.Context, a TranscodeArgs, existingID string, sum []byte, srcBytes int64) error {
	err := w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		// mezzanine_bytes is deliberately not copied: the file is the canonical
		// asset's and billing it twice would charge for one copy of the bytes twice.
		if _, err := tx.Exec(ctx,
			`update assets dst set
			     state = src.state, duration_sec = src.duration_sec,
			     width = src.width, height = src.height, frame_rate = src.frame_rate,
			     mezzanine_key = src.mezzanine_key, complexity = src.complexity,
			     dash_skeleton = src.dash_skeleton, source_bytes = nullif($4,0)::bigint,
			     source_sha256 = $3, deduplicated_from = src.id, updated_at = now()
			   from assets src
			  where dst.id = $1 and src.id = $2`,
			a.AssetID, existingID, sum, srcBytes); err != nil {
			return err
		}
		// No rendition rows of its own. They describe media this asset does not own,
		// and a copied pending rung queues an encode published where nothing reads it;
		// the API resolves renditions through deduplicated_from instead.
		return nil
	})
	if err != nil {
		return err
	}
	w.emit(ctx, a, "asset.ready", map[string]any{
		"asset_id": a.AssetID, "deduplicated_from": existingID,
	})
	return nil
}

// applySourceRetention removes the original unless the tenant pays to keep it.
//
// Safe only because the mezzanine is already stored: every rendition, including one
// generated on demand much later, is built from the mezzanine and never from the
// original. Failure here is logged into the asset, not fatal -- the video is fine,
// there is simply a file left behind for a later sweep.
func (w *TranscodeWorker) applySourceRetention(ctx context.Context, a TranscodeArgs, sourceKey, mezzKey string) {
	if sourceKey == "" || mezzKey == "" {
		return // nothing uploaded by us, or no mezzanine to fall back on
	}

	var retain bool
	if err := w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, `select retain_original from tenant_limits`).Scan(&retain)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil // no limits row: plan default, do not retain
		}
		return err
	}); err != nil || retain {
		return
	}

	if err := w.Store.Delete(ctx, sourceKey); err != nil {
		// Not fatal: the video is fine, there is just an object left to sweep. But it
		// must be visible, because silently leaked originals are exactly the cost this
		// is meant to remove.
		_ = w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
			_, e := tx.Exec(ctx,
				`update assets set last_retention_error = $2 where id = $1`,
				a.AssetID, err.Error())
			return e
		})
		return
	}
	_ = w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`update assets set source_deleted_at = now(), source_key = null
			  where id = $1`, a.AssetID)
		return err
	})
}

// durationBuckets span a few seconds to a few hours: a short clip and a three-hour
// lecture are both normal here.
var durationBuckets = []float64{10, 30, 60, 300, 900, 1800, 3600, 7200}

func (w *TranscodeWorker) observeEncode(a TranscodeArgs) func() {
	if w.Metrics == nil {
		return func() {}
	}
	return w.Metrics.Timer("alchemist_encode_seconds",
		"wall time for a full transcode", durationBuckets, "tenant", a.TenantID)
}

// CountJob records a job outcome. Kept on the worker so every queue reports the same
// shape, which is what makes a single alert rule cover all of them.
func (w *TranscodeWorker) CountJob(kind, result string) {
	if w.Metrics == nil {
		return
	}
	w.Metrics.Inc("alchemist_jobs_total", "jobs processed by kind and result", 1,
		"kind", kind, "result", result)
}
