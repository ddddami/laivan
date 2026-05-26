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
	RoomTypeCount       int
	AvailableOfferCount int
	LowestPriceKobo     int
}

type PropertyDetail struct {
	Property
	RoomTypes []RoomTypeDetail
}

type RoomTypeDetail struct {
	RoomType
	AgentOffers []AgentOffer
}
