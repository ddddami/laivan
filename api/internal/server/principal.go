package server

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"

	"github.com/ddddami/laivan/internal/auth"
	"github.com/ddddami/laivan/internal/domain"
)

type principal struct {
	User      domain.User
	Access    domain.EffectiveAccess
	CSRFToken string
}

type principalContextKey struct{}

func (app *app) requireAuthenticatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.auth == nil || app.applications == nil {
			app.errorResponse(w, r, http.StatusUnauthorized, "unauthenticated", "Authentication is required")
			return
		}

		sessionCookie, err := r.Cookie(app.cfg.Auth.SessionCookieName)
		if err != nil || sessionCookie.Value == "" {
			app.errorResponse(w, r, http.StatusUnauthorized, "unauthenticated", "Authentication is required")
			return
		}
		csrfCookie, _ := r.Cookie(app.cfg.Auth.CSRFCookieName)
		csrfCookieValue := ""
		if csrfCookie != nil {
			csrfCookieValue = csrfCookie.Value
		}
		result, err := app.auth.GetSession(r.Context(), sessionCookie.Value, csrfCookieValue)
		if err != nil {
			if errors.Is(err, auth.ErrUnauthenticated) {
				app.errorResponse(w, r, http.StatusUnauthorized, "unauthenticated", "Authentication is required")
				return
			}
			app.serverErrorResponse(w, r, err)
			return
		}

		access, err := app.applications.GetEffectiveAccess(r.Context(), result.User.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		ctx := context.WithValue(r.Context(), principalContextKey{}, principal{User: result.User, Access: access, CSRFToken: result.CSRFToken})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *app) requireCampusOperator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := principalFromContext(r.Context())
		if !ok || (!hasRole(p.Access, "campus_operator") && !p.Access.GlobalAdmin) {
			app.errorResponse(w, r, http.StatusForbidden, "forbidden", "Campus operator access is required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (app *app) requireCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		p, ok := principalFromContext(r.Context())
		if !ok || r.Header.Get("Origin") != app.cfg.Auth.WebOrigin || !validCSRFHeader(r.Header.Get("X-CSRF-Token"), p.CSRFToken) {
			app.errorResponse(w, r, http.StatusForbidden, "forbidden", "A valid CSRF token and request origin are required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func hasRole(access domain.EffectiveAccess, role string) bool {
	for _, candidate := range access.Roles {
		if candidate == role {
			return true
		}
	}
	return false
}

func validCSRFHeader(value, expected string) bool {
	if value == "" || expected == "" || len(value) != len(expected) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(value), []byte(expected)) == 1
}

func principalFromContext(ctx context.Context) (principal, bool) {
	value, ok := ctx.Value(principalContextKey{}).(principal)
	return value, ok
}
