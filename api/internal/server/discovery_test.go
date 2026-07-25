package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ddddami/laivan/internal/domain"
	"github.com/ddddami/laivan/internal/repo"
)

func TestDiscoveryReturnsResults(t *testing.T) {
	app := testAppWithRepo()
	req := httptest.NewRequest(http.MethodGet, "/v1/discovery?campus_id=550e8400-e29b-41d4-a716-446655440002&category=self_contained", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
	}

	var body struct {
		Results []struct {
			Property struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Area string `json:"area"`
			} `json:"property"`
			UnitType struct {
				Category     string `json:"category"`
				Name         string `json:"name"`
				BedroomCount int    `json:"bedroom_count"`
			} `json:"unit_type"`
			Pricing struct {
				LowestPriceNaira int `json:"lowest_price_naira"`
			} `json:"pricing"`
			OfferSummary struct {
				AvailableOfferCount int `json:"available_offer_count"`
			} `json:"offer_summary"`
		} `json:"results"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if len(body.Results) != 1 {
		t.Fatalf("results length = %d, want 1", len(body.Results))
	}
	if body.Results[0].Property.Name != "Alice Lodge" {
		t.Fatalf("property name = %q, want Alice Lodge", body.Results[0].Property.Name)
	}
	if body.Results[0].UnitType.Category != "self_contained" {
		t.Fatalf("category = %q, want self_contained", body.Results[0].UnitType.Category)
	}
	if body.Results[0].UnitType.Name != "Self-contained" {
		t.Fatalf("unit type name = %q, want Self-contained", body.Results[0].UnitType.Name)
	}
	if body.Results[0].Pricing.LowestPriceNaira != 350000 {
		t.Fatalf("lowest_price_naira = %d, want 350000", body.Results[0].Pricing.LowestPriceNaira)
	}
	if body.Results[0].OfferSummary.AvailableOfferCount != 2 {
		t.Fatalf("available_offer_count = %d, want 2", body.Results[0].OfferSummary.AvailableOfferCount)
	}
}

func TestDiscoveryValidationErrors(t *testing.T) {
	app := testAppWithRepo()
	req := httptest.NewRequest(http.MethodGet, "/v1/discovery?campus_id=bad-uuid&category=self_contained&has_parlour=yes", nil)
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
	if body.Error.Fields["has_parlour"] == "" {
		t.Fatal("has_parlour validation error missing")
	}
}

func TestDiscoveryDefaultsToAvailableRecommendedResults(t *testing.T) {
	app := testApp()
	store := &spyPropertyRepo{stub: &stubPropertyRepo{}}
	app.propertyRepo = store
	req := httptest.NewRequest(http.MethodGet, "/v1/discovery?campus_id=550e8400-e29b-41d4-a716-446655440002", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
	}
	if store.discoveryFilter.Availability != repo.DiscoveryAvailabilityAvailable {
		t.Fatalf("availability = %q, want %q", store.discoveryFilter.Availability, repo.DiscoveryAvailabilityAvailable)
	}
	if store.discoveryFilter.Filters.Sort != "recommended" {
		t.Fatalf("sort = %q, want recommended", store.discoveryFilter.Filters.Sort)
	}
}

func TestDiscoveryRejectsInvalidAvailability(t *testing.T) {
	app := testAppWithRepo()
	req := httptest.NewRequest(http.MethodGet, "/v1/discovery?campus_id=550e8400-e29b-41d4-a716-446655440002&availability=stale", nil)
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
	if body.Error.Fields["availability"] == "" {
		t.Fatal("availability validation error missing")
	}
}

func TestDiscoveryAcceptsAllAvailability(t *testing.T) {
	app := testApp()
	store := &spyPropertyRepo{stub: &stubPropertyRepo{}}
	app.propertyRepo = store
	req := httptest.NewRequest(http.MethodGet, "/v1/discovery?campus_id=550e8400-e29b-41d4-a716-446655440002&availability=all", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
	}
	if store.discoveryFilter.Availability != repo.DiscoveryAvailabilityAll {
		t.Fatalf("availability = %q, want %q", store.discoveryFilter.Availability, repo.DiscoveryAvailabilityAll)
	}
}

func TestDiscoveryResultWithoutAvailableOfferHasNullPrice(t *testing.T) {
	app := testApp()
	response := app.discoveryResultResponse(domain.DiscoveryResult{
		AvailableOfferCount: 0,
	})

	pricing, ok := response["pricing"].(map[string]any)
	if !ok {
		t.Fatalf("pricing = %#v, want response object", response["pricing"])
	}
	if pricing["lowest_price_naira"] != nil {
		t.Fatalf("lowest_price_naira = %#v, want nil", pricing["lowest_price_naira"])
	}
}
