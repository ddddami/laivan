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
)

func (app *app) discover(w http.ResponseWriter, r *http.Request) {
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
		"lowest_price_naira", "-lowest_price_naira",
	}
	filters.SortColumnMap = map[string]string{
		"lowest_price_naira": "lowest_price_kobo",
	}

	data.ValidateFilters(v, filters)

	categories := readCSV(qs, "category", nil)
	for _, c := range categories {
		v.Check(validUnitCategory(c), "category", "Invalid category: "+c)
	}

	area := readString(qs, "area", "")
	bathroomType := readString(qs, "bathroom_type", "")
	kitchenType := readString(qs, "kitchen_type", "")
	hasParlour := readBool(qs, "has_parlour", nil, v)
	minPrice := readInt(qs, "min_price", 0, v)
	maxPrice := readInt(qs, "max_price", 0, v)

	if !v.Valid() {
		app.validationFailedResponse(w, r, v.FieldErrors)
		return
	}

	filter := repo.DiscoveryFilter{
		CampusID:     domain.ID(campusID),
		Categories:   categories,
		Area:         area,
		BathroomType: bathroomType,
		KitchenType:  kitchenType,
		HasParlour:   hasParlour,
		Filters:      filters,
	}
	if minPrice > 0 {
		filter.MinPrice = &minPrice
	}
	if maxPrice > 0 {
		filter.MaxPrice = &maxPrice
	}

	results, totalRecords, err := app.propertyRepo.Discover(r.Context(), filter)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("discover: %w", err))
		return
	}

	data := envelope{
		"results":  app.discoveryResultsResponse(results),
		"metadata": data.CalculateMetadata(totalRecords, filters.Page, filters.PageSize),
	}

	if err := writeJSON(w, http.StatusOK, data, nil); err != nil {
		app.logger.Error("write discover response", "error", err)
	}
}

func (app *app) discoveryResultsResponse(results []domain.DiscoveryResult) []map[string]any {
	result := make([]map[string]any, 0, len(results))
	for _, item := range results {
		result = append(result, app.discoveryResultResponse(item))
	}
	return result
}

func (app *app) discoveryResultResponse(r domain.DiscoveryResult) map[string]any {
	return map[string]any{
		"property": map[string]any{
			"id":       string(r.PropertyID),
			"name":     r.PropertyName,
			"area":     r.PropertyArea,
			"landmark": nullableString(r.PropertyLandmark),
		},
		"unit_type": map[string]any{
			"id":            string(r.UnitTypeID),
			"category":      string(r.UnitTypeCategory),
			"name":          unitTypeDisplayNameFromResult(r),
			"description":   nullableString(r.UnitTypeDescription),
			"notes":         nullableString(r.UnitTypeNotes),
			"bedroom_count": r.Structure.BedroomCount,
			"has_parlour":   r.Structure.HasParlour,
			"bathroom_type": nullableString(r.Structure.BathroomType),
			"kitchen_type":  nullableString(r.Structure.KitchenType),
		},
		"pricing": map[string]any{
			"lowest_price_naira": r.LowestPrice.Naira(),
		},
		"offer_summary": map[string]any{
			"available_offer_count": r.AvailableOfferCount,
		},
		"thumbnail_url": app.thumbnailURL(r.ThumbnailURL),
		"created_at":    r.CreatedAt.Format(time.RFC3339),
	}
}

// unitTypeDisplayNameFromResult is the DiscoveryResult variant of unitTypeDisplayName.
func unitTypeDisplayNameFromResult(r domain.DiscoveryResult) string {
	if r.UnitTypeName != "" {
		return r.UnitTypeName
	}
	return unitCategoryDisplayName(r.UnitTypeCategory)
}
