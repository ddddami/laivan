package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ddddami/laivan/internal/domain"
	"github.com/ddddami/laivan/internal/repo"
	"github.com/ddddami/laivan/internal/validator"
	"github.com/go-chi/chi/v5"
)

func (app *app) getCampusBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	v := validator.New()
	v.Check(validator.NotBlank(slug), "slug", "Slug is required")
	v.Check(validator.MaxChars(slug, 100), "slug", "Slug must not exceed 100 characters")
	if !v.Valid() {
		app.validationFailedResponse(w, r, v.FieldErrors)
		return
	}

	campus, err := app.propertyRepo.GetCampusBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			app.notFoundResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, fmt.Errorf("get campus by slug: %w", err))
		return
	}

	data := envelope{"campus": campusResponse(campus)}
	if err := writeJSON(w, http.StatusOK, data, nil); err != nil {
		app.logger.Error("write campus response", "error", err)
	}
}

func campusResponse(campus domain.Campus) map[string]any {
	return map[string]any{
		"id":         string(campus.ID),
		"slug":       campus.Slug,
		"name":       campus.Name,
		"short_name": campus.ShortName,
	}
}
