package domain

import "time"

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
	Version            int
	ArchivedAt         *time.Time
	Timestamps
}

type AgentOfferPatch struct {
	Title       *string
	Description *string
	Notes       *string
	PriceKobo   *int
	Status      *AgentOfferStatus
}

type AgentOfferDetail struct {
	AgentOffer
	Agent AgentSummary
	Media []Media
}
