package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
	setUnitTypeETag(w, created)

	if err := writeJSON(w, http.StatusCreated, data, nil); err != nil {
		app.logger.Error("write property unit type response", "error", err)
	}
}

type optionalUnitString struct {
	value   string
	present bool
}

func (s *optionalUnitString) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return errors.New("unit type patch fields must not be null")
	}

	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return errors.New("unit type patch fields must be strings")
	}
	s.value = value
	s.present = true
	return nil
}

type optionalUnitInt struct {
	value   int
	present bool
}

func (i *optionalUnitInt) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return errors.New("unit type patch fields must not be null")
	}

	var value int
	if err := json.Unmarshal(data, &value); err != nil {
		return errors.New("unit type patch fields must be numbers")
	}
	i.value = value
	i.present = true
	return nil
}

type optionalUnitBool struct {
	value   bool
	present bool
}

func (b *optionalUnitBool) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return errors.New("unit type patch fields must not be null")
	}

	var value bool
	if err := json.Unmarshal(data, &value); err != nil {
		return errors.New("unit type patch fields must be booleans")
	}
	b.value = value
	b.present = true
	return nil
}

func (app *app) updatePropertyUnitType(w http.ResponseWriter, r *http.Request) {
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
	expectedVersion, ok := parseUnitTypeETag(ifMatch, domain.ID(id))
	if !ok {
		app.preconditionFailedResponse(w, r)
		return
	}

	var input struct {
		Category     optionalUnitString `json:"category"`
		Name         optionalUnitString `json:"name"`
		Description  optionalUnitString `json:"description"`
		Notes        optionalUnitString `json:"notes"`
		BedroomCount optionalUnitInt    `json:"bedroom_count"`
		HasParlour   optionalUnitBool   `json:"has_parlour"`
		BathroomType optionalUnitString `json:"bathroom_type"`
		KitchenType  optionalUnitString `json:"kitchen_type"`
	}
	if err := readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v.Check(input.Category.present || input.Name.present || input.Description.present || input.Notes.present || input.BedroomCount.present || input.HasParlour.present || input.BathroomType.present || input.KitchenType.present, "body", "At least one unit type field is required")
	if input.Category.present {
		v.Check(validUnitCategory(input.Category.value), "category", "Category is invalid")
	}
	if input.Name.present {
		v.Check(validator.NotBlank(input.Name.value), "name", "Name is required")
		v.Check(validator.MaxChars(input.Name.value, 100), "name", "Name must not exceed 100 characters")
	}
	if input.Description.present {
		v.Check(validator.MaxChars(input.Description.value, 1000), "description", "Description must not exceed 1000 characters")
	}
	if input.Notes.present {
		v.Check(validator.MaxChars(input.Notes.value, 2000), "notes", "Notes must not exceed 2000 characters")
	}
	if input.BedroomCount.present {
		v.Check(input.BedroomCount.value >= 0, "bedroom_count", "Bedroom count must be greater than or equal to 0")
	}
	if input.BathroomType.present {
		v.Check(input.BathroomType.value == "" || validBathroomType(input.BathroomType.value), "bathroom_type", "Bathroom type is invalid")
	}
	if input.KitchenType.present {
		v.Check(input.KitchenType.value == "" || validKitchenType(input.KitchenType.value), "kitchen_type", "Kitchen type is invalid")
	}
	if !v.Valid() {
		app.validationFailedResponse(w, r, v.FieldErrors)
		return
	}

	unitType, err := app.propertyRepo.GetPropertyUnitType(r.Context(), domain.ID(id))
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			app.notFoundResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, fmt.Errorf("get property unit type for update: %w", err))
		return
	}
	property, err := app.propertyRepo.Get(r.Context(), unitType.PropertyID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			app.notFoundResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, fmt.Errorf("get property for unit type update: %w", err))
		return
	}
	p, ok := principalFromContext(r.Context())
	if !ok || (!p.Access.GlobalAdmin && !containsID(p.Access.CampusOperatorIDs, property.CampusID)) {
		app.errorResponse(w, r, http.StatusForbidden, "forbidden", "Campus operator access is required for this unit type")
		return
	}

	updated, err := app.propertyRepo.UpdatePropertyUnitType(r.Context(), domain.ID(id), expectedVersion, domain.PropertyUnitTypePatch{
		Category:     optionalUnitCategoryValue(input.Category),
		Name:         optionalUnitStringValue(input.Name),
		Description:  optionalUnitStringValue(input.Description),
		Notes:        optionalUnitStringValue(input.Notes),
		BedroomCount: optionalUnitIntValue(input.BedroomCount),
		HasParlour:   optionalUnitBoolValue(input.HasParlour),
		BathroomType: optionalUnitStringValue(input.BathroomType),
		KitchenType:  optionalUnitStringValue(input.KitchenType),
	})
	if err != nil {
		if errors.Is(err, repo.ErrStaleUpdate) {
			app.preconditionFailedResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, fmt.Errorf("update property unit type: %w", err))
		return
	}

	setUnitTypeETag(w, updated)
	if err := writeJSON(w, http.StatusOK, envelope{"unit_type": propertyUnitTypeResponse(updated)}, nil); err != nil {
		app.logger.Error("write updated property unit type response", "error", err)
	}
}

func optionalUnitStringValue(value optionalUnitString) *string {
	if !value.present {
		return nil
	}
	return &value.value
}

func optionalUnitCategoryValue(value optionalUnitString) *domain.UnitCategory {
	if !value.present {
		return nil
	}
	category := domain.UnitCategory(value.value)
	return &category
}

func optionalUnitIntValue(value optionalUnitInt) *int {
	if !value.present {
		return nil
	}
	return &value.value
}

func optionalUnitBoolValue(value optionalUnitBool) *bool {
	if !value.present {
		return nil
	}
	return &value.value
}

func unitTypeETag(unitType domain.PropertyUnitType) string {
	return fmt.Sprintf(`"property-unit-type-%s-%d"`, unitType.ID, unitType.Version)
}

func setUnitTypeETag(w http.ResponseWriter, unitType domain.PropertyUnitType) {
	w.Header().Set("ETag", unitTypeETag(unitType))
}

func parseUnitTypeETag(value string, id domain.ID) (int, bool) {
	prefix := fmt.Sprintf(`"property-unit-type-%s-`, id)
	if !strings.HasPrefix(value, prefix) || !strings.HasSuffix(value, `"`) {
		return 0, false
	}
	version, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(value, prefix), `"`))
	if err != nil || version < 1 {
		return 0, false
	}
	return version, true
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
