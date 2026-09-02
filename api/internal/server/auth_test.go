package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

func TestGoogleAuthRoutesRateLimitByRemoteAddress(t *testing.T) {
	app := testApp()
	app.oidcRateLimiter = newRateLimiter(1, time.Minute)

	first := httptest.NewRequest(http.MethodGet, "/v1/auth/google/start", nil)
	first.RemoteAddr = "192.0.2.1:1234"
	firstResponse := httptest.NewRecorder()
	app.routes().ServeHTTP(firstResponse, first)
	if firstResponse.Code != http.StatusServiceUnavailable {
		t.Fatalf("first status code = %d, want %d", firstResponse.Code, http.StatusServiceUnavailable)
	}

	second := httptest.NewRequest(http.MethodGet, "/v1/auth/google/start", nil)
	second.RemoteAddr = "192.0.2.1:1234"
	secondResponse := httptest.NewRecorder()
	app.routes().ServeHTTP(secondResponse, second)

	assertErrorResponse(t, secondResponse, http.StatusTooManyRequests, "rate_limited", "Too many authentication attempts. Try again later")
	if secondResponse.Header().Get("Retry-After") == "" {
		t.Fatal("Retry-After header is empty")
	}
}
