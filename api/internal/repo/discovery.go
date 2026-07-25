package repo

import (
	"context"
	"fmt"

	"github.com/ddddami/laivan/internal/data"
	generateddb "github.com/ddddami/laivan/internal/db/generated"
	"github.com/ddddami/laivan/internal/domain"
)

type DiscoveryAvailability string

const (
	DiscoveryAvailabilityAvailable DiscoveryAvailability = "available"
	DiscoveryAvailabilityAll       DiscoveryAvailability = "all"
)

type DiscoveryFilter struct {
	CampusID     domain.ID
	Categories   []string
	Area         string
	BathroomType string
	KitchenType  string
	HasParlour   *bool
	MinPrice     *int
	MaxPrice     *int
	Availability DiscoveryAvailability
	Filters      data.Filters
}

func (r *PropertyRepository) Discover(ctx context.Context, filter DiscoveryFilter) ([]domain.DiscoveryResult, int, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	campusUUID, err := uuidParam(filter.CampusID)
	if err != nil {
		return nil, 0, err
	}

	sort := filter.Filters.SortColumn()
	if filter.Filters.SortDirection() == "DESC" {
		sort = "-" + sort
	}

	params := generateddb.DiscoverPropertiesParams{
		Sort:          sort,
		ResultOffset:  filter.Filters.Offset(),
		ResultLimit:   filter.Filters.Limit(),
		CampusID:      campusUUID,
		Categories:    []string{},
		Area:          filter.Area,
		BathroomType:  filter.BathroomType,
		KitchenType:   filter.KitchenType,
		HasParlourSet: filter.HasParlour != nil,
		Availability:  string(filter.Availability),
	}
	if len(filter.Categories) > 0 {
		params.Categories = filter.Categories
	}
	if filter.HasParlour != nil {
		params.HasParlour = *filter.HasParlour
	}
	if filter.MinPrice != nil {
		params.MinPriceSet = true
		params.MinPriceKobo = int(domain.Kobo(*filter.MinPrice))
	}
	if filter.MaxPrice != nil {
		params.MaxPriceSet = true
		params.MaxPriceKobo = int(domain.Kobo(*filter.MaxPrice))
	}

	rows, err := r.queries.DiscoverProperties(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("discover: %w", err)
	}

	results := make([]domain.DiscoveryResult, 0, len(rows))
	for _, row := range rows {
		results = append(results, discoveryResultFromRow(row))
	}

	totalCount := 0
	if len(rows) > 0 {
		totalCount = int(rows[0].TotalCount)
	}

	return results, totalCount, nil
}

func discoveryResultFromRow(row generateddb.DiscoverPropertiesRow) domain.DiscoveryResult {
	return domain.DiscoveryResult{
		PropertyID:          domain.ID(uuidString(row.PropertyID)),
		PropertyName:        row.PropertyName,
		PropertyArea:        row.PropertyArea,
		PropertyLandmark:    textString(row.PropertyLandmark),
		UnitTypeID:          domain.ID(uuidString(row.UnitTypeID)),
		UnitTypeCategory:    domain.UnitCategory(row.Category),
		UnitTypeName:        row.UnitTypeName,
		UnitTypeDescription: textString(row.Description),
		UnitTypeNotes:       textString(row.Notes),
		Structure: domain.UnitStructure{
			BedroomCount: intPointer(row.BedroomCount),
			HasParlour:   boolPointer(row.HasParlour),
			BathroomType: textString(row.BathroomType),
			KitchenType:  textString(row.KitchenType),
		},
		LowestPrice:         domain.Money{AmountKobo: row.LowestPriceKobo},
		AvailableOfferCount: row.AvailableOfferCount,
		ThumbnailURL:        row.ThumbnailUrl,
		CreatedAt:           row.CreatedAt.Time,
		UpdatedAt:           row.UpdatedAt.Time,
	}
}
