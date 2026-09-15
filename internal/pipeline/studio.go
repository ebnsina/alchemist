package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"

	"github.com/ebnsina/alchemist/internal/platform/db"
	"github.com/ebnsina/alchemist/internal/platform/media"
	"github.com/ebnsina/alchemist/internal/platform/storage"
)

type EditArgs struct {
	EditID   string `json:"edit_id"`
	TenantID string `json:"tenant_id"`
}

func (EditArgs) Kind() string { return "studio_edit" }

func (EditArgs) InsertOpts() river.InsertOpts {
	// Encode queue, not IO: this runs ffmpeg and should compete for the same slots as
	// everything else that does, or a studio full of reels starves the main pipeline.
	return river.InsertOpts{Queue: QueueEncode, MaxAttempts: 3}
}

type EditWorker struct {
	river.WorkerDefaults[EditArgs]
	DB      *db.DB
	Store   *storage.Store
	River   *river.Client[pgx.Tx]
	WorkDir string
}

func (w *EditWorker) Timeout(*river.Job[EditArgs]) time.Duration { return 30 * time.Minute }

// Work renders one edit into a brand new asset. The source asset is read and never
// written: somebody may already have handed out a link to it.
func (w *EditWorker) Work(ctx context.Context, job *river.Job[EditArgs]) (err error) {
	a := job.Args

	defer func() {
		if err != nil && job.Attempt >= job.MaxAttempts {
			w.failEdit(ctx, a, "processing_failed")
		}
	}()

	var rawOps []byte
	var sourceAssetID, profile string
	var mezzKey *string
	var duration *float64
	err = w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`select e.ops, e.source_asset_id::text, a.mezzanine_key, a.ladder_profile,
			        a.duration_sec::float8
			   from edits e join assets a on a.id = e.source_asset_id
			  where e.id = $1`, a.EditID).
			Scan(&rawOps, &sourceAssetID, &mezzKey, &profile, &duration)
	})
	if err != nil {
		return fmt.Errorf("load edit %s: %w", a.EditID, err)
	}

	var edit media.Edit
	if err := json.Unmarshal(rawOps, &edit); err != nil {
		return w.failEdit(ctx, a, "invalid_request")
	}
	// Nothing to render from. The original may be long deleted, which is normal and
	// exactly why edits work off the mezzanine.
	if mezzKey == nil || *mezzKey == "" {
		return w.failEdit(ctx, a, "not_ready")
	}

	if err := w.setEditState(ctx, a, "rendering"); err != nil {
		return err
	}

	dir := filepath.Join(w.WorkDir, "edit-"+a.EditID)
	defer os.RemoveAll(dir)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}

	src := filepath.Join(dir, "mezzanine.mp4")
	if err := w.download(ctx, *mezzKey, src); err != nil {
		return fmt.Errorf("fetch mezzanine: %w", err)
	}

	dst := filepath.Join(dir, "edited.mp4")
	var dur float64
	if duration != nil {
		dur = *duration
	}

	// Overlays become files before ffmpeg is asked to composite them: text is drawn
	// here, images come out of storage. This package owns storage; the media package
	// must not.
	prepared, err := w.prepareOverlays(ctx, a, edit, dir)
	if err != nil {
		if errors.Is(err, media.ErrBadEdit) {
			return w.failEdit(ctx, a, "invalid_edit")
		}
		return err
	}

	if err := media.RenderWith(ctx, edit, src, dst, dur, prepared); err != nil {
		// A request that cannot be rendered is the customer's to fix, so it stops
		// here rather than retrying twice more to fail the same way.
		if errors.Is(err, media.ErrBadEdit) {
			return w.failEdit(ctx, a, "invalid_request")
		}
		return err
	}

	// The output becomes an ordinary asset: everything downstream — probe, ladder,
	// encode, package, playback — is the pipeline that already exists.
	var outID string
	err = w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`insert into assets (tenant_id, state, ladder_profile)
			 values ($1, 'uploading', $2) returning id::text`, a.TenantID, profile).Scan(&outID)
	})
	if err != nil {
		return fmt.Errorf("create edited asset: %w", err)
	}

	key := fmt.Sprintf("src/%s/%s/original", a.TenantID, outID)
	f, err := os.Open(dst)
	if err != nil {
		return err
	}
	if err := w.Store.Put(ctx, key, f, "video/mp4"); err != nil {
		f.Close()
		return fmt.Errorf("store edited video: %w", err)
	}
	f.Close()

	err = w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		if _, e := tx.Exec(ctx,
			`update assets set source_key = $2, state = 'uploaded' where id = $1`,
			outID, key); e != nil {
			return e
		}
		_, e := tx.Exec(ctx,
			`update edits set output_asset_id = $2, state = 'done', updated_at = now()
			  where id = $1`, a.EditID, outID)
		return e
	})
	if err != nil {
		return fmt.Errorf("finish edit: %w", err)
	}

	if _, err := w.River.Insert(ctx,
		TranscodeArgs{AssetID: outID, TenantID: a.TenantID}, nil); err != nil {
		return fmt.Errorf("queue edited transcode: %w", err)
	}
	return nil
}

