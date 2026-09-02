//go:build integration

package repo

import (
	"context"
	"errors"
	"testing"
	"time"

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

	activated, err := repository.ActivateApplication(ctx, application.ID, operatorID, "approved after review")
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
	if len(access.AgentCampusIDs) != 1 || access.AgentCampusIDs[0] != campusID {
		t.Fatalf("active agent campuses = %v, want [%s]", access.AgentCampusIDs, campusID)
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
	_, _ = db.Exec(ctx, `INSERT INTO users (email, display_name) VALUES ('dummy@example.com', 'Existing Phone Agent'); INSERT INTO agents (user_id, display_name, phone_number, status) VALUES ((SELECT id FROM users WHERE email='dummy@example.com'), 'Existing Phone Agent', '+2348061234567', 'active')`)
	repository := NewAgentApplicationRepository(db)
	application, err := repository.CreateApplication(ctx, domain.AgentApplication{ApplicantUserID: applicantID, CampusID: campusID, Name: "Phone Conflict", PhoneNumber: "+2348061234567"})
	if err != nil {
		t.Fatalf("create phone conflict application: %v", err)
	}
	if _, err := repository.ActivateApplication(ctx, application.ID, operatorID, ""); !errors.Is(err, ErrPhoneConflict) {
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

func TestAgentLifecycleRepositoryIsScopedTransactionalAndRevokesSessions(t *testing.T) {
	ctx := context.Background()
	db := openIntegrationDB(t, ctx)
	t.Cleanup(db.Close)
	resetAgentApplicationTables(t, ctx, db)
	t.Cleanup(func() { resetAgentApplicationTables(t, ctx, db) })

	campusID := testCampusID(t, ctx, db)
	operatorID := insertIdentityUser(t, ctx, db, "lifecycle-operator@example.com", "Lifecycle Operator")
	applicantID := insertIdentityUser(t, ctx, db, "lifecycle-applicant@example.com", "Lifecycle Applicant")
	if _, err := db.Exec(ctx, `INSERT INTO campus_operators (user_id, campus_id) VALUES ($1, $2)`, string(operatorID), string(campusID)); err != nil {
		t.Fatalf("insert lifecycle operator: %v", err)
	}

	repository := NewAgentApplicationRepository(db)
	application, err := repository.CreateApplication(ctx, domain.AgentApplication{ApplicantUserID: applicantID, CampusID: campusID, Name: "Lifecycle Agent", PhoneNumber: "+2348071234567"})
	if err != nil {
		t.Fatalf("create lifecycle application: %v", err)
	}
	activated, err := repository.ActivateApplication(ctx, application.ID, operatorID, "approved")
	if err != nil || activated.AgentID == nil {
		t.Fatalf("activate lifecycle application = %#v, error = %v", activated, err)
	}

	var associationCount int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM agent_campuses WHERE agent_id = $1 AND campus_id = $2`, string(*activated.AgentID), string(campusID)).Scan(&associationCount); err != nil {
		t.Fatalf("count activated agent campus associations: %v", err)
	}
	if associationCount != 1 {
		t.Fatalf("activated agent campus associations = %d, want 1", associationCount)
	}
	if _, err := repository.ActivateApplication(ctx, application.ID, operatorID, "duplicate activation"); !errors.Is(err, ErrApplicationResolved) {
		t.Fatalf("duplicate activation error = %v, want %v", err, ErrApplicationResolved)
	}

	identityRepository := NewIdentityRepository(db)
	if _, err := identityRepository.CreateSession(ctx, applicantID, []byte("12345678901234567890123456789012"), []byte("abcdefghijklmnopqrstuvwxyz123456"), time.Now().UTC().Add(time.Hour)); err != nil {
		t.Fatalf("create lifecycle session: %v", err)
	}

	suspended, err := repository.SuspendAgent(ctx, *activated.AgentID, operatorID, "reviewed reports")
	if err != nil {
		t.Fatalf("suspend agent: %v", err)
	}
	if suspended.Status != domain.AgentStatusSuspended || suspended.ID != *activated.AgentID {
		t.Fatalf("suspended agent = %#v", suspended)
	}
	var agentStatus, applicationStatus string
	if err := db.QueryRow(ctx, `SELECT status FROM agents WHERE id = $1`, string(*activated.AgentID)).Scan(&agentStatus); err != nil {
		t.Fatalf("load suspended agent: %v", err)
	}
	if err := db.QueryRow(ctx, `SELECT status FROM agent_applications WHERE id = $1`, string(application.ID)).Scan(&applicationStatus); err != nil {
		t.Fatalf("load suspended application: %v", err)
	}
	if agentStatus != "suspended" || applicationStatus != "suspended" {
		t.Fatalf("suspended statuses = %s/%s, want suspended/suspended", agentStatus, applicationStatus)
	}
	var revokedCount int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM sessions WHERE user_id = $1 AND revoked_at IS NOT NULL`, string(applicantID)).Scan(&revokedCount); err != nil {
		t.Fatalf("count revoked lifecycle sessions: %v", err)
	}
	if revokedCount != 1 {
		t.Fatalf("revoked lifecycle sessions = %d, want 1", revokedCount)
	}
	if _, found, err := identityRepository.GetSession(ctx, []byte("12345678901234567890123456789012")); err != nil || found {
		t.Fatalf("revoked session lookup found=%t error=%v, want anonymous", found, err)
	}
	access, err := repository.GetEffectiveAccess(ctx, applicantID)
	if err != nil {
		t.Fatalf("get suspended effective access: %v", err)
	}
	activeAgentRole := false
	for _, role := range access.Roles {
		if role == "active_agent" {
			activeAgentRole = true
		}
	}
	if access.Agent == nil || access.Agent.Status != domain.AgentStatusSuspended || activeAgentRole {
		t.Fatalf("suspended effective access = %#v, want no active_agent role", access)
	}
	assertAuditEvent(t, ctx, db, "agent_suspended", *activated.AgentID)

	reinstated, err := repository.ReinstateAgent(ctx, *activated.AgentID, operatorID, "review complete")
	if err != nil {
		t.Fatalf("reinstate agent: %v", err)
	}
	if reinstated.Status != domain.AgentStatusActive {
		t.Fatalf("reinstated agent = %#v", reinstated)
	}
	if err := db.QueryRow(ctx, `SELECT status FROM agent_applications WHERE id = $1`, string(application.ID)).Scan(&applicationStatus); err != nil {
		t.Fatalf("load reinstated application: %v", err)
	}
	if applicationStatus != "active" {
		t.Fatalf("reinstated application status = %s, want active", applicationStatus)
	}
	if _, found, err := identityRepository.GetSession(ctx, []byte("12345678901234567890123456789012")); err != nil || found {
		t.Fatalf("reinstated session lookup found=%t error=%v, want no automatic session", found, err)
	}
	assertAuditEvent(t, ctx, db, "agent_reinstated", *activated.AgentID)
	if _, err := repository.ReinstateAgent(ctx, *activated.AgentID, operatorID, "duplicate reinstatement"); !errors.Is(err, ErrAgentLifecycleConflict) {
		t.Fatalf("duplicate reinstatement error = %v, want %v", err, ErrAgentLifecycleConflict)
	}
}

func TestAgentLifecycleRepositoryCampusScopeGlobalAdminAndUnlinkedAgent(t *testing.T) {
	ctx := context.Background()
	db := openIntegrationDB(t, ctx)
	t.Cleanup(db.Close)
	resetAgentApplicationTables(t, ctx, db)
	t.Cleanup(func() { resetAgentApplicationTables(t, ctx, db) })

	campusID := testCampusID(t, ctx, db)
	otherCampusID := createTestCampus(t, ctx, db, "lifecycle-other-campus")
	operatorID := insertIdentityUser(t, ctx, db, "scoped-lifecycle-operator@example.com", "Scoped Operator")
	otherOperatorID := insertIdentityUser(t, ctx, db, "other-lifecycle-operator@example.com", "Other Operator")
	applicantID := insertIdentityUser(t, ctx, db, "other-lifecycle-applicant@example.com", "Other Applicant")
	if _, err := db.Exec(ctx, `INSERT INTO campus_operators (user_id, campus_id) VALUES ($1, $2), ($3, $4)`, string(operatorID), string(campusID), string(otherOperatorID), string(otherCampusID)); err != nil {
		t.Fatalf("insert scoped lifecycle operators: %v", err)
	}

	repository := NewAgentApplicationRepository(db)
	application, err := repository.CreateApplication(ctx, domain.AgentApplication{ApplicantUserID: applicantID, CampusID: otherCampusID, Name: "Other Campus Agent", PhoneNumber: "+2348081234567"})
	if err != nil {
		t.Fatalf("create other campus application: %v", err)
	}
	activated, err := repository.ActivateApplication(ctx, application.ID, otherOperatorID, "approved")
	if err != nil || activated.AgentID == nil {
		t.Fatalf("activate other campus application = %#v, error = %v", activated, err)
	}
	if _, err := repository.SuspendAgent(ctx, *activated.AgentID, operatorID, "cross-campus"); !errors.Is(err, ErrCampusForbidden) {
		t.Fatalf("cross-campus suspend error = %v, want %v", err, ErrCampusForbidden)
	}
	if _, err := db.Exec(ctx, `INSERT INTO global_admin_roles (user_id) VALUES ($1)`, string(operatorID)); err != nil {
		t.Fatalf("insert global admin role: %v", err)
	}
	if _, err := repository.SuspendAgent(ctx, *activated.AgentID, operatorID, "global review"); err != nil {
		t.Fatalf("global admin cross-campus suspend: %v", err)
	}

}

func TestAgentCampusBackfillFollowsOfferCampusWithoutChangingPublicRecords(t *testing.T) {
	ctx := context.Background()
	db := openIntegrationDB(t, ctx)
	t.Cleanup(db.Close)
	resetAgentApplicationTables(t, ctx, db)
	t.Cleanup(func() { resetAgentApplicationTables(t, ctx, db) })

	campusID := testCampusID(t, ctx, db)
	agentID := domain.ID("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	propertyID := domain.ID("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	unitTypeID := domain.ID("cccccccc-cccc-4ccc-8ccc-cccccccccccc")
	offerID := domain.ID("dddddddd-dddd-4ddd-8ddd-dddddddddddd")

	var userID string
	if err := db.QueryRow(ctx, `INSERT INTO users (email, display_name) VALUES ('backfill@example.com', 'Backfill Agent') RETURNING id::text`).Scan(&userID); err != nil {
		t.Fatalf("insert backfill user: %v", err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO agents (id, user_id, display_name, phone_number) VALUES ($1, $2, $3, $4)`, string(agentID), userID, "Backfill Agent", "+2348091111111"); err != nil {
		t.Fatalf("insert backfill agent: %v", err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO properties (id, campus_id, name, area) VALUES ($1, $2, $3, $4)`, string(propertyID), string(campusID), "Backfill Property", "Backfill Area"); err != nil {
		t.Fatalf("insert backfill property: %v", err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO property_unit_types (id, property_id, category, name) VALUES ($1, $2, $3, $4)`, string(unitTypeID), string(propertyID), "self_contained", "Backfill Unit"); err != nil {
		t.Fatalf("insert backfill unit type: %v", err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO agent_offers (id, property_unit_type_id, agent_id, title, price_kobo) VALUES ($1, $2, $3, $4, $5)`, string(offerID), string(unitTypeID), string(agentID), "Backfill Offer", 100000); err != nil {
		t.Fatalf("insert backfill offer: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db.Exec(ctx, `DELETE FROM properties WHERE id = $1`, string(propertyID)); err != nil {
			t.Errorf("cleanup backfill property: %v", err)
		}
	})

	if _, err := db.Exec(ctx, `INSERT INTO agent_campuses (agent_id, campus_id) SELECT DISTINCT ao.agent_id, p.campus_id FROM agent_offers ao JOIN property_unit_types put ON put.id = ao.property_unit_type_id JOIN properties p ON p.id = put.property_id ON CONFLICT (agent_id, campus_id) DO NOTHING`); err != nil {
		t.Fatalf("run agent campus backfill: %v", err)
	}
	var associationCount int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM agent_campuses WHERE agent_id = $1 AND campus_id = $2`, string(agentID), string(campusID)).Scan(&associationCount); err != nil {
		t.Fatalf("count backfilled association: %v", err)
	}
	if associationCount != 1 {
		t.Fatalf("backfilled associations = %d, want 1", associationCount)
	}
	var persistedPropertyID, persistedUnitTypeID, persistedOfferAgentID string
	if err := db.QueryRow(ctx, `SELECT p.id::text, put.id::text, ao.agent_id::text FROM properties p JOIN property_unit_types put ON put.property_id = p.id JOIN agent_offers ao ON ao.property_unit_type_id = put.id WHERE ao.id = $1`, string(offerID)).Scan(&persistedPropertyID, &persistedUnitTypeID, &persistedOfferAgentID); err != nil {
		t.Fatalf("load backfill public records: %v", err)
	}
	if persistedPropertyID != string(propertyID) || persistedUnitTypeID != string(unitTypeID) || persistedOfferAgentID != string(agentID) {
		t.Fatalf("backfill public relationships = %s/%s/%s, want %s/%s/%s", persistedPropertyID, persistedUnitTypeID, persistedOfferAgentID, propertyID, unitTypeID, agentID)
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
