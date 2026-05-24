package domain

import "time"

type MediaKind string

const (
	MediaKindImage MediaKind = "image"
	MediaKindVideo MediaKind = "video"
)

type Media struct {
	ID                ID
	PropertyID        ID
	RoomTypeID        ID
	AgentOfferID      ID
	UploadedByAgentID ID
	URL               string
	Kind              MediaKind
	Caption           string
	CreatedAt         time.Time
}
