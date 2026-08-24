package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/ddddami/laivan/internal/domain"
)

var (
	ErrAuthUnavailable = errors.New("authentication is not configured")
	ErrInvalidAttempt  = errors.New("invalid authentication callback")
	ErrUnauthenticated = errors.New("authentication is required")
	ErrInvalidCSRF     = errors.New("invalid csrf token")
	ErrInvalidIdentity = errors.New("identity provider returned an invalid identity")
)

type Config struct {
	StateSigningKey string
	StateDuration   time.Duration
	SessionDuration time.Duration
}

type Store interface {
	UpsertUser(ctx context.Context, identity domain.ExternalIdentity, email, displayName string) (domain.User, error)
	CreateSession(ctx context.Context, userID domain.ID, tokenHash, csrfTokenHash []byte, expiresAt time.Time) (domain.Session, error)
	GetSession(ctx context.Context, tokenHash []byte) (domain.Session, bool, error)
	UpdateSessionCSRFToken(ctx context.Context, tokenHash, csrfTokenHash []byte) error
	RevokeSession(ctx context.Context, tokenHash []byte) error
}

type Provider interface {
	Name() domain.IdentityProvider
	AuthorizationURL(state, nonce, codeChallenge string) string
	Authenticate(ctx context.Context, code, codeVerifier string) (IdentityProfile, error)
}

type IdentityProfile struct {
	Subject       string
	Email         string
	DisplayName   string
	EmailVerified bool
	Nonce         string
}

type SessionResult struct {
	User         domain.User
	SessionToken string
	CSRFToken    string
}

type Service struct {
	cfg      Config
	store    Store
	provider Provider
	now      func() time.Time
	rand     func([]byte) error
}

func NewService(cfg Config, store Store, provider Provider) *Service {
	return &Service{
		cfg:      cfg,
		store:    store,
		provider: provider,
		now:      time.Now,
		rand:     func(value []byte) error { _, err := rand.Read(value); return err },
	}
}

func (s *Service) Begin() (redirectURL, cookieValue string, err error) {
	if s.provider == nil || s.cfg.StateSigningKey == "" {
		return "", "", ErrAuthUnavailable
	}

	state, err := s.randomString(32)
	if err != nil {
		return "", "", fmt.Errorf("generate oauth state: %w", err)
	}
	nonce, err := s.randomString(32)
	if err != nil {
		return "", "", fmt.Errorf("generate oauth nonce: %w", err)
	}
	verifier, err := s.randomString(32)
	if err != nil {
		return "", "", fmt.Errorf("generate pkce verifier: %w", err)
	}

	attempt := authAttempt{
		State:     state,
		Nonce:     nonce,
		Verifier:  verifier,
		ExpiresAt: s.now().Add(s.cfg.StateDuration).Unix(),
	}
	cookieValue, err = s.signAttempt(attempt)
	if err != nil {
		return "", "", fmt.Errorf("sign oauth attempt: %w", err)
	}

	challengeBytes := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(challengeBytes[:])
	redirectURL = s.provider.AuthorizationURL(state, nonce, challenge)

	return redirectURL, cookieValue, nil
}

func (s *Service) Complete(ctx context.Context, code, state, cookieValue string) (SessionResult, error) {
	attempt, err := s.verifyAttempt(cookieValue, state)
	if err != nil {
		return SessionResult{}, err
	}
	if code == "" || s.provider == nil {
		return SessionResult{}, ErrInvalidAttempt
	}

	identity, err := s.provider.Authenticate(ctx, code, attempt.Verifier)
	if err != nil {
		return SessionResult{}, fmt.Errorf("%w: authenticate with provider: %v", ErrInvalidAttempt, err)
	}
	if identity.Nonce != attempt.Nonce || !identity.EmailVerified {
		return SessionResult{}, ErrInvalidIdentity
	}

	email, err := normalizeEmail(identity.Email)
	if err != nil || identity.Subject == "" || s.provider.Name() == "" {
		return SessionResult{}, ErrInvalidIdentity
	}
	displayName := strings.TrimSpace(identity.DisplayName)
	if displayName == "" {
		displayName = email
	}

	user, err := s.store.UpsertUser(ctx, domain.ExternalIdentity{
		Provider: s.provider.Name(),
		Subject:  identity.Subject,
	}, email, displayName)
	if err != nil {
		return SessionResult{}, fmt.Errorf("upsert authenticated user: %w", err)
	}

	sessionToken, err := s.randomString(32)
	if err != nil {
		return SessionResult{}, fmt.Errorf("generate session token: %w", err)
	}
	csrfToken, err := s.randomString(32)
	if err != nil {
		return SessionResult{}, fmt.Errorf("generate csrf token: %w", err)
	}

	if _, err := s.store.CreateSession(ctx, user.ID, hashToken(sessionToken), hashToken(csrfToken), s.now().Add(s.cfg.SessionDuration)); err != nil {
		return SessionResult{}, fmt.Errorf("persist authenticated session: %w", err)
	}

	return SessionResult{User: user, SessionToken: sessionToken, CSRFToken: csrfToken}, nil
}

