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

func (app *app) createPropertyUnitType(w http.ResponseWriter, r *http.Request) {
	propertyID := chi.URLParam(r, "id")

	var input struct {
		Category     string `json:"category"`
		Name         string `json:"name"`
		Description  string `json:"description"`
		Notes        string `json:"notes"`
		BedroomCount *int   `json:"bedroom_count"`
		HasParlour   *bool  `json:"has_parlour"`
		BathroomType string `json:"bathroom_type"`
		KitchenType  string `json:"kitchen_type"`
	}

	if err := readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Name == "" {
		input.Name = unitCategoryDisplayName(domain.UnitCategory(input.Category))
	}

	v := validator.New()
	v.Check(validator.NotBlank(propertyID), "id", "ID is required")
	v.Check(validator.ValidUUID(propertyID), "id", "ID must be a valid UUID")
	v.Check(validUnitCategory(input.Category), "category", "Category must be single_room, self_contained, room_and_parlour, one_bedroom_flat, two_bedroom_flat, three_bedroom_flat, or other")
	v.Check(validator.MaxChars(input.Name, 100), "name", "Name must not exceed 100 characters")
	v.Check(validator.MaxChars(input.Description, 1000), "description", "Description must not exceed 1000 characters")
	v.Check(validator.MaxChars(input.Notes, 2000), "notes", "Notes must not exceed 2000 characters")
	if input.BedroomCount != nil {
		v.Check(*input.BedroomCount >= 0, "bedroom_count", "Bedroom count must be greater than or equal to 0")
	}
	v.Check(input.BathroomType == "" || validBathroomType(input.BathroomType), "bathroom_type", "Bathroom type must be private, shared, or unknown")
	v.Check(input.KitchenType == "" || validKitchenType(input.KitchenType), "kitchen_type", "Kitchen type must be private, shared, none, or unknown")

	if !v.Valid() {
		app.validationFailedResponse(w, r, v.FieldErrors)
		return
	}

	p, ok := principalFromContext(r.Context())
	if !ok {
		app.errorResponse(w, r, http.StatusUnauthorized, "unauthenticated", "Authentication is required")
		return
	}
	property, err := app.propertyRepo.Get(r.Context(), domain.ID(propertyID))
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			app.notFoundResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, fmt.Errorf("get property for unit type contribution: %w", err))
		return
	}
	if !canContributeToCampus(p.Access, property.CampusID) {
		app.errorResponse(w, r, http.StatusForbidden, "forbidden", "You are not authorized to contribute to this campus")
		return
	}

	created, err := app.propertyRepo.CreatePropertyUnitType(r.Context(), domain.PropertyUnitType{
		PropertyID:  domain.ID(propertyID),
		Category:    domain.UnitCategory(input.Category),
		Name:        input.Name,
		Description: input.Description,
		Notes:       input.Notes,
		Structure: domain.UnitStructure{
			BedroomCount: input.BedroomCount,
			HasParlour:   input.HasParlour,
			BathroomType: input.BathroomType,
			KitchenType:  input.KitchenType,
		},
	})
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			app.notFoundResponse(w, r)
			return
		}

		app.serverErrorResponse(w, r, fmt.Errorf("create property unit type: %w", err))
		return
	}

	data := envelope{
		"unit_type": propertyUnitTypeResponse(created),
	}

	if err := writeJSON(w, http.StatusCreated, data, nil); err != nil {
		app.logger.Error("write property unit type response", "error", err)
	}
}

func (app *app) listPropertyUnitTypes(w http.ResponseWriter, r *http.Request) {
	propertyID := chi.URLParam(r, "id")

	v := validator.New()
	v.Check(validator.NotBlank(propertyID), "id", "ID is required")
	v.Check(validator.ValidUUID(propertyID), "id", "ID must be a valid UUID")
	if !v.Valid() {
		app.validationFailedResponse(w, r, v.FieldErrors)
		return
	}

	unitTypes, err := app.propertyRepo.ListPropertyUnitTypes(r.Context(), domain.ID(propertyID))
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			app.notFoundResponse(w, r)
			return
		}

		app.serverErrorResponse(w, r, fmt.Errorf("list property unit types: %w", err))
		return
	}

	data := envelope{
		"unit_types": propertyUnitTypesResponse(unitTypes),
	}

	if err := writeJSON(w, http.StatusOK, data, nil); err != nil {
		app.logger.Error("write property unit types response", "error", err)
	}
}

func propertyUnitTypesResponse(unitTypes []domain.PropertyUnitType) []map[string]any {
	return mapItems(unitTypes, propertyUnitTypeResponse)
}

func propertyUnitTypeResponse(ut domain.PropertyUnitType) map[string]any {
	return map[string]any{
		"id":            string(ut.ID),
		"property_id":   string(ut.PropertyID),
		"category":      string(ut.Category),
		"name":          unitTypeDisplayName(ut),
		"description":   ut.Description,
		"version":       ut.Version,
		"notes":         nullableString(ut.Notes),
		"bedroom_count": ut.Structure.BedroomCount,
		"has_parlour":   ut.Structure.HasParlour,
		"bathroom_type": nullableString(ut.Structure.BathroomType),
		"kitchen_type":  nullableString(ut.Structure.KitchenType),
		"created_at":    ut.CreatedAt.Format(time.RFC3339),
		"updated_at":    ut.UpdatedAt.Format(time.RFC3339),
	}
}
