package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ddddami/laivan/internal/domain"
	"github.com/ddddami/laivan/internal/repo"
)

func TestAgentLifecycleRoutesRejectAnonymousAndMissingCSRF(t *testing.T) {
	tests := []struct {
		name string
		path string
		app  *app
		req  *http.Request
		code int
	}{
		{
			name: "anonymous suspend",
			path: "/v1/operator/agents/550e8400-e29b-41d4-a716-446655440040/suspend",
			app:  testApp(),
			req:  httptest.NewRequest(http.MethodPost, "/v1/operator/agents/550e8400-e29b-41d4-a716-446655440040/suspend", nil),
			code: http.StatusUnauthorized,
		},
		{
			name: "missing csrf",
			path: "/v1/operator/agents/550e8400-e29b-41d4-a716-446655440040/suspend",
			app:  authenticatedTestApp(domain.ID("550e8400-e29b-41d4-a716-446655440001"), &fakeAgentApplicationStore{access: domain.EffectiveAccess{Roles: []string{"campus_operator"}}}),
			req:  authenticatedRequest(http.MethodPost, "/v1/operator/agents/550e8400-e29b-41d4-a716-446655440040/suspend", `{}`, false),
			code: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			tt.app.routes().ServeHTTP(response, tt.req)
			if response.Code != tt.code {
				t.Fatalf("status code = %d, want %d", response.Code, tt.code)
			}
		})
	}
}

func TestAgentLifecycleRoutesRejectNonOperatorAndUnknownFields(t *testing.T) {
	userID := domain.ID("550e8400-e29b-41d4-a716-446655440001")

	operatorApp := authenticatedTestApp(userID, &fakeAgentApplicationStore{access: domain.EffectiveAccess{Roles: []string{"campus_operator"}}})
	unknownFieldRequest := authenticatedRequest(http.MethodPost, "/v1/operator/agents/550e8400-e29b-41d4-a716-446655440040/suspend", `{"unknown":"value"}`, true)
	unknownFieldResponse := httptest.NewRecorder()
	operatorApp.routes().ServeHTTP(unknownFieldResponse, unknownFieldRequest)
	assertErrorResponse(t, unknownFieldResponse, http.StatusBadRequest, "bad_request", `request body contains unknown field "unknown"`)

	nonOperatorApp := authenticatedTestApp(userID, &fakeAgentApplicationStore{})
	nonOperatorRequest := authenticatedRequest(http.MethodPost, "/v1/operator/agents/550e8400-e29b-41d4-a716-446655440040/suspend", `{}`, true)
	nonOperatorResponse := httptest.NewRecorder()
	nonOperatorApp.routes().ServeHTTP(nonOperatorResponse, nonOperatorRequest)
	assertErrorResponse(t, nonOperatorResponse, http.StatusForbidden, "forbidden", "Campus operator access is required")
}

func TestAgentLifecycleRoutesReturnAgentStatus(t *testing.T) {
	userID := domain.ID("550e8400-e29b-41d4-a716-446655440001")
	for _, action := range []string{"suspend", "reinstate"} {
		t.Run(action, func(t *testing.T) {
			store := &fakeAgentApplicationStore{
				access:         domain.EffectiveAccess{Roles: []string{"campus_operator"}},
				lifecycleAgent: domain.LinkedAgent{ID: domain.ID("550e8400-e29b-41d4-a716-446655440040"), Status: domain.AgentStatusSuspended},
			}
			app := authenticatedTestApp(userID, store)
			path := "/v1/operator/agents/550e8400-e29b-41d4-a716-446655440040/" + action
			request := authenticatedRequest(http.MethodPost, path, `{"operator_note":"reviewed reports"}`, true)
			response := httptest.NewRecorder()

			app.routes().ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("status code = %d, want %d", response.Code, http.StatusOK)
			}
			var body struct {
				Agent struct {
					ID     string `json:"id"`
					Status string `json:"status"`
				} `json:"agent"`
			}
			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatalf("decode lifecycle response: %v", err)
			}
			if body.Agent.ID != string(store.lifecycleAgent.ID) || body.Agent.Status != string(store.lifecycleAgent.Status) {
				t.Fatalf("agent response = %#v, want %#v", body.Agent, store.lifecycleAgent)
			}
		})
	}
}

func TestAgentLifecycleRoutesMapRepositoryErrors(t *testing.T) {
	userID := domain.ID("550e8400-e29b-41d4-a716-446655440001")
	for _, test := range []struct {
		name string
		err  error
		code int
	}{
		{name: "scope", err: repo.ErrCampusForbidden, code: http.StatusForbidden},
		{name: "transition conflict", err: repo.ErrAgentLifecycleConflict, code: http.StatusConflict},
		{name: "missing", err: repo.ErrAgentNotFound, code: http.StatusNotFound},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := &fakeAgentApplicationStore{access: domain.EffectiveAccess{Roles: []string{"campus_operator"}}, lifecycleErr: test.err}
			app := authenticatedTestApp(userID, store)
			request := authenticatedRequest(http.MethodPost, "/v1/operator/agents/550e8400-e29b-41d4-a716-446655440040/suspend", `{}`, true)
			response := httptest.NewRecorder()

			app.routes().ServeHTTP(response, request)

			if response.Code != test.code {
				t.Fatalf("status code = %d, want %d", response.Code, test.code)
			}
		})
	}
}
