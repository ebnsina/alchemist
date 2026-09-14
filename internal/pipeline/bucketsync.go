package pipeline

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"

	"github.com/ebnsina/alchemist/internal/platform/db"
	"github.com/ebnsina/alchemist/internal/platform/keys"
	"github.com/ebnsina/alchemist/internal/platform/storage"
)

// videoExtensions bounds what a reconcile will pick up. Customers keep PDFs, images
// and stray archives in the same bucket, and transcoding those wastes their money
// and ours.
var videoExtensions = map[string]bool{
	".mp4": true, ".mov": true, ".mkv": true, ".avi": true, ".webm": true,
	".m4v": true, ".mpg": true, ".mpeg": true, ".wmv": true, ".flv": true,
	".ts": true, ".mxf": true, ".m2ts": true,
}

type BucketSyncArgs struct {
	SourceID string `json:"source_id"`
	TenantID string `json:"tenant_id"`
}

func (BucketSyncArgs) Kind() string { return "bucket_sync" }

// Unique per source: overlapping reconciles of the same bucket would double-ingest
// and fight over the cursor.
func (BucketSyncArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       QueueIO,
		MaxAttempts: 3,
		// States are listed explicitly. River's default set includes Completed, and
		// completed jobs are retained for 24h -- so the default would let a source
		// sync once per day rather than once per reconcile, and new objects would
		// take up to 24h to appear. Block only while a sync is genuinely outstanding.
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
			ByState: []rivertype.JobState{
				rivertype.JobStateAvailable,
				rivertype.JobStatePending,
				rivertype.JobStateRunning,
				rivertype.JobStateRetryable,
				rivertype.JobStateScheduled,
			},
		},
	}
}

type BucketSyncWorker struct {
	river.WorkerDefaults[BucketSyncArgs]
	DB    *db.DB
	Keys  *keys.Wrapper
	River *river.Client[pgx.Tx]
	// PageSize bounds one pass. A reconcile that must walk a million objects before
	// making any progress never finishes; each run takes a page and advances.
	PageSize int32
}

func (w *BucketSyncWorker) Timeout(*river.Job[BucketSyncArgs]) time.Duration {
	return 10 * time.Minute
}

type bucketSource struct {
	Endpoint, Region, Bucket, Prefix string
	AccessKeyID                      string
	WrappedSecret, SecretNonce       []byte
	Cursor                           string
}

func (w *BucketSyncWorker) Work(ctx context.Context, job *river.Job[BucketSyncArgs]) error {
	a := job.Args

	var src bucketSource
	var cursor *string
	err := w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`select endpoint, region, bucket, prefix, access_key_id,
			        wrapped_secret, secret_nonce, cursor
			   from bucket_sources where id = $1 and active`, a.SourceID).
			Scan(&src.Endpoint, &src.Region, &src.Bucket, &src.Prefix,
				&src.AccessKeyID, &src.WrappedSecret, &src.SecretNonce, &cursor)
	})
	if err == pgx.ErrNoRows {
		return nil // source removed or deactivated since the job was queued
	}
	if err != nil {
		return fmt.Errorf("load bucket source: %w", err)
	}
	if cursor != nil {
		src.Cursor = *cursor
	}

	secret, err := w.Keys.Unwrap(src.WrappedSecret, src.SecretNonce, a.SourceID)
	if err != nil {
		return w.recordError(ctx, a, "credentials could not be read")
	}

	client, err := storage.NewClient(ctx, src.Endpoint, src.Region, src.Bucket,
		src.AccessKeyID, string(secret))
	if err != nil {
		return w.recordError(ctx, a, "could not connect to the bucket")
	}

	pageSize := w.PageSize
	if pageSize <= 0 {
		pageSize = 500
	}
	listing, err := client.List(ctx, src.Prefix, src.Cursor, pageSize)
	if err != nil {
		return w.recordError(ctx, a, "could not list the bucket")
	}

	imported := 0
	for _, obj := range listing.Objects {
		if !videoExtensions[strings.ToLower(path.Ext(obj.Key))] {
			continue
		}
		if obj.Size == 0 {
			continue
		}
		ok, err := w.ingest(ctx, a, src, obj)
		if err != nil {
			return err
		}
		if ok {
			imported++
		}
	}

	// Advance the cursor only after the page is handled, so a crash re-reads the page
	// rather than skipping it. Re-reading is harmless: bucket_objects makes ingest
	// idempotent.
	next := listing.Cursor
	if !listing.HasMore {
		next = "" // finished a full pass; the next run starts from the beginning
	}
	if err := w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`update bucket_sources
			    set cursor = nullif($2,''), last_synced_at = now(), last_error = null
			  where id = $1`, a.SourceID, next)
		return err
	}); err != nil {
		return err
	}

	// More pages waiting: queue the next immediately rather than idling until the
	// periodic run, so a first import of a large library completes promptly.
	if listing.HasMore && w.River != nil {
		if _, err := w.River.Insert(ctx, BucketSyncArgs{
			SourceID: a.SourceID, TenantID: a.TenantID,
		}, nil); err != nil {
			return err
		}
	}
	return nil
}

// ingest records the object and queues a transcode, unless this exact key and etag
// has been seen before. Returns whether anything new was queued.
func (w *BucketSyncWorker) ingest(ctx context.Context, a BucketSyncArgs, src bucketSource, obj storage.ListedObject) (bool, error) {
	var assetID string
	queued := false

	err := w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		var exists bool
		if err := tx.QueryRow(ctx,
			`select exists (select 1 from bucket_objects
			                 where source_id = $1 and object_key = $2 and etag = $3)`,
			a.SourceID, obj.Key, obj.ETag).Scan(&exists); err != nil {
			return err
		}
		if exists {
			return nil
		}

		if err := tx.QueryRow(ctx,
			`insert into assets (tenant_id, state, ladder_profile, source_url,
			                     bucket_source_id, source_object_key)
			 select $1, 'uploaded', ladder_profile, $2, $3, $4 from tenants where id = $1
			 returning id::text`,
			a.TenantID, "s3://"+src.Bucket+"/"+obj.Key, a.SourceID, obj.Key).
			Scan(&assetID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx,
			`insert into bucket_objects (source_id, object_key, etag, asset_id)
			 values ($1,$2,$3,$4)
			 on conflict (source_id, object_key, etag) do nothing`,
			a.SourceID, obj.Key, obj.ETag, assetID); err != nil {
			return err
		}
		queued = true
		return nil
	})
	if err != nil || !queued {
		return false, err
	}

	// The transcoder reads the object back with the source's own credentials, via
	// bucket_source_id, so the customer's bucket can stay private.
	if _, err := w.River.Insert(ctx, TranscodeArgs{
		AssetID: assetID, TenantID: a.TenantID,
	}, nil); err != nil {
		return false, err
	}
	return true, nil
}

func (w *BucketSyncWorker) recordError(ctx context.Context, a BucketSyncArgs, msg string) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	_ = w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`update bucket_sources set last_error = $2, last_synced_at = now()
			  where id = $1`, a.SourceID, msg)
		return err
	})
	return fmt.Errorf("bucket sync: %s", msg)
}
