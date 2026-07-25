package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ddddami/laivan/internal/config"
	"github.com/ddddami/laivan/internal/domain"
	"github.com/ddddami/laivan/internal/repo"
	"github.com/ddddami/laivan/internal/storage"
)

type app struct {
	cfg           config.Config
	logger        *slog.Logger
	version       string
	propertyRepo  PropertyStore
	mediaUploader storage.Uploader
	mediaURLs     MediaURLBuilder
}

type MediaURLBuilder interface {
	ThumbnailURL(sourceURL string) string
	MediumURL(sourceURL string) string
}

// NOTE: This repository boundary is intentionally consolidated for now.
// Split only when operational pressure becomes recurring
// (see gh issue:  #8 repository/interface split).
type PropertyStore interface {
	GetCampusBySlug(ctx context.Context, slug string) (domain.Campus, error)
	Create(ctx context.Context, property domain.Property) (domain.Property, error)
	Get(ctx context.Context, id domain.ID) (domain.Property, error)
	GetWithDetails(ctx context.Context, id domain.ID) (domain.PropertyDetail, error)
	ListWithSummary(ctx context.Context, filter repo.PropertyListFilter) ([]domain.PropertySummary, int, error)
	Discover(ctx context.Context, filter repo.DiscoveryFilter) ([]domain.DiscoveryResult, int, error)
	CreateMedia(ctx context.Context, media domain.Media) (domain.Media, error)
	ListMediaByProperty(ctx context.Context, propertyID domain.ID) ([]domain.Media, error)
	ListMediaByPropertyUnitType(ctx context.Context, propertyUnitTypeID domain.ID) ([]domain.Media, error)
	ListMediaByAgentOffer(ctx context.Context, agentOfferID domain.ID) ([]domain.Media, error)
	CreatePropertyUnitType(ctx context.Context, unitType domain.PropertyUnitType) (domain.PropertyUnitType, error)
	ListPropertyUnitTypes(ctx context.Context, propertyID domain.ID) ([]domain.PropertyUnitType, error)
	CreateAgentOffer(ctx context.Context, offer domain.AgentOffer) (domain.AgentOffer, error)
	ListAgentOffers(ctx context.Context, unitTypeID domain.ID) ([]domain.AgentOffer, error)
}

func New(cfg config.Config, logger *slog.Logger, version string, propertyRepo PropertyStore, mediaUploader storage.Uploader, mediaURLs MediaURLBuilder) *http.Server {
	app := &app{
		cfg:           cfg,
		logger:        logger,
		version:       version,
		propertyRepo:  propertyRepo,
		mediaUploader: mediaUploader,
		mediaURLs:     mediaURLs,
	}

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      app.routes(),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}
