package domain

type InquiryStatus string

const (
	InquiryStatusOpen   InquiryStatus = "open"
	InquiryStatusClosed InquiryStatus = "closed"
)

type Inquiry struct {
	ID            ID
	StudentUserID ID
	AgentOfferID  ID
	SubmissionID  ID
	Message       string
	Status        InquiryStatus
	Timestamps
}

type InquiryHandoff struct {
	AgentDisplayName string
	PhoneNumber      string
	WhatsAppNumber   string
	OfferTitle       string
	PropertyName     string
	UnitName         string
	Area             string
}

type InspectionRequestStatus string

const (
	InspectionRequestStatusPending   InspectionRequestStatus = "pending"
	InspectionRequestStatusAccepted  InspectionRequestStatus = "accepted"
	InspectionRequestStatusDeclined  InspectionRequestStatus = "declined"
	InspectionRequestStatusCompleted InspectionRequestStatus = "completed"
)

type InspectionRequest struct {
	ID           ID
	StudentID    ID
	AgentOfferID ID
	Status       InspectionRequestStatus
	Timestamps
}

type ReservationIntentStatus string

const (
	ReservationIntentStatusPending   ReservationIntentStatus = "pending"
	ReservationIntentStatusAccepted  ReservationIntentStatus = "accepted"
	ReservationIntentStatusDeclined  ReservationIntentStatus = "declined"
	ReservationIntentStatusCancelled ReservationIntentStatus = "cancelled"
)

type ReservationIntent struct {
	ID           ID
	StudentID    ID
	AgentOfferID ID
	Status       ReservationIntentStatus
	Timestamps
}
