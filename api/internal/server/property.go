package server

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ddddami/laivan/internal/data"
	"github.com/ddddami/laivan/internal/domain"
	"github.com/ddddami/laivan/internal/repo"
	"github.com/ddddami/laivan/internal/validator"
	"github.com/go-chi/chi/v5"
)

const (
	defaultPropertyListLimit = 20
	maxPropertyListLimit     = 100
)

func (app *app) listProperties(w http.ResponseWriter, r *http.Request) {
	if app.propertyRepo == nil {
		app.serverErrorResponse(w, r, errors.New("database not available"))
		return
	}

	qs := r.URL.Query()
	campusID := readString(qs, "campus_id", "")

	v := validator.New()
	v.Check(validator.NotBlank(campusID), "campus_id", "Campus ID is required")
	v.Check(validator.ValidUUID(campusID), "campus_id", "Campus ID must be a valid UUID")

	var filters data.Filters
	filters.Page = readInt(qs, "page", 1, v)
	filters.PageSize = readInt(qs, "page_size", defaultPropertyListLimit, v)
	filters.Sort = readString(qs, "sort", "-created_at")
	filters.SortSafelist = []string{
		"created_at", "-created_at",
		"name", "-name",
		"lowest_price_naira", "-lowest_price_naira",
		"available_offer_count", "-available_offer_count",
	}
	filters.SortColumnMap = map[string]string{
		"lowest_price_naira": "lowest_price_kobo",
	}

	data.ValidateFilters(v, filters)

	if !v.Valid() {
		app.validationFailedResponse(w, r, v.FieldErrors)
		return
	}

	area := readString(qs, "area", "")

	properties, totalRecords, err := app.propertyRepo.ListWithSummary(r.Context(), repo.PropertyListFilter{
		CampusID: domain.ID(campusID),
		Area:     area,
		Filters:  filters,
	})
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("list properties: %w", err))
		return
	}

	data := envelope{
		"properties": propertiesSummaryResponse(properties),
		"metadata":   data.CalculateMetadata(totalRecords, filters.Page, filters.PageSize),
	}

	if err := writeJSON(w, http.StatusOK, data, nil); err != nil {
		app.logger.Error("write properties response", "error", err)
	}
}

func (app *app) createProperty(w http.ResponseWriter, r *http.Request) {
	if app.propertyRepo == nil {
		app.serverErrorResponse(w, r, errors.New("database not available"))
		return
	}

	var input struct {
		CampusID    string `json:"campus_id"`
		Name        string `json:"name"`
		Area        string `json:"area"`
		Landmark    string `json:"landmark"`
		Description string `json:"description"`
	}

	if err := readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	v.Check(validator.NotBlank(input.CampusID), "campus_id", "Campus ID is required")
	v.Check(validator.ValidUUID(input.CampusID), "campus_id", "Campus ID must be a valid UUID")
	v.Check(validator.NotBlank(input.Name), "name", "Name is required")
	v.Check(validator.MaxChars(input.Name, 255), "name", "Name must not exceed 255 characters")
	v.Check(validator.NotBlank(input.Area), "area", "Area is required")
	v.Check(validator.MaxChars(input.Area, 100), "area", "Area must not exceed 100 characters")
	v.Check(validator.MaxChars(input.Landmark, 100), "landmark", "Landmark must not exceed 100 characters")
	v.Check(validator.MaxChars(input.Description, 1000), "description", "Description must not exceed 1000 characters")

	if !v.Valid() {
		app.validationFailedResponse(w, r, v.FieldErrors)
		return
	}

	property := domain.Property{
		CampusID: domain.ID(input.CampusID),
		Name:     input.Name,
		Location: domain.ApproxLocation{
			Area:     input.Area,
			Landmark: input.Landmark,
		},
		Description: input.Description,
	}

	created, err := app.propertyRepo.Create(r.Context(), property)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("create property: %w", err))
		return
	}

	data := envelope{
		"property": propertyResponse(created),
	}

	if err := writeJSON(w, http.StatusCreated, data, nil); err != nil {
		app.logger.Error("write property response", "error", err)
	}
}

func (app *app) getProperty(w http.ResponseWriter, r *http.Request) {
	if app.propertyRepo == nil {
		app.serverErrorResponse(w, r, errors.New("database not available"))
		return
	}

	id := chi.URLParam(r, "id")

	v := validator.New()
	v.Check(validator.NotBlank(id), "id", "ID is required")
	v.Check(validator.ValidUUID(id), "id", "ID must be a valid UUID")
	if !v.Valid() {
		app.validationFailedResponse(w, r, v.FieldErrors)
		return
	}

	property, err := app.propertyRepo.GetWithDetails(r.Context(), domain.ID(id))
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			app.notFoundResponse(w, r)
			return
		}

		app.serverErrorResponse(w, r, fmt.Errorf("get property: %w", err))
		return
	}

	data := envelope{
		"property": propertyDetailResponse(property),
	}

	if err := writeJSON(w, http.StatusOK, data, nil); err != nil {
		app.logger.Error("write property response", "error", err)
	}
}

