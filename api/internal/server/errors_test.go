package server

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ddddami/laivan/internal/config"
)

func TestNotFoundResponse(t *testing.T) {
	app := testApp()
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusNotFound, "not_found", "The requested resource could not be found")
}

func TestMethodNotAllowedResponse(t *testing.T) {
	app := testApp()
	req := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusMethodNotAllowed, "method_not_allowed", "The request method is not supported for this resource")
}

func TestBadRequestResponse(t *testing.T) {
	app := testApp()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rr := httptest.NewRecorder()

	app.badRequestResponse(rr, req, errors.New("request body must not be empty"))

	assertErrorResponse(t, rr, http.StatusBadRequest, "bad_request", "request body must not be empty")
}

func TestValidationFailedResponse(t *testing.T) {
	app := testApp()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rr := httptest.NewRecorder()

	app.validationFailedResponse(rr, req, map[string]string{"name": "Name is required"})

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}

	var body struct {
		Error struct {
			Code    string            `json:"code"`
			Message string            `json:"message"`
			Fields  map[string]string `json:"fields"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if body.Error.Code != "validation_failed" {
		t.Fatalf("code = %q, want validation_failed", body.Error.Code)
	}

	if body.Error.Fields["name"] != "Name is required" {
		t.Fatalf("name field error = %q, want Name is required", body.Error.Fields["name"])
	}
}

func assertErrorResponse(t *testing.T, rr *httptest.ResponseRecorder, status int, code, message string) {
	t.Helper()

	if rr.Code != status {
		t.Fatalf("status code = %d, want %d", rr.Code, status)
	}

	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if body.Error.Code != code {
		t.Fatalf("code = %q, want %q", body.Error.Code, code)
	}

	if body.Error.Message != message {
		t.Fatalf("message = %q, want %q", body.Error.Message, message)
	}
}

func testApp() *app {
	return &app{
		cfg: config.Config{
			Env:             config.EnvTest,
			Port:            4000,
			AllowedOrigins:  []string{"http://localhost:5173"},
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    10 * time.Second,
			IdleTimeout:     time.Minute,
			ShutdownTimeout: 10 * time.Second,
			Media: config.MediaConfig{
				MaxUploadBytes: 10 << 20,
			},
		},
		logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
		version: "test-version",
	}
}
