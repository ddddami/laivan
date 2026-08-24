package server

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (app *app) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		app.logger.Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"duration", time.Since(start),
			"request_id", middleware.GetReqID(r.Context()),
		)
	})
}

func (app *app) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				requestID := middleware.GetReqID(r.Context())
				app.logger.Error("panic recovered",
					"panic", recovered,
					"stack", string(debug.Stack()),
					"method", r.Method,
					"path", r.URL.Path,
					"request_id", requestID,
				)
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("panic: %v", recovered))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *app) routes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(app.logRequest)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   app.cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(app.recoverPanic)
	r.NotFound(app.notFoundResponse)
	r.MethodNotAllowed(app.methodNotAllowedResponse)

	r.Get("/healthz", app.healthz)
	r.Get("/docs", app.docs)
	r.Get("/docs/", app.docs)
	r.Get("/openapi.yaml", app.openapi)

	r.Route("/v1/auth", func(r chi.Router) {
		r.Get("/google/start", app.googleAuthStart)
		r.Get("/google/callback", app.googleAuthCallback)
		r.Get("/session", app.authSession)
		r.Post("/logout", app.authLogout)
	})

	r.Route("/v1/agent-applications", func(r chi.Router) {
		r.Use(app.requireAuthenticatedUser, app.requireCSRF)
		r.Post("/", app.createAgentApplication)
		r.Get("/", app.listAgentApplications)
	})

	r.Route("/v1/operator/agent-applications", func(r chi.Router) {
		r.Use(app.requireAuthenticatedUser, app.requireCampusOperator, app.requireCSRF)
		r.Get("/", app.listOperatorAgentApplications)
		r.Post("/{id}/activate", app.activateAgentApplication)
		r.Post("/{id}/decline", app.declineAgentApplication)
	})

	r.Get("/v1/campuses/{slug}", app.getCampusBySlug)

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
