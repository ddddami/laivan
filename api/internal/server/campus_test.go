package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetCampusBySlugReturnsPublicCampus(t *testing.T) {
	app := testAppWithRepo()
	req := httptest.NewRequest(http.MethodGet, "/v1/campuses/futa", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
	}

	var body struct {
		Campus struct {
			ID        string `json:"id"`
			Slug      string `json:"slug"`
			Name      string `json:"name"`
			ShortName string `json:"short_name"`
		} `json:"campus"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if body.Campus.ID != "550e8400-e29b-41d4-a716-446655440002" {
		t.Fatalf("campus ID = %q, want 550e8400-e29b-41d4-a716-446655440002", body.Campus.ID)
	}
	if body.Campus.Slug != "futa" {
		t.Fatalf("campus slug = %q, want futa", body.Campus.Slug)
	}
	if body.Campus.Name != "Federal University of Technology, Akure" {
		t.Fatalf("campus name = %q, want Federal University of Technology, Akure", body.Campus.Name)
	}
	if body.Campus.ShortName != "FUTA" {
		t.Fatalf("campus short name = %q, want FUTA", body.Campus.ShortName)
	}
}

func TestGetCampusBySlugNotFound(t *testing.T) {
	app := testAppWithRepo()
	req := httptest.NewRequest(http.MethodGet, "/v1/campuses/missing-campus", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusNotFound, "not_found", "The requested resource could not be found")
}
