package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ddddami/laivan/internal/repo"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	dbURL := os.Getenv("LAIVAN_DB_URL")
	if dbURL == "" {
		fmt.Fprintln(os.Stderr, "LAIVAN_DB_URL environment variable is required")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "failed to ping database: %v\n", err)
		os.Exit(1)
	}

	agentAppRepo := repo.NewAgentApplicationRepository(pool)

	cmd := os.Args[1]
	switch cmd {
	case "grant-role":
		runGrantRole(ctx, pool, os.Args[2:])
	case "list-applications":
		runListApplications(ctx, pool, os.Args[2:])
	case "approve-agent":
		runApproveAgent(ctx, agentAppRepo, os.Args[2:])
	case "suspend-agent":
		runSuspendAgent(ctx, pool, agentAppRepo, os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: admin <command> [flags]")
	fmt.Println("\nCommands:")
	fmt.Println("  grant-role         Grant a user the global_admin or campus_operator role")
	fmt.Println("  list-applications  List pending agent applications")
	fmt.Println("  approve-agent      Approve a pending agent application")
	fmt.Println("  suspend-agent      Suspend an active agent")
}
