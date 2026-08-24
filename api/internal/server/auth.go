package server

import (
	"errors"
	"net/http"

	"github.com/ddddami/laivan/internal/auth"
	"github.com/ddddami/laivan/internal/domain"
)

func (app *app) googleAuthStart(w http.ResponseWriter, r *http.Request) {
	if app.auth == nil {
		app.authUnavailableResponse(w, r)
		return
	}

	redirectURL, cookieValue, err := app.auth.Begin()
	if err != nil {
		if errors.Is(err, auth.ErrAuthUnavailable) {
			app.authUnavailableResponse(w, r)
			return
		}

		app.serverErrorResponse(w, r, err)
		return
	}

	app.setAuthAttemptCookie(w, cookieValue)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (app *app) googleAuthCallback(w http.ResponseWriter, r *http.Request) {
	app.clearAuthAttemptCookie(w)
	if app.auth == nil {
		app.authUnavailableResponse(w, r)
		return
	}

	attemptCookie, err := r.Cookie(app.cfg.Auth.OIDCStateCookieName())
	if err != nil {
		app.authCallbackError(w, r)
		return
	}

	result, err := app.auth.Complete(r.Context(), r.URL.Query().Get("code"), r.URL.Query().Get("state"), attemptCookie.Value)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidAttempt) || errors.Is(err, auth.ErrInvalidIdentity) {
			app.authCallbackError(w, r)
			return
		}

		app.serverErrorResponse(w, r, err)
		return
	}

	app.setSessionCookie(w, result.SessionToken)
	app.setCSRFCookie(w, result.CSRFToken)
	http.Redirect(w, r, app.cfg.Auth.WebOrigin, http.StatusFound)
}

func (app *app) authSession(w http.ResponseWriter, r *http.Request) {
	if app.auth == nil {
		app.writeAnonymousSession(w, r)
		return
	}

	sessionCookie, err := r.Cookie(app.cfg.Auth.SessionCookieName)
	if err != nil {
		app.writeAnonymousSession(w, r)
		return
	}
	csrfCookie, _ := r.Cookie(app.cfg.Auth.CSRFCookieName)
	csrfToken := ""
	if csrfCookie != nil {
		csrfToken = csrfCookie.Value
	}

	result, err := app.auth.GetSession(r.Context(), sessionCookie.Value, csrfToken)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthenticated) {
			app.clearSessionCookie(w)
			app.clearCSRFCookie(w)
			app.writeAnonymousSession(w, r)
			return
		}

		app.serverErrorResponse(w, r, err)
		return
	}

	app.setCSRFCookie(w, result.CSRFToken)
	data := envelope{
		"authenticated": true,
		"user":          userResponse(result.User),
		"roles":         []string{},
		"agent":         nil,
		"csrf_token":    result.CSRFToken,
	}
	if err := writeJSON(w, http.StatusOK, data, nil); err != nil {
		app.logger.Error("write auth session response", "error", err)
	}
}

func (app *app) authLogout(w http.ResponseWriter, r *http.Request) {
	if app.auth == nil {
		app.authUnavailableResponse(w, r)
		return
	}

	if r.Header.Get("Origin") != app.cfg.Auth.WebOrigin {
		app.errorResponse(w, r, http.StatusForbidden, "forbidden", "The request origin is not allowed")
		return
	}

	sessionCookie, err := r.Cookie(app.cfg.Auth.SessionCookieName)
	if err != nil {
		app.errorResponse(w, r, http.StatusUnauthorized, "unauthenticated", "Authentication is required")
		return
	}

	if err := app.auth.Logout(r.Context(), sessionCookie.Value, r.Header.Get("X-CSRF-Token")); err != nil {
		switch {
		case errors.Is(err, auth.ErrUnauthenticated):
			app.errorResponse(w, r, http.StatusUnauthorized, "unauthenticated", "Authentication is required")
		case errors.Is(err, auth.ErrInvalidCSRF):
			app.errorResponse(w, r, http.StatusForbidden, "forbidden", "A valid CSRF token is required")
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	app.clearSessionCookie(w)
	app.clearCSRFCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (app *app) writeAnonymousSession(w http.ResponseWriter, r *http.Request) {
	data := envelope{
		"authenticated": false,
		"user":          nil,
		"roles":         []string{},
		"agent":         nil,
		"csrf_token":    nil,
	}
	if err := writeJSON(w, http.StatusOK, data, nil); err != nil {
		app.logger.Error("write anonymous session response", "error", err)
	}
}

func (app *app) authUnavailableResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, http.StatusServiceUnavailable, "auth_unavailable", "Authentication is not configured")
}

func (app *app) authCallbackError(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, http.StatusBadRequest, "invalid_auth_callback", "The Google sign-in attempt is invalid or has expired")
}

func (app *app) setAuthAttemptCookie(w http.ResponseWriter, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     app.cfg.Auth.OIDCStateCookieName(),
		Value:    value,
		Path:     "/",
		MaxAge:   int(app.cfg.Auth.OIDCStateDuration.Seconds()),
		HttpOnly: true,
		Secure:   app.cfg.Auth.SecureCookies,
		SameSite: http.SameSiteLaxMode,
	})
}

func (app *app) setSessionCookie(w http.ResponseWriter, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     app.cfg.Auth.SessionCookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   int(app.cfg.Auth.SessionDuration.Seconds()),
		HttpOnly: true,
		Secure:   app.cfg.Auth.SecureCookies,
		SameSite: http.SameSiteLaxMode,
	})
}

func (app *app) setCSRFCookie(w http.ResponseWriter, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     app.cfg.Auth.CSRFCookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   int(app.cfg.Auth.SessionDuration.Seconds()),
		Secure:   app.cfg.Auth.SecureCookies,
		SameSite: http.SameSiteLaxMode,
	})
}

func (app *app) clearAuthAttemptCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: app.cfg.Auth.OIDCStateCookieName(), Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: app.cfg.Auth.SecureCookies, SameSite: http.SameSiteLaxMode})
}

func (app *app) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: app.cfg.Auth.SessionCookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: app.cfg.Auth.SecureCookies, SameSite: http.SameSiteLaxMode})
}

func (app *app) clearCSRFCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: app.cfg.Auth.CSRFCookieName, Value: "", Path: "/", MaxAge: -1, Secure: app.cfg.Auth.SecureCookies, SameSite: http.SameSiteLaxMode})
}

func userResponse(user domain.User) envelope {
	return envelope{
		"id":           user.ID,
		"email":        user.Email,
		"display_name": user.DisplayName,
		"status":       user.Status,
	}
}
