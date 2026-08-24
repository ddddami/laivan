//go:build integration

package repo

import (
	"context"
	"errors"
	"testing"

	"github.com/ddddami/laivan/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestAgentApplicationRepositoryWorkflowIsScopedAndAudited(t *testing.T) {
	ctx := context.Background()
	db := openIntegrationDB(t, ctx)
	t.Cleanup(db.Close)
	resetAgentApplicationTables(t, ctx, db)
	t.Cleanup(func() { resetAgentApplicationTables(t, ctx, db) })

	campusID := testCampusID(t, ctx, db)
	operatorID := insertIdentityUser(t, ctx, db, "operator@example.com", "Campus Operator")
	applicantID := insertIdentityUser(t, ctx, db, "applicant@example.com", "Applicant Example")
	if _, err := db.Exec(ctx, `INSERT INTO campus_operators (user_id, campus_id) VALUES ($1, $2)`, string(operatorID), string(campusID)); err != nil {
		t.Fatalf("insert campus operator: %v", err)
	}

	repository := NewAgentApplicationRepository(db)
	application, err := repository.CreateApplication(ctx, domain.AgentApplication{
		ApplicantUserID: applicantID,
		CampusID:        campusID,
		Name:            "Applicant Agent",
		PhoneNumber:     "+2348031234567",
	})
	if err != nil {
		t.Fatalf("create application: %v", err)
	}
	if _, err := repository.CreateApplication(ctx, domain.AgentApplication{
		ApplicantUserID: applicantID,
		CampusID:        campusID,
		Name:            "Duplicate Application",
		PhoneNumber:     "+2348031234567",
	}); !errors.Is(err, ErrApplicationConflict) {
		t.Fatalf("duplicate application error = %v, want %v", err, ErrApplicationConflict)
	}

	activated, err := repository.ActivateApplication(ctx, application.ID, operatorID, nil, "approved after review")
	if err != nil {
		t.Fatalf("activate application: %v", err)
	}
	if activated.Status != domain.AgentApplicationStatusActive || activated.AgentID == nil {
		t.Fatalf("activated application = %#v, want active application with agent", activated)
	}
	var linkedUserID, agentStatus string
	if err := db.QueryRow(ctx, `SELECT user_id::text, status FROM agents WHERE id = $1`, string(*activated.AgentID)).Scan(&linkedUserID, &agentStatus); err != nil {
		t.Fatalf("load created agent: %v", err)
	}
	if linkedUserID != string(applicantID) || agentStatus != "active" {
		t.Fatalf("created agent user/status = %s/%s, want %s/active", linkedUserID, agentStatus, applicantID)
	}
	access, err := repository.GetEffectiveAccess(ctx, applicantID)
	if err != nil {
		t.Fatalf("get active agent access: %v", err)
	}
	if access.Agent == nil || access.Agent.Status != domain.AgentStatusActive || len(access.Roles) != 1 || access.Roles[0] != "active_agent" {
		t.Fatalf("active agent access = %#v, want active_agent role and active linked agent", access)
	}
	if _, err := db.Exec(ctx, `UPDATE agents SET status = 'suspended' WHERE id = $1`, string(*activated.AgentID)); err != nil {
		t.Fatalf("suspend linked agent for access test: %v", err)
	}
	access, err = repository.GetEffectiveAccess(ctx, applicantID)
	if err != nil {
		t.Fatalf("get suspended agent access: %v", err)
	}
	if access.Agent == nil || access.Agent.Status != domain.AgentStatusSuspended || len(access.Roles) != 0 {
		t.Fatalf("suspended agent access = %#v, want linked status without effective roles", access)
	}
	assertAuditEvent(t, ctx, db, "agent_application_activated", application.ID)

	otherCampusID := createTestCampus(t, ctx, db, "agent-application-other-campus")
	if _, err := repository.ListOperatorApplications(ctx, operatorID, otherCampusID, domain.AgentApplicationStatusPending); !errors.Is(err, ErrCampusForbidden) {
		t.Fatalf("cross-campus list error = %v, want %v", err, ErrCampusForbidden)
	}

	legacyApplicantID := insertIdentityUser(t, ctx, db, "legacy-applicant@example.com", "Legacy Applicant")
	legacyID := insertLegacyAgent(t, ctx, db, "Legacy Agent", "+2348041234567")
	legacyApplication, err := repository.CreateApplication(ctx, domain.AgentApplication{
		ApplicantUserID: legacyApplicantID,
		CampusID:        campusID,
		Name:            "Submitted Name Must Not Replace Legacy",
		PhoneNumber:     "+2348041234567",
	})
	if err != nil {
		t.Fatalf("create legacy application: %v", err)
	}
	linked, err := repository.ActivateApplication(ctx, legacyApplication.ID, operatorID, &legacyID, "linked after explicit review")
	if err != nil {
		t.Fatalf("explicitly link legacy agent: %v", err)
	}
	if linked.AgentID == nil || *linked.AgentID != legacyID {
		t.Fatalf("linked application agent ID = %v, want %s", linked.AgentID, legacyID)
	}
	var displayName, phoneNumber, userID string
	if err := db.QueryRow(ctx, `SELECT display_name, phone_number, user_id::text FROM agents WHERE id = $1`, string(legacyID)).Scan(&displayName, &phoneNumber, &userID); err != nil {
		t.Fatalf("load explicitly linked agent: %v", err)
	}
	if displayName != "Legacy Agent" || phoneNumber != "+2348041234567" || userID != string(legacyApplicantID) {
		t.Fatalf("linked legacy agent = %s/%s/%s, public fields were overwritten", displayName, phoneNumber, userID)
	}
	assertAuditEvent(t, ctx, db, "agent_application_linked", legacyApplication.ID)

	declineApplicantID := insertIdentityUser(t, ctx, db, "decline@example.com", "Decline Applicant")
	toDecline, err := repository.CreateApplication(ctx, domain.AgentApplication{ApplicantUserID: declineApplicantID, CampusID: campusID, Name: "Decline Me", PhoneNumber: "+2348051234567"})
	if err != nil {
		t.Fatalf("create application to decline: %v", err)
	}
	declined, err := repository.DeclineApplication(ctx, toDecline.ID, operatorID, "not enough information")
	if err != nil {
		t.Fatalf("decline application: %v", err)
	}
	if declined.Status != domain.AgentApplicationStatusDeclined || declined.AgentID != nil {
		t.Fatalf("declined application = %#v", declined)
	}
	assertAuditEvent(t, ctx, db, "agent_application_declined", toDecline.ID)
}

