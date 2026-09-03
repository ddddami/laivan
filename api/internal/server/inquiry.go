package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ddddami/laivan/internal/domain"
	"github.com/ddddami/laivan/internal/repo"
	"github.com/ddddami/laivan/internal/validator"
	"github.com/go-chi/chi/v5"
)

func (app *app) createInquiry(w http.ResponseWriter, r *http.Request) {
	p, ok := principalFromContext(r.Context())
	if !ok {
		app.errorResponse(w, r, http.StatusUnauthorized, "unauthenticated", "Authentication is required")
		return
	}

	offerID := chi.URLParam(r, "id")
	var input struct {
		Message      string `json:"message"`
		SubmissionID string `json:"submission_id"`
	}
	if err := readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	input.Message = strings.TrimSpace(input.Message)
	v := validator.New()
	v.Check(validator.ValidUUID(offerID), "id", "Offer ID must be a valid UUID")
	v.Check(validator.NotBlank(input.Message), "message", "Message is required")
	v.Check(validator.MaxChars(input.Message, 1000), "message", "Message must not exceed 1000 characters")
	v.Check(validator.ValidUUID(input.SubmissionID), "submission_id", "Submission ID must be a valid UUID")
	if !v.Valid() {
		app.validationFailedResponse(w, r, v.FieldErrors)
		return
	}

	if app.inquiries == nil {
		app.serverErrorResponse(w, r, errors.New("inquiry store is not configured"))
		return
	}
	inquiry, handoff, created, err := app.inquiries.Submit(r.Context(), p.User.ID, domain.ID(offerID), domain.ID(input.SubmissionID), input.Message)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			app.notFoundResponse(w, r)
		case errors.Is(err, repo.ErrOfferUnavailable):
			app.errorResponse(w, r, http.StatusConflict, "offer_unavailable", "This offer is no longer accepting questions")
		case errors.Is(err, repo.ErrOwnOffer):
			app.errorResponse(w, r, http.StatusForbidden, "forbidden", "You cannot ask a question about your own offer")
		default:
			app.serverErrorResponse(w, r, fmt.Errorf("create inquiry: %w", err))
		}
		return
	}

	status := http.StatusCreated
	if !created {
		status = http.StatusOK
	}
	data := envelope{
		"inquiry": inquiryResponse(inquiry),
		"handoff": inquiryHandoffResponse(inquiry, handoff),
	}
	if err := writeJSON(w, status, data, nil); err != nil {
		app.logger.Error("write inquiry response", "error", err)
	}
}

func inquiryResponse(inquiry domain.Inquiry) map[string]any {
	return map[string]any{
		"id":             string(inquiry.ID),
		"agent_offer_id": string(inquiry.AgentOfferID),
		"message":        inquiry.Message,
		"status":         string(inquiry.Status),
		"created_at":     inquiry.CreatedAt.Format(time.RFC3339),
	}
}

func inquiryHandoffResponse(inquiry domain.Inquiry, handoff domain.InquiryHandoff) map[string]any {
	return map[string]any{
		"channel": "whatsapp",
		"url":     whatsappURL(inquiry, handoff),
	}
}

func whatsappURL(inquiry domain.Inquiry, handoff domain.InquiryHandoff) string {
	number := handoff.WhatsAppNumber
	if number == "" {
		number = handoff.PhoneNumber
	}
	number = strings.TrimPrefix(number, "+")
	message := fmt.Sprintf("Hello %s, I have a question about %q at %s, %s in %s.\n\nMy question: %s\n\nSent through Laivan inquiry %s.", handoff.AgentDisplayName, handoff.OfferTitle, handoff.PropertyName, handoff.UnitName, handoff.Area, inquiry.Message, inquiry.ID)
	return "https://wa.me/" + number + "?text=" + url.QueryEscape(message)
}
