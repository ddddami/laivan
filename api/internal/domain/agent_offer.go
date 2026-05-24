package domain

type AgentOfferStatus string

const (
	AgentOfferStatusAvailable   AgentOfferStatus = "available"
	AgentOfferStatusUnavailable AgentOfferStatus = "unavailable"
	AgentOfferStatusPaused      AgentOfferStatus = "paused"
)

type AgentOffer struct {
	ID          ID
	RoomTypeID  ID
	AgentID     ID
	Title       string
	Description string
	Price       Money
	Status      AgentOfferStatus
	Timestamps
}
