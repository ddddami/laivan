package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ddddami/laivan/internal/storage"
)

type stubUploader struct {
	input storage.UploadInput
}

func (s *stubUploader) Upload(ctx context.Context, input storage.UploadInput) (string, error) {
	s.input = input
	return "https://media.example.test/" + input.Key, nil
}

func TestUploadMediaValidationRequiresOneTarget(t *testing.T) {
	app := testAppWithRepo()
	app.mediaUploader = &stubUploader{}

	body, contentType := multipartBody(t, map[string]string{
		"property_id":           "550e8400-e29b-41d4-a716-446655440000",
		"property_unit_type_id": "550e8400-e29b-41d4-a716-446655440020",
		"uploaded_by_agent_id":  "550e8400-e29b-41d4-a716-446655440040",
	}, "room.jpg", tinyJPEG())
	req := httptest.NewRequest(http.MethodPost, "/v1/media", body)
	req.Header.Set("Content-Type", contentType)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}

	var decoded struct {
		Error struct {
			Fields map[string]string `json:"fields"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&decoded); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if decoded.Error.Fields["target"] == "" {
		t.Fatal("target validation error missing")
	}
}

func TestUploadMediaRejectsUnsupportedFileType(t *testing.T) {
	app := testAppWithRepo()
	app.mediaUploader = &stubUploader{}

	body, contentType := multipartBody(t, map[string]string{
		"property_id":          "550e8400-e29b-41d4-a716-446655440000",
		"uploaded_by_agent_id": "550e8400-e29b-41d4-a716-446655440040",
	}, "room.txt", []byte("not an image"))
	req := httptest.NewRequest(http.MethodPost, "/v1/media", body)
	req.Header.Set("Content-Type", contentType)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}
}

func TestUploadMediaCreatesMediaRecord(t *testing.T) {
	uploader := &stubUploader{}
	app := testAppWithRepo()
	app.mediaUploader = uploader

	body, contentType := multipartBody(t, map[string]string{
		"property_id":          "550e8400-e29b-41d4-a716-446655440000",
		"uploaded_by_agent_id": "550e8400-e29b-41d4-a716-446655440040",
		"caption":              "Front view",
	}, "room.jpg", tinyJPEG())
	req := httptest.NewRequest(http.MethodPost, "/v1/media", body)
	req.Header.Set("Content-Type", contentType)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusCreated)
	}
	if uploader.input.Key == "" {
		t.Fatal("upload key is empty")
	}
	if uploader.input.ContentType != "image/jpeg" {
		t.Fatalf("content type = %q, want image/jpeg", uploader.input.ContentType)
	}

	var decoded struct {
		Media []struct {
			URL          string `json:"url"`
			ThumbnailURL string `json:"thumbnail_url"`
			Caption      string `json:"caption"`
		} `json:"media"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&decoded); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(decoded.Media) != 1 {
		t.Fatalf("media length = %d, want 1", len(decoded.Media))
	}
	if decoded.Media[0].URL == "" || decoded.Media[0].ThumbnailURL == "" {
		t.Fatalf("media URL fields must be present: %#v", decoded.Media[0])
	}
	if decoded.Media[0].Caption != "Front view" {
		t.Fatalf("caption = %q, want Front view", decoded.Media[0].Caption)
	}
}

func TestUploadMediaRejectsEmptyFile(t *testing.T) {
	app := testAppWithRepo()
	app.mediaUploader = &stubUploader{}

	body, contentType := multipartBody(t, map[string]string{
		"property_id":          "550e8400-e29b-41d4-a716-446655440000",
		"uploaded_by_agent_id": "550e8400-e29b-41d4-a716-446655440040",
	}, "room.jpg", []byte{})
	req := httptest.NewRequest(http.MethodPost, "/v1/media", body)
	req.Header.Set("Content-Type", contentType)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}
}

func TestUploadMediaRejectsFileTooLarge(t *testing.T) {
	app := testAppWithRepo()
	app.cfg.Media.MaxUploadBytes = 100
	app.mediaUploader = &stubUploader{}

	body, contentType := multipartBody(t, map[string]string{
		"property_id":          "550e8400-e29b-41d4-a716-446655440000",
		"uploaded_by_agent_id": "550e8400-e29b-41d4-a716-446655440040",
	}, "room.jpg", make([]byte, 101))
	req := httptest.NewRequest(http.MethodPost, "/v1/media", body)
	req.Header.Set("Content-Type", contentType)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}
}

