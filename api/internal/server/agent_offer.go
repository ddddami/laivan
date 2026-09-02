package server

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ddddami/laivan/internal/domain"
	"github.com/ddddami/laivan/internal/repo"
	"github.com/ddddami/laivan/internal/validator"
	"github.com/go-chi/chi/v5"
)

func (app *app) createAgentOffer(w http.ResponseWriter, r *http.Request) {
	unitTypeID := chi.URLParam(r, "id")
	p, ok := principalFromContext(r.Context())
	if !ok || p.Access.Agent == nil {
		app.errorResponse(w, r, http.StatusForbidden, "agent_required", "An active agent profile is required")
		return
	}

	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Notes       string `json:"notes"`
		PriceNaira  int    `json:"price_naira"`
		Status      string `json:"status"`
	}

	if err := readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Status == "" {
		input.Status = string(domain.AgentOfferStatusAvailable)
	}

	v := validator.New()
	v.Check(validator.NotBlank(unitTypeID), "id", "ID is required")
	v.Check(validator.ValidUUID(unitTypeID), "id", "ID must be a valid UUID")
	v.Check(validator.NotBlank(input.Title), "title", "Title is required")
	v.Check(validator.MaxChars(input.Title, 255), "title", "Title must not exceed 255 characters")
	v.Check(validator.MaxChars(input.Description, 1000), "description", "Description must not exceed 1000 characters")
	v.Check(validator.MaxChars(input.Notes, 2000), "notes", "Notes must not exceed 2000 characters")
	v.Check(input.PriceNaira > 0, "price_naira", "Price must be greater than 0")
	v.Check(validator.MaxValue(input.PriceNaira, domain.MaxNaira), "price_naira", "Price exceeds maximum allowed value")
	v.Check(validAgentOfferStatus(input.Status), "status", "Status must be available, unavailable, or paused")

	if !v.Valid() {
		app.validationFailedResponse(w, r, v.FieldErrors)
		return
	}

	created, err := app.propertyRepo.CreateAgentOffer(r.Context(), domain.AgentOffer{
		PropertyUnitTypeID: domain.ID(unitTypeID),
		AgentID:            p.Access.Agent.ID,
		Title:              input.Title,
		Description:        input.Description,
		Notes:              input.Notes,
		Price:              domain.Money{AmountKobo: domain.Kobo(input.PriceNaira)},
		Status:             domain.AgentOfferStatus(input.Status),
	})
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrUnitTypeNotFound):
			app.notFoundResponse(w, r)
			return
		case errors.Is(err, repo.ErrAgentNotFound):
			app.errorResponse(w, r, http.StatusForbidden, "agent_required", "An active agent profile is required")
			return
		case errors.Is(err, repo.ErrAgentForbidden):
			app.errorResponse(w, r, http.StatusForbidden, "forbidden", "The agent is not authorized for this campus")
			return
		case errors.Is(err, repo.ErrNotFound):
			app.notFoundResponse(w, r)
			return
		case errors.Is(err, repo.ErrDuplicate):
			app.conflictResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, fmt.Errorf("create agent offer: %w", err))
			return
		}
	}

	data := envelope{
		"agent_offer": agentOfferResponse(created),
	}
	setETag(w, "agent-offer", created.ID, created.Version)

	if err := writeJSON(w, http.StatusCreated, data, nil); err != nil {
		app.logger.Error("write agent offer response", "error", err)
	}
}

func (app *app) updateAgentOffer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	v := validator.New()
	v.Check(validator.NotBlank(id), "id", "ID is required")
	v.Check(validator.ValidUUID(id), "id", "ID must be a valid UUID")
	if !v.Valid() {
		app.validationFailedResponse(w, r, v.FieldErrors)
		return
	}

	ifMatch := r.Header.Get("If-Match")
	if ifMatch == "" {
		app.preconditionRequiredResponse(w, r)
		return
	}
	expectedVersion, ok := parseETag(ifMatch, "agent-offer", domain.ID(id))
	if !ok {
		app.preconditionFailedResponse(w, r)
		return
	}

	var input struct {
		Title       patchField[string] `json:"title"`
		Description patchField[string] `json:"description"`
		Notes       patchField[string] `json:"notes"`
		PriceNaira  patchField[int]    `json:"price_naira"`
		Status      patchField[string] `json:"status"`
	}
	if err := readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v.Check(input.Title.Present || input.Description.Present || input.Notes.Present || input.PriceNaira.Present || input.Status.Present, "body", "At least one agent offer field is required")
	if input.Title.Present {
		v.Check(validator.NotBlank(input.Title.Value), "title", "Title is required")
		v.Check(validator.MaxChars(input.Title.Value, 255), "title", "Title must not exceed 255 characters")
	}
	if input.Description.Present {
		v.Check(validator.MaxChars(input.Description.Value, 1000), "description", "Description must not exceed 1000 characters")
	}
	if input.Notes.Present {
		v.Check(validator.MaxChars(input.Notes.Value, 2000), "notes", "Notes must not exceed 2000 characters")
	}
	if input.PriceNaira.Present {
		v.Check(input.PriceNaira.Value > 0, "price_naira", "Price must be greater than 0")
		v.Check(validator.MaxValue(input.PriceNaira.Value, domain.MaxNaira), "price_naira", "Price exceeds maximum allowed value")
	}
	if input.Status.Present {
		v.Check(validAgentOfferStatus(input.Status.Value), "status", "Status must be available, unavailable, or paused")
	}
	if !v.Valid() {
		app.validationFailedResponse(w, r, v.FieldErrors)
		return
	}

	p, ok := principalFromContext(r.Context())
	if !ok {
		app.errorResponse(w, r, http.StatusUnauthorized, "unauthenticated", "Authentication is required")
		return
	}

	var priceKobo *int
	if input.PriceNaira.Present {
		value := domain.Kobo(input.PriceNaira.Value)
		priceKobo = &value
	}
	var status *domain.AgentOfferStatus
	if input.Status.Present {
		value := domain.AgentOfferStatus(input.Status.Value)
		status = &value
	}
	updated, err := app.propertyRepo.UpdateAgentOffer(r.Context(), domain.ID(id), expectedVersion, domain.AgentOfferPatch{
		Title:       patchValue(input.Title),
		Description: patchValue(input.Description),
		Notes:       patchValue(input.Notes),
		PriceKobo:   priceKobo,
		Status:      status,
	}, p.User.ID)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			app.notFoundResponse(w, r)
		case errors.Is(err, repo.ErrStaleUpdate):
			app.preconditionFailedResponse(w, r)
		case errors.Is(err, repo.ErrAgentForbidden):
			app.errorResponse(w, r, http.StatusForbidden, "forbidden", "You are not authorized to update this agent offer")
		default:
			app.serverErrorResponse(w, r, fmt.Errorf("update agent offer: %w", err))
		}
		return
	}

	setETag(w, "agent-offer", updated.ID, updated.Version)
	if err := writeJSON(w, http.StatusOK, envelope{"agent_offer": agentOfferResponse(updated)}, nil); err != nil {
		app.logger.Error("write updated agent offer response", "error", err)
	}
}

