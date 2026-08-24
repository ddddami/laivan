package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ddddami/laivan/internal/domain"
)

func TestCreatePropertyReturnsProperty(t *testing.T) {
	app := testAppWithRepo()
	body := `{"campus_id":"550e8400-e29b-41d4-a716-446655440000","name":"Alice Lodge","area":"Obanla","landmark":"South Gate","description":"A nice lodge"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/properties", strings.NewReader(body))
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusCreated)
	}

	var bodyDecoded struct {
		Property struct {
			ID        string `json:"id"`
			CampusID  string `json:"campus_id"`
			Name      string `json:"name"`
			Area      string `json:"area"`
			Landmark  string `json:"landmark"`
			CreatedAt string `json:"created_at"`
		} `json:"property"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&bodyDecoded); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if bodyDecoded.Property.ID == "" {
		t.Fatal("property ID is empty")
	}
	if bodyDecoded.Property.Name != "Alice Lodge" {
		t.Fatalf("property name = %q, want Alice Lodge", bodyDecoded.Property.Name)
	}
	if bodyDecoded.Property.Area != "Obanla" {
		t.Fatalf("property area = %q, want Obanla", bodyDecoded.Property.Area)
	}
	if bodyDecoded.Property.CampusID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("campus ID = %q, want 550e8400-e29b-41d4-a716-446655440000", bodyDecoded.Property.CampusID)
	}
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
				ID    string `json:"id"`
				Name  string `json:"name"`
				Media []struct {
					URL     string `json:"url"`
					Caption string `json:"caption"`
				} `json:"media"`
				AgentOffers []struct {
					Title      string `json:"title"`
					PriceNaira int    `json:"price_naira"`
					Agent      struct {
						ID          string `json:"id"`
						DisplayName string `json:"display_name"`
					} `json:"agent"`
					Media []struct {
						URL string `json:"url"`
					} `json:"media"`
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
	if len(body.Property.UnitTypes[0].Media) != 1 {
		t.Fatalf("unit media length = %d, want 1", len(body.Property.UnitTypes[0].Media))
	}
	if body.Property.UnitTypes[0].Media[0].URL != "https://media.example.test/unit.jpg" {
		t.Fatalf("unit media URL = %q, want https://media.example.test/unit.jpg", body.Property.UnitTypes[0].Media[0].URL)
	}
	if body.Property.UnitTypes[0].Media[0].Caption != "Unit media" {
		t.Fatalf("unit media caption = %q, want Unit media", body.Property.UnitTypes[0].Media[0].Caption)
	}
	if len(body.Property.UnitTypes[0].AgentOffers) != 1 {
		t.Fatalf("agent offers length = %d, want 1", len(body.Property.UnitTypes[0].AgentOffers))
	}
	if body.Property.UnitTypes[0].AgentOffers[0].PriceNaira != 350000 {
		t.Fatalf("price_naira = %d, want 350000", body.Property.UnitTypes[0].AgentOffers[0].PriceNaira)
	}
	if body.Property.UnitTypes[0].AgentOffers[0].Agent.ID != "550e8400-e29b-41d4-a716-446655440040" {
		t.Fatalf("agent ID = %q, want 550e8400-e29b-41d4-a716-446655440040", body.Property.UnitTypes[0].AgentOffers[0].Agent.ID)
	}
	if body.Property.UnitTypes[0].AgentOffers[0].Agent.DisplayName != "Bisi Housing Connect" {
		t.Fatalf("agent display name = %q, want Bisi Housing Connect", body.Property.UnitTypes[0].AgentOffers[0].Agent.DisplayName)
	}
	if len(body.Property.UnitTypes[0].AgentOffers[0].Media) != 1 {
		t.Fatalf("offer media length = %d, want 1", len(body.Property.UnitTypes[0].AgentOffers[0].Media))
	}
	if body.Property.UnitTypes[0].AgentOffers[0].Media[0].URL != "https://media.example.test/offer.jpg" {
		t.Fatalf("offer media URL = %q, want https://media.example.test/offer.jpg", body.Property.UnitTypes[0].AgentOffers[0].Media[0].URL)
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

func TestThumbnailURLNilMediaURLs(t *testing.T) {
	app := testApp()
	if got := app.thumbnailURL("https://example.com/img.jpg"); got != "https://example.com/img.jpg" {
		t.Fatalf("thumbnailURL with nil mediaURLs = %v, want %q", got, "https://example.com/img.jpg")
	}
}

func TestValidUnitCategory(t *testing.T) {
	categories := []struct {
		value string
		valid bool
	}{
		{"single_room", true},
		{"self_contained", true},
		{"room_and_parlour", true},
		{"one_bedroom_flat", true},
		{"two_bedroom_flat", true},
		{"three_bedroom_flat", true},
		{"other", true},
		{"invalid", false},
		{"", false},
	}
	for _, tc := range categories {
		t.Run(tc.value, func(t *testing.T) {
			if got := validUnitCategory(tc.value); got != tc.valid {
				t.Fatalf("validUnitCategory(%q) = %v, want %v", tc.value, got, tc.valid)
			}
		})
	}
}

func TestValidBathroomType(t *testing.T) {
	values := []struct {
		value string
		valid bool
	}{
		{"private", true},
		{"shared", true},
		{"unknown", true},
		{"invalid", false},
		{"", false},
	}
	for _, tc := range values {
		t.Run(tc.value, func(t *testing.T) {
			if got := validBathroomType(tc.value); got != tc.valid {
				t.Fatalf("validBathroomType(%q) = %v, want %v", tc.value, got, tc.valid)
			}
		})
	}
}

func TestValidKitchenType(t *testing.T) {
	values := []struct {
		value string
		valid bool
	}{
		{"private", true},
		{"shared", true},
		{"none", true},
		{"unknown", true},
		{"invalid", false},
		{"", false},
	}
	for _, tc := range values {
		t.Run(tc.value, func(t *testing.T) {
			if got := validKitchenType(tc.value); got != tc.valid {
				t.Fatalf("validKitchenType(%q) = %v, want %v", tc.value, got, tc.valid)
			}
		})
	}
}

func TestUnitCategoryDisplayName(t *testing.T) {
	tests := []struct {
		category domain.UnitCategory
		want     string
	}{
		{domain.UnitCategorySingleRoom, "Single Room"},
		{domain.UnitCategorySelfContained, "Self-contained"},
		{domain.UnitCategoryRoomAndParlour, "Room and Parlour"},
		{domain.UnitCategoryOneBedroomFlat, "One-bedroom Flat"},
		{domain.UnitCategoryTwoBedroomFlat, "Two-bedroom Flat"},
		{domain.UnitCategoryThreeBedroomFlat, "Three-bedroom Flat"},
		{domain.UnitCategoryOther, "Other"},
	}
	for _, tt := range tests {
		t.Run(string(tt.category), func(t *testing.T) {
			if got := unitCategoryDisplayName(tt.category); got != tt.want {
				t.Fatalf("unitCategoryDisplayName(%q) = %q, want %q", tt.category, got, tt.want)
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