func (s *Service) GetSession(ctx context.Context, sessionToken, csrfCookie string) (SessionResult, error) {
	if sessionToken == "" {
		return SessionResult{}, ErrUnauthenticated
	}

	session, found, err := s.store.GetSession(ctx, hashToken(sessionToken))
	if err != nil {
		return SessionResult{}, fmt.Errorf("load session: %w", err)
	}
	if !found {
		return SessionResult{}, ErrUnauthenticated
	}

	csrfToken := csrfCookie
	if !validTokenHash(csrfCookie, session.CSRFTokenHash) {
		csrfToken, err = s.randomString(32)
		if err != nil {
			return SessionResult{}, fmt.Errorf("rotate csrf token: %w", err)
		}
		if err := s.store.UpdateSessionCSRFToken(ctx, session.TokenHash, hashToken(csrfToken)); err != nil {
			return SessionResult{}, fmt.Errorf("persist rotated csrf token: %w", err)
		}
	}

	return SessionResult{User: session.User, CSRFToken: csrfToken}, nil
}

func (s *Service) Logout(ctx context.Context, sessionToken, csrfToken string) error {
	if sessionToken == "" {
		return ErrUnauthenticated
	}

	session, found, err := s.store.GetSession(ctx, hashToken(sessionToken))
	if err != nil {
		return fmt.Errorf("load session for logout: %w", err)
	}
	if !found {
		return ErrUnauthenticated
	}
	if !validTokenHash(csrfToken, session.CSRFTokenHash) {
		return ErrInvalidCSRF
	}

	if err := s.store.RevokeSession(ctx, session.TokenHash); err != nil {
		return fmt.Errorf("revoke authenticated session: %w", err)
	}

	return nil
}

type authAttempt struct {
	State     string `json:"state"`
	Nonce     string `json:"nonce"`
	Verifier  string `json:"verifier"`
	ExpiresAt int64  `json:"expires_at"`
}

func (s *Service) signAttempt(attempt authAttempt) (string, error) {
	payload, err := json.Marshal(attempt)
	if err != nil {
		return "", err
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(s.cfg.StateSigningKey))
	_, _ = mac.Write([]byte(encodedPayload))
	encodedSignature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return encodedPayload + "." + encodedSignature, nil
}

func (s *Service) verifyAttempt(cookieValue, state string) (authAttempt, error) {
	parts := strings.Split(cookieValue, ".")
	if len(parts) != 2 || state == "" {
		return authAttempt{}, ErrInvalidAttempt
	}

	mac := hmac.New(sha256.New, []byte(s.cfg.StateSigningKey))
	_, _ = mac.Write([]byte(parts[0]))
	expectedSignature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(expectedSignature, mac.Sum(nil)) {
		return authAttempt{}, ErrInvalidAttempt
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return authAttempt{}, ErrInvalidAttempt
	}
	var attempt authAttempt
	if err := json.Unmarshal(payload, &attempt); err != nil || attempt.State != state || attempt.ExpiresAt <= s.now().Unix() {
		return authAttempt{}, ErrInvalidAttempt
	}

	return attempt, nil
}

func (s *Service) randomString(size int) (string, error) {
	value := make([]byte, size)
	if err := s.rand(value); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(value), nil
}

func hashToken(token string) []byte {
	hash := sha256.Sum256([]byte(token))
	return hash[:]
}

func validTokenHash(token string, expected []byte) bool {
	return token != "" && hmac.Equal(hashToken(token), expected)
}

func normalizeEmail(value string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	parsed, err := mail.ParseAddress(normalized)
	if err != nil || parsed.Address != normalized {
		return "", ErrInvalidIdentity
	}

	return normalized, nil
}
