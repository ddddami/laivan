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
		"area", "-area",
		"lowest_price_naira", "-lowest_price_naira",
		"available_offer_count", "-available_offer_count",
	}
	filters.SortColumnMap = map[string]string{
		"lowest_price_naira": "lowest_price_kobo",
	}

	data.ValidateFilters(v, filters)

	area := readString(qs, "area", "")
	name := readString(qs, "name", "")
	hasOffers := readBool(qs, "has_offers", nil, v)
	minPrice := readInt(qs, "min_price", 0, v)
	maxPrice := readInt(qs, "max_price", 0, v)

	if !v.Valid() {
		app.validationFailedResponse(w, r, v.FieldErrors)
		return
	}

	filter := repo.PropertyListFilter{
		CampusID:  domain.ID(campusID),
		Area:      area,
		Name:      name,
		HasOffers: hasOffers,
		Filters:   filters,
	}
	if minPrice > 0 {
		filter.MinPrice = &minPrice
	}
	if maxPrice > 0 {
		filter.MaxPrice = &maxPrice
	}

	properties, totalRecords, err := app.propertyRepo.ListWithSummary(r.Context(), filter)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("list properties: %w", err))
		return
	}

	data := envelope{
		"properties": app.propertiesSummaryResponse(properties),
		"metadata":   data.CalculateMetadata(totalRecords, filters.Page, filters.PageSize),
	}

	if err := writeJSON(w, http.StatusOK, data, nil); err != nil {
		app.logger.Error("write properties response", "error", err)
	}
}

func (app *app) createProperty(w http.ResponseWriter, r *http.Request) {
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

	p, ok := principalFromContext(r.Context())
	if !ok || !canContributeToCampus(p.Access, domain.ID(input.CampusID)) {
		app.errorResponse(w, r, http.StatusForbidden, "forbidden", "You are not authorized to contribute to this campus")
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
	setETag(w, "property", created.ID, created.Version)

	data := envelope{
		"property": propertyResponse(created),
	}

	if err := writeJSON(w, http.StatusCreated, data, nil); err != nil {
		app.logger.Error("write property response", "error", err)
	}
}

func (app *app) getProperty(w http.ResponseWriter, r *http.Request) {
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
	setETag(w, "property", property.ID, property.Version)

	data := envelope{
		"property": app.propertyDetailResponse(property),
	}

	if err := writeJSON(w, http.StatusOK, data, nil); err != nil {
		app.logger.Error("write property response", "error", err)
	}
}

func (app *app) updateProperty(w http.ResponseWriter, r *http.Request) {
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
	expectedVersion, ok := parseETag(ifMatch, "property", domain.ID(id))
	if !ok {
		app.preconditionFailedResponse(w, r)
		return
	}

	var input struct {
		Name        patchField[string] `json:"name"`
		Area        patchField[string] `json:"area"`
		Landmark    patchField[string] `json:"landmark"`
		Description patchField[string] `json:"description"`
	}
	if err := readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v.Check(input.Name.Present || input.Area.Present || input.Landmark.Present || input.Description.Present, "body", "At least one property field is required")
	if input.Name.Present {
		v.Check(validator.NotBlank(input.Name.Value), "name", "Name is required")
		v.Check(validator.MaxChars(input.Name.Value, 255), "name", "Name must not exceed 255 characters")
	}
	if input.Area.Present {
		v.Check(validator.NotBlank(input.Area.Value), "area", "Area is required")
		v.Check(validator.MaxChars(input.Area.Value, 100), "area", "Area must not exceed 100 characters")
	}
	if input.Landmark.Present {
		v.Check(validator.MaxChars(input.Landmark.Value, 100), "landmark", "Landmark must not exceed 100 characters")
	}
	if input.Description.Present {
		v.Check(validator.MaxChars(input.Description.Value, 1000), "description", "Description must not exceed 1000 characters")
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

	updated, err := app.propertyRepo.Update(r.Context(), domain.ID(id), expectedVersion, domain.PropertyPatch{
		Name:        patchValue(input.Name),
		Area:        patchValue(input.Area),
		Landmark:    patchValue(input.Landmark),
		Description: patchValue(input.Description),
	}, p.User.ID)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			app.notFoundResponse(w, r)
		case errors.Is(err, repo.ErrStaleUpdate):
			app.preconditionFailedResponse(w, r)
		case errors.Is(err, repo.ErrCampusForbidden):
			app.errorResponse(w, r, http.StatusForbidden, "forbidden", "Campus operator access is required for this property")
		default:
			app.serverErrorResponse(w, r, fmt.Errorf("update property: %w", err))
		}
		return
	}

	setETag(w, "property", updated.ID, updated.Version)
	if err := writeJSON(w, http.StatusOK, envelope{"property": propertyResponse(updated)}, nil); err != nil {
		app.logger.Error("write updated property response", "error", err)
	}
}

func propertyResponse(p domain.Property) map[string]any {
	return map[string]any{
		"id":          string(p.ID),
		"campus_id":   string(p.CampusID),
		"name":        p.Name,
		"area":        p.Location.Area,
		"landmark":    p.Location.Landmark,
		"description": p.Description,
		"version":     p.Version,
		"created_at":  p.CreatedAt.Format(time.RFC3339),
		"updated_at":  p.UpdatedAt.Format(time.RFC3339),
	}
}

func (app *app) propertiesSummaryResponse(properties []domain.PropertySummary) []map[string]any {
	result := make([]map[string]any, 0, len(properties))
	for _, property := range properties {
		result = append(result, app.propertySummaryResponse(property))
	}
	return result
}

