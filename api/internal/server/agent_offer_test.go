package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ddddami/laivan/internal/domain"
)

func TestCreateAgentOfferValidationErrors(t *testing.T) {
	app := testAppWithActiveAgentRepo()
	body := `{"title":"","price_naira":0,"status":"gone"}`
	req := authenticatedRequest(http.MethodPost, "/v1/unit-types/not-a-uuid/agent-offers", body, true)
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

	for _, field := range []string{"id", "title", "price_naira", "status"} {
		if bodyDecoded.Error.Fields[field] == "" {
			t.Fatalf("%s validation error missing", field)
		}
	}
}

func TestCreateAgentOfferReturnsAgentOffer(t *testing.T) {
	app := testAppWithActiveAgentRepo()
	body := `{"title":"Fresh self-contained room","description":"Recently painted room with private bathroom.","price_naira":350000}`
	req := authenticatedRequest(http.MethodPost, "/v1/unit-types/550e8400-e29b-41d4-a716-446655440020/agent-offers", body, true)
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
			Version            int    `json:"version"`
		} `json:"agent_offer"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&bodyDecoded); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if bodyDecoded.AgentOffer.Title != "Fresh self-contained room" {
		t.Fatalf("title = %q, want Fresh self-contained room", bodyDecoded.AgentOffer.Title)
	}
	if bodyDecoded.AgentOffer.AgentID != "550e8400-e29b-41d4-a716-446655440040" {
		t.Fatalf("agent_id = %q, want authenticated agent", bodyDecoded.AgentOffer.AgentID)
	}
	if bodyDecoded.AgentOffer.PriceNaira != 350000 {
		t.Fatalf("price_naira = %d, want 350000", bodyDecoded.AgentOffer.PriceNaira)
	}
	if bodyDecoded.AgentOffer.Status != "available" {
		t.Fatalf("status = %q, want available", bodyDecoded.AgentOffer.Status)
	}
	if bodyDecoded.AgentOffer.Version != 1 {
		t.Fatalf("version = %d, want 1", bodyDecoded.AgentOffer.Version)
	}
}

func TestCreateAgentOfferConvertsNairaToKobo(t *testing.T) {
	spy := &spyPropertyRepo{stubPropertyRepo: &stubPropertyRepo{}}
	app := authenticatedTestApp(domain.ID("550e8400-e29b-41d4-a716-446655440001"), &fakeAgentApplicationStore{access: domain.EffectiveAccess{Agent: &domain.LinkedAgent{ID: domain.ID("550e8400-e29b-41d4-a716-446655440040"), Status: domain.AgentStatusActive}}})
	app.propertyRepo = spy

	body := `{"title":"Fresh self-contained room","price_naira":350000}`
	req := authenticatedRequest(http.MethodPost, "/v1/unit-types/550e8400-e29b-41d4-a716-446655440020/agent-offers", body, true)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusCreated)
	}

	if spy.createdOffer.Price.AmountKobo != 35000000 {
		t.Fatalf("stored price = %d kobo, want 35000000", spy.createdOffer.Price.AmountKobo)
	}
}

func TestCreateAgentOfferDuplicateReturnsConflict(t *testing.T) {
	dupRepo := &duplicateAgentOfferRepo{stubPropertyRepo: &stubPropertyRepo{}}
	app := authenticatedTestApp(domain.ID("550e8400-e29b-41d4-a716-446655440001"), &fakeAgentApplicationStore{access: domain.EffectiveAccess{Agent: &domain.LinkedAgent{ID: domain.ID("550e8400-e29b-41d4-a716-446655440040"), Status: domain.AgentStatusActive}}})
	app.propertyRepo = dupRepo

	body := `{"title":"Duplicate offer","price_naira":350000}`
	req := authenticatedRequest(http.MethodPost, "/v1/unit-types/550e8400-e29b-41d4-a716-446655440020/agent-offers", body, true)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusConflict, "conflict", "The resource already exists")
}

func TestCreateAgentOfferUnitTypeNotFound(t *testing.T) {
	app := testAppWithActiveAgentRepo()
	body := `{"title":"Fresh self-contained room","price_naira":350000}`
	req := authenticatedRequest(http.MethodPost, "/v1/unit-types/11111111-1111-1111-1111-111111111111/agent-offers", body, true)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusNotFound, "not_found", "The requested resource could not be found")
}

func TestCreateAgentOfferRequiresActiveAgent(t *testing.T) {
	app := authenticatedTestApp(domain.ID("550e8400-e29b-41d4-a716-446655440001"), &fakeAgentApplicationStore{})
	app.propertyRepo = &stubPropertyRepo{}
	body := `{"title":"Fresh self-contained room","price_naira":350000}`
	req := authenticatedRequest(http.MethodPost, "/v1/unit-types/550e8400-e29b-41d4-a716-446655440020/agent-offers", body, true)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusForbidden, "agent_required", "An active agent profile is required")
}

func TestCreateAgentOfferRejectsPriceOverflow(t *testing.T) {
	app := testAppWithActiveAgentRepo()
	body := `{"title":"Premium room","price_naira":21474837}`
	req := authenticatedRequest(http.MethodPost, "/v1/unit-types/550e8400-e29b-41d4-a716-446655440020/agent-offers", body, true)
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
	if bodyDecoded.Error.Fields["price_naira"] == "" {
		t.Fatal("price_naira validation error missing")
	}
}

func TestCreateAgentOfferAcceptsMaxPrice(t *testing.T) {
	spy := &spyPropertyRepo{stubPropertyRepo: &stubPropertyRepo{}}
	app := authenticatedTestApp(domain.ID("550e8400-e29b-41d4-a716-446655440001"), &fakeAgentApplicationStore{access: domain.EffectiveAccess{Agent: &domain.LinkedAgent{ID: domain.ID("550e8400-e29b-41d4-a716-446655440040"), Status: domain.AgentStatusActive}}})
	app.propertyRepo = spy

	body := `{"title":"Premium room","price_naira":21474836}`
	req := authenticatedRequest(http.MethodPost, "/v1/unit-types/550e8400-e29b-41d4-a716-446655440020/agent-offers", body, true)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusCreated)
	}

	if spy.createdOffer.Price.AmountKobo != 2147483600 {
		t.Fatalf("stored price = %d kobo, want 2147483600", spy.createdOffer.Price.AmountKobo)
	}
}

func TestCreateAgentOfferRejectsClientSuppliedAgentID(t *testing.T) {
	app := testAppWithActiveAgentRepo()
	body := `{"agent_id":"550e8400-e29b-41d4-a716-446655440099","title":"Premium room","price_naira":350000}`
	req := authenticatedRequest(http.MethodPost, "/v1/unit-types/550e8400-e29b-41d4-a716-446655440020/agent-offers", body, true)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestUpdateAgentOfferRequiresOwnership(t *testing.T) {
	app := authenticatedTestApp(domain.ID("550e8400-e29b-41d4-a716-446655440001"), &fakeAgentApplicationStore{access: domain.EffectiveAccess{
		Agent: &domain.LinkedAgent{ID: domain.ID("550e8400-e29b-41d4-a716-446655440041"), Status: domain.AgentStatusActive},
	}})
	app.propertyRepo = &stubPropertyRepo{}
	req := authenticatedRequest(http.MethodPatch, "/v1/agent-offers/550e8400-e29b-41d4-a716-446655440030", `{"title":"Updated offer"}`, true)
	req.Header.Set("If-Match", `"agent-offer-550e8400-e29b-41d4-a716-446655440030-1"`)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusForbidden, "forbidden", "You are not authorized to update this agent offer")
}

func TestUpdateAgentOfferRequiresIfMatch(t *testing.T) {
	app := testAppWithActiveAgentRepo()
	req := authenticatedRequest(http.MethodPatch, "/v1/agent-offers/550e8400-e29b-41d4-a716-446655440030", `{"title":"Updated offer"}`, true)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusPreconditionRequired, "precondition_required", "If-Match is required")
}

func TestUpdateAgentOfferReturnsNewVersionAndETag(t *testing.T) {
	app := testAppWithActiveAgentRepo()
	req := authenticatedRequest(http.MethodPatch, "/v1/agent-offers/550e8400-e29b-41d4-a716-446655440030", `{"title":"Updated offer","price_naira":360000}`, true)
	req.Header.Set("If-Match", `"agent-offer-550e8400-e29b-41d4-a716-446655440030-1"`)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
	}
	if got := rr.Header().Get("ETag"); got != `"agent-offer-550e8400-e29b-41d4-a716-446655440030-2"` {
		t.Fatalf("ETag = %q, want updated agent offer ETag", got)
	}
	var body struct {
		AgentOffer struct {
			Title      string `json:"title"`
			PriceNaira int    `json:"price_naira"`
			Version    int    `json:"version"`
		} `json:"agent_offer"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if body.AgentOffer.Title != "Updated offer" || body.AgentOffer.PriceNaira != 360000 {
		t.Fatalf("updated offer = %#v, want changed fields", body.AgentOffer)
	}
	if body.AgentOffer.Version != 2 {
		t.Fatalf("version = %d, want 2", body.AgentOffer.Version)
	}
}

func TestArchiveAgentOfferReturnsArchivedVersionAndETag(t *testing.T) {
	app := testAppWithActiveAgentRepo()
	req := authenticatedRequest(http.MethodPost, "/v1/agent-offers/550e8400-e29b-41d4-a716-446655440030/archive", "", true)
	req.Header.Set("If-Match", `"agent-offer-550e8400-e29b-41d4-a716-446655440030-1"`)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
	}
	if got := rr.Header().Get("ETag"); got != `"agent-offer-550e8400-e29b-41d4-a716-446655440030-2"` {
		t.Fatalf("ETag = %q, want archived agent offer ETag", got)
	}
	var body struct {
		AgentOffer struct {
			Status     string `json:"status"`
			Version    int    `json:"version"`
			ArchivedAt string `json:"archived_at"`
		} `json:"agent_offer"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if body.AgentOffer.Status != "unavailable" {
		t.Fatalf("status = %q, want unavailable", body.AgentOffer.Status)
	}
	if body.AgentOffer.Version != 2 || body.AgentOffer.ArchivedAt == "" {
		t.Fatalf("archived offer = %#v, want version 2 and archive timestamp", body.AgentOffer)
	}
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
