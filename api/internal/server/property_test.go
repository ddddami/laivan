package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ddddami/laivan/internal/domain"
	"github.com/ddddami/laivan/internal/repo"
)

type stubPropertyRepo struct{}

func (s *stubPropertyRepo) Create(ctx context.Context, property domain.Property) (domain.Property, error) {
	property.ID = domain.ID("550e8400-e29b-41d4-a716-446655440001")
	property.CreatedAt = time.Now()
	property.UpdatedAt = time.Now()
	return property, nil
}

func (s *stubPropertyRepo) Get(ctx context.Context, id domain.ID) (domain.Property, error) {
	if string(id) == "550e8400-e29b-41d4-a716-446655440000" {
		return domain.Property{
			ID:       id,
			CampusID: domain.ID("550e8400-e29b-41d4-a716-446655440002"),
			Name:     "Alice Lodge",
			Location: domain.ApproxLocation{Area: "Obanla"},
		}, nil
	}
	return domain.Property{}, repo.ErrNotFound
}

func (s *stubPropertyRepo) List(ctx context.Context, filter repo.PropertyListFilter) ([]domain.Property, error) {
	return []domain.Property{
		{
			ID:          domain.ID("550e8400-e29b-41d4-a716-446655440010"),
			CampusID:    filter.CampusID,
			Name:        "Alice Lodge",
			Location:    domain.ApproxLocation{Area: "Obanla", Landmark: "Near South Gate"},
			Description: "Self-contained rooms close to FUTA.",
			Timestamps: domain.Timestamps{
				CreatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
			},
		},
	}, nil
}

func (s *stubPropertyRepo) CreateRoomType(ctx context.Context, roomType domain.RoomType) (domain.RoomType, error) {
	if string(roomType.PropertyID) != "550e8400-e29b-41d4-a716-446655440000" {
		return domain.RoomType{}, repo.ErrNotFound
	}

	roomType.ID = domain.ID("550e8400-e29b-41d4-a716-446655440020")
	roomType.CreatedAt = time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC)
	roomType.UpdatedAt = time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC)
	return roomType, nil
}

func (s *stubPropertyRepo) ListRoomTypes(ctx context.Context, propertyID domain.ID) ([]domain.RoomType, error) {
	if string(propertyID) != "550e8400-e29b-41d4-a716-446655440000" {
		return nil, repo.ErrNotFound
	}

	return []domain.RoomType{
		{
			ID:          domain.ID("550e8400-e29b-41d4-a716-446655440020"),
			PropertyID:  propertyID,
			Name:        "Self-contained",
			Description: "Private room with bathroom and kitchenette.",
			Timestamps: domain.Timestamps{
				CreatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
			},
		},
	}, nil
}

func testAppWithRepo() *app {
	a := testApp()
	a.propertyRepo = &stubPropertyRepo{}
	return a
}

func TestCreatePropertyUnknownField(t *testing.T) {
	app := testAppWithRepo()
	body := `{"campus_id":"550e8400-e29b-41d4-a716-446655440000","name":"Alice Lodge","area":"Obanla","unknown":"field"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/properties", strings.NewReader(body))
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusBadRequest, "bad_request", "request body contains unknown field \"unknown\"")
}

func TestCreatePropertyValidationErrors(t *testing.T) {
	app := testAppWithRepo()
	body := `{"campus_id":"","name":"","area":""}`
	req := httptest.NewRequest(http.MethodPost, "/v1/properties", strings.NewReader(body))
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}

	var bodyDecoded struct {
		Error struct {
			Code    string            `json:"code"`
			Message string            `json:"message"`
			Fields  map[string]string `json:"fields"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&bodyDecoded); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if bodyDecoded.Error.Code != "validation_failed" {
		t.Fatalf("code = %q, want validation_failed", bodyDecoded.Error.Code)
	}

	if bodyDecoded.Error.Fields["campus_id"] == "" {
		t.Fatal("campus_id validation error missing")
	}

	if bodyDecoded.Error.Fields["name"] == "" {
		t.Fatal("name validation error missing")
	}

	if bodyDecoded.Error.Fields["area"] == "" {
		t.Fatal("area validation error missing")
	}
}

