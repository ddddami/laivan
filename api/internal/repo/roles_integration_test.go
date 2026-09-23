//go:build integration

package repo

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ddddami/laivan/internal/domain"
	"github.com/jackc/pgx/v5"
)

func TestRoleRepositoryGrantRevokeLifecycle(t *testing.T) {
	ctx := context.Background()
	db := openIntegrationDB(t, ctx)
	t.Cleanup(db.Close)
	if _, err := db.Exec(ctx, `TRUNCATE sessions, audit_events, campus_operators, global_admin_roles, agents, users CASCADE`); err != nil {
		t.Fatalf("truncate role tables: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db.Exec(context.Background(), `TRUNCATE sessions, audit_events, campus_operators, global_admin_roles, agents, users CASCADE`); err != nil {
			t.Errorf("cleanup role tables: %v", err)
		}
	})

	campusID := testCampusID(t, ctx, db)
	targetID := insertRoleUser(t, ctx, db, "role.target@example.com", "Role Target")
	actorID := insertRoleUser(t, ctx, db, "role.actor@example.com", "Role Actor")
	roles := NewRoleRepository(db)

	change, err := roles.Grant(ctx, "role.target@example.com", "global_admin", "", "")
	if err != nil {
		t.Fatalf("bootstrap global admin: %v", err)
	}
	if !change.Changed || change.TargetUserID != targetID {
		t.Fatalf("bootstrap change = %#v, want changed target", change)
	}
	if _, err := roles.Grant(ctx, "role.actor@example.com", "global_admin", "", actorID); !errors.Is(err, ErrRoleForbidden) {
		t.Fatalf("non-admin grant error = %v, want %v", err, ErrRoleForbidden)
	}
	if _, err := roles.Grant(ctx, "role.actor@example.com", "global_admin", "", targetID); err != nil {
		t.Fatalf("grant actor global admin: %v", err)
	}

	if _, err := db.Exec(ctx, `INSERT INTO sessions (user_id, token_hash, csrf_token_hash, expires_at) VALUES ($1, $2, $3, $4)`, actorID, []byte("01234567890123456789012345678901"), []byte("12345678901234567890123456789012"), time.Now().UTC().Add(time.Hour)); err != nil {
		t.Fatalf("insert target session: %v", err)
	}
	change, err = roles.Grant(ctx, "role.actor@example.com", "campus_operator", campusID, targetID)
	if err != nil {
		t.Fatalf("grant campus operator: %v", err)
	}
	if !change.Changed {
		t.Fatal("campus operator grant was not recorded")
	}

	change, err = roles.Revoke(ctx, "role.actor@example.com", "campus_operator", campusID, targetID)
	if err != nil {
		t.Fatalf("revoke campus operator: %v", err)
	}
	if !change.Changed {
		t.Fatal("campus operator revoke was not recorded")
	}

	var revoked bool
	if err := db.QueryRow(ctx, `SELECT revoked_at IS NOT NULL FROM sessions WHERE user_id = $1`, actorID).Scan(&revoked); err != nil {
		t.Fatalf("read revoked role session: %v", err)
	}
	if !revoked {
		t.Fatal("role revocation did not revoke the affected session")
	}

	var roleAuditCount, sessionAuditCount int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM audit_events WHERE action = 'role_revoked' AND resource_id = $1`, string(actorID)).Scan(&roleAuditCount); err != nil {
		t.Fatalf("count role audit events: %v", err)
	}
	if err := db.QueryRow(ctx, `SELECT count(*) FROM audit_events WHERE action = 'session_revoked_by_role_change'`).Scan(&sessionAuditCount); err != nil {
		t.Fatalf("count role session audit events: %v", err)
	}
	if roleAuditCount != 1 || sessionAuditCount != 1 {
		t.Fatalf("role audits = %d and session audits = %d, want 1 and 1", roleAuditCount, sessionAuditCount)
	}
}

func TestRoleRepositoryConcurrentBootstrapGrantsOnlyOneAdmin(t *testing.T) {
	ctx := t.Context()
	db := openIntegrationDB(t, ctx)
	t.Cleanup(db.Close)
	if _, err := db.Exec(ctx, `TRUNCATE sessions, audit_events, campus_operators, global_admin_roles, agents, users CASCADE`); err != nil {
		t.Fatalf("truncate role tables: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db.Exec(context.Background(), `TRUNCATE sessions, audit_events, campus_operators, global_admin_roles, agents, users CASCADE`); err != nil {
			t.Errorf("cleanup role tables: %v", err)
		}
	})

	insertRoleUser(t, ctx, db, "first.bootstrap@example.com", "First Bootstrap")
	insertRoleUser(t, ctx, db, "second.bootstrap@example.com", "Second Bootstrap")
	repository := NewRoleRepository(db)
	start := make(chan struct{})
	results := make(chan error, 2)
	var group sync.WaitGroup
	for _, email := range []string{"first.bootstrap@example.com", "second.bootstrap@example.com"} {
		group.Go(func() {
			<-start
			_, err := repository.Grant(ctx, email, "global_admin", "", "")
			results <- err
		})
	}
	close(start)
	group.Wait()
	close(results)

	var granted, actorRequired int
	for err := range results {
		switch {
		case err == nil:
			granted++
		case errors.Is(err, ErrRoleActorRequired):
			actorRequired++
		default:
			t.Fatalf("unexpected bootstrap result: %v", err)
		}
	}
	if granted != 1 || actorRequired != 1 {
		t.Fatalf("bootstrap outcomes = %d granted, %d denied; want one of each", granted, actorRequired)
	}
	var adminCount int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM global_admin_roles`).Scan(&adminCount); err != nil {
		t.Fatalf("count global admins: %v", err)
	}
	if adminCount != 1 {
		t.Fatalf("global admin count = %d, want 1", adminCount)
	}
}

func insertRoleUser(t *testing.T, ctx context.Context, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, email, displayName string) domain.ID {
	t.Helper()
	var id string
	if err := db.QueryRow(ctx, `INSERT INTO users (email, display_name) VALUES ($1, $2) RETURNING id::text`, email, displayName).Scan(&id); err != nil {
		t.Fatalf("insert role user: %v", err)
	}
	return domain.ID(id)
}
