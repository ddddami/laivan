package auth

import (
	"bytes"
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ddddami/laivan/internal/domain"
)

func TestServiceCompletesProviderSignInAndManagesSession(t *testing.T) {
	store := &fakeStore{user: domain.User{
		ID:          domain.ID("550e8400-e29b-41d4-a716-446655440001"),
		Email:       "person@example.com",
		DisplayName: "Person Example",
		Status:      domain.UserStatusActive,
	}}
	provider := &fakeProvider{name: "test-provider", identity: IdentityProfile{
		Subject:       "google-subject",
		Email:         "Person@Example.com",
		DisplayName:   "Person Example",
		EmailVerified: true,
	}}
	service := newTestService(store, provider)

	_, attemptCookie, err := service.Begin("/properties/property-1/unit-types/unit-1")
	if err != nil {
		t.Fatalf("Begin returned error: %v", err)
	}
	state := provider.state
	attempt, err := service.verifyAttempt(attemptCookie, state)
	if err != nil {
		t.Fatalf("verify test attempt: %v", err)
	}
	provider.identity.Nonce = attempt.Nonce

	result, err := service.Complete(context.Background(), "authorization-code", state, attemptCookie)
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}
	if result.User.Email != "person@example.com" {
		t.Fatalf("user email = %q, want normalized email", result.User.Email)
	}
	if result.SessionToken == "" || result.CSRFToken == "" {
		t.Fatal("Complete did not return browser tokens")
	}
	if result.ReturnTo != "/properties/property-1/unit-types/unit-1" {
		t.Fatalf("return path = %q, want original path", result.ReturnTo)
	}
	if store.createdUserEmail != "person@example.com" || store.createdIdentity != (domain.ExternalIdentity{Provider: "test-provider", Subject: "google-subject"}) {
		t.Fatalf("stored identity = %q/%#v, want normalized email/provider identity", store.createdUserEmail, store.createdIdentity)
	}

	session, err := service.GetSession(context.Background(), result.SessionToken, "")
	if err != nil {
		t.Fatalf("GetSession returned error: %v", err)
	}
	if session.CSRFToken == "" || session.CSRFToken == result.CSRFToken {
		t.Fatal("GetSession did not rotate a missing CSRF cookie")
	}

	if err := service.Logout(context.Background(), result.SessionToken, session.CSRFToken); err != nil {
		t.Fatalf("Logout returned error: %v", err)
	}
	if !store.revoked {
		t.Fatal("Logout did not revoke the session")
	}
}

func TestServiceRejectsMismatchedCallbackState(t *testing.T) {
	service := newTestService(&fakeStore{}, &fakeProvider{})
	_, attemptCookie, err := service.Begin("/")
	if err != nil {
		t.Fatalf("Begin returned error: %v", err)
	}

	_, err = service.Complete(context.Background(), "authorization-code", "wrong-state", attemptCookie)
	if !errors.Is(err, ErrInvalidAttempt) {
		t.Fatalf("Complete error = %v, want ErrInvalidAttempt", err)
	}
}

func TestServiceRejectsInvalidCSRF(t *testing.T) {
	store := &fakeStore{session: domain.Session{
		TokenHash:     hashToken("session-token"),
		CSRFTokenHash: hashToken("csrf-token"),
		User:          domain.User{Status: domain.UserStatusActive},
	}}
	service := newTestService(store, &fakeProvider{})

	err := service.Logout(context.Background(), "session-token", "wrong-csrf")
	if !errors.Is(err, ErrInvalidCSRF) {
		t.Fatalf("Logout error = %v, want ErrInvalidCSRF", err)
	}
}

type fakeProvider struct {
	name     string
	state    string
	identity IdentityProfile
	err      error
}

func (f *fakeProvider) Name() domain.IdentityProvider {
	if f.name == "" {
		return "test-provider"
	}
	return domain.IdentityProvider(f.name)
}

func (f *fakeProvider) AuthorizationURL(state, _, _ string) string {
	f.state = state
	return "https://accounts.example.test/auth?state=" + url.QueryEscape(state)
}

func (f *fakeProvider) Authenticate(context.Context, string, string) (IdentityProfile, error) {
	return f.identity, f.err
}

type fakeStore struct {
	user             domain.User
	createdUserEmail string
	createdIdentity  domain.ExternalIdentity
	session          domain.Session
	revoked          bool
}

func (f *fakeStore) UpsertUser(_ context.Context, identity domain.ExternalIdentity, email, _ string) (domain.User, error) {
	f.createdUserEmail = email
	f.createdIdentity = identity
	return f.user, nil
}

func (f *fakeStore) CreateSession(_ context.Context, _ domain.ID, tokenHash, csrfTokenHash []byte, expiresAt time.Time) (domain.Session, error) {
	f.session = domain.Session{TokenHash: tokenHash, CSRFTokenHash: csrfTokenHash, ExpiresAt: expiresAt, User: f.user}
	return f.session, nil
}

func (f *fakeStore) GetSession(_ context.Context, tokenHash []byte) (domain.Session, bool, error) {
	if f.revoked || !bytes.Equal(tokenHash, f.session.TokenHash) {
		return domain.Session{}, false, nil
	}

	return f.session, true, nil
}

func (f *fakeStore) UpdateSessionCSRFToken(_ context.Context, tokenHash, csrfTokenHash []byte) error {
	if !bytes.Equal(tokenHash, f.session.TokenHash) {
		return errors.New("session not found")
	}
	f.session.CSRFTokenHash = csrfTokenHash
	return nil
}

func (f *fakeStore) RevokeSession(_ context.Context, tokenHash []byte) error {
	if !bytes.Equal(tokenHash, f.session.TokenHash) {
		return errors.New("session not found")
	}
	f.revoked = true
	return nil
}

func newTestService(store Store, provider Provider) *Service {
	service := NewService(Config{
		StateSigningKey: strings.Repeat("a", 32),
		StateDuration:   10 * time.Minute,
		SessionDuration: time.Hour,
	}, store, provider)

	service.now = func() time.Time { return time.Date(2026, time.August, 24, 12, 0, 0, 0, time.UTC) }
	counter := byte(0)
	service.rand = func(value []byte) error {
		counter++
		for i := range value {
			value[i] = counter
		}
		return nil
	}

	return service
}
