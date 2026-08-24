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
