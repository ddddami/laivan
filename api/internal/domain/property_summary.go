package domain

type PropertySummary struct {
	Property
	RoomTypeCount       int32
	AvailableOfferCount int32
	LowestPriceKobo     int32
}
