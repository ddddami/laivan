package server

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func (app *app) docs(w http.ResponseWriter, r *http.Request) {
	path, err := repoFilePath("docs/api/index.html")
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("find API docs: %w", err))
		return
	}

	http.ServeFile(w, r, path)
}

func (app *app) openapi(w http.ResponseWriter, r *http.Request) {
	path, err := repoFilePath("openapi/openapi.yaml")
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("find OpenAPI contract: %w", err))
		return
	}

	w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
	http.ServeFile(w, r, path)
}

func repoFilePath(name string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for dir := wd; ; dir = filepath.Dir(dir) {
		candidate := filepath.Join(dir, name)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("file not found")
		}
	}
}
