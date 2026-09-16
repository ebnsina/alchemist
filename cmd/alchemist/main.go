package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertype"

	"github.com/ebnsina/alchemist/internal/adapters"
	"github.com/ebnsina/alchemist/internal/api"
	"github.com/ebnsina/alchemist/internal/modules/delivery"
	"github.com/ebnsina/alchemist/internal/modules/live"
	"github.com/ebnsina/alchemist/internal/pipeline"
	"github.com/ebnsina/alchemist/internal/platform/config"
	"github.com/ebnsina/alchemist/internal/platform/db"
	"github.com/ebnsina/alchemist/internal/platform/fetch"
	"github.com/ebnsina/alchemist/internal/platform/keys"
	"github.com/ebnsina/alchemist/internal/platform/metrics"
	"github.com/ebnsina/alchemist/internal/platform/signing"
	"github.com/ebnsina/alchemist/internal/platform/storage"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// One binary with a subcommand rather than a second one: the seed needs the same
	// password hashing as signup, and a copy of that drifts.
	if len(os.Args) > 1 && os.Args[1] == "create-admin" {
		createAdmin(log, os.Args[2:])
		return
	}

	cfg, err := config.Load()
	if err != nil {
		log.Error("configuration", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	database, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("database", "err", err)
		os.Exit(1)
	}
	defer database.Close()

	store, err := storage.New(ctx, cfg.S3Endpoint, cfg.S3Region, cfg.S3Bucket,
		cfg.S3AccessKey, cfg.S3SecretKey)
	if err != nil {
		log.Error("storage", "err", err)
		os.Exit(1)
	}

	keyWrapper, err := keys.NewWrapper(cfg.KEK)
	if err != nil {
		log.Error("key wrapper", "err", err)
		os.Exit(1)
	}

	if len(cfg.FetchAllowlist) > 0 {
		fetch.SetDevAllowlist(cfg.FetchAllowlist)
		log.Warn("SSRF address guard bypassed for specific hosts -- never set ALCHEMIST_FETCH_ALLOWLIST in production",
			"hosts", cfg.FetchAllowlist)
	}

	signer, err := signing.NewKeyring(cfg.PlaybackKeys)
	if err != nil {
		log.Error("playback keyring", "err", err)
		os.Exit(1)
	}

	reg := metrics.New()

	workers := river.NewWorkers()
	transcoder := &pipeline.TranscodeWorker{
		DB: database, Store: store, Keys: keyWrapper, Metrics: reg, WorkDir: cfg.WorkDir,
	}
	river.AddWorker(workers, transcoder)
	river.AddWorker(workers, &pipeline.WebhookWorker{DB: database})
	bucketSync := &pipeline.BucketSyncWorker{DB: database, Keys: keyWrapper}
	river.AddWorker(workers, bucketSync)
	migrator := &pipeline.MigrateWorker{DB: database, Keys: keyWrapper}
	river.AddWorker(workers, migrator)
	editor := &pipeline.EditWorker{DB: database, Store: store, WorkDir: cfg.WorkDir}
	river.AddWorker(workers, editor)
	reconciler := &pipeline.ReconcileWorker{DB: database}
	river.AddWorker(workers, reconciler)
	river.AddWorker(workers, &pipeline.JITWorker{TranscodeWorker: transcoder})
	river.AddWorker(workers, &pipeline.SweepWorker{DB: database, Store: store})
	river.AddWorker(workers, &pipeline.ReclaimWorker{Store: store})
	river.AddWorker(workers, &pipeline.StorageWorker{DB: database})
	liveBroadcast := &live.Worker{DB: database, Store: adapters.ObjectStore{Store: store},
		Assets: adapters.LiveAssets{DB: database}, Ladder: adapters.LiveLadder{DB: database},
		PullBase: cfg.LivePullBase, WorkDir: cfg.WorkDir}
	river.AddWorker(workers, &pipeline.LiveWorker{Live: liveBroadcast})
	river.AddWorker(workers, &pipeline.LiveReapWorker{DB: database, Live: liveBroadcast})

	riverClient, err := river.NewClient(riverpgxv5.New(database.Pool()), &river.Config{
		Queues: map[string]river.QueueConfig{
			pipeline.QueueEncode: {MaxWorkers: cfg.EncodeWorkers},
			pipeline.QueueIO:     {MaxWorkers: 8},
			// One slot per port: a live job holds its worker for the whole broadcast.
			pipeline.QueueLive: {MaxWorkers: max(1, cfg.LiveMaxStreams)},
		},
		Workers:      workers,
		PeriodicJobs: pipeline.PeriodicJobs(),
		Middleware:   []rivertype.Middleware{pipeline.JobMetrics(reg)},
	})
	if err != nil {
		log.Error("river", "err", err)
		os.Exit(1)
	}
	// The worker emits webhook events through the same client it is registered on.
	transcoder.River = riverClient
	bucketSync.River = riverClient
	migrator.River = riverClient
	editor.River = riverClient
	reconciler.River = riverClient
	liveBroadcast.Events = adapters.LiveEvents{DB: database, River: riverClient}
	liveBroadcast.Queue = adapters.LiveQueue{River: riverClient}

	if err := riverClient.Start(ctx); err != nil {
		log.Error("river start", "err", err)
		os.Exit(1)
	}

	// Node-local, so a goroutine here rather than a queued job: River would run the
	// job on one node and leave every other node's disk filling.
	go pruneWorkDir(ctx, log, cfg.WorkDir)

	// Egress is the largest line on a video bill and the origin is the only place
	// that sees the bytes, so it is metered here and folded into a daily row.
	egress := &adapters.Usage{DB: database}
	go egress.Flush(ctx, time.Minute)

	deliveryModule := delivery.New(
		adapters.ObjectStore{Store: store},
		adapters.ContentKeys{DB: database, Wrapper: keyWrapper},
		signer,
	).WithObserver(adapters.LazyRenditions{DB: database, River: riverClient}).
		WithResolver(adapters.DedupResolver{DB: database}).
		WithMeter(egress)

	// No ingest host means no live module and no live routes: there is nowhere for
	// an encoder to connect, so serving them could only ever fail.
	var liveModule *live.Module
	if cfg.LiveIngestHost != "" {
		liveModule = live.New(database, cfg.LiveIngestHost,
			adapters.LiveAssets{DB: database}, adapters.LiveQueue{River: riverClient})
	}

	srv := &http.Server{
		Addr: cfg.HTTPAddr,
		Handler: api.New(database, store, riverClient, deliveryModule, keyWrapper, cfg.AdminKey, reg,
			api.Accounts{
				WebOrigins:    cfg.WebOrigins,
				SessionDomain: cfg.SessionDomain,
				SessionSecure: cfg.SessionSecure,
			},
			liveModule, cfg.PlayerURL).Routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Info("listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("serve", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown", "err", err)
	}
	if err := riverClient.Stop(shutdownCtx); err != nil {
		log.Error("river stop", "err", err)
	}
}

// pruneWorkDirInterval is how often abandoned working directories are swept. Nothing
// is urgent about it -- the point is that a leak is bounded, not that it is caught
// quickly -- and the age cutoff does the deciding.
const pruneWorkDirInterval = time.Hour

func pruneWorkDir(ctx context.Context, log *slog.Logger, dir string) {
	prune := func() {
		n, err := pipeline.PruneWorkDir(dir, pipeline.WorkDirTTL)
		if err != nil {
			log.Error("work directory prune", "dir", dir, "err", err)
		}
		if n > 0 {
			log.Info("removed abandoned working directories", "count", n, "dir", dir)
		}
	}
	prune()

	t := time.NewTicker(pruneWorkDirInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			prune()
		}
	}
}