func (app *app) createRoomType(w http.ResponseWriter, r *http.Request) {
	if app.propertyRepo == nil {
		app.serverErrorResponse(w, r, errors.New("database not available"))
		return
	}

	propertyID := chi.URLParam(r, "id")

	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	v.Check(validator.NotBlank(propertyID), "id", "ID is required")
	v.Check(validator.ValidUUID(propertyID), "id", "ID must be a valid UUID")
	v.Check(validator.NotBlank(input.Name), "name", "Name is required")
	v.Check(validator.MaxChars(input.Name, 100), "name", "Name must not exceed 100 characters")
	v.Check(validator.MaxChars(input.Description, 1000), "description", "Description must not exceed 1000 characters")

	if !v.Valid() {
		app.validationFailedResponse(w, r, v.FieldErrors)
		return
	}

	created, err := app.propertyRepo.CreateRoomType(r.Context(), domain.RoomType{
		PropertyID:  domain.ID(propertyID),
		Name:        input.Name,
		Description: input.Description,
	})
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			app.notFoundResponse(w, r)
			return
		}

		app.serverErrorResponse(w, r, fmt.Errorf("create room type: %w", err))
		return
	}

	data := envelope{
		"room_type": roomTypeResponse(created),
	}

	if err := writeJSON(w, http.StatusCreated, data, nil); err != nil {
		app.logger.Error("write room type response", "error", err)
	}
}

func (app *app) listRoomTypes(w http.ResponseWriter, r *http.Request) {
	if app.propertyRepo == nil {
		app.serverErrorResponse(w, r, errors.New("database not available"))
		return
	}

	propertyID := chi.URLParam(r, "id")

	v := validator.New()
	v.Check(validator.NotBlank(propertyID), "id", "ID is required")
	v.Check(validator.ValidUUID(propertyID), "id", "ID must be a valid UUID")
	if !v.Valid() {
		app.validationFailedResponse(w, r, v.FieldErrors)
		return
	}

	roomTypes, err := app.propertyRepo.ListRoomTypes(r.Context(), domain.ID(propertyID))
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			app.notFoundResponse(w, r)
			return
		}

		app.serverErrorResponse(w, r, fmt.Errorf("list room types: %w", err))
		return
	}

	data := envelope{
		"room_types": roomTypesResponse(roomTypes),
	}

	if err := writeJSON(w, http.StatusOK, data, nil); err != nil {
		app.logger.Error("write room types response", "error", err)
	}
}

func (app *app) createAgentOffer(w http.ResponseWriter, r *http.Request) {
	if app.propertyRepo == nil {
		app.serverErrorResponse(w, r, errors.New("database not available"))
		return
	}

	roomTypeID := chi.URLParam(r, "id")

	var input struct {
		AgentID     string `json:"agent_id"`
		Title       string `json:"title"`
		Description string `json:"description"`
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
	v.Check(validator.NotBlank(roomTypeID), "id", "ID is required")
	v.Check(validator.ValidUUID(roomTypeID), "id", "ID must be a valid UUID")
	v.Check(validator.NotBlank(input.AgentID), "agent_id", "Agent ID is required")
	v.Check(validator.ValidUUID(input.AgentID), "agent_id", "Agent ID must be a valid UUID")
	v.Check(validator.NotBlank(input.Title), "title", "Title is required")
	v.Check(validator.MaxChars(input.Title, 255), "title", "Title must not exceed 255 characters")
	v.Check(validator.MaxChars(input.Description, 1000), "description", "Description must not exceed 1000 characters")
	v.Check(input.PriceNaira > 0, "price_naira", "Price must be greater than 0")
	v.Check(validAgentOfferStatus(input.Status), "status", "Status must be available, unavailable, or paused")

	if !v.Valid() {
		app.validationFailedResponse(w, r, v.FieldErrors)
		return
	}

	created, err := app.propertyRepo.CreateAgentOffer(r.Context(), domain.AgentOffer{
		RoomTypeID:  domain.ID(roomTypeID),
		AgentID:     domain.ID(input.AgentID),
		Title:       input.Title,
		Description: input.Description,
		Price:       domain.Money{AmountKobo: input.PriceNaira * 100},
		Status:      domain.AgentOfferStatus(input.Status),
	})
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			app.notFoundResponse(w, r)
			return
		}

		app.serverErrorResponse(w, r, fmt.Errorf("create agent offer: %w", err))
		return
	}

	data := envelope{
		"agent_offer": agentOfferResponse(created),
	}

	if err := writeJSON(w, http.StatusCreated, data, nil); err != nil {
		app.logger.Error("write agent offer response", "error", err)
	}
}

