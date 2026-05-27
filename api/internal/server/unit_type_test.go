package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreatePropertyUnitTypeValidationErrors(t *testing.T) {
	app := testAppWithRepo()
	body := `{"category":"self_contained"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/properties/not-a-uuid/unit-types", strings.NewReader(body))
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}

	var bodyDecoded struct {
		Error struct {
			Code   string            `json:"code"`
			Fields map[string]string `json:"fields"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&bodyDecoded); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if bodyDecoded.Error.Code != "validation_failed" {
		t.Fatalf("code = %q, want validation_failed", bodyDecoded.Error.Code)
	}
	if bodyDecoded.Error.Fields["id"] == "" {
		t.Fatal("id validation error missing")
	}
}

func TestCreatePropertyUnitTypeDefaultsNameFromCategory(t *testing.T) {
	spy := &spyPropertyRepo{stub: &stubPropertyRepo{}}
	app := testApp()
	app.propertyRepo = spy

	body := `{"category":"self_contained","bedroom_count":1,"bathroom_type":"private","kitchen_type":"private"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/properties/550e8400-e29b-41d4-a716-446655440000/unit-types", strings.NewReader(body))
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusCreated)
	}

	if spy.createdUnitType.Name != "Self-contained" {
		t.Fatalf("name = %q, want Self-contained", spy.createdUnitType.Name)
	}
}

func TestCreatePropertyUnitTypeReturnsUnitType(t *testing.T) {
	app := testAppWithRepo()
	body := `{"category":"self_contained","name":"Self-contained","description":"Private room with bathroom and kitchenette.","bedroom_count":1,"has_parlour":false,"bathroom_type":"private","kitchen_type":"private"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/properties/550e8400-e29b-41d4-a716-446655440000/unit-types", strings.NewReader(body))
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusCreated)
	}

	var bodyDecoded struct {
		UnitType struct {
			ID           string `json:"id"`
			PropertyID   string `json:"property_id"`
			Category     string `json:"category"`
			Name         string `json:"name"`
			BedroomCount int    `json:"bedroom_count"`
			BathroomType string `json:"bathroom_type"`
			KitchenType  string `json:"kitchen_type"`
		} `json:"unit_type"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&bodyDecoded); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if bodyDecoded.UnitType.Category != "self_contained" {
		t.Fatalf("category = %q, want self_contained", bodyDecoded.UnitType.Category)
	}
	if bodyDecoded.UnitType.Name != "Self-contained" {
		t.Fatalf("unit type name = %q, want Self-contained", bodyDecoded.UnitType.Name)
	}
	if bodyDecoded.UnitType.BedroomCount != 1 {
		t.Fatalf("bedroom_count = %d, want 1", bodyDecoded.UnitType.BedroomCount)
	}
	if bodyDecoded.UnitType.BathroomType != "private" {
		t.Fatalf("bathroom_type = %q, want private", bodyDecoded.UnitType.BathroomType)
	}
	if bodyDecoded.UnitType.KitchenType != "private" {
		t.Fatalf("kitchen_type = %q, want private", bodyDecoded.UnitType.KitchenType)
	}
	if bodyDecoded.UnitType.PropertyID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("property ID = %q, want 550e8400-e29b-41d4-a716-446655440000", bodyDecoded.UnitType.PropertyID)
	}
}

func TestCreatePropertyUnitTypePropertyNotFound(t *testing.T) {
	app := testAppWithRepo()
	body := `{"category":"self_contained","name":"Self-contained"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/properties/11111111-1111-1111-1111-111111111111/unit-types", strings.NewReader(body))
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusNotFound, "not_found", "The requested resource could not be found")
}

func TestListPropertyUnitTypesReturnsUnitTypes(t *testing.T) {
	app := testAppWithRepo()
	req := httptest.NewRequest(http.MethodGet, "/v1/properties/550e8400-e29b-41d4-a716-446655440000/unit-types", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
	}

	var body struct {
		UnitTypes []struct {
			PropertyID string `json:"property_id"`
			Name       string `json:"name"`
		} `json:"unit_types"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if len(body.UnitTypes) != 1 {
		t.Fatalf("unit types length = %d, want 1", len(body.UnitTypes))
	}
	if body.UnitTypes[0].Name != "Self-contained" {
		t.Fatalf("unit type name = %q, want Self-contained", body.UnitTypes[0].Name)
	}
}

func TestListPropertyUnitTypesPropertyNotFound(t *testing.T) {
	app := testAppWithRepo()
	req := httptest.NewRequest(http.MethodGet, "/v1/properties/11111111-1111-1111-1111-111111111111/unit-types", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusNotFound, "not_found", "The requested resource could not be found")
}
