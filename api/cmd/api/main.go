package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/ddddami/laivan/internal/config"
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
	srv := server.New(cfg, logger, version)

	logger.Info("starting api server", "addr", srv.Addr, "env", cfg.Env, "version", version)

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("api server failed", "error", err)
		os.Exit(1)
	}
}

func newLogger(cfg config.Config) *slog.Logger {
	if cfg.IsDevelopment() {
		return slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}
