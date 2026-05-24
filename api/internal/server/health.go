package server

import "net/http"

func (app *app) healthz(w http.ResponseWriter, r *http.Request) {
	data := envelope{
		"status":      "ok",
		"environment": app.cfg.Env,
		"version":     app.version,
	}

	if err := writeJSON(w, http.StatusOK, data, nil); err != nil {
		app.logger.Error("write health response", "error", err)
	}
}
