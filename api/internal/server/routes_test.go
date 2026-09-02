package server

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ddddami/laivan/internal/config"
	"github.com/go-chi/chi/v5/middleware"
)

// capturingHandler is a slog.Handler that stores all records for inspection.
type capturingHandler struct {
	records []slog.Record
}

func (h *capturingHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }
func (h *capturingHandler) Handle(_ context.Context, r slog.Record) error {
	h.records = append(h.records, r)
	return nil
}
func (h *capturingHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *capturingHandler) WithGroup(_ string) slog.Handler      { return h }

func TestRoutesCORSAllowsIfMatch(t *testing.T) {
	app := testApp()
	req := httptest.NewRequest(http.MethodOptions, "/v1/properties/550e8400-e29b-41d4-a716-446655440000", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", http.MethodPatch)
	req.Header.Set("Access-Control-Request-Headers", "If-Match")
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
	}
	if got := rr.Header().Get("Access-Control-Allow-Headers"); !strings.Contains(got, "If-Match") {
		t.Fatalf("Access-Control-Allow-Headers = %q, want If-Match", got)
	}
}

func TestLogRequestCapturesStatusAndDuration(t *testing.T) {
	handler := &capturingHandler{}
	logger := slog.New(handler)

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
		logger:  logger,
		version: "test-version",
	}

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
	}

	var found bool
	for _, r := range handler.records {
		if r.Message == "request completed" {
			found = true
			var hasMethod, hasPath, hasStatus, hasDuration bool
			r.Attrs(func(a slog.Attr) bool {
				switch a.Key {
				case "method":
					hasMethod = a.Value.String() == "GET"
				case "path":
					hasPath = a.Value.String() == "/healthz"
				case "status":
					hasStatus = a.Value.Int64() == http.StatusOK
				case "duration":
					hasDuration = true
				}
				return true
			})
			if !hasMethod || !hasPath || !hasStatus || !hasDuration {
				t.Fatal("request log missing expected fields")
			}
		}
	}
	if !found {
		t.Fatal("request completed log not found")
	}
}

func TestRecoverPanicReturnsJSONServerError(t *testing.T) {
	handler := &capturingHandler{}
	app := testApp()
	app.logger = slog.New(handler)

	panickingHandler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})
	recoveredHandler := middleware.RequestID(app.recoverPanic(panickingHandler))
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rr := httptest.NewRecorder()

	recoveredHandler.ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusInternalServerError, "internal_server_error", "The server encountered a problem and could not process your request")
	if rr.Header().Get("Connection") != "close" {
		t.Fatalf("Connection header = %q, want close", rr.Header().Get("Connection"))
	}

	var found bool
	for _, record := range handler.records {
		if record.Message == "panic recovered" {
			found = true
		}
	}
	if !found {
		t.Fatal("panic recovery log not found")
	}
}
