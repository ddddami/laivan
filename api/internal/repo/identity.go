package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	generateddb "github.com/ddddami/laivan/internal/db/generated"
	"github.com/ddddami/laivan/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type IdentityRepository struct {
	queries *generateddb.Queries
	db      generateddb.DBTX
}

func NewIdentityRepository(database generateddb.DBTX) *IdentityRepository {
	return &IdentityRepository{queries: generateddb.New(database), db: database}
}

func (r *IdentityRepository) UpsertUser(ctx context.Context, identity domain.ExternalIdentity, email, displayName string) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	beginner, ok := r.db.(interface {
		Begin(context.Context) (pgx.Tx, error)
	})
	if !ok {
		return domain.User{}, errors.New("identity repository database does not support transactions")
	}

	tx, err := beginner.Begin(ctx)
	if err != nil {
		return domain.User{}, fmt.Errorf("begin upsert user transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	queries := r.queries.WithTx(tx)
	row, err := queries.GetUserByIdentity(ctx, generateddb.GetUserByIdentityParams{
		Provider: string(identity.Provider),
		Subject:  identity.Subject,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		row, err = queries.CreateUser(ctx, generateddb.CreateUserParams{
			Email:       email,
			DisplayName: displayName,
		})
		if err != nil {
			return domain.User{}, fmt.Errorf("create user: %w", err)
		}
		if err := queries.CreateUserIdentity(ctx, generateddb.CreateUserIdentityParams{
			UserID:   row.ID,
			Provider: string(identity.Provider),
			Subject:  identity.Subject,
		}); err != nil {
			return domain.User{}, fmt.Errorf("create user identity: %w", err)
		}
	} else if err != nil {
		return domain.User{}, fmt.Errorf("get user by identity: %w", err)
	} else {
		row, err = queries.UpdateUserProfile(ctx, generateddb.UpdateUserProfileParams{
			ID:          row.ID,
			Email:       email,
			DisplayName: displayName,
		})
		if err != nil {
			return domain.User{}, fmt.Errorf("update user profile: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.User{}, fmt.Errorf("commit upsert user transaction: %w", err)
	}

	return userFromRow(row), nil
}

func (r *IdentityRepository) CreateSession(ctx context.Context, userID domain.ID, tokenHash, csrfTokenHash []byte, expiresAt time.Time) (domain.Session, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	userUUID, err := uuidParam(userID)
	if err != nil {
		return domain.Session{}, err
	}

	beginner, ok := r.db.(interface {
		Begin(context.Context) (pgx.Tx, error)
	})
	if !ok {
		return domain.Session{}, errors.New("identity repository database does not support transactions")
	}

	tx, err := beginner.Begin(ctx)
	if err != nil {
		return domain.Session{}, fmt.Errorf("begin session transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)

	row, err := queries.CreateSession(ctx, generateddb.CreateSessionParams{
		UserID:        userUUID,
		TokenHash:     tokenHash,
		CsrfTokenHash: csrfTokenHash,
		ExpiresAt:     timestamptzParam(expiresAt),
	})
	if err != nil {
		return domain.Session{}, fmt.Errorf("create session: %w", err)
	}
	if err := queries.CreateAuditEvent(ctx, generateddb.CreateAuditEventParams{
		ActorUserID:  userUUID,
		Action:       "user_signed_in",
		ResourceType: "session",
		ResourceID:   row.ID,
		Metadata:     []byte(`{}`),
	}); err != nil {
		return domain.Session{}, fmt.Errorf("write sign-in audit event: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Session{}, fmt.Errorf("commit session transaction: %w", err)
	}

	return sessionFromRow(row), nil
}

func (r *IdentityRepository) GetSession(ctx context.Context, tokenHash []byte) (domain.Session, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	row, err := r.queries.GetActiveSession(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Session{}, false, nil
		}

		return domain.Session{}, false, fmt.Errorf("get active session: %w", err)
	}

	return sessionFromActiveRow(row), true, nil
}

func (r *IdentityRepository) UpdateSessionCSRFToken(ctx context.Context, tokenHash, csrfTokenHash []byte) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	if err := r.queries.UpdateSessionCSRFToken(ctx, generateddb.UpdateSessionCSRFTokenParams{
		TokenHash:     tokenHash,
		CsrfTokenHash: csrfTokenHash,
	}); err != nil {
		return fmt.Errorf("update session csrf token: %w", err)
	}

	return nil
}

func (r *IdentityRepository) RevokeSession(ctx context.Context, tokenHash []byte) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	beginner, ok := r.db.(interface {
		Begin(context.Context) (pgx.Tx, error)
	})
	if !ok {
		return errors.New("identity repository database does not support transactions")
	}
	tx, err := beginner.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin revoke session transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)

	row, err := queries.RevokeSession(ctx, tokenHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	if err := queries.CreateAuditEvent(ctx, generateddb.CreateAuditEventParams{
		ActorUserID:  row.UserID,
		Action:       "session_revoked",
		ResourceType: "session",
		ResourceID:   row.ID,
		Metadata:     []byte(`{}`),
	}); err != nil {
		return fmt.Errorf("write logout audit event: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit revoke session transaction: %w", err)
	}

	return nil
}

func userFromRow(row generateddb.User) domain.User {
	return domain.User{
		ID:          domain.ID(uuidString(row.ID)),
		Email:       row.Email,
		DisplayName: row.DisplayName,
		Status:      domain.UserStatus(row.Status),
		Timestamps: domain.Timestamps{
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		},
	}
}

func sessionFromRow(row generateddb.Session) domain.Session {
	return domain.Session{
		ID:            domain.ID(uuidString(row.ID)),
		TokenHash:     row.TokenHash,
		CSRFTokenHash: row.CsrfTokenHash,
		CreatedAt:     row.CreatedAt.Time,
		ExpiresAt:     row.ExpiresAt.Time,
		RevokedAt:     timestamptzPointer(row.RevokedAt),
		LastUsedAt:    timestamptzPointer(row.LastUsedAt),
	}
}

func sessionFromActiveRow(row generateddb.GetActiveSessionRow) domain.Session {
	return domain.Session{
		ID: domain.ID(uuidString(row.SessionID)),
		User: domain.User{
			ID:          domain.ID(uuidString(row.UserID)),
			Email:       row.Email,
			DisplayName: row.DisplayName,
			Status:      domain.UserStatus(row.Status),
			Timestamps: domain.Timestamps{
				CreatedAt: row.UserCreatedAt.Time,
				UpdatedAt: row.UserUpdatedAt.Time,
			},
		},
		TokenHash:     row.TokenHash,
		CSRFTokenHash: row.CsrfTokenHash,
		CreatedAt:     row.SessionCreatedAt.Time,
		ExpiresAt:     row.ExpiresAt.Time,
		RevokedAt:     timestamptzPointer(row.RevokedAt),
		LastUsedAt:    timestamptzPointer(row.LastUsedAt),
	}
}

func timestamptzParam(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: true}
}

func timestamptzPointer(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}

	result := value.Time
	return &result
}
