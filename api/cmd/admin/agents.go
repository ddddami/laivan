package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/ddddami/laivan/internal/domain"
	"github.com/ddddami/laivan/internal/repo"
)

// A hardcoded system user ID for CLI actions so audit logs don't break.
// In a real database, this might need to be a known global admin ID.
// For the CLI, we'll assume the admin supplies their own actor ID or we use a fallback.
const cliActorID = domain.ID("00000000-0000-0000-0000-000000000000")

func runListApplications(ctx context.Context, agentAppRepo *repo.AgentApplicationRepository, args []string) {
	fs := flag.NewFlagSet("list-applications", flag.ExitOnError)
	status := fs.String("status", "pending", "Filter by status (pending, approved, declined)")
	fs.Parse(args)

	apps, _, err := agentAppRepo.ListApplications(ctx, *status, 1, 50)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to list applications: %v\n", err)
		os.Exit(1)
	}

	if len(apps) == 0 {
		fmt.Printf("No %s applications found.\n", *status)
		return
	}

	fmt.Printf("Found %d %s applications:\n", len(apps), *status)
	for _, app := range apps {
		fmt.Printf("- ID: %s | Applicant: %s | Phone: %s | Campus: %s\n", app.ID, app.ApplicantName, app.ApplicantPhone, app.CampusID)
	}
}

func runApproveAgent(ctx context.Context, agentAppRepo *repo.AgentApplicationRepository, args []string) {
	fs := flag.NewFlagSet("approve-agent", flag.ExitOnError)
	appID := fs.String("id", "", "ID of the pending agent application")
	actorID := fs.String("actor", string(cliActorID), "Your User ID for the audit log")
	note := fs.String("note", "Approved via admin CLI", "Operator note")
	fs.Parse(args)

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

func runSuspendAgent(ctx context.Context, agentAppRepo *repo.AgentApplicationRepository, args []string) {
	fs := flag.NewFlagSet("suspend-agent", flag.ExitOnError)
	agentID := fs.String("id", "", "ID of the active agent")
	actorID := fs.String("actor", string(cliActorID), "Your User ID for the audit log")
	note := fs.String("note", "Suspended via admin CLI", "Operator note (reason for suspension)")
	fs.Parse(args)

	if *agentID == "" {
		fmt.Fprintln(os.Stderr, "Usage: admin suspend-agent --id=<agent_id>")
		os.Exit(1)
	}

	agent, err := agentAppRepo.SuspendAgent(ctx, domain.ID(*agentID), domain.ID(*actorID), *note)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to suspend agent: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully suspended agent %s. Current status: %s\n", *agentID, agent.Status)
}