func (app *app) propertySummaryResponse(p domain.PropertySummary) map[string]any {
	return map[string]any{
		"id":                    string(p.ID),
		"campus_id":             string(p.CampusID),
		"name":                  p.Name,
		"area":                  p.Location.Area,
		"landmark":              p.Location.Landmark,
		"description":           p.Description,
		"version":               p.Version,
		"unit_type_count":       p.UnitTypeCount,
		"available_offer_count": p.AvailableOfferCount,
		"lowest_price_naira":    p.LowestPrice.Naira(),
		"thumbnail_url":         app.thumbnailURL(p.ThumbnailURL),
		"created_at":            p.CreatedAt.Format(time.RFC3339),
		"updated_at":            p.UpdatedAt.Format(time.RFC3339),
	}
}

func (app *app) propertyDetailResponse(p domain.PropertyDetail) map[string]any {
	return map[string]any{
		"id":          string(p.ID),
		"campus_id":   string(p.CampusID),
		"name":        p.Name,
		"area":        p.Location.Area,
		"landmark":    p.Location.Landmark,
		"description": p.Description,
		"version":     p.Version,
		"media":       app.mediaListResponse(p.Media),
		"unit_types":  app.propertyUnitTypeDetailsResponse(p.UnitTypes),
		"created_at":  p.CreatedAt.Format(time.RFC3339),
		"updated_at":  p.UpdatedAt.Format(time.RFC3339),
	}
}

func (app *app) propertyUnitTypeDetailsResponse(unitTypes []domain.PropertyUnitTypeDetail) []map[string]any {
	result := make([]map[string]any, 0, len(unitTypes))
	for _, unitType := range unitTypes {
		result = append(result, app.propertyUnitTypeDetailResponse(unitType))
	}
	return result
}

func (app *app) propertyUnitTypeDetailResponse(ut domain.PropertyUnitTypeDetail) map[string]any {
	return map[string]any{
		"id":            string(ut.ID),
		"property_id":   string(ut.PropertyID),
		"category":      string(ut.Category),
		"name":          unitTypeDisplayName(ut.PropertyUnitType),
		"description":   ut.Description,
		"version":       ut.Version,
		"notes":         nullableString(ut.Notes),
		"bedroom_count": ut.Structure.BedroomCount,
		"has_parlour":   ut.Structure.HasParlour,
		"bathroom_type": nullableString(ut.Structure.BathroomType),
		"kitchen_type":  nullableString(ut.Structure.KitchenType),
		"media":         app.mediaListResponse(ut.Media),
		"agent_offers":  app.agentOfferDetailsResponse(ut.AgentOffers),
		"created_at":    ut.CreatedAt.Format(time.RFC3339),
		"updated_at":    ut.UpdatedAt.Format(time.RFC3339),
	}
}

func (app *app) agentOfferDetailsResponse(offers []domain.AgentOfferDetail) []map[string]any {
	return mapItems(offers, app.agentOfferDetailResponse)
}

func (app *app) agentOfferDetailResponse(offer domain.AgentOfferDetail) map[string]any {
	response := agentOfferResponse(offer.AgentOffer)
	response["agent"] = map[string]any{
		"id":           string(offer.Agent.ID),
		"display_name": offer.Agent.DisplayName,
	}
	response["media"] = app.mediaListResponse(offer.Media)
	return response
}

func (app *app) thumbnailURL(sourceURL string) any {
	if sourceURL == "" {
		return nil
	}
	if app.mediaURLs == nil {
		return sourceURL
	}
	return app.mediaURLs.ThumbnailURL(sourceURL)
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}

	return value
}

func validUnitCategory(category string) bool {
	switch domain.UnitCategory(category) {
	case domain.UnitCategorySingleRoom,
		domain.UnitCategorySelfContained,
		domain.UnitCategoryRoomAndParlour,
		domain.UnitCategoryOneBedroomFlat,
		domain.UnitCategoryTwoBedroomFlat,
		domain.UnitCategoryThreeBedroomFlat,
		domain.UnitCategoryOther:
		return true
	default:
		return false
	}
}

func unitCategoryDisplayName(category domain.UnitCategory) string {
	switch category {
	case domain.UnitCategorySingleRoom:
		return "Single Room"
	case domain.UnitCategorySelfContained:
		return "Self-contained"
	case domain.UnitCategoryRoomAndParlour:
		return "Room and Parlour"
	case domain.UnitCategoryOneBedroomFlat:
		return "One-bedroom Flat"
	case domain.UnitCategoryTwoBedroomFlat:
		return "Two-bedroom Flat"
	case domain.UnitCategoryThreeBedroomFlat:
		return "Three-bedroom Flat"
	default:
		return "Other"
	}
}

func validBathroomType(value string) bool {
	switch value {
	case "private", "shared", "unknown":
		return true
	default:
		return false
	}
}

func validKitchenType(value string) bool {
	switch value {
	case "private", "shared", "none", "unknown":
		return true
	default:
		return false
	}
}

// mapItems transforms a slice of domain items into a slice of response maps.
// It replaces the repetitive for...append pattern in response builders.
func mapItems[T any](items []T, fn func(T) map[string]any) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, fn(item))
	}
	return result
}

// unitTypeDisplayName returns the unit type name, defaulting to the category
// display name when no custom name is set.
func unitTypeDisplayName(ut domain.PropertyUnitType) string {
	if ut.Name != "" {
		return ut.Name
	}
	return unitCategoryDisplayName(ut.Category)
}
