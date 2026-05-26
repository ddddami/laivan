package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ddddami/laivan/internal/config"
	"github.com/ddddami/laivan/internal/domain"
	"github.com/ddddami/laivan/internal/repo"
)

type app struct {
	cfg          config.Config
	logger       *slog.Logger
	version      string
	propertyRepo PropertyStore
}

type PropertyStore interface {
	Create(ctx context.Context, property domain.Property) (domain.Property, error)
	Get(ctx context.Context, id domain.ID) (domain.Property, error)
	GetWithDetails(ctx context.Context, id domain.ID) (domain.PropertyDetail, error)
	ListWithSummary(ctx context.Context, filter repo.PropertyListFilter) ([]domain.PropertySummary, int, error)
	CreatePropertyUnitType(ctx context.Context, unitType domain.PropertyUnitType) (domain.PropertyUnitType, error)
	ListPropertyUnitTypes(ctx context.Context, propertyID domain.ID) ([]domain.PropertyUnitType, error)
	CreateAgentOffer(ctx context.Context, offer domain.AgentOffer) (domain.AgentOffer, error)
	ListAgentOffers(ctx context.Context, unitTypeID domain.ID) ([]domain.AgentOffer, error)
}

func New(cfg config.Config, logger *slog.Logger, version string, propertyRepo PropertyStore) *http.Server {
	app := &app{
		cfg:          cfg,
		logger:       logger,
		version:      version,
		propertyRepo: propertyRepo,
	}

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      app.routes(),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}
