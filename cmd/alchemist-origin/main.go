// Command alchemist-origin runs the delivery plane on its own.
//
// This binary exists to prove the seam is real rather than aspirational. It mounts
// the delivery module and nothing else: no control-plane routes, no job queue, no
// transcoding. Deploy it near storage and behind the BDIX edge and scale it on
// request volume, independently of the API.
//
// Extracting delivery into its own repository is then mechanical: take
// internal/modules/delivery, the adapters it uses, the platform packages those
// adapters touch, and this file.
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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"

	"github.com/ebnsina/alchemist/internal/adapters"
	"github.com/ebnsina/alchemist/internal/modules/delivery"
	"github.com/ebnsina/alchemist/internal/platform/config"
	"github.com/ebnsina/alchemist/internal/platform/db"
	"github.com/ebnsina/alchemist/internal/platform/keys"
	"github.com/ebnsina/alchemist/internal/platform/signing"
	"github.com/ebnsina/alchemist/internal/platform/storage"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

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

	signer, err := signing.NewKeyring(cfg.PlaybackKeys)
	if err != nil {
		log.Error("playback keyring", "err", err)
		os.Exit(1)
	}

	// The origin must queue deferred renditions too. Without this, an asset served
	// only by a standalone origin stays on its low ladder forever -- which silently
	// disables JIT in exactly the deployment this binary exists for.
	riverClient, err := river.NewClient(riverpgxv5.New(database.Pool()), &river.Config{})
	if err != nil {
		log.Error("river", "err", err)
		os.Exit(1)
	}

	module := delivery.New(
		adapters.ObjectStore{Store: store},
		adapters.ContentKeys{DB: database, Wrapper: keyWrapper},
		signer,
	).WithObserver(adapters.LazyRenditions{DB: database, River: riverClient}).
		WithResolver(adapters.DedupResolver{DB: database})

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer)
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	module.Routes(r)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Info("origin listening", "addr", cfg.HTTPAddr)
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
}
