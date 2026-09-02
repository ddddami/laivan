package server

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ddddami/laivan/internal/domain"
	"github.com/ddddami/laivan/internal/phone"
	"github.com/ddddami/laivan/internal/repo"
	"github.com/ddddami/laivan/internal/validator"
	"github.com/go-chi/chi/v5"
)

func (app *app) createAgentApplication(w http.ResponseWriter, r *http.Request) {
	p, ok := principalFromContext(r.Context())
	if !ok {
		app.errorResponse(w, r, http.StatusUnauthorized, "unauthenticated", "Authentication is required")
		return
	}
	var input struct {
		CampusID    string `json:"campus_id"`
		Name        string `json:"name"`
		PhoneNumber string `json:"phone_number"`
	}
	if err := readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	v := validator.New()
	v.Check(validator.NotBlank(input.CampusID), "campus_id", "Campus ID is required")
	v.Check(validator.ValidUUID(input.CampusID), "campus_id", "Campus ID must be a valid UUID")
	v.Check(validator.NotBlank(input.Name), "name", "Name is required")
	v.Check(validator.MaxChars(input.Name, 255), "name", "Name must not exceed 255 characters")
	v.Check(validator.NotBlank(input.PhoneNumber), "phone_number", "Phone number is required")

	normalizedPhone, err := phone.Normalize(input.PhoneNumber)
	if err != nil {
		v.AddFieldError("phone_number", "Phone number must be a valid Nigerian phone number")
	}
	if !v.Valid() {
		app.validationFailedResponse(w, r, v.FieldErrors)
		return
	}

	created, err := app.applications.CreateApplication(r.Context(), domain.AgentApplication{
		ApplicantUserID: p.User.ID,
		CampusID:        domain.ID(input.CampusID),
		Name:            input.Name,
		PhoneNumber:     normalizedPhone,
		Status:          domain.AgentApplicationStatusPending,
	})
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrAlreadyLinked), errors.Is(err, repo.ErrApplicationConflict):
			app.conflictResponse(w, r)
		case errors.Is(err, repo.ErrNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, fmt.Errorf("create agent application: %w", err))
		}
		return
	}
	if err := writeJSON(w, http.StatusCreated, envelope{"application": agentApplicationResponse(created)}, nil); err != nil {
		app.logger.Error("write agent application response", "error", err)
	}
}

func (app *app) listAgentApplications(w http.ResponseWriter, r *http.Request) {
	p, ok := principalFromContext(r.Context())
	if !ok {
		app.errorResponse(w, r, http.StatusUnauthorized, "unauthenticated", "Authentication is required")
		return
	}
	applications, err := app.applications.ListApplications(r.Context(), p.User.ID)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("list agent applications: %w", err))
		return
	}
	if err := writeJSON(w, http.StatusOK, envelope{"applications": agentApplicationsResponse(applications)}, nil); err != nil {
		app.logger.Error("write agent applications response", "error", err)
	}
}

func (app *app) listOperatorAgentApplications(w http.ResponseWriter, r *http.Request) {
	p, ok := principalFromContext(r.Context())
	if !ok {
		app.errorResponse(w, r, http.StatusUnauthorized, "unauthenticated", "Authentication is required")
		return
	}
	qs := r.URL.Query()
	campusID := qs.Get("campus_id")
	status := qs.Get("status")
	if status == "" {
		status = string(domain.AgentApplicationStatusPending)
	}
	v := validator.New()
	v.Check(validator.NotBlank(campusID), "campus_id", "Campus ID is required")
	v.Check(validator.ValidUUID(campusID), "campus_id", "Campus ID must be a valid UUID")
	v.Check(validApplicationStatus(status), "status", "Status must be pending, active, declined, or suspended")
	if !v.Valid() {
		app.validationFailedResponse(w, r, v.FieldErrors)
		return
	}

	applications, err := app.applications.ListOperatorApplications(r.Context(), p.User.ID, domain.ID(campusID), domain.AgentApplicationStatus(status))
	if err != nil {
		if errors.Is(err, repo.ErrCampusForbidden) {
			app.errorResponse(w, r, http.StatusForbidden, "forbidden", "The operator is not authorized for this campus")
			return
		}
		app.serverErrorResponse(w, r, fmt.Errorf("list operator agent applications: %w", err))
		return
	}
	if err := writeJSON(w, http.StatusOK, envelope{"applications": agentApplicationsResponse(applications)}, nil); err != nil {
		app.logger.Error("write operator agent applications response", "error", err)
	}
}

