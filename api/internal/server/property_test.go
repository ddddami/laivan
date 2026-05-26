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

func (s *stubPropertyRepo) GetWithDetails(ctx context.Context, id domain.ID) (domain.PropertyDetail, error) {
	if string(id) != "550e8400-e29b-41d4-a716-446655440000" {
		return domain.PropertyDetail{}, repo.ErrNotFound
	}

	return domain.PropertyDetail{
		Property: domain.Property{
			ID:          id,
			CampusID:    domain.ID("550e8400-e29b-41d4-a716-446655440002"),
			Name:        "Alice Lodge",
			Location:    domain.ApproxLocation{Area: "Obanla", Landmark: "Near South Gate"},
			Description: "Gated lodge with multiple room categories near campus.",
			Timestamps: domain.Timestamps{
				CreatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
			},
		},
		RoomTypes: []domain.RoomTypeDetail{
			{
				RoomType: domain.RoomType{
					ID:          domain.ID("550e8400-e29b-41d4-a716-446655440020"),
					PropertyID:  id,
					Name:        "Self-contained",
					Description: "Private room with bathroom and kitchenette.",
					Timestamps: domain.Timestamps{
						CreatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
					},
				},
				AgentOffers: []domain.AgentOffer{
					{
						ID:          domain.ID("550e8400-e29b-41d4-a716-446655440030"),
						RoomTypeID:  domain.ID("550e8400-e29b-41d4-a716-446655440020"),
						AgentID:     domain.ID("550e8400-e29b-41d4-a716-446655440040"),
						Title:       "Fresh self-contained room",
						Description: "Recently painted room with private bathroom.",
						Price:       domain.Money{AmountKobo: 35000000},
						Status:      domain.AgentOfferStatusAvailable,
						Timestamps: domain.Timestamps{
							CreatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
							UpdatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
						},
					},
				},
			},
		},
	}, nil
}

func (s *stubPropertyRepo) List(ctx context.Context, filter repo.PropertyListFilter) ([]domain.Property, error) {
	return []domain.Property{
		{
			ID:          domain.ID("550e8400-e29b-41d4-a716-446655440010"),
			CampusID:    filter.CampusID,
			Name:        "Alice Lodge",
			Location:    domain.ApproxLocation{Area: "Obanla", Landmark: "Near South Gate"},
			Description: "Gated lodge with multiple room categories near campus.",
			Timestamps: domain.Timestamps{
				CreatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
			},
		},
	}, nil
}