func TestUploadMediaRejectsMissingFilesField(t *testing.T) {
	app := testAppWithRepo()
	app.mediaUploader = &stubUploader{}

	body, contentType := multipartBody(t, map[string]string{
		"property_id":          "550e8400-e29b-41d4-a716-446655440000",
		"uploaded_by_agent_id": "550e8400-e29b-41d4-a716-446655440040",
	}, "", nil)
	req := httptest.NewRequest(http.MethodPost, "/v1/media", body)
	req.Header.Set("Content-Type", contentType)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}
}

func TestUploadMediaRejectsMissingAgentID(t *testing.T) {
	app := testAppWithRepo()
	app.mediaUploader = &stubUploader{}

	body, contentType := multipartBody(t, map[string]string{
		"property_id": "550e8400-e29b-41d4-a716-446655440000",
	}, "room.jpg", tinyJPEG())
	req := httptest.NewRequest(http.MethodPost, "/v1/media", body)
	req.Header.Set("Content-Type", contentType)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}

	var decoded struct {
		Error struct {
			Fields map[string]string `json:"fields"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&decoded); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if decoded.Error.Fields["uploaded_by_agent_id"] == "" {
		t.Fatal("uploaded_by_agent_id validation error missing")
	}
}

func TestUploadMediaReturnsServerErrorOnUploadFailure(t *testing.T) {
	app := testAppWithRepo()
	app.mediaUploader = &failUploader{err: errors.New("storage unavailable")}

	body, contentType := multipartBody(t, map[string]string{
		"property_id":          "550e8400-e29b-41d4-a716-446655440000",
		"uploaded_by_agent_id": "550e8400-e29b-41d4-a716-446655440040",
	}, "room.jpg", tinyJPEG())
	req := httptest.NewRequest(http.MethodPost, "/v1/media", body)
	req.Header.Set("Content-Type", contentType)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestUploadMediaRejectsCaptionTooLong(t *testing.T) {
	app := testAppWithRepo()
	app.mediaUploader = &stubUploader{}

	caption := strings.Repeat("a", 501)
	body, contentType := multipartBody(t, map[string]string{
		"property_id":          "550e8400-e29b-41d4-a716-446655440000",
		"uploaded_by_agent_id": "550e8400-e29b-41d4-a716-446655440040",
		"caption":              caption,
	}, "room.jpg", tinyJPEG())
	req := httptest.NewRequest(http.MethodPost, "/v1/media", body)
	req.Header.Set("Content-Type", contentType)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}

	var decoded struct {
		Error struct {
			Fields map[string]string `json:"fields"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&decoded); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if decoded.Error.Fields["caption"] == "" {
		t.Fatal("caption validation error missing")
	}
}

func TestUploadMediaRejectsInvalidTargetUUID(t *testing.T) {
	app := testAppWithRepo()
	app.mediaUploader = &stubUploader{}

	body, contentType := multipartBody(t, map[string]string{
		"property_id":          "not-a-uuid",
		"uploaded_by_agent_id": "550e8400-e29b-41d4-a716-446655440040",
	}, "room.jpg", tinyJPEG())
	req := httptest.NewRequest(http.MethodPost, "/v1/media", body)
	req.Header.Set("Content-Type", contentType)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}

	var decoded struct {
		Error struct {
			Fields map[string]string `json:"fields"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&decoded); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if decoded.Error.Fields["property_id"] == "" {
		t.Fatal("property_id validation error missing")
	}
}

type failUploader struct {
	err error
}

func (f *failUploader) Upload(ctx context.Context, input storage.UploadInput) (string, error) {
	return "", f.err
}

func multipartBody(t *testing.T, fields map[string]string, filename string, file []byte) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("write field: %v", err)
		}
	}
	if filename != "" || file != nil {
		part, err := writer.CreateFormFile(mediaFilesFormKey, filename)
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		if _, err := part.Write(file); err != nil {
			t.Fatalf("write file: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	return body, writer.FormDataContentType()
}

func tinyJPEG() []byte {
	return []byte{
		0xff, 0xd8, 0xff, 0xdb, 0x00, 0x43, 0x00,
		0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
		0xff, 0xd9,
	}
}