func (app *app) archiveAgentOffer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	v := validator.New()
	v.Check(validator.NotBlank(id), "id", "ID is required")
	v.Check(validator.ValidUUID(id), "id", "ID must be a valid UUID")
	if !v.Valid() {
		app.validationFailedResponse(w, r, v.FieldErrors)
		return
	}

	ifMatch := r.Header.Get("If-Match")
	if ifMatch == "" {
		app.preconditionRequiredResponse(w, r)
		return
	}
	expectedVersion, ok := parseETag(ifMatch, "agent-offer", domain.ID(id))
	if !ok {
		app.preconditionFailedResponse(w, r)
		return
	}
	p, ok := principalFromContext(r.Context())
	if !ok {
		app.errorResponse(w, r, http.StatusUnauthorized, "unauthenticated", "Authentication is required")
		return
	}
	archived, err := app.propertyRepo.ArchiveAgentOffer(r.Context(), domain.ID(id), expectedVersion, p.User.ID)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			app.notFoundResponse(w, r)
		case errors.Is(err, repo.ErrStaleUpdate):
			app.preconditionFailedResponse(w, r)
		case errors.Is(err, repo.ErrAgentForbidden):
			app.errorResponse(w, r, http.StatusForbidden, "forbidden", "You are not authorized to archive this agent offer")
		default:
			app.serverErrorResponse(w, r, fmt.Errorf("archive agent offer: %w", err))
		}
		return
	}

	setETag(w, "agent-offer", archived.ID, archived.Version)
	if err := writeJSON(w, http.StatusOK, envelope{"agent_offer": agentOfferResponse(archived)}, nil); err != nil {
		app.logger.Error("write archived agent offer response", "error", err)
	}
}

func (app *app) listAgentOffers(w http.ResponseWriter, r *http.Request) {
	unitTypeID := chi.URLParam(r, "id")

	v := validator.New()
	v.Check(validator.NotBlank(unitTypeID), "id", "ID is required")
	v.Check(validator.ValidUUID(unitTypeID), "id", "ID must be a valid UUID")
	if !v.Valid() {
		app.validationFailedResponse(w, r, v.FieldErrors)
		return
	}

	offers, err := app.propertyRepo.ListAgentOffers(r.Context(), domain.ID(unitTypeID))
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			app.notFoundResponse(w, r)
			return
		}

		app.serverErrorResponse(w, r, fmt.Errorf("list agent offers: %w", err))
		return
	}

	data := envelope{
		"agent_offers": agentOffersResponse(offers),
	}

	if err := writeJSON(w, http.StatusOK, data, nil); err != nil {
		app.logger.Error("write agent offers response", "error", err)
	}
}

func agentOffersResponse(offers []domain.AgentOffer) []map[string]any {
	return mapItems(offers, agentOfferResponse)
}

func agentOfferResponse(offer domain.AgentOffer) map[string]any {
	return map[string]any{
		"id":                    string(offer.ID),
		"property_unit_type_id": string(offer.PropertyUnitTypeID),
		"agent_id":              string(offer.AgentID),
		"title":                 offer.Title,
		"description":           offer.Description,
		"notes":                 nullableString(offer.Notes),
		"price_naira":           offer.Price.Naira(),
		"status":                string(offer.Status),
		"version":               offer.Version,
		"archived_at":           nullableTime(offer.ArchivedAt),
		"created_at":            offer.CreatedAt.Format(time.RFC3339),
		"updated_at":            offer.UpdatedAt.Format(time.RFC3339),
	}
}

func validAgentOfferStatus(status string) bool {
	switch domain.AgentOfferStatus(status) {
	case domain.AgentOfferStatusAvailable, domain.AgentOfferStatusUnavailable, domain.AgentOfferStatusPaused:
		return true
	default:
		return false
	}
}
