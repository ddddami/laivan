//go:build integration

package repo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestAgentUserForeignKeyRestrictsUserDeletion(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	db := openIntegrationDB(t, ctx)
	t.Cleanup(db.Close)
	resetAgentApplicationTables(t, ctx, db)
	t.Cleanup(func() { resetAgentApplicationTables(t, context.Background(), db) })

	userID := insertIdentityUser(t, ctx, db, "restrict-agent@example.com", "Restrict Agent")
	var agentID string
	if err := db.QueryRow(ctx, `INSERT INTO agents (user_id, display_name, phone_number) VALUES ($1, 'Restrict Agent', '+2348012345678') RETURNING id::text`, string(userID)).Scan(&agentID); err != nil {
		t.Fatalf("insert linked agent: %v", err)
	}

	if _, err := db.Exec(ctx, `DELETE FROM users WHERE id = $1`, string(userID)); !isForeignKeyViolation(err) {
		t.Fatalf("delete linked agent user error = %v, want foreign key violation", err)
	}
	var linkedUserID string
	if err := db.QueryRow(ctx, `SELECT user_id::text FROM agents WHERE id = $1`, agentID).Scan(&linkedUserID); err != nil {
		t.Fatalf("load linked agent after rejected user deletion: %v", err)
	}
	if linkedUserID != string(userID) {
		t.Fatalf("linked agent user ID = %s, want %s", linkedUserID, userID)
	}
}

