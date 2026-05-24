package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ddddami/laivan/internal/config"
)

type app struct {
	cfg     config.Config
	logger  *slog.Logger
	version string
}

func New(cfg config.Config, logger *slog.Logger, version string) *http.Server {
	app := &app{
		cfg:     cfg,
		logger:  logger,
		version: version,
	}

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      app.routes(),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}