func (s *stubPropertyRepo) ListWithSummary(ctx context.Context, filter repo.PropertyListFilter) ([]domain.PropertySummary, error) {
	return []domain.PropertySummary{
		{
			Property: domain.Property{
				ID:          domain.ID("550e8400-e29b-41d4-a716-446655440010"),
				CampusID:    filter.CampusID,
				Name:        "Alice Lodge",
				Location:    domain.ApproxLocation{Area: "Obanla", Landmark: "Near South Gate"},
				Description: "Gated lodge with multiple room categories near campus.",
				Timestamps: domain.Timestamps{
					CreatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
				},
			},
			RoomTypeCount:       3,
			AvailableOfferCount: 5,
			LowestPriceKobo:     25000000,
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

func (s *stubPropertyRepo) CreateAgentOffer(ctx context.Context, offer domain.AgentOffer) (domain.AgentOffer, error) {
	if string(offer.RoomTypeID) != "550e8400-e29b-41d4-a716-446655440020" {
		return domain.AgentOffer{}, repo.ErrNotFound
	}

	offer.ID = domain.ID("550e8400-e29b-41d4-a716-446655440030")
	offer.CreatedAt = time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC)
	offer.UpdatedAt = time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC)
	return offer, nil
}

func (s *stubPropertyRepo) ListAgentOffers(ctx context.Context, roomTypeID domain.ID) ([]domain.AgentOffer, error) {
	if string(roomTypeID) != "550e8400-e29b-41d4-a716-446655440020" {
		return nil, repo.ErrNotFound
	}

	return []domain.AgentOffer{
		{
			ID:          domain.ID("550e8400-e29b-41d4-a716-446655440030"),
			RoomTypeID:  roomTypeID,
			AgentID:     domain.ID("550e8400-e29b-41d4-a716-446655440040"),
			Title:       "Fresh self-contained room",
			Description: "Recently painted room with private bathroom.",
			Price:       domain.Money{AmountKobo: 35000000},
			Status:      domain.AgentOfferStatusAvailable,
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
			ID                  string `json:"id"`
			CampusID            string `json:"campus_id"`
			Name                string `json:"name"`
			Area                string `json:"area"`
			RoomTypeCount       int32  `json:"room_type_count"`
			AvailableOfferCount int32  `json:"available_offer_count"`
			LowestPriceKobo     int32  `json:"lowest_price_kobo"`
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
	if body.Properties[0].RoomTypeCount != 3 {
		t.Fatalf("room_type_count = %d, want 3", body.Properties[0].RoomTypeCount)
	}
	if body.Properties[0].AvailableOfferCount != 5 {
		t.Fatalf("available_offer_count = %d, want 5", body.Properties[0].AvailableOfferCount)
	}
	if body.Properties[0].LowestPriceKobo != 25000000 {
		t.Fatalf("lowest_price_kobo = %d, want 25000000", body.Properties[0].LowestPriceKobo)
	}
}

func TestGetPropertyReturnsPropertyWithDetails(t *testing.T) {
	app := testAppWithRepo()
	req := httptest.NewRequest(http.MethodGet, "/v1/properties/550e8400-e29b-41d4-a716-446655440000", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
	}

	var body struct {
		Property struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			RoomTypes []struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				AgentOffers []struct {
					Title     string `json:"title"`
					PriceKobo int32  `json:"price_kobo"`
				} `json:"agent_offers"`
			} `json:"room_types"`
		} `json:"property"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if body.Property.Name != "Alice Lodge" {
		t.Fatalf("property name = %q, want Alice Lodge", body.Property.Name)
	}
	if len(body.Property.RoomTypes) != 1 {
		t.Fatalf("room types length = %d, want 1", len(body.Property.RoomTypes))
	}
	if body.Property.RoomTypes[0].Name != "Self-contained" {
		t.Fatalf("room type name = %q, want Self-contained", body.Property.RoomTypes[0].Name)
	}
	if len(body.Property.RoomTypes[0].AgentOffers) != 1 {
		t.Fatalf("agent offers length = %d, want 1", len(body.Property.RoomTypes[0].AgentOffers))
	}
	if body.Property.RoomTypes[0].AgentOffers[0].PriceKobo != 35000000 {
		t.Fatalf("price_kobo = %d, want 35000000", body.Property.RoomTypes[0].AgentOffers[0].PriceKobo)
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

func TestCreateAgentOfferValidationErrors(t *testing.T) {
	app := testAppWithRepo()
	body := `{"agent_id":"","title":"","price_kobo":0,"status":"gone"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/room-types/not-a-uuid/agent-offers", strings.NewReader(body))
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

	for _, field := range []string{"id", "agent_id", "title", "price_kobo", "status"} {
		if bodyDecoded.Error.Fields[field] == "" {
			t.Fatalf("%s validation error missing", field)
		}
	}
}

func TestCreateAgentOfferReturnsAgentOffer(t *testing.T) {
	app := testAppWithRepo()
	body := `{"agent_id":"550e8400-e29b-41d4-a716-446655440040","title":"Fresh self-contained room","description":"Recently painted room with private bathroom.","price_kobo":35000000}`
	req := httptest.NewRequest(http.MethodPost, "/v1/room-types/550e8400-e29b-41d4-a716-446655440020/agent-offers", strings.NewReader(body))
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusCreated)
	}

	var bodyDecoded struct {
		AgentOffer struct {
			RoomTypeID string `json:"room_type_id"`
			AgentID    string `json:"agent_id"`
			Title      string `json:"title"`
			PriceKobo  int32  `json:"price_kobo"`
			Status     string `json:"status"`
		} `json:"agent_offer"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&bodyDecoded); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if bodyDecoded.AgentOffer.Title != "Fresh self-contained room" {
		t.Fatalf("title = %q, want Fresh self-contained room", bodyDecoded.AgentOffer.Title)
	}
	if bodyDecoded.AgentOffer.PriceKobo != 35000000 {
		t.Fatalf("price_kobo = %d, want 35000000", bodyDecoded.AgentOffer.PriceKobo)
	}
	if bodyDecoded.AgentOffer.Status != "available" {
		t.Fatalf("status = %q, want available", bodyDecoded.AgentOffer.Status)
	}
}

func TestCreateAgentOfferRoomTypeNotFound(t *testing.T) {
	app := testAppWithRepo()
	body := `{"agent_id":"550e8400-e29b-41d4-a716-446655440040","title":"Fresh self-contained room","price_kobo":35000000}`
	req := httptest.NewRequest(http.MethodPost, "/v1/room-types/11111111-1111-1111-1111-111111111111/agent-offers", strings.NewReader(body))
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusNotFound, "not_found", "The requested resource could not be found")
}

func TestListAgentOffersReturnsAgentOffers(t *testing.T) {
	app := testAppWithRepo()
	req := httptest.NewRequest(http.MethodGet, "/v1/room-types/550e8400-e29b-41d4-a716-446655440020/agent-offers", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
	}

	var body struct {
		AgentOffers []struct {
			RoomTypeID string `json:"room_type_id"`
			Title      string `json:"title"`
			PriceKobo  int32  `json:"price_kobo"`
		} `json:"agent_offers"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if len(body.AgentOffers) != 1 {
		t.Fatalf("agent offers length = %d, want 1", len(body.AgentOffers))
	}
	if body.AgentOffers[0].Title != "Fresh self-contained room" {
		t.Fatalf("title = %q, want Fresh self-contained room", body.AgentOffers[0].Title)
	}
}

func TestListAgentOffersRoomTypeNotFound(t *testing.T) {
	app := testAppWithRepo()
	req := httptest.NewRequest(http.MethodGet, "/v1/room-types/11111111-1111-1111-1111-111111111111/agent-offers", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusNotFound, "not_found", "The requested resource could not be found")
}
