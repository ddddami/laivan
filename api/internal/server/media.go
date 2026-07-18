package server

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/ddddami/laivan/internal/domain"
	"github.com/ddddami/laivan/internal/repo"
	"github.com/ddddami/laivan/internal/storage"
	"github.com/ddddami/laivan/internal/validator"
)

const mediaFilesFormKey = "files"

var allowedImageContentTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

func (app *app) uploadMedia(w http.ResponseWriter, r *http.Request) {
	if app.mediaUploader == nil {
		app.serverErrorResponse(w, r, errors.New("media storage not available"))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, app.cfg.Media.MaxUploadBytes+(1<<20))
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("request body must be multipart form-data within %d MB", app.cfg.Media.MaxUploadBytes/(1<<20)))
		return
	}

	propertyID := strings.TrimSpace(r.FormValue("property_id"))
	propertyUnitTypeID := strings.TrimSpace(r.FormValue("property_unit_type_id"))
	agentOfferID := strings.TrimSpace(r.FormValue("agent_offer_id"))
	uploadedByAgentID := strings.TrimSpace(r.FormValue("uploaded_by_agent_id"))
	caption := strings.TrimSpace(r.FormValue("caption"))

	v := validator.New()
	validateMediaUploadTarget(v, propertyID, propertyUnitTypeID, agentOfferID)
	v.Check(validator.NotBlank(uploadedByAgentID), "uploaded_by_agent_id", "Uploaded by agent ID is required")
	v.Check(validator.ValidUUID(uploadedByAgentID), "uploaded_by_agent_id", "Uploaded by agent ID must be a valid UUID")
	v.Check(validator.MaxChars(caption, 500), "caption", "Caption must not exceed 500 characters")

	files := r.MultipartForm.File[mediaFilesFormKey]
	v.Check(len(files) > 0, mediaFilesFormKey, "At least one image file is required")

	if !v.Valid() {
		app.validationFailedResponse(w, r, v.FieldErrors)
		return
	}

	targetType, targetID := mediaTarget(propertyID, propertyUnitTypeID, agentOfferID)
	validatedFiles := make([]validatedMediaFile, 0, len(files))
	for _, fileHeader := range files {
		data, contentType, err := readMediaFile(fileHeader, app.cfg.Media.MaxUploadBytes)
		if err != nil {
			var validationErr mediaValidationError
			if errors.As(err, &validationErr) {
				app.validationFailedResponse(w, r, map[string]string{mediaFilesFormKey: validationErr.Error()})
				return
			}

			app.serverErrorResponse(w, r, fmt.Errorf("read media file: %w", err))
			return
		}
		validatedFiles = append(validatedFiles, validatedMediaFile{
			Filename:    fileHeader.Filename,
			Data:        data,
			ContentType: contentType,
		})
	}

	created := make([]domain.Media, 0, len(files))
	for _, file := range validatedFiles {
		media, err := app.processMediaFile(r.Context(), file, mediaUploadTarget{
			TargetType:         targetType,
			TargetID:           targetID,
			PropertyID:         propertyID,
			PropertyUnitTypeID: propertyUnitTypeID,
			AgentOfferID:       agentOfferID,
			UploadedByAgentID:  uploadedByAgentID,
			Caption:            caption,
		})
		if err != nil {
			if errors.Is(err, repo.ErrForeignKeyViolation) {
				app.badRequestResponse(w, r, fmt.Errorf("referenced resource does not exist"))
				return
			}
			if errors.Is(err, repo.ErrNotFound) {
				app.notFoundResponse(w, r)
				return
			}

			app.serverErrorResponse(w, r, fmt.Errorf("upload media: %w", err))
			return
		}
		created = append(created, media)
	}

	data := envelope{"media": app.mediaListResponse(created)}
	if err := writeJSON(w, http.StatusCreated, data, nil); err != nil {
		app.logger.Error("write media upload response", "error", err)
	}
}

type mediaUploadTarget struct {
	TargetType         string
	TargetID           string
	PropertyID         string
	PropertyUnitTypeID string
	AgentOfferID       string
	UploadedByAgentID  string
	Caption            string
}

type mediaValidationError string

func (e mediaValidationError) Error() string {
	return string(e)
}

type validatedMediaFile struct {
	Filename    string
	Data        []byte
	ContentType string
}

