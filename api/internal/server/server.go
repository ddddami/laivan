package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ddddami/laivan/internal/auth"
	"github.com/ddddami/laivan/internal/config"
	"github.com/ddddami/laivan/internal/domain"
	"github.com/ddddami/laivan/internal/repo"
	"github.com/ddddami/laivan/internal/storage"
)

type app struct {
	cfg             config.Config
	logger          *slog.Logger
	version         string
	propertyRepo    PropertyStore
	mediaUploader   storage.ObjectStore
	mediaURLs       MediaURLBuilder
	auth            *auth.Service
	applications    AgentApplicationStore
	inquiries       InquiryStore
	oidcRateLimiter *rateLimiter
}

type MediaURLBuilder interface {
	ThumbnailURL(sourceURL string) string
	MediumURL(sourceURL string) string
}

type AgentApplicationStore interface {
	GetEffectiveAccess(ctx context.Context, userID domain.ID) (domain.EffectiveAccess, error)
	CreateApplication(ctx context.Context, application domain.AgentApplication) (domain.AgentApplication, error)
	ListApplications(ctx context.Context, applicantID domain.ID) ([]domain.AgentApplication, error)
	ListOperatorApplications(ctx context.Context, operatorID, campusID domain.ID, status domain.AgentApplicationStatus) ([]domain.AgentApplication, error)
	ActivateApplication(ctx context.Context, applicationID, operatorID domain.ID, operatorNote string) (domain.AgentApplication, error)
	DeclineApplication(ctx context.Context, applicationID, operatorID domain.ID, operatorNote string) (domain.AgentApplication, error)
	SuspendAgent(ctx context.Context, agentID, operatorID domain.ID, operatorNote string) (domain.LinkedAgent, error)
	ReinstateAgent(ctx context.Context, agentID, operatorID domain.ID, operatorNote string) (domain.LinkedAgent, error)
}

type InquiryStore interface {
	Submit(ctx context.Context, studentUserID, agentOfferID, submissionID domain.ID, message string) (domain.Inquiry, domain.InquiryHandoff, bool, error)
}

// NOTE: This repository boundary is intentionally consolidated for now.
// Split only when operational pressure becomes recurring
// (see gh issue:  #8 repository/interface split).
type PropertyStore interface {
	GetCampusBySlug(ctx context.Context, slug string) (domain.Campus, error)
	Create(ctx context.Context, property domain.Property, actorUserID domain.ID) (domain.Property, error)
	Get(ctx context.Context, id domain.ID) (domain.Property, error)
	Update(ctx context.Context, id domain.ID, expectedVersion int, patch domain.PropertyPatch, actorUserID domain.ID) (domain.Property, error)
	GetPropertyUnitType(ctx context.Context, id domain.ID) (domain.PropertyUnitType, error)
	UpdatePropertyUnitType(ctx context.Context, id domain.ID, expectedVersion int, patch domain.PropertyUnitTypePatch, actorUserID domain.ID) (domain.PropertyUnitType, error)
	GetWithDetails(ctx context.Context, id domain.ID) (domain.PropertyDetail, error)
	ListWithSummary(ctx context.Context, filter repo.PropertyListFilter) ([]domain.PropertySummary, int, error)
	Discover(ctx context.Context, filter repo.DiscoveryFilter) ([]domain.DiscoveryResult, int, error)
	CreateMediaBatch(ctx context.Context, media []domain.Media) ([]domain.Media, error)
	ListMediaByProperty(ctx context.Context, propertyID domain.ID) ([]domain.Media, error)
	ListMediaByPropertyUnitType(ctx context.Context, propertyUnitTypeID domain.ID) ([]domain.Media, error)
	ListMediaByAgentOffer(ctx context.Context, agentOfferID domain.ID) ([]domain.Media, error)
	GetMediaTarget(ctx context.Context, targetType string, id domain.ID) (repo.MediaTarget, error)
	CreatePropertyUnitType(ctx context.Context, unitType domain.PropertyUnitType, actorUserID domain.ID) (domain.PropertyUnitType, error)
	ListPropertyUnitTypes(ctx context.Context, propertyID domain.ID) ([]domain.PropertyUnitType, error)
	CreateAgentOffer(ctx context.Context, offer domain.AgentOffer) (domain.AgentOffer, error)
	ListAgentOffers(ctx context.Context, unitTypeID domain.ID) ([]domain.AgentOffer, error)
	GetAgentOffer(ctx context.Context, id domain.ID) (domain.AgentOffer, error)
	UpdateAgentOffer(ctx context.Context, id domain.ID, expectedVersion int, patch domain.AgentOfferPatch, actorUserID domain.ID) (domain.AgentOffer, error)
	ArchiveAgentOffer(ctx context.Context, id domain.ID, expectedVersion int, actorUserID domain.ID) (domain.AgentOffer, error)
}

func New(cfg config.Config, logger *slog.Logger, version string, propertyRepo PropertyStore, mediaUploader storage.ObjectStore, mediaURLs MediaURLBuilder, authService *auth.Service, applications AgentApplicationStore, inquiries InquiryStore) *http.Server {
	app := &app{
		cfg:             cfg,
		logger:          logger,
		version:         version,
		propertyRepo:    propertyRepo,
		mediaUploader:   mediaUploader,
		mediaURLs:       mediaURLs,
		auth:            authService,
		applications:    applications,
		inquiries:       inquiries,
		oidcRateLimiter: newRateLimiter(cfg.Auth.OIDCRateLimitRequests, cfg.Auth.OIDCRateLimitWindow),
	}

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      app.routes(),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}
