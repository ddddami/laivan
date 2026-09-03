package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ddddami/laivan/internal/domain"
	"github.com/ddddami/laivan/internal/repo"
)

func TestCreatePropertyUnitTypeRequiresAuthentication(t *testing.T) {
	app := testApp()
	req := httptest.NewRequest(http.MethodPost, "/v1/properties/550e8400-e29b-41d4-a716-446655440000/unit-types", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusUnauthorized, "unauthenticated", "Authentication is required")
}

func TestCreatePropertyUnitTypeInvalidBathroomType(t *testing.T) {
	app := testAppWithActiveAgentRepo()
	body := `{"category":"self_contained","bathroom_type":"invalid"}`
	req := authenticatedRequest(http.MethodPost, "/v1/properties/550e8400-e29b-41d4-a716-446655440000/unit-types", body, true)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}

	var bodyDecoded struct {
		Error struct {
			Fields map[string]string `json:"fields"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&bodyDecoded); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if bodyDecoded.Error.Fields["bathroom_type"] == "" {
		t.Fatal("bathroom_type validation error missing")
	}
}

func TestCreatePropertyUnitTypeInvalidKitchenType(t *testing.T) {
	app := testAppWithActiveAgentRepo()
	body := `{"category":"self_contained","kitchen_type":"invalid"}`
	req := authenticatedRequest(http.MethodPost, "/v1/properties/550e8400-e29b-41d4-a716-446655440000/unit-types", body, true)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}

	var bodyDecoded struct {
		Error struct {
			Fields map[string]string `json:"fields"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&bodyDecoded); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if bodyDecoded.Error.Fields["kitchen_type"] == "" {
		t.Fatal("kitchen_type validation error missing")
	}
}

func TestCreatePropertyUnitTypeValidationErrors(t *testing.T) {
	app := testAppWithActiveAgentRepo()
	body := `{"category":"self_contained"}`
	req := authenticatedRequest(http.MethodPost, "/v1/properties/not-a-uuid/unit-types", body, true)
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
	spy := &spyPropertyRepo{stubPropertyRepo: &stubPropertyRepo{}}
	app := testAppWithActiveAgentRepo()
	app.propertyRepo = spy

	body := `{"category":"self_contained","bedroom_count":1,"bathroom_type":"private","kitchen_type":"private"}`
	req := authenticatedRequest(http.MethodPost, "/v1/properties/550e8400-e29b-41d4-a716-446655440000/unit-types", body, true)
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
	app := testAppWithActiveAgentRepo()
	body := `{"category":"self_contained","name":"Self-contained","description":"Private room with bathroom and kitchenette.","bedroom_count":1,"has_parlour":false,"bathroom_type":"private","kitchen_type":"private"}`
	req := authenticatedRequest(http.MethodPost, "/v1/properties/550e8400-e29b-41d4-a716-446655440000/unit-types", body, true)
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
			Version      int    `json:"version"`
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
	if bodyDecoded.UnitType.Version != 1 {
		t.Fatalf("version = %d, want 1", bodyDecoded.UnitType.Version)
	}
}

func TestCreatePropertyUnitTypePropertyNotFound(t *testing.T) {
	app := testAppWithActiveAgentRepo()
	body := `{"category":"self_contained","name":"Self-contained"}`
	req := authenticatedRequest(http.MethodPost, "/v1/properties/11111111-1111-1111-1111-111111111111/unit-types", body, true)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusNotFound, "not_found", "The requested resource could not be found")
}