func TestStrictAgentIdentityMigrationRemovesUnlinkedAgentDependencies(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	db := openIntegrationDB(t, ctx)
	t.Cleanup(db.Close)
	conn, err := db.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire migration connection: %v", err)
	}
	t.Cleanup(conn.Release)

	schemaName := "strict_identity_" + migrationTestSchemaSuffix(t)
	schemaIdentifier := quoteIdentifier(schemaName)
	if _, err := conn.Exec(ctx, "CREATE SCHEMA "+schemaIdentifier); err != nil {
		t.Fatalf("create migration test schema: %v", err)
	}
	t.Cleanup(func() {
		if _, err := conn.Exec(context.Background(), "DROP SCHEMA "+schemaIdentifier+" CASCADE"); err != nil {
			t.Errorf("drop migration test schema: %v", err)
		}
	})
	if _, err := conn.Exec(ctx, "SET search_path TO "+schemaIdentifier+", public"); err != nil {
		t.Fatalf("set migration test search path: %v", err)
	}

	for version := 1; version <= 15; version++ {
		migrationPath := filepath.Join(migrationDirectory(t), fmt.Sprintf("%06d_*.sql", version))
		matches, err := filepath.Glob(migrationPath)
		if err != nil || len(matches) != 1 {
			t.Fatalf("find migration %06d: matches=%v error=%v", version, matches, err)
		}
		sql, err := os.ReadFile(matches[0])
		if err != nil {
			t.Fatalf("read migration %06d: %v", version, err)
		}
		if _, err := conn.Exec(ctx, migrationUpSQL(sql)); err != nil {
			t.Fatalf("apply migration %06d: %v", version, err)
		}
	}

	const (
		agentID     = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
		propertyID  = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
		unitTypeID  = "cccccccc-cccc-4ccc-8ccc-cccccccccccc"
		offerID     = "dddddddd-dddd-4ddd-8ddd-dddddddddddd"
		application = "eeeeeeee-eeee-4eee-8eee-eeeeeeeeeeee"
		mediaID     = "ffffffff-ffff-4fff-8fff-ffffffffffff"
	)
	var campusID, applicantID string
	if err := conn.QueryRow(ctx, `SELECT id::text FROM campuses WHERE slug = 'futa'`).Scan(&campusID); err != nil {
		t.Fatalf("load migrated FUTA campus: %v", err)
	}
	if err := conn.QueryRow(ctx, `INSERT INTO users (email, display_name) VALUES ('upgrade-applicant@example.com', 'Upgrade Applicant') RETURNING id::text`).Scan(&applicantID); err != nil {
		t.Fatalf("insert upgrade applicant: %v", err)
	}
	statements := []struct {
		name string
		sql  string
		args []any
	}{
		{"agent", `INSERT INTO agents (id, display_name, phone_number) VALUES ($1, 'Legacy Agent', '+2348012345678')`, []any{agentID}},
		{"property", `INSERT INTO properties (id, campus_id, name, area) VALUES ($1, $2, 'Upgrade Property', 'Obanla')`, []any{propertyID, campusID}},
		{"unit type", `INSERT INTO property_unit_types (id, property_id, category, name) VALUES ($1, $2, 'self_contained', 'Upgrade Unit')`, []any{unitTypeID, propertyID}},
		{"offer", `INSERT INTO agent_offers (id, property_unit_type_id, agent_id, title, price_kobo) VALUES ($1, $2, $3, 'Upgrade Offer', 100000)`, []any{offerID, unitTypeID, agentID}},
		{"agent campus", `INSERT INTO agent_campuses (agent_id, campus_id) VALUES ($1, $2)`, []any{agentID, campusID}},
		{"application", `INSERT INTO agent_applications (id, applicant_user_id, campus_id, name, phone_number, status, agent_id) VALUES ($1, $2, $3, 'Legacy Application', '+2348098765432', 'active', $4)`, []any{application, applicantID, campusID, agentID}},
		{"media", `INSERT INTO media (id, property_id, uploaded_by_agent_id, url, kind) VALUES ($1, $2, $3, 'https://example.com/upgrade.jpg', 'image')`, []any{mediaID, propertyID, agentID}},
	}
	for _, statement := range statements {
		if _, err := conn.Exec(ctx, statement.sql, statement.args...); err != nil {
			t.Fatalf("insert %s fixture: %v", statement.name, err)
		}
	}

	var agentCount, offerCount, campusAssociationCount, applicationCount int
	if err := conn.QueryRow(ctx, `SELECT count(*) FROM agents WHERE id = $1`, agentID).Scan(&agentCount); err != nil {
		t.Fatalf("count legacy agent before migration: %v", err)
	}
	if agentCount != 1 {
		t.Fatalf("legacy agent count before migration = %d, want 1", agentCount)
	}
	if err := conn.QueryRow(ctx, `SELECT count(*) FROM agent_applications WHERE id = $1`, application).Scan(&applicationCount); err != nil {
		t.Fatalf("count dependent application before migration: %v", err)
	}
	if applicationCount != 1 {
		t.Fatalf("dependent application count before migration = %d, want 1", applicationCount)
	}
	if err := conn.QueryRow(ctx, `SELECT count(*) FROM agent_offers WHERE id = $1`, offerID).Scan(&offerCount); err != nil {
		t.Fatalf("count legacy offer before migration: %v", err)
	}
	if err := conn.QueryRow(ctx, `SELECT count(*) FROM agent_campuses WHERE agent_id = $1`, agentID).Scan(&campusAssociationCount); err != nil {
		t.Fatalf("count legacy campus association before migration: %v", err)
	}
	if offerCount != 1 || campusAssociationCount != 1 {
		t.Fatalf("legacy dependencies before migration = offer %d, campus %d, want 1/1", offerCount, campusAssociationCount)
	}

	var migrationSQL []byte
	migrationSQL, err = os.ReadFile(filepath.Join(migrationDirectory(t), "000016_enforce_strict_agent_identity.sql"))
	if err != nil {
		t.Fatalf("read strict identity migration: %v", err)
	}
	if _, err := conn.Exec(ctx, migrationUpSQL(migrationSQL)); err != nil {
		t.Fatalf("apply strict identity migration: %v", err)
	}

	for name, query := range map[string]string{
		"agent":       `SELECT count(*) FROM agents WHERE id = $1`,
		"offer":       `SELECT count(*) FROM agent_offers WHERE id = $1`,
		"application": `SELECT count(*) FROM agent_applications WHERE id = $1`,
		"association": `SELECT count(*) FROM agent_campuses WHERE agent_id = $1`,
	} {
		var count int
		argument := agentID
		if name == "offer" {
			argument = offerID
		}
		if name == "application" {
			argument = application
		}
		if err := conn.QueryRow(ctx, query, argument).Scan(&count); err != nil {
			t.Fatalf("count %s after migration: %v", name, err)
		}
		if count != 0 {
			t.Fatalf("%s count after migration = %d, want 0", name, count)
		}
	}
	var mediaCount, mediaUploaderCount int
	if err := conn.QueryRow(ctx, `SELECT count(*), count(uploaded_by_agent_id) FROM media WHERE id = $1`, mediaID).Scan(&mediaCount, &mediaUploaderCount); err != nil {
		t.Fatalf("inspect retained media after migration: %v", err)
	}
	if mediaCount != 1 || mediaUploaderCount != 0 {
		t.Fatalf("retained media after migration = count %d, uploaded_by_agent count %d, want 1/0", mediaCount, mediaUploaderCount)
	}

	var nullable string
	if err := conn.QueryRow(ctx, `SELECT is_nullable FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = 'agents' AND column_name = 'user_id'`).Scan(&nullable); err != nil {
		t.Fatalf("inspect agents.user_id nullability: %v", err)
	}
	if nullable != "NO" {
		t.Fatalf("agents.user_id nullability = %q, want NO", nullable)
	}
}

func migrationDirectory(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve migration test source path")
	}
	return filepath.Join(filepath.Dir(filename), "../../migrations")
}

func migrationTestSchemaSuffix(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf("%d", os.Getpid())
}

func quoteIdentifier(identifier string) string {
	return `"` + identifier + `"`
}

func migrationUpSQL(migration []byte) string {
	_, up, _ := strings.Cut(string(migration), "-- +goose Up")
	up, _, _ = strings.Cut(up, "-- +goose Down")
	return up
}
