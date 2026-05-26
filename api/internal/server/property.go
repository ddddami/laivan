package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

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

	query := r.URL.Query()
	campusID := query.Get("campus_id")
	limitValue := query.Get("limit")
	limit := defaultPropertyListLimit

	v := validator.New()
	v.Check(validator.NotBlank(campusID), "campus_id", "Campus ID is required")
	v.Check(validator.ValidUUID(campusID), "campus_id", "Campus ID must be a valid UUID")

	if limitValue != "" {
		parsedLimit, err := strconv.Atoi(limitValue)
		if err != nil {
			v.AddFieldError("limit", "Limit must be an integer")
		} else {
			limit = parsedLimit
		}
	}

	v.Check(limit > 0, "limit", "Limit must be greater than 0")
	v.Check(limit <= maxPropertyListLimit, "limit", "Limit must not exceed 100")

	if !v.Valid() {
		app.validationFailedResponse(w, r, v.FieldErrors)
		return
	}

	properties, err := app.propertyRepo.List(r.Context(), repo.PropertyListFilter{
		CampusID: domain.ID(campusID),
		Limit:    int32(limit),
	})
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("list properties: %w", err))
		return
	}

	data := envelope{
		"properties": propertiesResponse(properties),
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

	property, err := app.propertyRepo.Get(r.Context(), domain.ID(id))
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			app.notFoundResponse(w, r)
			return
		}

		app.serverErrorResponse(w, r, fmt.Errorf("get property: %w", err))
		return
	}

	data := envelope{
		"property": propertyResponse(property),
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