func (app *app) activateAgentApplication(w http.ResponseWriter, r *http.Request) {
	app.respondToAgentApplication(w, r, true)
}

func (app *app) declineAgentApplication(w http.ResponseWriter, r *http.Request) {
	app.respondToAgentApplication(w, r, false)
}

func (app *app) respondToAgentApplication(w http.ResponseWriter, r *http.Request, activate bool) {
	p, ok := principalFromContext(r.Context())
	if !ok {
		app.errorResponse(w, r, http.StatusUnauthorized, "unauthenticated", "Authentication is required")
		return
	}
	applicationID := chi.URLParam(r, "id")
	v := validator.New()
	v.Check(validator.NotBlank(applicationID), "id", "ID is required")
	v.Check(validator.ValidUUID(applicationID), "id", "ID must be a valid UUID")

	var input struct {
		OperatorNote string `json:"operator_note"`
	}
	if r.Body != nil && r.Body != http.NoBody {
		if err := readJSON(w, r, &input); err != nil && err.Error() != "request body must not be empty" {
			app.badRequestResponse(w, r, err)
			return
		}
	}
	input.OperatorNote = strings.TrimSpace(input.OperatorNote)
	v.Check(validator.MaxChars(input.OperatorNote, 2000), "operator_note", "Operator note must not exceed 2000 characters")
	if !v.Valid() {
		app.validationFailedResponse(w, r, v.FieldErrors)
		return
	}

	var application domain.AgentApplication
	var err error
	if activate {
		application, err = app.applications.ActivateApplication(r.Context(), domain.ID(applicationID), p.User.ID, input.OperatorNote)
	} else {
		application, err = app.applications.DeclineApplication(r.Context(), domain.ID(applicationID), p.User.ID, input.OperatorNote)
	}
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrCampusForbidden):
			app.errorResponse(w, r, http.StatusForbidden, "forbidden", "The operator is not authorized for this campus")
		case errors.Is(err, repo.ErrNotFound):
			app.notFoundResponse(w, r)
		case errors.Is(err, repo.ErrAlreadyLinked), errors.Is(err, repo.ErrApplicationResolved), errors.Is(err, repo.ErrLegacyAgentConflict), errors.Is(err, repo.ErrPhoneConflict):
			app.conflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, fmt.Errorf("decide agent application: %w", err))
		}
		return
	}
	if err := writeJSON(w, http.StatusOK, envelope{"application": agentApplicationResponse(application)}, nil); err != nil {
		app.logger.Error("write agent application decision response", "error", err)
	}
}

func validApplicationStatus(value string) bool {
	switch domain.AgentApplicationStatus(value) {
	case domain.AgentApplicationStatusPending, domain.AgentApplicationStatusActive, domain.AgentApplicationStatusDeclined, domain.AgentApplicationStatusSuspended:
		return true
	default:
		return false
	}
}

func agentApplicationsResponse(applications []domain.AgentApplication) []map[string]any {
	result := make([]map[string]any, 0, len(applications))
	for _, application := range applications {
		result = append(result, agentApplicationResponse(application))
	}
	return result
}

func agentApplicationResponse(application domain.AgentApplication) map[string]any {
	response := map[string]any{
		"id":            string(application.ID),
		"campus_id":     string(application.CampusID),
		"name":          application.Name,
		"phone_number":  application.PhoneNumber,
		"status":        string(application.Status),
		"agent_id":      nullableApplicationID(application.AgentID),
		"operator_note": nullableString(application.OperatorNote),
		"created_at":    application.CreatedAt.Format(time.RFC3339),
		"updated_at":    application.UpdatedAt.Format(time.RFC3339),
		"decided_at":    nullableTime(application.DecidedAt),
	}
	if application.ApplicantUserID != "" {
		response["applicant"] = map[string]any{
			"id":           string(application.ApplicantUserID),
			"email":        application.ApplicantEmail,
			"display_name": application.ApplicantDisplayName,
		}
	}
	return response
}

func nullableApplicationID(value *domain.ID) any {
	if value == nil {
		return nil
	}
	return string(*value)
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.Format(time.RFC3339)
}