// prepareOverlays turns each overlay into a PNG on disk. An image overlay names a
// storage key — the account's own logo, usually — and is never a URL, so nothing here
// can be pointed at somewhere it should not go.
func (w *EditWorker) prepareOverlays(
	ctx context.Context, a EditArgs, e media.Edit, dir string,
) ([]media.PreparedOverlay, error) {
	out := make([]media.PreparedOverlay, 0, len(e.Overlays))
	for i, o := range e.Overlays {
		path := filepath.Join(dir, fmt.Sprintf("overlay-%d.png", i))

		switch o.Kind {
		case "text":
			colour, err := media.ParseHexColour(orDefault(o.Colour, "#ffffff"))
			if err != nil {
				return nil, err
			}
			// Sized against a 1080-tall frame and scaled to the real one by the
			// filter, so the same overlay reads the same on any size of output.
			px := o.Scale * 1080
			if px < 16 {
				px = 48
			}
			if _, _, err := media.RenderText(media.TextImage{
				Text: o.Text, PixelSize: px, Colour: colour, Shadow: o.Shadow,
			}, path); err != nil {
				return nil, err
			}
		case "image":
			// Resolved from the tenant's own row. The request never names a key.
			key, err := w.overlayKey(ctx, a, o.Source)
			if err != nil {
				return nil, err
			}
			if err := w.download(ctx, key, path); err != nil {
				return nil, fmt.Errorf("fetch overlay image: %w", err)
			}
		default:
			return nil, fmt.Errorf("%w: %q is not an overlay", media.ErrBadEdit, o.Kind)
		}
		out = append(out, media.PreparedOverlay{Overlay: o, Path: path})
	}
	return out, nil
}

// overlayKey turns a name into a storage key, scoped to this tenant. "logo" is the
// only one for now; anything else is refused rather than resolved.
func (w *EditWorker) overlayKey(ctx context.Context, a EditArgs, name string) (string, error) {
	if name != "logo" {
		return "", fmt.Errorf("%w: %q is not an image we can put on a video", media.ErrBadEdit, name)
	}
	var key *string
	err := w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		e := tx.QueryRow(ctx, `select logo_key from tenant_branding`).Scan(&key)
		if e == pgx.ErrNoRows {
			return nil
		}
		return e
	})
	if err != nil {
		return "", err
	}
	if key == nil || *key == "" {
		return "", fmt.Errorf("%w: this account has no logo to put on a video", media.ErrBadEdit)
	}
	return *key, nil
}

func orDefault(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func (w *EditWorker) download(ctx context.Context, key, dst string) error {
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

func (w *EditWorker) setEditState(ctx context.Context, a EditArgs, state string) error {
	return w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`update edits set state = $2, updated_at = now() where id = $1`, a.EditID, state)
		return err
	})
}

// failEdit records why and returns nil: the job is finished, and a retry would fail
// the same way. Returning the error instead would burn attempts for nothing.
func (w *EditWorker) failEdit(ctx context.Context, a EditArgs, code string) error {
	return w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`update edits set state = 'failed', error_code = $2, updated_at = now()
			  where id = $1`, a.EditID, code)
		return err
	})
}