func TestAgentApplicationRepositoryPhoneConflictLeavesApplicationPending(t *testing.T) {
	ctx := context.Background()
	db := openIntegrationDB(t, ctx)
	t.Cleanup(db.Close)
	resetAgentApplicationTables(t, ctx, db)
	t.Cleanup(func() { resetAgentApplicationTables(t, ctx, db) })

	campusID := testCampusID(t, ctx, db)
	operatorID := insertIdentityUser(t, ctx, db, "phone-operator@example.com", "Phone Operator")
	applicantID := insertIdentityUser(t, ctx, db, "phone-applicant@example.com", "Phone Applicant")
	if _, err := db.Exec(ctx, `INSERT INTO campus_operators (user_id, campus_id) VALUES ($1, $2)`, string(operatorID), string(campusID)); err != nil {
		t.Fatalf("insert campus operator: %v", err)
	}
	insertLegacyAgent(t, ctx, db, "Existing Phone Agent", "+2348061234567")
	repository := NewAgentApplicationRepository(db)
	application, err := repository.CreateApplication(ctx, domain.AgentApplication{ApplicantUserID: applicantID, CampusID: campusID, Name: "Phone Conflict", PhoneNumber: "+2348061234567"})
	if err != nil {
		t.Fatalf("create phone conflict application: %v", err)
	}
	if _, err := repository.ActivateApplication(ctx, application.ID, operatorID, nil, ""); !errors.Is(err, ErrPhoneConflict) {
		t.Fatalf("phone conflict activation error = %v, want %v", err, ErrPhoneConflict)
	}
	var status string
	if err := db.QueryRow(ctx, `SELECT status FROM agent_applications WHERE id = $1`, string(application.ID)).Scan(&status); err != nil {
		t.Fatalf("load pending application: %v", err)
	}
	if status != string(domain.AgentApplicationStatusPending) {
		t.Fatalf("application status = %q, want pending", status)
	}
	var auditCount int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM audit_events WHERE resource_id = $1`, string(application.ID)).Scan(&auditCount); err != nil {
		t.Fatalf("count phone conflict audit events: %v", err)
	}
	if auditCount != 0 {
		t.Fatalf("phone conflict audit count = %d, want 0", auditCount)
	}
}

func resetAgentApplicationTables(t *testing.T, ctx context.Context, db *pgxpool.Pool) {
	t.Helper()
	if _, err := db.Exec(ctx, `TRUNCATE agent_applications, audit_events, campus_operators, global_admin_roles, agents, users CASCADE`); err != nil {
		t.Fatalf("truncate agent application tables: %v", err)
	}
}

func insertIdentityUser(t *testing.T, ctx context.Context, db *pgxpool.Pool, email, displayName string) domain.ID {
	t.Helper()
	var id string
	if err := db.QueryRow(ctx, `INSERT INTO users (email, display_name) VALUES ($1, $2) RETURNING id::text`, email, displayName).Scan(&id); err != nil {
		t.Fatalf("insert identity user: %v", err)
	}
	return domain.ID(id)
}

func insertLegacyAgent(t *testing.T, ctx context.Context, db *pgxpool.Pool, displayName, phoneNumber string) domain.ID {
	t.Helper()
	var id string
	if err := db.QueryRow(ctx, `INSERT INTO agents (display_name, phone_number) VALUES ($1, $2) RETURNING id::text`, displayName, phoneNumber).Scan(&id); err != nil {
		t.Fatalf("insert legacy agent: %v", err)
	}
	return domain.ID(id)
}

func assertAuditEvent(t *testing.T, ctx context.Context, db *pgxpool.Pool, action string, resourceID domain.ID) {
	t.Helper()
	var count int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM audit_events WHERE action = $1 AND resource_id = $2`, action, string(resourceID)).Scan(&count); err != nil {
		t.Fatalf("count audit event: %v", err)
	}
	if count != 1 {
		t.Fatalf("audit event count for %s/%s = %d, want 1", action, resourceID, count)
	}
}
