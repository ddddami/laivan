package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *app) routes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.NotFound(app.notFoundResponse)
	r.MethodNotAllowed(app.methodNotAllowedResponse)

	r.Get("/healthz", app.healthz)
	r.Get("/docs", app.docs)
	r.Get("/docs/", app.docs)
	r.Get("/openapi.yaml", app.openapi)

	r.Route("/v1/properties", func(r chi.Router) {
		r.Get("/", app.listProperties)
		r.Post("/", app.createProperty)
		r.Get("/{id}", app.getProperty)
	})

	return r
}
