package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthSessionReturnsAnonymousStateWithoutCookie(t *testing.T) {
	app := testApp()
	req := httptest.NewRequest(http.MethodGet, "/v1/auth/session", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
	}

	var body struct {
		Authenticated bool     `json:"authenticated"`
		User          any      `json:"user"`
		Roles         []string `json:"roles"`
		CSRFToken     any      `json:"csrf_token"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if body.Authenticated {
		t.Fatal("anonymous session was authenticated")
	}
	if body.User != nil || body.CSRFToken != nil {
		t.Fatal("anonymous session exposed authenticated fields")
	}
	if len(body.Roles) != 0 {
		t.Fatalf("anonymous roles = %v, want empty", body.Roles)
	}
}

func TestGoogleAuthStartReturnsUnavailableWhenNotConfigured(t *testing.T) {
	app := testApp()
	req := httptest.NewRequest(http.MethodGet, "/v1/auth/google/start", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusServiceUnavailable, "auth_unavailable", "Authentication is not configured")
}
