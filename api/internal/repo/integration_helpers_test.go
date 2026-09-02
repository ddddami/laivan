//go:build integration

package repo

import (
	"context"
	"os"
	"testing"
	"time"

	appdb "github.com/ddddami/laivan/internal/db"
	"github.com/ddddami/laivan/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

func openIntegrationDB(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("LAIVAN_TEST_DB_URL")
	if databaseURL == "" {
		t.Fatal("LAIVAN_TEST_DB_URL is required for integration tests")
	}

	pool, err := appdb.Open(ctx, databaseURL, 25)
	if err != nil {
		t.Fatalf("open integration database: %v", err)
	}

	return pool
}

func truncateProperties(t *testing.T, ctx context.Context, db *pgxpool.Pool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if _, err := db.Exec(ctx, "TRUNCATE properties CASCADE"); err != nil {
		t.Fatalf("truncate properties: %v", err)
	}
}

func truncateAgents(t *testing.T, ctx context.Context, db *pgxpool.Pool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if _, err := db.Exec(ctx, "TRUNCATE agents CASCADE"); err != nil {
		t.Fatalf("truncate agents: %v", err)
	}
}

func testCampusID(t *testing.T, ctx context.Context, db *pgxpool.Pool) domain.ID {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var campusID string
	if err := db.QueryRow(ctx, "SELECT id::text FROM campuses WHERE slug = 'futa'").Scan(&campusID); err != nil {
		t.Fatalf("select FUTA campus: %v", err)
	}

	return domain.ID(campusID)
}

func createTestCampus(t *testing.T, ctx context.Context, db *pgxpool.Pool, slug string) domain.ID {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var campusID string
	err := db.QueryRow(ctx, `
		INSERT INTO campuses (slug, name, short_name)
		VALUES ($1, 'FUTA North', 'FUTA North')
		ON CONFLICT (slug) DO UPDATE SET slug = EXCLUDED.slug
		RETURNING id::text
	`, slug).Scan(&campusID)
	if err != nil {
		t.Fatalf("create test campus: %v", err)
	}

	return domain.ID(campusID)
}

func insertProperty(t *testing.T, ctx context.Context, db *pgxpool.Pool, campusID domain.ID, name string, createdAt time.Time) domain.ID {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var propertyID string
	err := db.QueryRow(ctx, `
		INSERT INTO properties (campus_id, name, area, landmark, description, created_at, updated_at)
		VALUES ($1, $2, 'Obanla', 'Near South Gate', 'Test property', $3, $3)
		RETURNING id::text
	`, string(campusID), name, createdAt).Scan(&propertyID)
	if err != nil {
		t.Fatalf("insert property: %v", err)
	}

	return domain.ID(propertyID)
}

func insertPropertyWithArea(t *testing.T, ctx context.Context, db *pgxpool.Pool, campusID domain.ID, name string, area string, createdAt time.Time) domain.ID {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var propertyID string
	err := db.QueryRow(ctx, `
		INSERT INTO properties (campus_id, name, area, landmark, description, created_at, updated_at)
		VALUES ($1, $2, $3, 'Near South Gate', 'Test property', $4, $4)
		RETURNING id::text
	`, string(campusID), name, area, createdAt).Scan(&propertyID)
	if err != nil {
		t.Fatalf("insert property with area: %v", err)
	}

	return domain.ID(propertyID)
}

func insertPropertyUnitType(t *testing.T, ctx context.Context, db *pgxpool.Pool, propertyID domain.ID, name string) domain.ID {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var unitTypeID string
	err := db.QueryRow(ctx, `
		INSERT INTO property_unit_types (property_id, name, description)
		VALUES ($1, $2, 'Private room with bathroom and kitchenette.')
		RETURNING id::text
	`, string(propertyID), name).Scan(&unitTypeID)
	if err != nil {
		t.Fatalf("insert property unit type: %v", err)
	}

	return domain.ID(unitTypeID)
}

func insertPropertyUnitTypeFull(t *testing.T, ctx context.Context, db *pgxpool.Pool, propertyID domain.ID, category, name string, bedroomCount int, hasParlour bool, bathroomType, kitchenType, notes string) domain.ID {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var unitTypeID string
	err := db.QueryRow(ctx, `
		INSERT INTO property_unit_types (property_id, category, name, description, notes, bedroom_count, has_parlour, bathroom_type, kitchen_type)
		VALUES ($1, $2, $3, 'Test unit type', $4, $5, $6, $7, $8)
		RETURNING id::text
	`, string(propertyID), category, name, notes, bedroomCount, hasParlour, bathroomType, kitchenType).Scan(&unitTypeID)
	if err != nil {
		t.Fatalf("insert property unit type: %v", err)
	}

	return domain.ID(unitTypeID)
}

func insertUser(t *testing.T, ctx context.Context, db *pgxpool.Pool, displayName string) domain.ID {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var userID string
	if err := db.QueryRow(ctx, `
		INSERT INTO users (email, display_name)
		VALUES (gen_random_uuid() || '@test.com', $1)
		RETURNING id::text
	`, displayName).Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	return domain.ID(userID)
}

func grantCampusOperator(t *testing.T, ctx context.Context, db *pgxpool.Pool, userID, campusID domain.ID) {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if _, err := db.Exec(ctx, `
		INSERT INTO campus_operators (user_id, campus_id)
		VALUES ($1, $2)
	`, string(userID), string(campusID)); err != nil {
		t.Fatalf("grant campus operator: %v", err)
	}
}

func grantGlobalAdmin(t *testing.T, ctx context.Context, db *pgxpool.Pool, userID domain.ID) {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if _, err := db.Exec(ctx, `
		INSERT INTO global_admin_roles (user_id)
		VALUES ($1)
	`, string(userID)); err != nil {
		t.Fatalf("grant global admin: %v", err)
	}
}

func insertAgent(t *testing.T, ctx context.Context, db *pgxpool.Pool, displayName string) domain.ID {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Email and phone must be unique per call. Email uses a random UUID suffix.
	// Phone derives a valid E.164 number from the user's ID hash so it stays unique
	// without needing a separate counter and satisfies the +234XXXXXXXXXX constraint.
	var userID string
	err := db.QueryRow(ctx, `
		INSERT INTO users (email, display_name)
		VALUES (gen_random_uuid() || '@test.com', $1)
		RETURNING id::text
	`, displayName).Scan(&userID)
	if err != nil {
		t.Fatalf("insert user for agent: %v", err)
	}

	var agentID string
	err = db.QueryRow(ctx, `
		INSERT INTO agents (user_id, display_name, phone_number)
		VALUES ($1::uuid, $2, '+234' || lpad((abs(hashtext($3)) % 10000000000)::text, 10, '0'))
		RETURNING id::text
	`, userID, displayName, userID).Scan(&agentID)
	if err != nil {
		t.Fatalf("insert agent: %v", err)
	}

	return domain.ID(agentID)
}

func agentUserID(t *testing.T, ctx context.Context, db *pgxpool.Pool, agentID domain.ID) domain.ID {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var userID string
	if err := db.QueryRow(ctx, "SELECT user_id::text FROM agents WHERE id = $1", string(agentID)).Scan(&userID); err != nil {
		t.Fatalf("get agent user: %v", err)
	}

	return domain.ID(userID)
}

func associateAgentWithCampus(t *testing.T, ctx context.Context, db *pgxpool.Pool, agentID, campusID domain.ID) {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if _, err := db.Exec(ctx, "INSERT INTO agent_campuses (agent_id, campus_id) VALUES ($1, $2)", string(agentID), string(campusID)); err != nil {
		t.Fatalf("associate agent with campus: %v", err)
	}
}

func insertAgentOffer(t *testing.T, ctx context.Context, db *pgxpool.Pool, unitTypeID domain.ID, agentID domain.ID, title string, priceKobo int) domain.ID {
	t.Helper()

	return insertAgentOfferWithStatus(t, ctx, db, unitTypeID, agentID, title, priceKobo, "available")
}

func insertAgentOfferWithStatus(t *testing.T, ctx context.Context, db *pgxpool.Pool, unitTypeID domain.ID, agentID domain.ID, title string, priceKobo int, status string) domain.ID {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var offerID string
	err := db.QueryRow(ctx, `
		INSERT INTO agent_offers (property_unit_type_id, agent_id, title, description, price_kobo, status)
		VALUES ($1, $2, $3, 'Test offer', $4, $5)
		RETURNING id::text
	`, string(unitTypeID), string(agentID), title, priceKobo, status).Scan(&offerID)
	if err != nil {
		t.Fatalf("insert agent offer: %v", err)
	}

	return domain.ID(offerID)
}
