package domain

import "time"

type IdentityProvider string

const IdentityProviderGoogle IdentityProvider = "google"

type ExternalIdentity struct {
	Provider IdentityProvider
	Subject  string
}

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
)

type User struct {
	ID          ID
	Email       string
	DisplayName string
	Status      UserStatus
	Timestamps
}

type Session struct {
	ID            ID
	User          User
	TokenHash     []byte
	CSRFTokenHash []byte
	CreatedAt     time.Time
	ExpiresAt     time.Time
	RevokedAt     *time.Time
	LastUsedAt    *time.Time
}

type EffectiveAccess struct {
	Roles             []string
	Agent             *LinkedAgent
	AgentCampusIDs    []ID
	CampusOperatorIDs []ID
	GlobalAdmin       bool
}

type AgentApplicationStatus string

const (
	AgentApplicationStatusPending   AgentApplicationStatus = "pending"
	AgentApplicationStatusActive    AgentApplicationStatus = "active"
	AgentApplicationStatusDeclined  AgentApplicationStatus = "declined"
	AgentApplicationStatusSuspended AgentApplicationStatus = "suspended"
)

type AgentApplication struct {
	ID                   ID
	ApplicantUserID      ID
	ApplicantEmail       string
	ApplicantDisplayName string
	CampusID             ID
	Name                 string
	PhoneNumber          string
	Status               AgentApplicationStatus
	ReviewerUserID       *ID
	AgentID              *ID
	OperatorNote         string
	CreatedAt            time.Time
	UpdatedAt            time.Time
	DecidedAt            *time.Time
}
