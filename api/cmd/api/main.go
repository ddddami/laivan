package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ddddami/laivan/internal/config"
	"github.com/ddddami/laivan/internal/db"
	"github.com/ddddami/laivan/internal/server"
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

	if cfg.DatabaseURL != "" {
		pool, err := db.Open(context.Background(), cfg.DatabaseURL)
		if err != nil {
			logger.Error("open database", "error", err)
			os.Exit(1)
		}
		defer pool.Close()

		logger.Info("database connection pool ready")
	}

	srv := server.New(cfg, logger, version)

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

func newLogger(cfg config.Config) *slog.Logger {
	if cfg.IsDevelopment() {
		return slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}
