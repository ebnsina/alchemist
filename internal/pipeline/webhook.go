package pipeline

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"

	"github.com/ebnsina/alchemist/internal/platform/db"
	"github.com/ebnsina/alchemist/internal/platform/fetch"
)

type WebhookArgs struct {
	EndpointID string          `json:"endpoint_id"`
	TenantID   string          `json:"tenant_id"`
	Event      string          `json:"event"`
	Payload    json.RawMessage `json:"payload"`
}

func (WebhookArgs) Kind() string { return "webhook" }

// Retries stretch over hours: a customer endpoint is usually down for a deploy, not
// forever, and giving up in ninety seconds turns a blip into lost data.
func (WebhookArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: QueueIO, MaxAttempts: 8}
}

type WebhookWorker struct {
	river.WorkerDefaults[WebhookArgs]
	DB *db.DB
}

func (w *WebhookWorker) Timeout(*river.Job[WebhookArgs]) time.Duration {
	return 60 * time.Second
}

func (w *WebhookWorker) Work(ctx context.Context, job *river.Job[WebhookArgs]) error {
	a := job.Args

	var url string
	var secret []byte
	var active bool
	err := w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`select url, secret, active from webhook_endpoints where id = $1`,
			a.EndpointID).Scan(&url, &secret, &active)
	})
	if err == pgx.ErrNoRows || (err == nil && !active) {
		return nil // endpoint deleted or disabled between enqueue and delivery
	}
	if err != nil {
		return fmt.Errorf("load endpoint: %w", err)
	}

	body, err := json.Marshal(map[string]any{
		"event":     a.Event,
		"data":      a.Payload,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return err
	}

	status, deliverErr := w.post(ctx, url, secret, body)
	w.record(ctx, a, body, status, deliverErr, job.Attempt)

	if deliverErr != nil {
		return deliverErr
	}
	// Anything outside 2xx is retried; the customer's 500 may well be transient.
	if status < 200 || status > 299 {
		return fmt.Errorf("endpoint returned %d", status)
	}
	return nil
}

// post signs the body so the receiver can verify it came from us. The timestamp is
// inside the signed payload, which is what stops a captured delivery being replayed
// indefinitely.
func (w *WebhookWorker) post(ctx context.Context, url string, secret, body []byte) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write(body)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Alchemist/1.0")
	req.Header.Set("X-Alchemist-Signature", "sha256="+hex.EncodeToString(mac.Sum(nil)))

	// Customer URLs are untrusted input just like pull-from-URL sources, so the same
	// guard applies: a webhook endpoint must not be usable to probe our network.
	resp, err := fetch.Client(30 * time.Second).Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	return resp.StatusCode, nil
}

func (w *WebhookWorker) record(ctx context.Context, a WebhookArgs, payload []byte, status int, derr error, attempt int) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()

	var errText *string
	if derr != nil {
		s := derr.Error()
		errText = &s
	}
	var statusPtr *int
	if status > 0 {
		statusPtr = &status
	}
	_ = w.DB.AsTenant(ctx, a.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`insert into webhook_deliveries
			   (endpoint_id, tenant_id, event, payload, status_code, error, attempt, delivered_at)
			 values ($1,$2,$3,$4,$5,$6,$7, case when $5 between 200 and 299 then now() end)`,
			a.EndpointID, a.TenantID, a.Event, payload, statusPtr, errText, attempt)
		return err
	})
}

// Emit queues an event to every active endpoint subscribed to it.
func Emit(ctx context.Context, database *db.DB, rc *river.Client[pgx.Tx], tenantID, event string, data any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	var endpointIDs []string
	if err := database.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx,
			`select id::text from webhook_endpoints
			  where active and $1 = any(events)`, event)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return err
			}
			endpointIDs = append(endpointIDs, id)
		}
		return rows.Err()
	}); err != nil {
		return err
	}

	for _, id := range endpointIDs {
		if _, err := rc.Insert(ctx, WebhookArgs{
			EndpointID: id, TenantID: tenantID, Event: event, Payload: payload,
		}, nil); err != nil {
			return err
		}
	}
	return nil
}