func (app *app) listAgentOffers(w http.ResponseWriter, r *http.Request) {
	if app.propertyRepo == nil {
		app.serverErrorResponse(w, r, errors.New("database not available"))
		return
	}

	roomTypeID := chi.URLParam(r, "id")

	v := validator.New()
	v.Check(validator.NotBlank(roomTypeID), "id", "ID is required")
	v.Check(validator.ValidUUID(roomTypeID), "id", "ID must be a valid UUID")
	if !v.Valid() {
		app.validationFailedResponse(w, r, v.FieldErrors)
		return
	}

	offers, err := app.propertyRepo.ListAgentOffers(r.Context(), domain.ID(roomTypeID))
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

func propertiesResponse(properties []domain.Property) []map[string]any {
	response := make([]map[string]any, 0, len(properties))
	for _, property := range properties {
		response = append(response, propertyResponse(property))
	}

	return response
}

func propertyResponse(p domain.Property) map[string]any {
	return map[string]any{
		"id":          string(p.ID),
		"campus_id":   string(p.CampusID),
		"name":        p.Name,
		"area":        p.Location.Area,
		"landmark":    p.Location.Landmark,
		"description": p.Description,
		"created_at":  p.CreatedAt.Format(time.RFC3339),
		"updated_at":  p.UpdatedAt.Format(time.RFC3339),
	}
}

func roomTypesResponse(roomTypes []domain.RoomType) []map[string]any {
	response := make([]map[string]any, 0, len(roomTypes))
	for _, roomType := range roomTypes {
		response = append(response, roomTypeResponse(roomType))
	}

	return response
}

func roomTypeResponse(rt domain.RoomType) map[string]any {
	return map[string]any{
		"id":          string(rt.ID),
		"property_id": string(rt.PropertyID),
		"name":        rt.Name,
		"description": rt.Description,
		"created_at":  rt.CreatedAt.Format(time.RFC3339),
		"updated_at":  rt.UpdatedAt.Format(time.RFC3339),
	}
}

func propertiesSummaryResponse(properties []domain.PropertySummary) []map[string]any {
	response := make([]map[string]any, 0, len(properties))
	for _, property := range properties {
		response = append(response, propertySummaryResponse(property))
	}

	return response
}

func propertySummaryResponse(p domain.PropertySummary) map[string]any {
	return map[string]any{
		"id":                    string(p.ID),
		"campus_id":             string(p.CampusID),
		"name":                  p.Name,
		"area":                  p.Location.Area,
		"landmark":              p.Location.Landmark,
		"description":           p.Description,
		"room_type_count":       p.RoomTypeCount,
		"available_offer_count": p.AvailableOfferCount,
		"lowest_price_naira":    p.LowestPriceKobo / 100,
		"created_at":            p.CreatedAt.Format(time.RFC3339),
		"updated_at":            p.UpdatedAt.Format(time.RFC3339),
	}
}

func propertyDetailResponse(p domain.PropertyDetail) map[string]any {
	return map[string]any{
		"id":          string(p.ID),
		"campus_id":   string(p.CampusID),
		"name":        p.Name,
		"area":        p.Location.Area,
		"landmark":    p.Location.Landmark,
		"description": p.Description,
		"room_types":  roomTypeDetailsResponse(p.RoomTypes),
		"created_at":  p.CreatedAt.Format(time.RFC3339),
		"updated_at":  p.UpdatedAt.Format(time.RFC3339),
	}
}

func roomTypeDetailsResponse(roomTypes []domain.RoomTypeDetail) []map[string]any {
	response := make([]map[string]any, 0, len(roomTypes))
	for _, rt := range roomTypes {
		response = append(response, roomTypeDetailResponse(rt))
	}

	return response
}

func roomTypeDetailResponse(rt domain.RoomTypeDetail) map[string]any {
	return map[string]any{
		"id":           string(rt.ID),
		"property_id":  string(rt.PropertyID),
		"name":         rt.Name,
		"description":  rt.Description,
		"agent_offers": agentOffersResponse(rt.AgentOffers),
		"created_at":   rt.CreatedAt.Format(time.RFC3339),
		"updated_at":   rt.UpdatedAt.Format(time.RFC3339),
	}
}

func agentOffersResponse(offers []domain.AgentOffer) []map[string]any {
	response := make([]map[string]any, 0, len(offers))
	for _, offer := range offers {
		response = append(response, agentOfferResponse(offer))
	}

	return response
}

func agentOfferResponse(offer domain.AgentOffer) map[string]any {
	return map[string]any{
		"id":           string(offer.ID),
		"room_type_id": string(offer.RoomTypeID),
		"agent_id":     string(offer.AgentID),
		"title":        offer.Title,
		"description":  offer.Description,
		"price_naira":  offer.Price.AmountKobo / 100,
		"status":       string(offer.Status),
		"created_at":   offer.CreatedAt.Format(time.RFC3339),
		"updated_at":   offer.UpdatedAt.Format(time.RFC3339),
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
