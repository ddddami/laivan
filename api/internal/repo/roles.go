package repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ddddami/laivan/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type RoleRepository struct {
	db generatedDB
}

type generatedDB interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type RoleChange struct {
	TargetUserID domain.ID
	Changed      bool
}

func NewRoleRepository(database generatedDB) *RoleRepository {
	return &RoleRepository{db: database}
}

func (r *RoleRepository) Grant(ctx context.Context, email, role string, campusID, actorID domain.ID) (RoleChange, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	role = strings.TrimSpace(role)
	if role != "global_admin" && role != "campus_operator" {
		return RoleChange{}, ErrInvalidRole
	}
	if role == "campus_operator" && campusID == "" {
		return RoleChange{}, fmt.Errorf("campus is required for %s", role)
	}

	tx, err := r.begin(ctx, "grant role")
	if err != nil {
		return RoleChange{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	targetID, err := findUserID(ctx, tx, email)
	if err != nil {
		return RoleChange{}, err
	}
	actorUUID, err := optionalRoleActorID(actorID)
	if err != nil {
		return RoleChange{}, err
	}
	if !actorUUID.Valid {
		if role != "global_admin" || !firstGlobalAdmin(ctx, tx) {
			return RoleChange{}, ErrRoleActorRequired
		}
	} else if !isGlobalAdmin(ctx, tx, actorUUID) {
		return RoleChange{}, ErrRoleForbidden
	}

	var commandTag pgconn.CommandTag
	switch role {
	case "global_admin":
		commandTag, err = tx.Exec(ctx, `INSERT INTO global_admin_roles (user_id) VALUES ($1) ON CONFLICT DO NOTHING`, targetID)
	case "campus_operator":
		campusUUID, parseErr := uuidParam(campusID)
		if parseErr != nil {
			return RoleChange{}, parseErr
		}
		commandTag, err = tx.Exec(ctx, `INSERT INTO campus_operators (user_id, campus_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, targetID, campusUUID)
	}
	if err != nil {
		return RoleChange{}, fmt.Errorf("grant %s: %w", role, err)
	}

	changed := commandTag.RowsAffected() == 1
	if changed && actorUUID.Valid {
		if err := writeRoleAudit(ctx, tx, actorUUID, "role_granted", targetID, role, campusID); err != nil {
			return RoleChange{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return RoleChange{}, fmt.Errorf("commit role grant: %w", err)
	}
	return RoleChange{TargetUserID: domain.ID(uuidString(targetID)), Changed: changed}, nil
}

func (r *RoleRepository) Revoke(ctx context.Context, email, role string, campusID, actorID domain.ID) (RoleChange, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	role = strings.TrimSpace(role)
	if role != "global_admin" && role != "campus_operator" {
		return RoleChange{}, ErrInvalidRole
	}
	if role == "campus_operator" && campusID == "" {
		return RoleChange{}, fmt.Errorf("campus is required for %s", role)
	}
	actorUUID, err := uuidParam(actorID)
	if err != nil {
		return RoleChange{}, ErrRoleActorRequired
	}

	tx, err := r.begin(ctx, "revoke role")
	if err != nil {
		return RoleChange{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if !isGlobalAdmin(ctx, tx, actorUUID) {
		return RoleChange{}, ErrRoleForbidden
	}
	targetID, err := findUserID(ctx, tx, email)
	if err != nil {
		return RoleChange{}, err
	}

	var commandTag pgconn.CommandTag
	switch role {
	case "global_admin":
		commandTag, err = tx.Exec(ctx, `DELETE FROM global_admin_roles WHERE user_id = $1`, targetID)
	case "campus_operator":
		campusUUID, parseErr := uuidParam(campusID)
		if parseErr != nil {
			return RoleChange{}, parseErr
		}
		commandTag, err = tx.Exec(ctx, `DELETE FROM campus_operators WHERE user_id = $1 AND campus_id = $2`, targetID, campusUUID)
	}
	if err != nil {
		return RoleChange{}, fmt.Errorf("revoke %s: %w", role, err)
	}
	changed := commandTag.RowsAffected() == 1
	if changed {
		if err := writeRoleAudit(ctx, tx, actorUUID, "role_revoked", targetID, role, campusID); err != nil {
			return RoleChange{}, err
		}
		if err := revokeRoleSessions(ctx, tx, actorUUID, targetID); err != nil {
			return RoleChange{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return RoleChange{}, fmt.Errorf("commit role revoke: %w", err)
	}
	return RoleChange{TargetUserID: domain.ID(uuidString(targetID)), Changed: changed}, nil
}

func (r *RoleRepository) begin(ctx context.Context, operation string) (pgx.Tx, error) {
	beginner, ok := r.db.(interface {
		Begin(context.Context) (pgx.Tx, error)
	})
	if !ok {
		return nil, fmt.Errorf("%s database does not support transactions", operation)
	}
	tx, err := beginner.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin %s transaction: %w", operation, err)
	}
	return tx, nil
}

func findUserID(ctx context.Context, tx pgx.Tx, email string) (pgtype.UUID, error) {
	var id pgtype.UUID
	err := tx.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, strings.ToLower(strings.TrimSpace(email))).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return pgtype.UUID{}, ErrNotFound
	}
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("find role target user: %w", err)
	}
	return id, nil
}

func optionalRoleActorID(id domain.ID) (pgtype.UUID, error) {
	if id == "" {
		return pgtype.UUID{}, nil
	}
	return uuidParam(id)
}

func firstGlobalAdmin(ctx context.Context, tx pgx.Tx) bool {
	var count int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM global_admin_roles`).Scan(&count); err != nil {
		return false
	}
	return count == 0
}

func isGlobalAdmin(ctx context.Context, tx pgx.Tx, userID pgtype.UUID) bool {
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM global_admin_roles WHERE user_id = $1)`, userID).Scan(&exists); err != nil {
		return false
	}
	return exists
}

func writeRoleAudit(ctx context.Context, tx pgx.Tx, actorID pgtype.UUID, action string, targetID pgtype.UUID, role string, campusID domain.ID) error {
	metadata, err := json.Marshal(map[string]string{"role": role, "campus_id": string(campusID)})
	if err != nil {
		return fmt.Errorf("encode role audit metadata: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO audit_events (actor_user_id, action, resource_type, resource_id, metadata) VALUES ($1, $2, 'user_role', $3, $4)`, actorID, action, targetID, metadata); err != nil {
		return fmt.Errorf("write role audit event: %w", err)
	}
	return nil
}

func revokeRoleSessions(ctx context.Context, tx pgx.Tx, actorID, targetID pgtype.UUID) error {
	rows, err := tx.Query(ctx, `UPDATE sessions SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL RETURNING id`, targetID)
	if err != nil {
		return fmt.Errorf("revoke role sessions: %w", err)
	}
	var sessionIDs []pgtype.UUID
	for rows.Next() {
		var sessionID pgtype.UUID
		if err := rows.Scan(&sessionID); err != nil {
			rows.Close()
			return fmt.Errorf("scan revoked role session: %w", err)
		}
		sessionIDs = append(sessionIDs, sessionID)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	for _, sessionID := range sessionIDs {
		if _, err := tx.Exec(ctx, `INSERT INTO audit_events (actor_user_id, action, resource_type, resource_id, metadata) VALUES ($1, 'session_revoked_by_role_change', 'session', $2, '{"reason":"role_revoked"}'::jsonb)`, actorID, sessionID); err != nil {
			return fmt.Errorf("write role session audit event: %w", err)
		}
	}
	return nil
}
