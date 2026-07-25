package domain

type Agent struct {
	ID          ID
	DisplayName string
	PhoneNumber string
	// Optional override when the agent's WhatsApp number differs from PhoneNumber.
	WhatsAppNumber string
	Timestamps
}

type AgentSummary struct {
	ID          ID
	DisplayName string
}

type Student struct {
	ID          ID
	DisplayName string
	PhoneNumber string
	Timestamps
}

type Admin struct {
	ID          ID
	DisplayName string
	Email       string
	Timestamps
}
