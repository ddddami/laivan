//go:build integration

package repo

import (
	"context"
	"testing"
	"time"

	"github.com/ddddami/laivan/internal/domain"
)

func TestIdentityRepositorySessionLifecycle(t *testing.T) {
	ctx := context.Background()
	db := openIntegrationDB(t, ctx)
	t.Cleanup(db.Close)

	if _, err := db.Exec(ctx, "TRUNCATE sessions, users CASCADE"); err != nil {
		t.Fatalf("truncate identity tables: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db.Exec(ctx, "TRUNCATE sessions, users CASCADE"); err != nil {
			t.Errorf("cleanup identity tables: %v", err)
		}
	})

	repository := NewIdentityRepository(db)
	user, err := repository.UpsertUser(ctx, domain.ExternalIdentity{
		Provider: domain.IdentityProviderGoogle,
		Subject:  "provider-subject",
	}, "person@example.com", "Person Example")
	if err != nil {
		t.Fatalf("upsert user: %v", err)
	}
	if user.ID == "" || user.Status != domain.UserStatusActive {
		t.Fatalf("user = %#v, want generated active user", user)
	}

	updated, err := repository.UpsertUser(ctx, domain.ExternalIdentity{
		Provider: domain.IdentityProviderGoogle,
		Subject:  "provider-subject",
	}, "person@example.com", "Updated Person")
	if err != nil {
		t.Fatalf("update user through existing identity: %v", err)
	}
	if updated.ID != user.ID || updated.DisplayName != "Updated Person" {
		t.Fatalf("updated user = %#v, want same user with updated profile", updated)
	}

	if _, err := repository.UpsertUser(ctx, domain.ExternalIdentity{
		Provider: "other-provider",
		Subject:  "other-subject",
	}, "person@example.com", "Other Provider Person"); err == nil {
		t.Fatal("upserted a second provider identity into an existing email without explicit linking")
	}

	otherUser, err := repository.UpsertUser(ctx, domain.ExternalIdentity{
		Provider: "other-provider",
		Subject:  "other-subject",
	}, "other@example.com", "Other Provider Person")
	if err != nil {
		t.Fatalf("create user for second provider: %v", err)
	}
	if otherUser.ID == user.ID {
		t.Fatal("second provider identity unexpectedly shared the first user")
	}

	tokenHash := []byte("12345678901234567890123456789012")
	csrfHash := []byte("abcdefghijklmnopqrstuvwxyz123456")
	expiresAt := time.Now().UTC().Truncate(time.Microsecond).Add(time.Hour)
	created, err := repository.CreateSession(ctx, user.ID, tokenHash, csrfHash, expiresAt)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if created.ID == "" || !created.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("created session = %#v, want generated session with expiry", created)
	}

	fetched, found, err := repository.GetSession(ctx, tokenHash)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if !found || fetched.User.Email != user.Email || string(fetched.CSRFTokenHash) != string(csrfHash) {
		t.Fatalf("fetched session = %#v, want active session for user", fetched)
	}

	newCSRFHash := []byte("98765432109876543210987654321098")
	if err := repository.UpdateSessionCSRFToken(ctx, tokenHash, newCSRFHash); err != nil {
		t.Fatalf("update csrf token: %v", err)
	}
	if err := repository.RevokeSession(ctx, tokenHash); err != nil {
		t.Fatalf("revoke session: %v", err)
	}

	_, found, err = repository.GetSession(ctx, tokenHash)
	if err != nil {
		t.Fatalf("get revoked session: %v", err)
	}
	if found {
		t.Fatal("revoked session was returned as active")
	}
}