func TestCreatePropertyUnitTypeRejectsUnassignedCampus(t *testing.T) {
	app := authenticatedTestApp(domain.ID("550e8400-e29b-41d4-a716-446655440001"), &fakeAgentApplicationStore{access: domain.EffectiveAccess{
		Agent:          &domain.LinkedAgent{ID: domain.ID("550e8400-e29b-41d4-a716-446655440040"), Status: domain.AgentStatusActive},
		AgentCampusIDs: []domain.ID{"550e8400-e29b-41d4-a716-446655440003"},
	}})
	app.propertyRepo = &stubPropertyRepo{}
	body := `{"category":"self_contained"}`
	req := authenticatedRequest(http.MethodPost, "/v1/properties/550e8400-e29b-41d4-a716-446655440000/unit-types", body, true)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusForbidden, "forbidden", "You are not authorized to contribute to this campus")
}

func TestUpdatePropertyUnitTypeRequiresOperatorAccess(t *testing.T) {
	spy := &spyPropertyRepo{stubPropertyRepo: &stubPropertyRepo{}}
	app := testAppWithActiveAgentRepo()
	app.propertyRepo = spy
	req := authenticatedRequest(http.MethodPatch, "/v1/unit-types/550e8400-e29b-41d4-a716-446655440020", `{"description":"Updated"}`, true)
	req.Header.Set("If-Match", `"property-unit-type-550e8400-e29b-41d4-a716-446655440020-1"`)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	assertErrorCodeResponse(t, rr, http.StatusForbidden, "forbidden")
	if spy.updateUnitTypeCalls != 0 {
		t.Fatalf("update unit type calls = %d, want 0", spy.updateUnitTypeCalls)
	}
}

func TestUpdatePropertyUnitTypeRequiresIfMatch(t *testing.T) {
	spy := &spyPropertyRepo{stubPropertyRepo: &stubPropertyRepo{}}
	app := testAppWithCampusOperatorRepo(spy)
	req := authenticatedRequest(http.MethodPatch, "/v1/unit-types/550e8400-e29b-41d4-a716-446655440020", `{"description":"Updated"}`, true)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	assertErrorCodeResponse(t, rr, http.StatusPreconditionRequired, "precondition_required")
	if spy.updateUnitTypeCalls != 0 {
		t.Fatalf("update unit type calls = %d, want 0", spy.updateUnitTypeCalls)
	}
}

