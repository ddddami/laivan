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
		UnitTypes: []domain.PropertyUnitTypeDetail{
			{
				PropertyUnitType: domain.PropertyUnitType{
					ID:          domain.ID("550e8400-e29b-41d4-a716-446655440020"),
					PropertyID:  id,
					Category:    domain.UnitCategorySelfContained,
					Name:        "Self-contained",
					Description: "Private room with bathroom and kitchenette.",
					Structure: domain.UnitStructure{
						BedroomCount: intPointer(1),
						HasParlour:   boolPointer(false),
						BathroomType: "private",
						KitchenType:  "private",
					},
					Timestamps: domain.Timestamps{
						CreatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
					},
				},
				AgentOffers: []domain.AgentOffer{
					{
						ID:                 domain.ID("550e8400-e29b-41d4-a716-446655440030"),
						PropertyUnitTypeID: domain.ID("550e8400-e29b-41d4-a716-446655440020"),
						AgentID:            domain.ID("550e8400-e29b-41d4-a716-446655440040"),
						Title:              "Fresh self-contained room",
						Description:        "Recently painted room with private bathroom.",
						Price:              domain.Money{AmountKobo: 35000000},
						Status:             domain.AgentOfferStatusAvailable,
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

func (s *stubPropertyRepo) ListWithSummary(ctx context.Context, filter repo.PropertyListFilter) ([]domain.PropertySummary, int, error) {
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
			UnitTypeCount:       3,
			AvailableOfferCount: 5,
			LowestPrice:         domain.Money{AmountKobo: 25000000},
		},
	}, 1, nil
}

func (s *stubPropertyRepo) Discover(ctx context.Context, filter repo.DiscoveryFilter) ([]domain.DiscoveryResult, int, error) {
	return []domain.DiscoveryResult{
		{
			PropertyID:       domain.ID("550e8400-e29b-41d4-a716-446655440010"),
			PropertyName:     "Alice Lodge",
			PropertyArea:     "Obanla",
			PropertyLandmark: "Near South Gate",
			UnitTypeID:       domain.ID("550e8400-e29b-41d4-a716-446655440020"),
			UnitTypeCategory: domain.UnitCategorySelfContained,
			UnitTypeName:     "Self-contained",
			Structure: domain.UnitStructure{
				BedroomCount: intPointer(1),
				HasParlour:   boolPointer(false),
				BathroomType: "private",
				KitchenType:  "private",
			},
			LowestPrice:         domain.Money{AmountKobo: 35000000},
			AvailableOfferCount: 2,
			CreatedAt:           time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
		},
	}, 1, nil
}

func (s *stubPropertyRepo) CreatePropertyUnitType(ctx context.Context, unitType domain.PropertyUnitType) (domain.PropertyUnitType, error) {
	if string(unitType.PropertyID) != "550e8400-e29b-41d4-a716-446655440000" {
		return domain.PropertyUnitType{}, repo.ErrNotFound
	}

	unitType.ID = domain.ID("550e8400-e29b-41d4-a716-446655440020")
	unitType.CreatedAt = time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC)
	unitType.UpdatedAt = time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC)
	return unitType, nil
}

func (s *stubPropertyRepo) ListPropertyUnitTypes(ctx context.Context, propertyID domain.ID) ([]domain.PropertyUnitType, error) {
	if string(propertyID) != "550e8400-e29b-41d4-a716-446655440000" {
		return nil, repo.ErrNotFound
	}

	return []domain.PropertyUnitType{
		{
			ID:          domain.ID("550e8400-e29b-41d4-a716-446655440020"),
			PropertyID:  propertyID,
			Category:    domain.UnitCategorySelfContained,
			Name:        "Self-contained",
			Description: "Private room with bathroom and kitchenette.",
			Structure: domain.UnitStructure{
				BedroomCount: intPointer(1),
				HasParlour:   boolPointer(false),
				BathroomType: "private",
				KitchenType:  "private",
			},
			Timestamps: domain.Timestamps{
				CreatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
			},
		},
	}, nil
}