func (app *app) processMediaFile(ctx context.Context, file validatedMediaFile, target mediaUploadTarget) (domain.Media, error) {
	objectID, err := randomUUIDString()
	if err != nil {
		return domain.Media{}, fmt.Errorf("generate media object id: %w", err)
	}
	objectKey := storage.MediaObjectKey(target.TargetType, target.TargetID, objectID, file.Filename)

	url, err := app.mediaUploader.Upload(ctx, storage.UploadInput{
		Key:         objectKey,
		Body:        bytes.NewReader(file.Data),
		ContentType: file.ContentType,
	})
	if err != nil {
		return domain.Media{}, fmt.Errorf("upload object: %w", err)
	}

	media, err := app.propertyRepo.CreateMedia(ctx, domain.Media{
		PropertyID:         domain.ID(target.PropertyID),
		PropertyUnitTypeID: domain.ID(target.PropertyUnitTypeID),
		AgentOfferID:       domain.ID(target.AgentOfferID),
		UploadedByAgentID:  domain.ID(target.UploadedByAgentID),
		URL:                url,
		ObjectKey:          objectKey,
		Kind:               domain.MediaKindImage,
		Caption:            target.Caption,
		ContentType:        file.ContentType,
		SizeBytes:          int64(len(file.Data)),
	})
	if err != nil {
		app.logger.Error("media object orphaned after database insert failed",
			"object_key", objectKey,
			"target_type", target.TargetType,
			"target_id", target.TargetID,
			"error", err,
		)
		return domain.Media{}, err
	}

	return media, nil
}

func readMediaFile(fileHeader *multipart.FileHeader, maxBytes int64) ([]byte, string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, "", fmt.Errorf("open uploaded file: %w", err)
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		return nil, "", fmt.Errorf("read uploaded file: %w", err)
	}
	if int64(len(data)) > maxBytes {
		return nil, "", mediaValidationError(fmt.Sprintf("Each image file must not exceed %d bytes", maxBytes))
	}
	if len(data) == 0 {
		return nil, "", mediaValidationError("Image file must not be empty")
	}

	contentType := http.DetectContentType(data)
	if !allowedImageContentTypes[contentType] {
		return nil, "", mediaValidationError("Only JPEG, PNG, and WebP images are supported")
	}

	return data, contentType, nil
}

func validateMediaUploadTarget(v *validator.Validator, propertyID, propertyUnitTypeID, agentOfferID string) {
	targetCount := 0
	if propertyID != "" {
		targetCount++
		v.Check(validator.ValidUUID(propertyID), "property_id", "Property ID must be a valid UUID")
	}
	if propertyUnitTypeID != "" {
		targetCount++
		v.Check(validator.ValidUUID(propertyUnitTypeID), "property_unit_type_id", "Property unit type ID must be a valid UUID")
	}
	if agentOfferID != "" {
		targetCount++
		v.Check(validator.ValidUUID(agentOfferID), "agent_offer_id", "Agent offer ID must be a valid UUID")
	}
	v.Check(targetCount == 1, "target", "Exactly one media target is required")
}

func mediaTarget(propertyID, propertyUnitTypeID, agentOfferID string) (string, string) {
	if propertyID != "" {
		return "property", propertyID
	}
	if propertyUnitTypeID != "" {
		return "property_unit_type", propertyUnitTypeID
	}
	return "agent_offer", agentOfferID
}

func randomUUIDString() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	buf := make([]byte, 36)
	hex.Encode(buf[0:8], b[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], b[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], b[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], b[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:36], b[10:16])
	return string(buf), nil
}

func (app *app) mediaListResponse(items []domain.Media) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, app.mediaResponse(item))
	}
	return result
}

func (app *app) mediaResponse(media domain.Media) map[string]any {
	thumbnailURL := media.URL
	mediumURL := media.URL
	if app.mediaURLs != nil {
		thumbnailURL = app.mediaURLs.ThumbnailURL(media.URL)
		mediumURL = app.mediaURLs.MediumURL(media.URL)
	}

	return map[string]any{
		"id":                    string(media.ID),
		"property_id":           nullableID(media.PropertyID),
		"property_unit_type_id": nullableID(media.PropertyUnitTypeID),
		"agent_offer_id":        nullableID(media.AgentOfferID),
		"uploaded_by_agent_id":  nullableID(media.UploadedByAgentID),
		"url":                   media.URL,
		"thumbnail_url":         thumbnailURL,
		"medium_url":            mediumURL,
		"kind":                  string(media.Kind),
		"caption":               nullableString(media.Caption),
		"content_type":          nullableString(media.ContentType),
		"size_bytes":            media.SizeBytes,
		"created_at":            media.CreatedAt.Format(time.RFC3339),
	}
}

func nullableID(id domain.ID) any {
	if id == "" {
		return nil
	}
	return string(id)
}
