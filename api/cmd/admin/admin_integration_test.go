//go:build integration

package main

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func openTestDB(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()
	dbURL := os.Getenv("LAIVAN_TEST_DB_URL")
	if dbURL == "" {
		t.Skip("LAIVAN_TEST_DB_URL is required for integration tests")
	}

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		t.Fatalf("parse db url: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("connect to db: %v", err)
	}

	return pool
}

func TestGrantRole(t *testing.T) {
	ctx := context.Background()
	pool := openTestDB(t, ctx)
	t.Cleanup(pool.Close)

	// Clean up user
	if _, err := pool.Exec(ctx, "DELETE FROM users WHERE email = 'test_admin_cli@example.com'"); err != nil {
		t.Fatalf("clean up test admin user: %v", err)
	}

	var userID string
	err := pool.QueryRow(ctx, "INSERT INTO users (email, display_name) VALUES ('test_admin_cli@example.com', 'CLI Test') RETURNING id::text").Scan(&userID)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	runGrantRole(ctx, pool, []string{"--email=test_admin_cli@example.com", "--role=global_admin"})

	var exists bool
	err = pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM global_admin_roles WHERE user_id = $1)", userID).Scan(&exists)
	if err != nil {
		t.Fatalf("query global_admin_roles: %v", err)
	}

	if !exists {
		t.Errorf("expected user to be in global_admin_roles table")
	}
}
