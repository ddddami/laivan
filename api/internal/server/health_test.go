package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ddddami/laivan/internal/config"
)

func TestHealthz(t *testing.T) {
	cfg := config.Config{
		Env:             config.EnvTest,
		Port:            4000,
		AllowedOrigins:  []string{"http://localhost:5173"},
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     time.Minute,
		ShutdownTimeout: 10 * time.Second,
	}
	app := &app{
		cfg:     cfg,
		logger:  slog.New(slog.NewTextHandler(httptest.NewRecorder(), nil)),
		version: "test-version",
	}

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
	}

	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	var body struct {
		Status      string `json:"status"`
		Environment string `json:"environment"`
		Version     string `json:"version"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if body.Status != "ok" {
		t.Fatalf("status = %q, want ok", body.Status)
	}

	if body.Environment != config.EnvTest {
		t.Fatalf("environment = %q, want %q", body.Environment, config.EnvTest)
	}

	if body.Version != "test-version" {
		t.Fatalf("version = %q, want test-version", body.Version)
	}
}
