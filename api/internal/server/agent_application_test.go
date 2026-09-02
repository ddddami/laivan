package server

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ddddami/laivan/internal/auth"
	"github.com/ddddami/laivan/internal/domain"
	"github.com/ddddami/laivan/internal/repo"
)

func TestAgentApplicationRoutesRejectAnonymousUsers(t *testing.T) {
	app := testApp()
	request := httptest.NewRequest(http.MethodPost, "/v1/agent-applications", strings.NewReader(`{"campus_id":"550e8400-e29b-41d4-a716-446655440002","name":"Agent","phone_number":"08031234567"}`))
	response := httptest.NewRecorder()

	app.routes().ServeHTTP(response, request)

	assertErrorResponse(t, response, http.StatusUnauthorized, "unauthenticated", "Authentication is required")
}

func TestAgentApplicationListUsesAuthenticatedApplicantOwnership(t *testing.T) {
	userID := domain.ID("550e8400-e29b-41d4-a716-446655440001")
	store := &fakeAgentApplicationStore{applications: []domain.AgentApplication{{ID: domain.ID("550e8400-e29b-41d4-a716-446655440010"), ApplicantUserID: userID, CampusID: domain.ID("550e8400-e29b-41d4-a716-446655440002"), Name: "My Application", PhoneNumber: "+2348031234567", Status: domain.AgentApplicationStatusPending}}}
	app := authenticatedTestApp(userID, store)
	request := authenticatedRequest(http.MethodGet, "/v1/agent-applications", "", false)
	response := httptest.NewRecorder()

	app.routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusOK)
	}
	if store.listApplicantID != userID {
		t.Fatalf("list applicant ID = %q, want %q", store.listApplicantID, userID)
	}
	var body struct {
		Applications []map[string]any `json:"applications"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode application response: %v", err)
	}
	if len(body.Applications) != 1 || body.Applications[0]["name"] != "My Application" {
		t.Fatalf("applications = %#v, want applicant-owned application", body.Applications)
	}
}

func TestAuthSessionEnrichesEffectiveRolesAndLinkedAgent(t *testing.T) {
	userID := domain.ID("550e8400-e29b-41d4-a716-446655440001")
	store := &fakeAgentApplicationStore{access: domain.EffectiveAccess{
		Roles: []string{"campus_operator"},
		Agent: &domain.LinkedAgent{ID: domain.ID("550e8400-e29b-41d4-a716-446655440040"), Status: domain.AgentStatusSuspended},
	}}
	app := authenticatedTestApp(userID, store)
	request := authenticatedRequest(http.MethodGet, "/v1/auth/session", "", false)
	response := httptest.NewRecorder()

	app.routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusOK)
	}
	var body struct {
		Roles []string `json:"roles"`
		Agent struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"agent"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode session response: %v", err)
	}
	if len(body.Roles) != 1 || body.Roles[0] != "campus_operator" {
		t.Fatalf("roles = %v, want effective roles", body.Roles)
	}
	if body.Agent.ID != "550e8400-e29b-41d4-a716-446655440040" || body.Agent.Status != "suspended" {
		t.Fatalf("agent = %#v, want linked suspended agent", body.Agent)
	}
}

func TestOperatorListRejectsCampusOutsideOperatorScope(t *testing.T) {
	userID := domain.ID("550e8400-e29b-41d4-a716-446655440001")
	store := &fakeAgentApplicationStore{access: domain.EffectiveAccess{Roles: []string{"campus_operator"}, CampusOperatorIDs: []domain.ID{"550e8400-e29b-41d4-a716-446655440002"}}, listErr: repo.ErrCampusForbidden}
	app := authenticatedTestApp(userID, store)
	request := authenticatedRequest(http.MethodGet, "/v1/operator/agent-applications?campus_id=550e8400-e29b-41d4-a716-446655440003", "", false)
	response := httptest.NewRecorder()

	app.routes().ServeHTTP(response, request)

	assertErrorResponse(t, response, http.StatusForbidden, "forbidden", "The operator is not authorized for this campus")
}

type fakeAgentApplicationStore struct {
	access          domain.EffectiveAccess
	applications    []domain.AgentApplication
	listApplicantID domain.ID
	listErr         error
	lifecycleAgent  domain.LinkedAgent
	lifecycleErr    error
}

func (f *fakeAgentApplicationStore) GetEffectiveAccess(context.Context, domain.ID) (domain.EffectiveAccess, error) {
	return f.access, nil
}

func (f *fakeAgentApplicationStore) CreateApplication(context.Context, domain.AgentApplication) (domain.AgentApplication, error) {
	return domain.AgentApplication{}, nil
}

func (f *fakeAgentApplicationStore) ListApplications(_ context.Context, applicantID domain.ID) ([]domain.AgentApplication, error) {
	f.listApplicantID = applicantID
	return f.applications, f.listErr
}

func (f *fakeAgentApplicationStore) ListOperatorApplications(context.Context, domain.ID, domain.ID, domain.AgentApplicationStatus) ([]domain.AgentApplication, error) {
	return nil, f.listErr
}

func (f *fakeAgentApplicationStore) ActivateApplication(context.Context, domain.ID, domain.ID, string) (domain.AgentApplication, error) {
	return domain.AgentApplication{}, nil
}

func (f *fakeAgentApplicationStore) DeclineApplication(context.Context, domain.ID, domain.ID, string) (domain.AgentApplication, error) {
	return domain.AgentApplication{}, nil
}

func (f *fakeAgentApplicationStore) SuspendAgent(context.Context, domain.ID, domain.ID, string) (domain.LinkedAgent, error) {
	return f.lifecycleAgent, f.lifecycleErr
}

func (f *fakeAgentApplicationStore) ReinstateAgent(context.Context, domain.ID, domain.ID, string) (domain.LinkedAgent, error) {
	return f.lifecycleAgent, f.lifecycleErr
}

type fakePrincipalSessionStore struct {
	session domain.Session
}

func (f *fakePrincipalSessionStore) UpsertUser(context.Context, domain.ExternalIdentity, string, string) (domain.User, error) {
	return f.session.User, nil
}

func (f *fakePrincipalSessionStore) CreateSession(context.Context, domain.ID, []byte, []byte, time.Time) (domain.Session, error) {
	return f.session, nil
}

func (f *fakePrincipalSessionStore) GetSession(context.Context, []byte) (domain.Session, bool, error) {
	return f.session, true, nil
}

func (f *fakePrincipalSessionStore) UpdateSessionCSRFToken(context.Context, []byte, []byte) error {
	return nil
}

func (f *fakePrincipalSessionStore) RevokeSession(context.Context, []byte) error {
	return nil
}

func authenticatedTestApp(userID domain.ID, applications AgentApplicationStore) *app {
	testApp := testApp()
	testApp.cfg.Auth.SessionCookieName = "laivan_session"
	testApp.cfg.Auth.CSRFCookieName = "laivan_csrf"
	testApp.cfg.Auth.WebOrigin = "http://localhost:5173"
	testApp.auth = auth.NewService(auth.Config{}, &fakePrincipalSessionStore{session: domain.Session{
		TokenHash:     testTokenHash("session-token"),
		CSRFTokenHash: testTokenHash("csrf-token"),
		User:          domain.User{ID: userID, Email: "person@example.com", DisplayName: "Person Example", Status: domain.UserStatusActive},
	}}, nil)
	testApp.applications = applications
	return testApp
}

func authenticatedRequest(method, path, body string, unsafe bool) *http.Request {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.AddCookie(&http.Cookie{Name: "laivan_session", Value: "session-token"})
	request.AddCookie(&http.Cookie{Name: "laivan_csrf", Value: "csrf-token"})
	if unsafe {
		request.Header.Set("Origin", "http://localhost:5173")
		request.Header.Set("X-CSRF-Token", "csrf-token")
	}
	return request
}

func testTokenHash(value string) []byte {
	hash := sha256.Sum256([]byte(value))
	return hash[:]
}
