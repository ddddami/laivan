//go:build integration

package main

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	appdb "github.com/ddddami/laivan/internal/db"
	"github.com/jackc/pgx/v5"
)

func TestSeedAgentsDoesNotReplaceExistingUserLink(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	databaseURL := os.Getenv("LAIVAN_TEST_DB_URL")
	if databaseURL == "" {
		t.Skip("LAIVAN_TEST_DB_URL is required for integration tests")
	}
	db, err := appdb.Open(ctx, databaseURL, 5)
	if err != nil {
		t.Fatalf("open integration database: %v", err)
	}
	t.Cleanup(db.Close)

	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("begin seed safety transaction: %v", err)
	}
	t.Cleanup(func() {
		if err := tx.Rollback(context.Background()); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			t.Errorf("rollback seed safety transaction: %v", err)
		}
	})

	var userID string
	if err := tx.QueryRow(ctx, `INSERT INTO users (email, display_name) VALUES ('real-seed-agent@example.com', 'Real Seed Agent') RETURNING id::text`).Scan(&userID); err != nil {
		t.Fatalf("insert existing agent user: %v", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO agents (id, user_id, display_name, phone_number) VALUES ($1, $2, 'Real Seed Agent', '+2348099999999')`, agents[0].ID, userID); err != nil {
		t.Fatalf("insert existing agent: %v", err)
	}

	err = seedAgents(ctx, tx)
	if err == nil || !strings.Contains(err.Error(), "existing agent is linked to another user") {
		t.Fatalf("seed existing agent error = %v, want linkage conflict", err)
	}

	var persistedUserID string
	if err := tx.QueryRow(ctx, `SELECT user_id::text FROM agents WHERE id = $1`, agents[0].ID).Scan(&persistedUserID); err != nil {
		t.Fatalf("load existing agent after rejected seed: %v", err)
	}
	if persistedUserID != userID {
		t.Fatalf("existing agent user ID = %s, want %s", persistedUserID, userID)
	}
}

func TestSeededAvailableOffersHaveFutaInquiryEligibility(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	databaseURL := os.Getenv("LAIVAN_TEST_DB_URL")
	if databaseURL == "" {
		t.Skip("LAIVAN_TEST_DB_URL is required for integration tests")
	}
	db, err := appdb.Open(ctx, databaseURL, 5)
	if err != nil {
		t.Fatalf("open integration database: %v", err)
	}
	t.Cleanup(db.Close)

	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("begin seed campus transaction: %v", err)
	}
	t.Cleanup(func() {
		if err := tx.Rollback(context.Background()); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			t.Errorf("rollback seed campus transaction: %v", err)
		}
	})

	var campusID string
	if err := tx.QueryRow(ctx, `SELECT id::text FROM campuses WHERE slug = 'futa'`).Scan(&campusID); err != nil {
		t.Fatalf("load FUTA campus: %v", err)
	}
	if err := seedAgents(ctx, tx); err != nil {
		t.Fatalf("seed agents: %v", err)
	}
	if err := seedAgentCampuses(ctx, tx, campusID); err != nil {
		t.Fatalf("seed agent campuses: %v", err)
	}
	if err := seedProperties(ctx, tx, campusID); err != nil {
		t.Fatalf("seed properties: %v", err)
	}
	if err := seedPropertyUnitTypes(ctx, tx); err != nil {
		t.Fatalf("seed property unit types: %v", err)
	}
	if err := seedAgentOffers(ctx, tx); err != nil {
		t.Fatalf("seed agent offers: %v", err)
	}

	for _, agent := range agents {
		var associated bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM agent_campuses
				WHERE agent_id = $1 AND campus_id = $2
			)
		`, agent.ID, campusID).Scan(&associated); err != nil {
			t.Fatalf("check campus association for agent %s: %v", agent.ID, err)
		}
		if !associated {
			t.Errorf("agent %s is not associated with FUTA", agent.ID)
		}
	}

	var ineligibleOffers int
	if err := tx.QueryRow(ctx, `
		SELECT count(*)
		FROM agent_offers ao
		JOIN property_unit_types put ON put.id = ao.property_unit_type_id
		JOIN properties p ON p.id = put.property_id
		JOIN agents a ON a.id = ao.agent_id
		WHERE ao.status = 'available'
		  AND ao.archived_at IS NULL
		  AND (a.status <> 'active' OR NOT EXISTS (
			  SELECT 1
			  FROM agent_campuses ac
			  WHERE ac.agent_id = a.id AND ac.campus_id = p.campus_id
		  ))
	`).Scan(&ineligibleOffers); err != nil {
		t.Fatalf("count ineligible available offers: %v", err)
	}
	if ineligibleOffers != 0 {
		t.Fatalf("ineligible available offers = %d, want 0", ineligibleOffers)
	}
}
