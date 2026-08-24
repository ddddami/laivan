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

	"github.com/ddddami/laivan/internal/auth"
	"github.com/ddddami/laivan/internal/auth/google"
	"github.com/ddddami/laivan/internal/config"
	"github.com/ddddami/laivan/internal/db"
	"github.com/ddddami/laivan/internal/media"
	"github.com/ddddami/laivan/internal/repo"
	"github.com/ddddami/laivan/internal/server"
	"github.com/ddddami/laivan/internal/storage"
)

var version = "dev"

func main() {
	cfg, err := config.Load()
	if err != nil {
		logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	logger := newLogger(cfg)

	var mediaUploader storage.ObjectStore
	var mediaURLs server.MediaURLBuilder
	pool, err := db.Open(context.Background(), cfg.DatabaseURL, cfg.DBMaxConns)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	logger.Info("database connection pool ready")
	propertyRepo := repo.NewPropertyRepository(pool)
	identityRepo := repo.NewIdentityRepository(pool)
	applicationRepo := repo.NewAgentApplicationRepository(pool)
	authService, err := newAuthService(cfg, identityRepo)
	if err != nil {
		logger.Error("create auth service", "error", err)
		os.Exit(1)
	}
	if cfg.Media.Enabled {
		uploader, err := storage.NewS3Uploader(context.Background(), cfg.Media)
		if err != nil {
			logger.Error("create media uploader", "error", err)
			os.Exit(1)
		}
		urlBuilder, err := media.NewURLBuilder(cfg.Media)
		if err != nil {
			logger.Error("create media url builder", "error", err)
			os.Exit(1)
		}
		mediaUploader = uploader
		mediaURLs = urlBuilder
	}

	srv := server.New(cfg, logger, version, propertyRepo, mediaUploader, mediaURLs, authService, applicationRepo)

	logger.Info("starting api server", "addr", srv.Addr, "env", cfg.Env, "version", version)

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- srv.ListenAndServe()
	}()

	shutdownSignals := make(chan os.Signal, 1)
	signal.Notify(shutdownSignals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(shutdownSignals)

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("api server failed", "error", err)
			os.Exit(1)
		}
	case sig := <-shutdownSignals:
		logger.Info("shutting down api server", "signal", sig.String())

		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("api server shutdown failed", "error", err)
			os.Exit(1)
		}

		if err := <-serverErrors; err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("api server failed during shutdown", "error", err)
			os.Exit(1)
		}

		logger.Info("api server stopped")
	}
}

func newAuthService(cfg config.Config, store auth.Store) (*auth.Service, error) {
	if !cfg.Auth.IsConfigured() {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	provider, err := google.New(ctx, google.Config{
		ClientID:     cfg.Auth.GoogleClientID,
		ClientSecret: cfg.Auth.GoogleClientSecret,
		RedirectURL:  cfg.Auth.GoogleRedirectURL,
	})
	if err != nil {
		return nil, err
	}

	return auth.NewService(auth.Config{
		StateSigningKey: cfg.Auth.OIDCStateSigningKey,
		StateDuration:   cfg.Auth.OIDCStateDuration,
		SessionDuration: cfg.Auth.SessionDuration,
	}, store, provider), nil
}

func newLogger(cfg config.Config) *slog.Logger {
	if cfg.IsDevelopment() {
		return slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}
