package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (app *app) routes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   app.cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
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
		r.Get("/{id}/unit-types", app.listPropertyUnitTypes)
		r.Post("/{id}/unit-types", app.createPropertyUnitType)
	})

	r.Route("/v1/unit-types", func(r chi.Router) {
		r.Get("/{id}/agent-offers", app.listAgentOffers)
		r.Post("/{id}/agent-offers", app.createAgentOffer)
	})

	r.Get("/v1/discovery", app.discover)
	r.Post("/v1/media", app.uploadMedia)

	return r
}
