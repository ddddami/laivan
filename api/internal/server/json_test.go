package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	headers := http.Header{"X-Test": []string{"ok"}}

	err := writeJSON(rr, http.StatusCreated, envelope{"status": "created"}, headers)
	if err != nil {
		t.Fatalf("writeJSON returned error: %v", err)
	}

	if rr.Code != http.StatusCreated {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusCreated)
	}

	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	if got := rr.Header().Get("X-Test"); got != "ok" {
		t.Fatalf("X-Test = %q, want ok", got)
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if body.Status != "created" {
		t.Fatalf("status = %q, want created", body.Status)
	}
}

func TestReadJSON(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantName string
		wantErr  string
	}{
		{name: "valid", body: `{"name":"Alice Lodge"}`, wantName: "Alice Lodge"},
		{name: "malformed", body: `{"name":`, wantErr: "request body contains badly-formed JSON"},
		{name: "unknown field", body: `{"name":"Alice Lodge","extra":"ignored"}`, wantErr: "request body contains unknown field \"extra\""},
		{name: "wrong type", body: `{"name":42}`, wantErr: "request body contains an incorrect JSON type for field \"name\""},
		{name: "empty", body: ``, wantErr: "request body must not be empty"},
		{name: "multiple values", body: `{"name":"Alice Lodge"} {"name":"Other"}`, wantErr: "request body must only contain a single JSON value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input struct {
				Name string `json:"name"`
			}

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			err := readJSON(rr, req, &input)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("readJSON returned nil error")
				}

				if err.Error() != tt.wantErr {
					t.Fatalf("error = %q, want %q", err.Error(), tt.wantErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("readJSON returned error: %v", err)
			}

			if input.Name != tt.wantName {
				t.Fatalf("name = %q, want %q", input.Name, tt.wantName)
			}
		})
	}
}

func TestReadJSONRejectsLargeBody(t *testing.T) {
	var input struct {
		Name string `json:"name"`
	}

	body := `{"name":"` + strings.Repeat("a", maxJSONBodySize) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rr := httptest.NewRecorder()

	err := readJSON(rr, req, &input)
	if err == nil {
		t.Fatal("readJSON returned nil error")
	}

	want := "request body must not be larger than 1048576 bytes"
	if err.Error() != want {
		t.Fatalf("error = %q, want %q", err.Error(), want)
	}
}
