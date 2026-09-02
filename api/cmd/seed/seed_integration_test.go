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
