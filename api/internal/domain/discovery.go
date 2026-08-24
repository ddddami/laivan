package domain

import "time"

type DiscoveryThumbnailSource string

const (
	DiscoveryThumbnailSourceUnitType DiscoveryThumbnailSource = "unit_type"
	DiscoveryThumbnailSourceProperty DiscoveryThumbnailSource = "property"
	DiscoveryThumbnailSourceNone     DiscoveryThumbnailSource = "none"
)

// DiscoveryResult represents a searchable rentable opportunity.
type DiscoveryResult struct {
	PropertyID          ID
	PropertyName        string
	PropertyArea        string
	PropertyLandmark    string
	UnitTypeID          ID
	UnitTypeCategory    UnitCategory
	UnitTypeName        string
	UnitTypeDescription string
	UnitTypeNotes       string
	Structure           UnitStructure
	LowestPrice         Money
	AvailableOfferCount int
	ThumbnailURL        string
	ThumbnailSource     DiscoveryThumbnailSource
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
