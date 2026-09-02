package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/ddddami/laivan/internal/domain"
	"github.com/ddddami/laivan/internal/repo"
	"github.com/jackc/pgx/v5/pgxpool"
)

// cliActorID represents the CLI action runner
const cliActorID = domain.ID("00000000-0000-0000-0000-000000000000")

func runListApplications(ctx context.Context, pool *pgxpool.Pool, args []string) {
	fs := flag.NewFlagSet("list-applications", flag.ExitOnError)
	status := fs.String("status", "pending", "Filter by status (pending, active, declined)")
	_ = fs.Parse(args)

	rows, err := pool.Query(ctx, "SELECT id, name, phone_number, campus_id FROM agent_applications WHERE status = $1 ORDER BY created_at DESC", *status)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to list applications: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	var apps []domain.AgentApplication
	for rows.Next() {
		var app domain.AgentApplication
		if err := rows.Scan(&app.ID, &app.Name, &app.PhoneNumber, &app.CampusID); err != nil {
			fmt.Fprintf(os.Stderr, "failed to scan application: %v\n", err)
			os.Exit(1)
		}
		apps = append(apps, app)
	}

	if len(apps) == 0 {
		fmt.Printf("No %s applications found.\n", *status)
		return
	}

	fmt.Printf("Found %d %s applications:\n", len(apps), *status)
	for _, app := range apps {
		fmt.Printf("- ID: %s | Applicant: %s | Phone: %s | Campus: %s\n", app.ID, app.Name, app.PhoneNumber, app.CampusID)
	}
}

func runApproveAgent(ctx context.Context, agentAppRepo *repo.AgentApplicationRepository, args []string) {
	fs := flag.NewFlagSet("approve-agent", flag.ExitOnError)
	appID := fs.String("id", "", "ID of the pending agent application")
	actorID := fs.String("actor", string(cliActorID), "Your User ID for the audit log")
	note := fs.String("note", "Approved via admin CLI", "Operator note")
	_ = fs.Parse(args)

	if *appID == "" {
		fmt.Fprintln(os.Stderr, "Usage: admin approve-agent --id=<application_id>")
		os.Exit(1)
	}

	agent, err := agentAppRepo.ActivateApplication(ctx, domain.ID(*appID), domain.ID(*actorID), nil, *note)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to approve application: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully approved application %s. New Agent ID: %s\n", *appID, agent.ID)
}

func runSuspendAgent(ctx context.Context, pool *pgxpool.Pool, agentAppRepo *repo.AgentApplicationRepository, args []string) {
	fs := flag.NewFlagSet("suspend-agent", flag.ExitOnError)
	email := fs.String("email", "", "Email of the active agent to suspend")
	actorID := fs.String("actor", string(cliActorID), "Your User ID for the audit log")
	note := fs.String("note", "Suspended via admin CLI", "Operator note (reason for suspension)")
	_ = fs.Parse(args)

	if *email == "" {
		fmt.Fprintln(os.Stderr, "Usage: admin suspend-agent --email=<agent_email>")
		os.Exit(1)
	}

	var agentID string
	query := `
		SELECT a.id 
		FROM agents a 
		JOIN users u ON u.id = a.user_id 
		WHERE u.email = $1 AND a.status = 'active'`

	err := pool.QueryRow(ctx, query, *email).Scan(&agentID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to find active agent with email %q: %v\n", *email, err)
		os.Exit(1)
	}

	agent, err := agentAppRepo.SuspendAgent(ctx, domain.ID(agentID), domain.ID(*actorID), *note)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to suspend agent: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully suspended agent %s (Email: %s). Current status: %s\n", agentID, *email, agent.Status)
}
