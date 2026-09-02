package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ddddami/laivan/internal/domain"
	"github.com/ddddami/laivan/internal/repo"
	"github.com/ddddami/laivan/internal/storage"
)

type stubUploader struct {
	input        storage.UploadInput
	inputs       []storage.UploadInput
	deletedKeys  []string
	failUploadAt int
	deleteErr    error
}

func TestUploadMediaRequiresAuthentication(t *testing.T) {
	app := testApp()
	app.mediaUploader = &stubUploader{}
	body, contentType := multipartBody(t, map[string]string{
		"property_id": "550e8400-e29b-41d4-a716-446655440000",
	}, "room.jpg", tinyJPEG())
	req := httptest.NewRequest(http.MethodPost, "/v1/media", body)
	req.Header.Set("Content-Type", contentType)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusUnauthorized, "unauthenticated", "Authentication is required")
}

func (s *stubUploader) Upload(ctx context.Context, input storage.UploadInput) (string, error) {
	if s.failUploadAt == len(s.inputs)+1 {
		return "", errors.New("storage unavailable")
	}
	s.input = input
	s.inputs = append(s.inputs, input)
	return "https://media.example.test/" + input.Key, nil
}

func (s *stubUploader) Delete(ctx context.Context, key string) error {
	s.deletedKeys = append(s.deletedKeys, key)
	return s.deleteErr
}

func TestUploadMediaValidationRequiresOneTarget(t *testing.T) {
	app := testAppWithActiveAgentRepo()
	app.mediaUploader = &stubUploader{}

	body, contentType := multipartBody(t, map[string]string{
		"property_id":           "550e8400-e29b-41d4-a716-446655440000",
		"property_unit_type_id": "550e8400-e29b-41d4-a716-446655440020",
		"uploaded_by_agent_id":  "550e8400-e29b-41d4-a716-446655440040",
	}, "room.jpg", tinyJPEG())
	req := authenticatedMultipartRequest(body, contentType)
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
	app := testAppWithActiveAgentRepo()
	app.mediaUploader = &stubUploader{}

	body, contentType := multipartBody(t, map[string]string{
		"property_id":          "550e8400-e29b-41d4-a716-446655440000",
		"uploaded_by_agent_id": "550e8400-e29b-41d4-a716-446655440040",
	}, "room.txt", []byte("not an image"))
	req := authenticatedMultipartRequest(body, contentType)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}
}

func TestUploadMediaValidatesAllFilesBeforeUploading(t *testing.T) {
	uploader := &stubUploader{}
	repository := &spyPropertyRepo{stub: &stubPropertyRepo{}}
	app := testAppWithActiveAgentRepo()
	app.propertyRepo = repository
	app.mediaUploader = uploader

	body, contentType := multipartBodyFiles(t, map[string]string{
		"property_id":          "550e8400-e29b-41d4-a716-446655440000",
		"uploaded_by_agent_id": "550e8400-e29b-41d4-a716-446655440040",
	}, []multipartTestFile{
		{filename: "room.jpg", data: tinyJPEG()},
		{filename: "room.gif", data: []byte("GIF89a")},
	})
	req := authenticatedMultipartRequest(body, contentType)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}
	if len(uploader.inputs) != 0 {
		t.Fatalf("uploads = %d, want 0", len(uploader.inputs))
	}
	if len(repository.createdMedia) != 0 {
		t.Fatalf("created media = %d, want 0", len(repository.createdMedia))
	}
}

