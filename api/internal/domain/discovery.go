package domain

import "time"

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
	CreatedAt           time.Time
}