func TestUpdatePropertyUnitTypeReturnsNewVersionAndETag(t *testing.T) {
	app := authenticatedTestApp(domain.ID("550e8400-e29b-41d4-a716-446655440001"), &fakeAgentApplicationStore{access: domain.EffectiveAccess{
		Roles:             []string{"campus_operator"},
		CampusOperatorIDs: []domain.ID{"550e8400-e29b-41d4-a716-446655440002"},
	}})
	app.propertyRepo = &stubPropertyRepo{}
	req := authenticatedRequest(http.MethodPatch, "/v1/unit-types/550e8400-e29b-41d4-a716-446655440020", `{"description":"Updated unit"}`, true)
	req.Header.Set("If-Match", `"property-unit-type-550e8400-e29b-41d4-a716-446655440020-1"`)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
	}
	if got := rr.Header().Get("ETag"); got != `"property-unit-type-550e8400-e29b-41d4-a716-446655440020-2"` {
		t.Fatalf("ETag = %q, want updated unit type ETag", got)
	}
	var body struct {
		UnitType struct {
			Description string `json:"description"`
			Version     int    `json:"version"`
		} `json:"unit_type"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if body.UnitType.Description != "Updated unit" {
		t.Fatalf("description = %q, want Updated unit", body.UnitType.Description)
	}
	if body.UnitType.Version != 2 {
		t.Fatalf("version = %d, want 2", body.UnitType.Version)
	}
}

func TestUpdatePropertyUnitTypeRejectsInvalidRequestsWithoutMutation(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		body     string
		ifMatch  string
		wantCode string
	}{
		{name: "malformed id", path: "/v1/unit-types/not-a-uuid", body: `{"description":"Updated"}`, wantCode: "validation_failed"},
		{name: "invalid body", path: "/v1/unit-types/550e8400-e29b-41d4-a716-446655440020", body: `{`, ifMatch: `"property-unit-type-550e8400-e29b-41d4-a716-446655440020-1"`, wantCode: "bad_request"},
		{name: "missing if-match", path: "/v1/unit-types/550e8400-e29b-41d4-a716-446655440020", body: `{"description":"Updated"}`, wantCode: "precondition_required"},
		{name: "malformed if-match", path: "/v1/unit-types/550e8400-e29b-41d4-a716-446655440020", body: `{"description":"Updated"}`, ifMatch: "not-an-etag", wantCode: "precondition_failed"},
		{name: "non-matching if-match", path: "/v1/unit-types/550e8400-e29b-41d4-a716-446655440020", body: `{"description":"Updated"}`, ifMatch: `"property-unit-type-550e8400-e29b-41d4-a716-446655440021-1"`, wantCode: "precondition_failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spy := &spyPropertyRepo{stubPropertyRepo: &stubPropertyRepo{}}
			app := testAppWithCampusOperatorRepo(spy)
			req := authenticatedRequest(http.MethodPatch, tt.path, tt.body, true)
			if tt.ifMatch != "" {
				req.Header.Set("If-Match", tt.ifMatch)
			}
			rr := httptest.NewRecorder()

			app.routes().ServeHTTP(rr, req)

			wantStatus := http.StatusUnprocessableEntity
			if tt.wantCode == "bad_request" {
				wantStatus = http.StatusBadRequest
			} else if tt.wantCode == "precondition_required" {
				wantStatus = http.StatusPreconditionRequired
			} else if tt.wantCode == "precondition_failed" {
				wantStatus = http.StatusPreconditionFailed
			}
			assertErrorCodeResponse(t, rr, wantStatus, tt.wantCode)
			if spy.updateUnitTypeCalls != 0 {
				t.Fatalf("update unit type calls = %d, want 0", spy.updateUnitTypeCalls)
			}
		})
	}
}

func TestUpdatePropertyUnitTypeRequiresAuthenticationWithoutMutation(t *testing.T) {
	spy := &spyPropertyRepo{stubPropertyRepo: &stubPropertyRepo{}}
	app := testApp()
	app.propertyRepo = spy
	req := httptest.NewRequest(http.MethodPatch, "/v1/unit-types/550e8400-e29b-41d4-a716-446655440020", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	assertErrorCodeResponse(t, rr, http.StatusUnauthorized, "unauthenticated")
	if spy.updateUnitTypeCalls != 0 {
		t.Fatalf("update unit type calls = %d, want 0", spy.updateUnitTypeCalls)
	}
}

func TestUpdatePropertyUnitTypeMapsRepositoryErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		status   int
		wantCode string
	}{
		{name: "not found", err: repo.ErrNotFound, status: http.StatusNotFound, wantCode: "not_found"},
		{name: "stale update", err: repo.ErrStaleUpdate, status: http.StatusPreconditionFailed, wantCode: "precondition_failed"},
		{name: "campus forbidden", err: repo.ErrCampusForbidden, status: http.StatusForbidden, wantCode: "forbidden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spy := &spyPropertyRepo{stubPropertyRepo: &stubPropertyRepo{}, updateUnitTypeErr: tt.err}
			app := testAppWithCampusOperatorRepo(spy)
			req := authenticatedRequest(http.MethodPatch, "/v1/unit-types/550e8400-e29b-41d4-a716-446655440020", `{"description":"Updated"}`, true)
			req.Header.Set("If-Match", `"property-unit-type-550e8400-e29b-41d4-a716-446655440020-1"`)
			rr := httptest.NewRecorder()

			app.routes().ServeHTTP(rr, req)

			assertErrorCodeResponse(t, rr, tt.status, tt.wantCode)
			if spy.updateUnitTypeCalls != 1 {
				t.Fatalf("update unit type calls = %d, want 1", spy.updateUnitTypeCalls)
			}
		})
	}
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
