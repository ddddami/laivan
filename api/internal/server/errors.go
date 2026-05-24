package server

import "net/http"

type errorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func (app *app) errorResponse(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	data := envelope{"error": errorBody{Code: code, Message: message}}

	if err := writeJSON(w, status, data, nil); err != nil {
		app.logger.Error("write error response", "method", r.Method, "path", r.URL.Path, "error", err)
	}
}

func (app *app) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, "bad_request", err.Error())
}

func (app *app) validationFailedResponse(w http.ResponseWriter, r *http.Request, fieldErrors map[string]string) {
	data := envelope{
		"error": errorBody{
			Code:    "validation_failed",
			Message: "Request validation failed",
			Fields:  fieldErrors,
		},
	}

	if err := writeJSON(w, http.StatusUnprocessableEntity, data, nil); err != nil {
		app.logger.Error("write validation error response", "method", r.Method, "path", r.URL.Path, "error", err)
	}
}

func (app *app) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, http.StatusNotFound, "not_found", "The requested resource could not be found")
}

func (app *app) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "The request method is not supported for this resource")
}

func (app *app) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Error("internal server error", "method", r.Method, "path", r.URL.Path, "error", err)
	app.errorResponse(w, r, http.StatusInternalServerError, "internal_server_error", "The server encountered a problem and could not process your request")
}
