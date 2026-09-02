package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ddddami/laivan/internal/domain"
	"github.com/ddddami/laivan/internal/repo"
	"github.com/jackc/pgx/v5/pgxpool"
)

func runGrantRole(ctx context.Context, pool *pgxpool.Pool, repo *repo.IdentityRepository, args []string) {
	fs := flag.NewFlagSet("grant-role", flag.ExitOnError)
	email := fs.String("email", "", "Email of the user")
	role := fs.String("role", "", "Role to grant (global_admin or campus_operator)")
	campusID := fs.String("campus", "", "Campus ID (required for campus_operator)")

	fs.Parse(args)

	if *email == "" || *role == "" {
		fmt.Fprintln(os.Stderr, "Usage: admin grant-role --email=<email> --role=<role> [--campus=<id>]")
		os.Exit(1)
	}

	*email = strings.ToLower(strings.TrimSpace(*email))

	// Ensure user exists and get their ID
	var userID domain.ID
	err := pool.QueryRow(ctx, "SELECT id FROM users WHERE email = $1", *email).Scan(&userID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to find user by email %q: %v\n", *email, err)
		os.Exit(1)
	}

	if *role == "global_admin" {
		_, err = pool.Exec(ctx, "UPDATE users SET roles = array_append(roles, 'global_admin') WHERE id = $1 AND NOT ('global_admin' = ANY(roles))", userID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to grant global_admin: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully granted global_admin to %s\n", *email)
		return
	}

	if *role == "campus_operator" {
		if *campusID == "" {
			fmt.Fprintln(os.Stderr, "campus is required when granting campus_operator")
			os.Exit(1)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to start transaction: %v\n", err)
			os.Exit(1)
		}
		defer tx.Rollback(ctx)

		_, err = tx.Exec(ctx, "UPDATE users SET roles = array_append(roles, 'campus_operator') WHERE id = $1 AND NOT ('campus_operator' = ANY(roles))", userID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to grant campus_operator: %v\n", err)
			os.Exit(1)
		}

		_, err = tx.Exec(ctx, "INSERT INTO campus_operators (user_id, campus_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, *campusID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to link operator to campus: %v\n", err)
			os.Exit(1)
		}

		if err := tx.Commit(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "failed to commit: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully granted campus_operator to %s for campus %s\n", *email, *campusID)
		return
	}

	fmt.Fprintf(os.Stderr, "unknown role: %s\n", *role)
	os.Exit(1)
}
