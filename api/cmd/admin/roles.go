package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ddddami/laivan/internal/domain"
	"github.com/ddddami/laivan/internal/repo"
	"github.com/jackc/pgx/v5/pgxpool"
)

func runGrantRole(ctx context.Context, pool *pgxpool.Pool, args []string) {
	fs := flag.NewFlagSet("grant-role", flag.ExitOnError)
	email := fs.String("email", "", "Email of the user")
	role := fs.String("role", "", "Role to grant (global_admin or campus_operator)")
	campusID := fs.String("campus", "", "Campus ID (required for campus_operator)")
	actorID := fs.String("actor", "", "Existing global admin user ID; omit only for first global admin bootstrap")

	_ = fs.Parse(args)

	if *email == "" || *role == "" {
		fmt.Fprintln(os.Stderr, "Usage: admin grant-role --email=<email> --role=<role> [--campus=<id>]")
		os.Exit(1)
	}

	change, err := repo.NewRoleRepository(pool).Grant(ctx, strings.ToLower(strings.TrimSpace(*email)), *role, domain.ID(*campusID), domain.ID(*actorID))
	if err != nil {
		printRoleError(err)
	}
	if !change.Changed {
		fmt.Printf("Role %s was already granted to %s\n", *role, *email)
		return
	}
	fmt.Printf("Successfully granted %s to %s\n", *role, *email)
}

func runRevokeRole(ctx context.Context, pool *pgxpool.Pool, args []string) {
	fs := flag.NewFlagSet("revoke-role", flag.ExitOnError)
	email := fs.String("email", "", "Email of the user")
	role := fs.String("role", "", "Role to revoke (global_admin or campus_operator)")
	campusID := fs.String("campus", "", "Campus ID (required for campus_operator)")
	actorID := fs.String("actor", "", "Existing global admin user ID")
	_ = fs.Parse(args)

	if *email == "" || *role == "" || *actorID == "" {
		fmt.Fprintln(os.Stderr, "Usage: admin revoke-role --email=<email> --role=<role> --actor=<global_admin_user_id> [--campus=<id>]")
		os.Exit(1)
	}
	change, err := repo.NewRoleRepository(pool).Revoke(ctx, strings.ToLower(strings.TrimSpace(*email)), *role, domain.ID(*campusID), domain.ID(*actorID))
	if err != nil {
		printRoleError(err)
	}
	if !change.Changed {
		fmt.Printf("Role %s was not assigned to %s\n", *role, *email)
		return
	}
	fmt.Printf("Successfully revoked %s from %s\n", *role, *email)
}

func printRoleError(err error) {
	switch {
	case errors.Is(err, repo.ErrNotFound):
		fmt.Fprintln(os.Stderr, "target user was not found")
	case errors.Is(err, repo.ErrRoleActorRequired):
		fmt.Fprintln(os.Stderr, "an existing global admin actor is required for this role change")
	case errors.Is(err, repo.ErrRoleForbidden):
		fmt.Fprintln(os.Stderr, "the actor is not a global admin")
	default:
		fmt.Fprintf(os.Stderr, "role change failed: %v\n", err)
	}
	os.Exit(1)
}
