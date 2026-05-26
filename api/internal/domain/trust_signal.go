package domain

import "time"

type TrustSignalSubject string

const (
	TrustSignalSubjectProperty   TrustSignalSubject = "property"
	TrustSignalSubjectUnitType   TrustSignalSubject = "property_unit_type"
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
