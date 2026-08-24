package server

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ddddami/laivan/internal/domain"
	"github.com/ddddami/laivan/internal/repo"
	"github.com/ddddami/laivan/internal/validator"
	"github.com/go-chi/chi/v5"
)

func (app *app) suspendAgent(w http.ResponseWriter, r *http.Request) {
	app.changeAgentLifecycle(w, r, true)
}

func (app *app) reinstateAgent(w http.ResponseWriter, r *http.Request) {
	app.changeAgentLifecycle(w, r, false)
}

func (app *app) changeAgentLifecycle(w http.ResponseWriter, r *http.Request, suspend bool) {
	p, ok := principalFromContext(r.Context())
	if !ok {
		app.errorResponse(w, r, http.StatusUnauthorized, "unauthenticated", "Authentication is required")
		return
	}

	agentID := chi.URLParam(r, "id")
	operatorNote, err := readOperatorNote(w, r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	v := validator.New()
	v.Check(validator.NotBlank(agentID), "id", "ID is required")
	v.Check(validator.ValidUUID(agentID), "id", "ID must be a valid UUID")
	v.Check(validator.MaxChars(operatorNote, 2000), "operator_note", "Operator note must not exceed 2000 characters")
	if !v.Valid() {
		app.validationFailedResponse(w, r, v.FieldErrors)
		return
	}

	var agent domain.LinkedAgent
	if suspend {
		agent, err = app.applications.SuspendAgent(r.Context(), domain.ID(agentID), p.User.ID, operatorNote)
	} else {
		agent, err = app.applications.ReinstateAgent(r.Context(), domain.ID(agentID), p.User.ID, operatorNote)
	}
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrAgentNotFound):
			app.notFoundResponse(w, r)
		case errors.Is(err, repo.ErrCampusForbidden):
			app.errorResponse(w, r, http.StatusForbidden, "forbidden", "The operator is not authorized for this agent's campus")
		case errors.Is(err, repo.ErrAgentLifecycleConflict):
			app.conflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, fmt.Errorf("change agent lifecycle: %w", err))
		}
		return
	}

	if err := writeJSON(w, http.StatusOK, envelope{"agent": agentLifecycleResponse(agent)}, nil); err != nil {
		app.logger.Error("write agent lifecycle response", "error", err)
	}
}

func readOperatorNote(w http.ResponseWriter, r *http.Request) (string, error) {
	var input struct {
		OperatorNote string `json:"operator_note"`
	}
	if r.Body != nil && r.Body != http.NoBody {
		if err := readJSON(w, r, &input); err != nil && err.Error() != "request body must not be empty" {
			return "", err
		}
	}
	return strings.TrimSpace(input.OperatorNote), nil
}

func agentLifecycleResponse(agent domain.LinkedAgent) map[string]string {
	return map[string]string{"id": string(agent.ID), "status": string(agent.Status)}
}
