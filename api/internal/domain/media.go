package domain

import "time"

type MediaKind string

const (
	MediaKindImage MediaKind = "image"
	MediaKindVideo MediaKind = "video"
)

type Media struct {
	ID                 ID
	PropertyID         ID
	PropertyUnitTypeID ID
	AgentOfferID       ID
	UploadedByAgentID  ID
	URL                string
	ObjectKey          string
	Kind               MediaKind
	Caption            string
	ContentType        string
	SizeBytes          int64
	CreatedAt          time.Time
	RemovedAt          *time.Time
	RemovedByUserID    ID
}
