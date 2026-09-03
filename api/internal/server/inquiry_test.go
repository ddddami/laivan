package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ddddami/laivan/internal/domain"
	"github.com/ddddami/laivan/internal/repo"
)

func TestCreateInquiryReturnsPersistedInquiryAndOpaqueHandoff(t *testing.T) {
	studentID := domain.ID("550e8400-e29b-41d4-a716-446655440001")
	store := &fakeInquiryStore{
		created: true,
		inquiry: domain.Inquiry{
			ID:            domain.ID("550e8400-e29b-41d4-a716-446655440010"),
			StudentUserID: studentID,
			AgentOfferID:  domain.ID("550e8400-e29b-41d4-a716-446655440030"),
			SubmissionID:  domain.ID("550e8400-e29b-41d4-a716-446655440020"),
			Message:       "Is the kitchen private?",
			Status:        domain.InquiryStatusOpen,
			Timestamps: domain.Timestamps{
				CreatedAt: time.Date(2026, time.September, 3, 14, 0, 0, 0, time.UTC),
			},
		},
		handoff: domain.InquiryHandoff{
			AgentDisplayName: "Bisi Housing Connect",
			PhoneNumber:      "+2348031234567",
			OfferTitle:       "Fresh self-contained room",
			PropertyName:     "Alice Lodge",
			UnitName:         "Self-contained",
			Area:             "Obanla",
		},
	}
	app := authenticatedTestApp(studentID, &fakeAgentApplicationStore{})
	app.inquiries = store
	request := authenticatedRequest(http.MethodPost, "/v1/agent-offers/550e8400-e29b-41d4-a716-446655440030/inquiries", `{"message":"  Is the kitchen private?  ","submission_id":"550e8400-e29b-41d4-a716-446655440020"}`, true)
	response := httptest.NewRecorder()

	app.routes().ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusCreated)
	}
	if store.studentUserID != studentID || store.agentOfferID != domain.ID("550e8400-e29b-41d4-a716-446655440030") {
		t.Fatalf("store arguments = student %q, offer %q", store.studentUserID, store.agentOfferID)
	}
	if store.message != "Is the kitchen private?" {
		t.Fatalf("stored message = %q, want trimmed message", store.message)
	}
	if strings.Contains(response.Body.String(), "phone_number") || strings.Contains(response.Body.String(), "whatsapp_number") {
		t.Fatal("response exposed a raw contact field")
	}

	var body struct {
		Inquiry struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"inquiry"`
		Handoff struct {
			Channel string `json:"channel"`
			URL     string `json:"url"`
		} `json:"handoff"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Inquiry.ID != string(store.inquiry.ID) || body.Inquiry.Status != "open" {
		t.Fatalf("inquiry response = %#v", body.Inquiry)
	}
	if body.Handoff.Channel != "whatsapp" || !strings.HasPrefix(body.Handoff.URL, "https://wa.me/2348031234567?text=") {
		t.Fatalf("handoff response = %#v", body.Handoff)
	}
}

func TestCreateInquiryReplayReturnsOK(t *testing.T) {
	store := &fakeInquiryStore{created: false}
	app := authenticatedTestApp(domain.ID("550e8400-e29b-41d4-a716-446655440001"), &fakeAgentApplicationStore{})
	app.inquiries = store
	request := authenticatedRequest(http.MethodPost, "/v1/agent-offers/550e8400-e29b-41d4-a716-446655440030/inquiries", `{"message":"Is it available?","submission_id":"550e8400-e29b-41d4-a716-446655440020"}`, true)
	response := httptest.NewRecorder()

	app.routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusOK)
	}
}

