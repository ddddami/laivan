package domain

type Property struct {
	ID          ID
	CampusID    ID
	Name        string
	Location    ApproxLocation
	Description string
	Version     int
	Timestamps
}

type PropertyPatch struct {
	Name        *string
	Area        *string
	Landmark    *string
	Description *string
}

type PropertySummary struct {
	Property
	UnitTypeCount       int
	AvailableOfferCount int
	LowestPrice         Money
	ThumbnailURL        string
}

type PropertyDetail struct {
	Property
	Media     []Media
	UnitTypes []PropertyUnitTypeDetail
}

type PropertyUnitTypeDetail struct {
	PropertyUnitType
	Media       []Media
	AgentOffers []AgentOfferDetail
}