func TestUploadMediaCreatesMediaRecord(t *testing.T) {
	uploader := &stubUploader{}
	app := testAppWithActiveAgentRepo()
	app.mediaUploader = uploader

	body, contentType := multipartBody(t, map[string]string{
		"property_id":          "550e8400-e29b-41d4-a716-446655440000",
		"uploaded_by_agent_id": "550e8400-e29b-41d4-a716-446655440040",
		"caption":              "Front view",
	}, "room.jpg", tinyJPEG())
	req := authenticatedMultipartRequest(body, contentType)
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

func TestUploadMediaCreatesAllMediaRecordsInRequestOrder(t *testing.T) {
	uploader := &stubUploader{}
	repository := &spyPropertyRepo{stub: &stubPropertyRepo{}}
	app := testAppWithActiveAgentRepo()
	app.propertyRepo = repository
	app.mediaUploader = uploader

	body, contentType := multipartBodyFiles(t, map[string]string{
		"property_id":          "550e8400-e29b-41d4-a716-446655440000",
		"uploaded_by_agent_id": "550e8400-e29b-41d4-a716-446655440040",
	}, []multipartTestFile{
		{filename: "first.jpg", data: tinyJPEG()},
		{filename: "second.jpg", data: tinyJPEG()},
	})
	req := authenticatedMultipartRequest(body, contentType)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusCreated)
	}
	if len(repository.createdMedia) != 2 {
		t.Fatalf("created media = %d, want 2", len(repository.createdMedia))
	}

	var decoded struct {
		Media []struct {
			URL string `json:"url"`
		} `json:"media"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&decoded); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(decoded.Media) != 2 {
		t.Fatalf("media length = %d, want 2", len(decoded.Media))
	}
	for i, input := range uploader.inputs {
		want := "https://media.example.test/" + input.Key
		if decoded.Media[i].URL != want {
			t.Fatalf("media URL %d = %q, want %q", i, decoded.Media[i].URL, want)
		}
	}
}

func TestUploadMediaRemovesUploadedObjectsAfterLaterUploadFails(t *testing.T) {
	uploader := &stubUploader{failUploadAt: 2}
	repository := &spyPropertyRepo{stub: &stubPropertyRepo{}}
	app := testAppWithActiveAgentRepo()
	app.propertyRepo = repository
	app.mediaUploader = uploader

	body, contentType := multipartBodyFiles(t, map[string]string{
		"property_id":          "550e8400-e29b-41d4-a716-446655440000",
		"uploaded_by_agent_id": "550e8400-e29b-41d4-a716-446655440040",
	}, []multipartTestFile{
		{filename: "first.jpg", data: tinyJPEG()},
		{filename: "second.jpg", data: tinyJPEG()},
	})
	req := authenticatedMultipartRequest(body, contentType)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
	if len(repository.createdMedia) != 0 {
		t.Fatalf("created media = %d, want 0", len(repository.createdMedia))
	}
	if len(uploader.deletedKeys) != 1 || uploader.deletedKeys[0] != uploader.inputs[0].Key {
		t.Fatalf("deleted keys = %#v, want uploaded object %q", uploader.deletedKeys, uploader.inputs[0].Key)
	}
}

func TestUploadMediaRemovesUploadedObjectsWhenPersistenceFails(t *testing.T) {
	uploader := &stubUploader{}
	app := testAppWithActiveAgentRepo()
	app.propertyRepo = &fkViolationRepo{stub: &stubPropertyRepo{}}
	app.mediaUploader = uploader

	body, contentType := multipartBodyFiles(t, map[string]string{
		"property_id":          "550e8400-e29b-41d4-a716-446655440000",
		"uploaded_by_agent_id": "550e8400-e29b-41d4-a716-446655440040",
	}, []multipartTestFile{
		{filename: "first.jpg", data: tinyJPEG()},
		{filename: "second.jpg", data: tinyJPEG()},
	})
	req := authenticatedMultipartRequest(body, contentType)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusBadRequest)
	}
	if len(uploader.deletedKeys) != 2 {
		t.Fatalf("deleted keys = %#v, want 2", uploader.deletedKeys)
	}
	for i, input := range uploader.inputs {
		if uploader.deletedKeys[i] != input.Key {
			t.Fatalf("deleted key %d = %q, want %q", i, uploader.deletedKeys[i], input.Key)
		}
	}
}

func TestUploadMediaRejectsEmptyFile(t *testing.T) {
	app := testAppWithActiveAgentRepo()
	app.mediaUploader = &stubUploader{}

	body, contentType := multipartBody(t, map[string]string{
		"property_id":          "550e8400-e29b-41d4-a716-446655440000",
		"uploaded_by_agent_id": "550e8400-e29b-41d4-a716-446655440040",
	}, "room.jpg", []byte{})
	req := authenticatedMultipartRequest(body, contentType)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}
}

func TestUploadMediaRejectsFileTooLarge(t *testing.T) {
	app := testAppWithActiveAgentRepo()
	app.cfg.Media.MaxUploadBytes = 100
	app.mediaUploader = &stubUploader{}

	body, contentType := multipartBody(t, map[string]string{
		"property_id":          "550e8400-e29b-41d4-a716-446655440000",
		"uploaded_by_agent_id": "550e8400-e29b-41d4-a716-446655440040",
	}, "room.jpg", make([]byte, 101))
	req := authenticatedMultipartRequest(body, contentType)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}
}

func TestUploadMediaRejectsMissingFilesField(t *testing.T) {
	app := testAppWithActiveAgentRepo()
	app.mediaUploader = &stubUploader{}

	body, contentType := multipartBody(t, map[string]string{
		"property_id":          "550e8400-e29b-41d4-a716-446655440000",
		"uploaded_by_agent_id": "550e8400-e29b-41d4-a716-446655440040",
	}, "", nil)
	req := authenticatedMultipartRequest(body, contentType)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}
}

func TestUploadMediaRejectsTooManyFiles(t *testing.T) {
	app := testAppWithActiveAgentRepo()
	app.mediaUploader = &stubUploader{}

	files := make([]multipartTestFile, 11)
	for i := range files {
		files[i] = multipartTestFile{filename: "room.jpg", data: tinyJPEG()}
	}
	body, contentType := multipartBodyFiles(t, map[string]string{
		"property_id":          "550e8400-e29b-41d4-a716-446655440000",
		"uploaded_by_agent_id": "550e8400-e29b-41d4-a716-446655440040",
	}, files)
	req := authenticatedMultipartRequest(body, contentType)
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
	want := "At most 10 image files are allowed per upload"
	if decoded.Error.Fields[mediaFilesFormKey] != want {
		t.Fatalf("files error = %q, want %q", decoded.Error.Fields[mediaFilesFormKey], want)
	}
}

func TestUploadMediaDerivesAgentProvenanceFromSession(t *testing.T) {
	repository := &spyPropertyRepo{stub: &stubPropertyRepo{}}
	app := testAppWithActiveAgentRepo()
	app.propertyRepo = repository
	app.mediaUploader = &stubUploader{}

	body, contentType := multipartBody(t, map[string]string{
		"property_id":          "550e8400-e29b-41d4-a716-446655440000",
		"uploaded_by_agent_id": "550e8400-e29b-41d4-a716-446655440099",
	}, "room.jpg", tinyJPEG())
	req := authenticatedMultipartRequest(body, contentType)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusCreated)
	}
	if len(repository.createdMedia) != 1 {
		t.Fatalf("created media = %d, want 1", len(repository.createdMedia))
	}
	if got := repository.createdMedia[0].UploadedByAgentID; got != "550e8400-e29b-41d4-a716-446655440040" {
		t.Fatalf("uploaded by agent ID = %q, want authenticated agent", got)
	}
}

func TestUploadMediaRejectsAnotherAgentsOffer(t *testing.T) {
	uploader := &stubUploader{}
	app := authenticatedTestApp(domain.ID("550e8400-e29b-41d4-a716-446655440001"), &fakeAgentApplicationStore{access: domain.EffectiveAccess{
		Agent:          &domain.LinkedAgent{ID: domain.ID("550e8400-e29b-41d4-a716-446655440041"), Status: domain.AgentStatusActive},
		AgentCampusIDs: []domain.ID{"550e8400-e29b-41d4-a716-446655440002"},
	}})
	app.propertyRepo = &spyPropertyRepo{stub: &stubPropertyRepo{}}
	app.mediaUploader = uploader
	body, contentType := multipartBody(t, map[string]string{
		"agent_offer_id": "550e8400-e29b-41d4-a716-446655440030",
	}, "room.jpg", tinyJPEG())
	req := authenticatedMultipartRequest(body, contentType)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusForbidden, "forbidden", "You are not authorized to add media to this resource")
	if len(uploader.inputs) != 0 {
		t.Fatalf("uploads = %d, want 0", len(uploader.inputs))
	}
}

func TestUploadMediaReturnsServerErrorOnUploadFailure(t *testing.T) {
	app := testAppWithActiveAgentRepo()
	app.mediaUploader = &failUploader{err: errors.New("storage unavailable")}

	body, contentType := multipartBody(t, map[string]string{
		"property_id":          "550e8400-e29b-41d4-a716-446655440000",
		"uploaded_by_agent_id": "550e8400-e29b-41d4-a716-446655440040",
	}, "room.jpg", tinyJPEG())
	req := authenticatedMultipartRequest(body, contentType)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestUploadMediaRejectsCaptionTooLong(t *testing.T) {
	app := testAppWithActiveAgentRepo()
	app.mediaUploader = &stubUploader{}

	caption := strings.Repeat("a", 501)
	body, contentType := multipartBody(t, map[string]string{
		"property_id":          "550e8400-e29b-41d4-a716-446655440000",
		"uploaded_by_agent_id": "550e8400-e29b-41d4-a716-446655440040",
		"caption":              caption,
	}, "room.jpg", tinyJPEG())
	req := authenticatedMultipartRequest(body, contentType)
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

func TestUploadMediaReturnsBadRequestOnForeignKeyViolation(t *testing.T) {
	app := testAppWithActiveAgentRepo()
	app.propertyRepo = &fkViolationRepo{stub: &stubPropertyRepo{}}
	app.mediaUploader = &stubUploader{}

	body, contentType := multipartBody(t, map[string]string{
		"property_id":          "550e8400-e29b-41d4-a716-446655440000",
		"uploaded_by_agent_id": "550e8400-e29b-41d4-a716-446655440040",
	}, "room.jpg", tinyJPEG())
	req := authenticatedMultipartRequest(body, contentType)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	assertErrorResponse(t, rr, http.StatusBadRequest, "bad_request", "referenced resource does not exist")
}

func TestUploadMediaRejectsInvalidTargetUUID(t *testing.T) {
	app := testAppWithActiveAgentRepo()
	app.mediaUploader = &stubUploader{}

	body, contentType := multipartBody(t, map[string]string{
		"property_id":          "not-a-uuid",
		"uploaded_by_agent_id": "550e8400-e29b-41d4-a716-446655440040",
	}, "room.jpg", tinyJPEG())
	req := authenticatedMultipartRequest(body, contentType)
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

func (f *failUploader) Delete(ctx context.Context, key string) error {
	return f.err
}

type fkViolationRepo struct {
	stub *stubPropertyRepo
}

func (s *fkViolationRepo) GetCampusBySlug(ctx context.Context, slug string) (domain.Campus, error) {
	return s.stub.GetCampusBySlug(ctx, slug)
}

func (s *fkViolationRepo) Create(ctx context.Context, property domain.Property) (domain.Property, error) {
	return s.stub.Create(ctx, property)
}
func (s *fkViolationRepo) Get(ctx context.Context, id domain.ID) (domain.Property, error) {
	return s.stub.Get(ctx, id)
}
func (s *fkViolationRepo) Update(ctx context.Context, id domain.ID, expectedVersion int, patch domain.PropertyPatch) (domain.Property, error) {
	return s.stub.Update(ctx, id, expectedVersion, patch)
}
func (s *fkViolationRepo) GetWithDetails(ctx context.Context, id domain.ID) (domain.PropertyDetail, error) {
	return s.stub.GetWithDetails(ctx, id)
}
func (s *fkViolationRepo) GetMediaTarget(ctx context.Context, targetType string, id domain.ID) (repo.MediaTarget, error) {
	return s.stub.GetMediaTarget(ctx, targetType, id)
}
func (s *fkViolationRepo) ListWithSummary(ctx context.Context, filter repo.PropertyListFilter) ([]domain.PropertySummary, int, error) {
	return s.stub.ListWithSummary(ctx, filter)
}
func (s *fkViolationRepo) Discover(ctx context.Context, filter repo.DiscoveryFilter) ([]domain.DiscoveryResult, int, error) {
	return s.stub.Discover(ctx, filter)
}
func (s *fkViolationRepo) CreateMedia(ctx context.Context, media domain.Media) (domain.Media, error) {
	return domain.Media{}, repo.ErrForeignKeyViolation
}
func (s *fkViolationRepo) CreateMediaBatch(ctx context.Context, media []domain.Media) ([]domain.Media, error) {
	return nil, repo.ErrForeignKeyViolation
}
func (s *fkViolationRepo) ListMediaByProperty(ctx context.Context, propertyID domain.ID) ([]domain.Media, error) {
	return s.stub.ListMediaByProperty(ctx, propertyID)
}
func (s *fkViolationRepo) ListMediaByPropertyUnitType(ctx context.Context, propertyUnitTypeID domain.ID) ([]domain.Media, error) {
	return s.stub.ListMediaByPropertyUnitType(ctx, propertyUnitTypeID)
}
func (s *fkViolationRepo) ListMediaByAgentOffer(ctx context.Context, agentOfferID domain.ID) ([]domain.Media, error) {
	return s.stub.ListMediaByAgentOffer(ctx, agentOfferID)
}
func (s *fkViolationRepo) CreatePropertyUnitType(ctx context.Context, unitType domain.PropertyUnitType) (domain.PropertyUnitType, error) {
	return s.stub.CreatePropertyUnitType(ctx, unitType)
}
func (s *fkViolationRepo) ListPropertyUnitTypes(ctx context.Context, propertyID domain.ID) ([]domain.PropertyUnitType, error) {
	return s.stub.ListPropertyUnitTypes(ctx, propertyID)
}
func (s *fkViolationRepo) CreateAgentOffer(ctx context.Context, offer domain.AgentOffer) (domain.AgentOffer, error) {
	return s.stub.CreateAgentOffer(ctx, offer)
}
func (s *fkViolationRepo) ListAgentOffers(ctx context.Context, unitTypeID domain.ID) ([]domain.AgentOffer, error) {
	return s.stub.ListAgentOffers(ctx, unitTypeID)
}

func multipartBody(t *testing.T, fields map[string]string, filename string, file []byte) (*bytes.Buffer, string) {
	t.Helper()

	files := []multipartTestFile(nil)
	if filename != "" || file != nil {
		files = append(files, multipartTestFile{filename: filename, data: file})
	}
	return multipartBodyFiles(t, fields, files)
}

func authenticatedMultipartRequest(body io.Reader, contentType string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "/v1/media", body)
	request.AddCookie(&http.Cookie{Name: "laivan_session", Value: "session-token"})
	request.AddCookie(&http.Cookie{Name: "laivan_csrf", Value: "csrf-token"})
	request.Header.Set("Origin", "http://localhost:5173")
	request.Header.Set("X-CSRF-Token", "csrf-token")
	request.Header.Set("Content-Type", contentType)
	return request
}

type multipartTestFile struct {
	filename string
	data     []byte
}

func multipartBodyFiles(t *testing.T, fields map[string]string, files []multipartTestFile) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("write field: %v", err)
		}
	}
	for _, file := range files {
		part, err := writer.CreateFormFile(mediaFilesFormKey, file.filename)
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		if _, err := part.Write(file.data); err != nil {
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
