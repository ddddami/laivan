package domain

import "time"

type TrustSignalSubject string

const (
	TrustSignalSubjectProperty   TrustSignalSubject = "property"
	TrustSignalSubjectRoomType   TrustSignalSubject = "room_type"
	TrustSignalSubjectAgentOffer TrustSignalSubject = "agent_offer"
	TrustSignalSubjectAgent      TrustSignalSubject = "agent"
)

type TrustSignal struct {
	ID        ID
	Subject   TrustSignalSubject
	SubjectID ID
	Kind      string
	Value     string
	CreatedAt time.Time
}
