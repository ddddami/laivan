package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDocsServesScalarReference(t *testing.T) {
	app := testApp()
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "Scalar.createApiReference") {
		t.Fatal("docs response does not include Scalar reference")
	}
}

func TestOpenAPIServesContract(t *testing.T) {
	app := testApp()
	req := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
	}
	if got := rr.Header().Get("Content-Type"); got != "text/yaml; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want text/yaml; charset=utf-8", got)
	}
	if !strings.Contains(rr.Body.String(), "openapi: 3.1.0") {
		t.Fatal("OpenAPI response does not include contract version")
	}
}