func TestListPropertiesValidationErrors(t *testing.T) {
	app := testAppWithRepo()
	req := httptest.NewRequest(http.MethodGet, "/v1/properties?limit=0", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}

	var body struct {
		Error struct {
			Code   string            `json:"code"`
			Fields map[string]string `json:"fields"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if body.Error.Code != "validation_failed" {
		t.Fatalf("code = %q, want validation_failed", body.Error.Code)
	}
	if body.Error.Fields["campus_id"] == "" {
		t.Fatal("campus_id validation error missing")
	}
	if body.Error.Fields["limit"] == "" {
		t.Fatal("limit validation error missing")
	}
}

func TestListPropertiesInvalidLimit(t *testing.T) {
	app := testAppWithRepo()
	req := httptest.NewRequest(http.MethodGet, "/v1/properties?campus_id=550e8400-e29b-41d4-a716-446655440002&limit=many", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}

	var body struct {
		Error struct {
			Fields map[string]string `json:"fields"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if body.Error.Fields["limit"] != "Limit must be an integer" {
		t.Fatalf("limit error = %q, want Limit must be an integer", body.Error.Fields["limit"])
	}
}

func TestListPropertiesReturnsProperties(t *testing.T) {
	app := testAppWithRepo()
	req := httptest.NewRequest(http.MethodGet, "/v1/properties?campus_id=550e8400-e29b-41d4-a716-446655440002", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
	}

	var body struct {
		Properties []struct {
			ID       string `json:"id"`
			CampusID string `json:"campus_id"`
			Name     string `json:"name"`
			Area     string `json:"area"`
		} `json:"properties"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if len(body.Properties) != 1 {
		t.Fatalf("properties length = %d, want 1", len(body.Properties))
	}
	if body.Properties[0].Name != "Alice Lodge" {
		t.Fatalf("property name = %q, want Alice Lodge", body.Properties[0].Name)
	}
	if body.Properties[0].CampusID != "550e8400-e29b-41d4-a716-446655440002" {
		t.Fatalf("campus ID = %q, want 550e8400-e29b-41d4-a716-446655440002", body.Properties[0].CampusID)
	}
}

func TestGetPropertyInvalidID(t *testing.T) {
	app := testAppWithRepo()
	req := httptest.NewRequest(http.MethodGet, "/v1/properties/not-a-uuid", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

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

	if body.Error.Fields["id"] == "" {
		t.Fatal("id validation error missing")
	}
}

func TestGetPropertyNotFound(t *testing.T) {
	app := testAppWithRepo()
	req := httptest.NewRequest(http.MethodGet, "/v1/properties/11111111-1111-1111-1111-111111111111", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusNotFound, "not_found", "The requested resource could not be found")
}

func TestCreateRoomTypeValidationErrors(t *testing.T) {
	app := testAppWithRepo()
	body := `{"name":""}`
	req := httptest.NewRequest(http.MethodPost, "/v1/properties/not-a-uuid/room-types", strings.NewReader(body))
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
	if bodyDecoded.Error.Fields["name"] == "" {
		t.Fatal("name validation error missing")
	}
}

func TestCreateRoomTypeReturnsRoomType(t *testing.T) {
	app := testAppWithRepo()
	body := `{"name":"Self-contained","description":"Private room with bathroom and kitchenette."}`
	req := httptest.NewRequest(http.MethodPost, "/v1/properties/550e8400-e29b-41d4-a716-446655440000/room-types", strings.NewReader(body))
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusCreated)
	}

	var bodyDecoded struct {
		RoomType struct {
			ID         string `json:"id"`
			PropertyID string `json:"property_id"`
			Name       string `json:"name"`
		} `json:"room_type"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&bodyDecoded); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if bodyDecoded.RoomType.Name != "Self-contained" {
		t.Fatalf("room type name = %q, want Self-contained", bodyDecoded.RoomType.Name)
	}
	if bodyDecoded.RoomType.PropertyID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("property ID = %q, want 550e8400-e29b-41d4-a716-446655440000", bodyDecoded.RoomType.PropertyID)
	}
}

func TestCreateRoomTypePropertyNotFound(t *testing.T) {
	app := testAppWithRepo()
	body := `{"name":"Self-contained"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/properties/11111111-1111-1111-1111-111111111111/room-types", strings.NewReader(body))
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusNotFound, "not_found", "The requested resource could not be found")
}

func TestListRoomTypesReturnsRoomTypes(t *testing.T) {
	app := testAppWithRepo()
	req := httptest.NewRequest(http.MethodGet, "/v1/properties/550e8400-e29b-41d4-a716-446655440000/room-types", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
	}

	var body struct {
		RoomTypes []struct {
			PropertyID string `json:"property_id"`
			Name       string `json:"name"`
		} `json:"room_types"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if len(body.RoomTypes) != 1 {
		t.Fatalf("room types length = %d, want 1", len(body.RoomTypes))
	}
	if body.RoomTypes[0].Name != "Self-contained" {
		t.Fatalf("room type name = %q, want Self-contained", body.RoomTypes[0].Name)
	}
}

func TestListRoomTypesPropertyNotFound(t *testing.T) {
	app := testAppWithRepo()
	req := httptest.NewRequest(http.MethodGet, "/v1/properties/11111111-1111-1111-1111-111111111111/room-types", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusNotFound, "not_found", "The requested resource could not be found")
}
