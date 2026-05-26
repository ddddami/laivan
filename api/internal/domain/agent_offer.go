package domain

type AgentOfferStatus string

const (
	AgentOfferStatusAvailable   AgentOfferStatus = "available"
	AgentOfferStatusUnavailable AgentOfferStatus = "unavailable"
	AgentOfferStatusPaused      AgentOfferStatus = "paused"
)

type AgentOffer struct {
	ID                 ID
	PropertyUnitTypeID ID
	AgentID            ID
	Title              string
	Description        string
	Notes              string
	Price              Money
	Status             AgentOfferStatus
	Timestamps
}