func TestCreateInquiryRejectsInvalidRequestsBeforeStore(t *testing.T) {
	cases := []struct {
		name string
		body string
		path string
		want int
	}{
		{name: "invalid offer id", body: `{"message":"Question","submission_id":"550e8400-e29b-41d4-a716-446655440020"}`, path: "/v1/agent-offers/not-a-uuid/inquiries", want: http.StatusUnprocessableEntity},
		{name: "blank message", body: `{"message":"  ","submission_id":"550e8400-e29b-41d4-a716-446655440020"}`, path: "/v1/agent-offers/550e8400-e29b-41d4-a716-446655440030/inquiries", want: http.StatusUnprocessableEntity},
		{name: "unknown field", body: `{"message":"Question","submission_id":"550e8400-e29b-41d4-a716-446655440020","extra":true}`, path: "/v1/agent-offers/550e8400-e29b-41d4-a716-446655440030/inquiries", want: http.StatusBadRequest},
		{name: "multiple values", body: `{"message":"Question","submission_id":"550e8400-e29b-41d4-a716-446655440020"}{}`, path: "/v1/agent-offers/550e8400-e29b-41d4-a716-446655440030/inquiries", want: http.StatusBadRequest},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &fakeInquiryStore{}
			app := authenticatedTestApp(domain.ID("550e8400-e29b-41d4-a716-446655440001"), &fakeAgentApplicationStore{})
			app.inquiries = store
			request := authenticatedRequest(http.MethodPost, testCase.path, testCase.body, true)
			response := httptest.NewRecorder()

			app.routes().ServeHTTP(response, request)

			if response.Code != testCase.want {
				t.Fatalf("status code = %d, want %d", response.Code, testCase.want)
			}
			if store.calls != 0 {
				t.Fatalf("store calls = %d, want 0", store.calls)
			}
		})
	}
}

func TestCreateInquiryMapsDomainErrors(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
		code string
	}{
		{name: "not found", err: repo.ErrNotFound, want: http.StatusNotFound, code: "not_found"},
		{name: "unavailable", err: repo.ErrOfferUnavailable, want: http.StatusConflict, code: "offer_unavailable"},
		{name: "own offer", err: repo.ErrOwnOffer, want: http.StatusForbidden, code: "forbidden"},
		{name: "unexpected", err: errors.New("database failed"), want: http.StatusInternalServerError, code: "internal_server_error"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &fakeInquiryStore{err: testCase.err}
			app := authenticatedTestApp(domain.ID("550e8400-e29b-41d4-a716-446655440001"), &fakeAgentApplicationStore{})
			app.inquiries = store
			request := authenticatedRequest(http.MethodPost, "/v1/agent-offers/550e8400-e29b-41d4-a716-446655440030/inquiries", `{"message":"Question","submission_id":"550e8400-e29b-41d4-a716-446655440020"}`, true)
			response := httptest.NewRecorder()

			app.routes().ServeHTTP(response, request)

			assertErrorCodeResponse(t, response, testCase.want, testCase.code)
		})
	}
}

func TestCreateInquiryRejectsAnonymousUsers(t *testing.T) {
	app := testApp()
	app.inquiries = &fakeInquiryStore{}
	request := httptest.NewRequest(http.MethodPost, "/v1/agent-offers/550e8400-e29b-41d4-a716-446655440030/inquiries", strings.NewReader(`{"message":"Question","submission_id":"550e8400-e29b-41d4-a716-446655440020"}`))
	response := httptest.NewRecorder()

	app.routes().ServeHTTP(response, request)

	assertErrorResponse(t, response, http.StatusUnauthorized, "unauthenticated", "Authentication is required")
}

func TestCreateInquiryRejectsInvalidCSRFBeforeStore(t *testing.T) {
	store := &fakeInquiryStore{}
	app := authenticatedTestApp(domain.ID("550e8400-e29b-41d4-a716-446655440001"), &fakeAgentApplicationStore{})
	app.inquiries = store
	request := authenticatedRequest(http.MethodPost, "/v1/agent-offers/550e8400-e29b-41d4-a716-446655440030/inquiries", `{"message":"Question","submission_id":"550e8400-e29b-41d4-a716-446655440020"}`, false)
	response := httptest.NewRecorder()

	app.routes().ServeHTTP(response, request)

	assertErrorResponse(t, response, http.StatusForbidden, "forbidden", "A valid CSRF token and request origin are required")
	if store.calls != 0 {
		t.Fatalf("store calls = %d, want 0", store.calls)
	}
}

type fakeInquiryStore struct {
	inquiry       domain.Inquiry
	handoff       domain.InquiryHandoff
	created       bool
	err           error
	calls         int
	studentUserID domain.ID
	agentOfferID  domain.ID
	submissionID  domain.ID
	message       string
}

func (f *fakeInquiryStore) Submit(_ context.Context, studentUserID, agentOfferID, submissionID domain.ID, message string) (domain.Inquiry, domain.InquiryHandoff, bool, error) {
	f.calls++
	f.studentUserID = studentUserID
	f.agentOfferID = agentOfferID
	f.submissionID = submissionID
	f.message = message
	return f.inquiry, f.handoff, f.created, f.err
}
