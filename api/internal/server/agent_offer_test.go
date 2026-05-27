package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateAgentOfferValidationErrors(t *testing.T) {
	app := testAppWithRepo()
	body := `{"agent_id":"","title":"","price_naira":0,"status":"gone"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/unit-types/not-a-uuid/agent-offers", strings.NewReader(body))
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

	for _, field := range []string{"id", "agent_id", "title", "price_naira", "status"} {
		if bodyDecoded.Error.Fields[field] == "" {
			t.Fatalf("%s validation error missing", field)
		}
	}
}

func TestCreateAgentOfferReturnsAgentOffer(t *testing.T) {
	app := testAppWithRepo()
	body := `{"agent_id":"550e8400-e29b-41d4-a716-446655440040","title":"Fresh self-contained room","description":"Recently painted room with private bathroom.","price_naira":350000}`
	req := httptest.NewRequest(http.MethodPost, "/v1/unit-types/550e8400-e29b-41d4-a716-446655440020/agent-offers", strings.NewReader(body))
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusCreated)
	}

	var bodyDecoded struct {
		AgentOffer struct {
			PropertyUnitTypeID string `json:"property_unit_type_id"`
			AgentID            string `json:"agent_id"`
			Title              string `json:"title"`
			PriceNaira         int    `json:"price_naira"`
			Status             string `json:"status"`
		} `json:"agent_offer"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&bodyDecoded); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if bodyDecoded.AgentOffer.Title != "Fresh self-contained room" {
		t.Fatalf("title = %q, want Fresh self-contained room", bodyDecoded.AgentOffer.Title)
	}
	if bodyDecoded.AgentOffer.PriceNaira != 350000 {
		t.Fatalf("price_naira = %d, want 350000", bodyDecoded.AgentOffer.PriceNaira)
	}
	if bodyDecoded.AgentOffer.Status != "available" {
		t.Fatalf("status = %q, want available", bodyDecoded.AgentOffer.Status)
	}
}

func TestCreateAgentOfferConvertsNairaToKobo(t *testing.T) {
	spy := &spyPropertyRepo{stub: &stubPropertyRepo{}}
	app := testApp()
	app.propertyRepo = spy

	body := `{"agent_id":"550e8400-e29b-41d4-a716-446655440040","title":"Fresh self-contained room","price_naira":350000}`
	req := httptest.NewRequest(http.MethodPost, "/v1/unit-types/550e8400-e29b-41d4-a716-446655440020/agent-offers", strings.NewReader(body))
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusCreated)
	}

	if spy.createdOffer.Price.AmountKobo != 35000000 {
		t.Fatalf("stored price = %d kobo, want 35000000", spy.createdOffer.Price.AmountKobo)
	}
}

func TestCreateAgentOfferUnitTypeNotFound(t *testing.T) {
	app := testAppWithRepo()
	body := `{"agent_id":"550e8400-e29b-41d4-a716-446655440040","title":"Fresh self-contained room","price_naira":350000}`
	req := httptest.NewRequest(http.MethodPost, "/v1/unit-types/11111111-1111-1111-1111-111111111111/agent-offers", strings.NewReader(body))
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusNotFound, "not_found", "The requested resource could not be found")
}

func TestListAgentOffersReturnsAgentOffers(t *testing.T) {
	app := testAppWithRepo()
	req := httptest.NewRequest(http.MethodGet, "/v1/unit-types/550e8400-e29b-41d4-a716-446655440020/agent-offers", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
	}

	var body struct {
		AgentOffers []struct {
			PropertyUnitTypeID string `json:"property_unit_type_id"`
			Title              string `json:"title"`
			PriceNaira         int    `json:"price_naira"`
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

func TestListAgentOffersUnitTypeNotFound(t *testing.T) {
	app := testAppWithRepo()
	req := httptest.NewRequest(http.MethodGet, "/v1/unit-types/11111111-1111-1111-1111-111111111111/agent-offers", nil)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusNotFound, "not_found", "The requested resource could not be found")
}