func (s *stubPropertyRepo) CreateAgentOffer(ctx context.Context, offer domain.AgentOffer) (domain.AgentOffer, error) {
	if string(offer.PropertyUnitTypeID) != "550e8400-e29b-41d4-a716-446655440020" {
		return domain.AgentOffer{}, repo.ErrNotFound
	}

	offer.ID = domain.ID("550e8400-e29b-41d4-a716-446655440030")
	offer.CreatedAt = time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC)
	offer.UpdatedAt = time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC)
	return offer, nil
}

func (s *stubPropertyRepo) ListAgentOffers(ctx context.Context, unitTypeID domain.ID) ([]domain.AgentOffer, error) {
	if string(unitTypeID) != "550e8400-e29b-41d4-a716-446655440020" {
		return nil, repo.ErrNotFound
	}

	return []domain.AgentOffer{
		{
			ID:                 domain.ID("550e8400-e29b-41d4-a716-446655440030"),
			PropertyUnitTypeID: domain.ID(unitTypeID),
			AgentID:            domain.ID("550e8400-e29b-41d4-a716-446655440040"),
			Title:              "Fresh self-contained room",
			Description:        "Recently painted room with private bathroom.",
			Price:              domain.Money{AmountKobo: 35000000},
			Status:             domain.AgentOfferStatusAvailable,
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
	req := httptest.NewRequest(http.MethodGet, "/v1/properties?page_size=0", nil)
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
	if body.Error.Fields["page_size"] == "" {
		t.Fatal("page_size validation error missing")
	}
}

func TestListPropertiesInvalidPageSize(t *testing.T) {
	app := testAppWithRepo()
	req := httptest.NewRequest(http.MethodGet, "/v1/properties?campus_id=550e8400-e29b-41d4-a716-446655440002&page_size=many", nil)
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
	if body.Error.Fields["page_size"] != "must be an integer value" {
		t.Fatalf("page_size error = %q, want must be an integer value", body.Error.Fields["page_size"])
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
			UnitTypeCount       int    `json:"unit_type_count"`
			AvailableOfferCount int    `json:"available_offer_count"`
			LowestPriceNaira    int    `json:"lowest_price_naira"`
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
	if body.Properties[0].UnitTypeCount != 3 {
		t.Fatalf("unit_type_count = %d, want 3", body.Properties[0].UnitTypeCount)
	}
	if body.Properties[0].AvailableOfferCount != 5 {
		t.Fatalf("available_offer_count = %d, want 5", body.Properties[0].AvailableOfferCount)
	}
	if body.Properties[0].LowestPriceNaira != 250000 {
		t.Fatalf("lowest_price_naira = %d, want 250000", body.Properties[0].LowestPriceNaira)
	}
}

func TestListPropertiesBadBooleanFilter(t *testing.T) {
	app := testAppWithRepo()
	req := httptest.NewRequest(http.MethodGet, "/v1/properties?campus_id=550e8400-e29b-41d4-a716-446655440002&has_offers=yes", nil)
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
	if body.Error.Fields["has_offers"] != "must be a boolean value" {
		t.Fatalf("has_offers error = %q, want must be a boolean value", body.Error.Fields["has_offers"])
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
			UnitTypes []struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				AgentOffers []struct {
					Title      string `json:"title"`
					PriceNaira int    `json:"price_naira"`
				} `json:"agent_offers"`
			} `json:"unit_types"`
		} `json:"property"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if body.Property.Name != "Alice Lodge" {
		t.Fatalf("property name = %q, want Alice Lodge", body.Property.Name)
	}
	if len(body.Property.UnitTypes) != 1 {
		t.Fatalf("unit types length = %d, want 1", len(body.Property.UnitTypes))
	}
	if body.Property.UnitTypes[0].Name != "Self-contained" {
		t.Fatalf("unit type name = %q, want Self-contained", body.Property.UnitTypes[0].Name)
	}
	if len(body.Property.UnitTypes[0].AgentOffers) != 1 {
		t.Fatalf("agent offers length = %d, want 1", len(body.Property.UnitTypes[0].AgentOffers))
	}
	if body.Property.UnitTypes[0].AgentOffers[0].PriceNaira != 350000 {
		t.Fatalf("price_naira = %d, want 350000", body.Property.UnitTypes[0].AgentOffers[0].PriceNaira)
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

type spyPropertyRepo struct {
	stub            *stubPropertyRepo
	createdUnitType domain.PropertyUnitType
	createdOffer    domain.AgentOffer
}

func (s *spyPropertyRepo) Create(ctx context.Context, property domain.Property) (domain.Property, error) {
	return s.stub.Create(ctx, property)
}

func (s *spyPropertyRepo) Get(ctx context.Context, id domain.ID) (domain.Property, error) {
	return s.stub.Get(ctx, id)
}

func (s *spyPropertyRepo) GetWithDetails(ctx context.Context, id domain.ID) (domain.PropertyDetail, error) {
	return s.stub.GetWithDetails(ctx, id)
}

func (s *spyPropertyRepo) ListWithSummary(ctx context.Context, filter repo.PropertyListFilter) ([]domain.PropertySummary, int, error) {
	return s.stub.ListWithSummary(ctx, filter)
}

func (s *spyPropertyRepo) Discover(ctx context.Context, filter repo.DiscoveryFilter) ([]domain.DiscoveryResult, int, error) {
	return s.stub.Discover(ctx, filter)
}

func (s *spyPropertyRepo) CreatePropertyUnitType(ctx context.Context, unitType domain.PropertyUnitType) (domain.PropertyUnitType, error) {
	s.createdUnitType = unitType
	return s.stub.CreatePropertyUnitType(ctx, unitType)
}

func (s *spyPropertyRepo) ListPropertyUnitTypes(ctx context.Context, propertyID domain.ID) ([]domain.PropertyUnitType, error) {
	return s.stub.ListPropertyUnitTypes(ctx, propertyID)
}

func (s *spyPropertyRepo) CreateAgentOffer(ctx context.Context, offer domain.AgentOffer) (domain.AgentOffer, error) {
	s.createdOffer = offer
	return s.stub.CreateAgentOffer(ctx, offer)
}

func (s *spyPropertyRepo) ListAgentOffers(ctx context.Context, unitTypeID domain.ID) ([]domain.AgentOffer, error) {
	return s.stub.ListAgentOffers(ctx, unitTypeID)
}
func TestUnitTypeDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		unitType domain.PropertyUnitType
		want     string
	}{
		{"custom name preserved", domain.PropertyUnitType{Name: "Premium Self-con", Category: domain.UnitCategorySelfContained}, "Premium Self-con"},
		{"empty name defaults to category", domain.PropertyUnitType{Name: "", Category: domain.UnitCategorySelfContained}, "Self-contained"},
		{"single room category", domain.PropertyUnitType{Name: "", Category: domain.UnitCategorySingleRoom}, "Single Room"},
		{"room and parlour category", domain.PropertyUnitType{Name: "", Category: domain.UnitCategoryRoomAndParlour}, "Room and Parlour"},
		{"other category", domain.PropertyUnitType{Name: "", Category: domain.UnitCategoryOther}, "Other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unitTypeDisplayName(tt.unitType); got != tt.want {
				t.Fatalf("unitTypeDisplayName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNairaConversion(t *testing.T) {
	m := domain.Money{AmountKobo: 35000000}
	if got := m.Naira(); got != 350000 {
		t.Fatalf("naira(35000000 kobo) = %d, want 350000", got)
	}
	m = domain.Money{AmountKobo: 0}
	if got := m.Naira(); got != 0 {
		t.Fatalf("naira(0 kobo) = %d, want 0", got)
	}
	if got := domain.Kobo(350000); got != 35000000 {
		t.Fatalf("kobo(350000 naira) = %d, want 35000000", got)
	}
	if got := domain.Kobo(0); got != 0 {
		t.Fatalf("kobo(0 naira) = %d, want 0", got)
	}
}

func intPointer(value int) *int {
	return &value
}

func boolPointer(value bool) *bool {
	return &value
}
