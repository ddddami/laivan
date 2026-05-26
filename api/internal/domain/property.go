package domain

type Property struct {
	ID          ID
	CampusID    ID
	Name        string
	Location    ApproxLocation
	Description string
	Timestamps
}

type PropertySummary struct {
	Property
	UnitTypeCount       int
	AvailableOfferCount int
	LowestPriceKobo     int
}

type PropertyDetail struct {
	Property
	UnitTypes []PropertyUnitTypeDetail
}

type PropertyUnitTypeDetail struct {
	PropertyUnitType
	AgentOffers []AgentOffer
}
