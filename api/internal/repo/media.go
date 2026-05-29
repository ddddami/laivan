package repo

import (
	"context"
	"errors"
	"fmt"

	generateddb "github.com/ddddami/laivan/internal/db/generated"
	"github.com/ddddami/laivan/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (r *PropertyRepository) CreateMedia(ctx context.Context, media domain.Media) (domain.Media, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	uploadedByAgentID, err := uuidParam(media.UploadedByAgentID)
	if err != nil {
		return domain.Media{}, err
	}

	row, err := r.queries.CreateMedia(ctx, generateddb.CreateMediaParams{
		PropertyID:         optionalUUIDParam(media.PropertyID),
		PropertyUnitTypeID: optionalUUIDParam(media.PropertyUnitTypeID),
		AgentOfferID:       optionalUUIDParam(media.AgentOfferID),
		UploadedByAgentID:  uploadedByAgentID,
		Url:                media.URL,
		ObjectKey:          textParam(media.ObjectKey),
		Kind:               string(media.Kind),
		Caption:            textParam(media.Caption),
		ContentType:        textParam(media.ContentType),
		SizeBytes:          int32Param(media.SizeBytes),
	})
	if err != nil {
		if isForeignKeyViolation(err) {
			return domain.Media{}, ErrNotFound
		}
		return domain.Media{}, fmt.Errorf("create media: %w", err)
	}

	return mediaFromCreateRow(row), nil
}

func (r *PropertyRepository) ListMediaByProperty(ctx context.Context, propertyID domain.ID) ([]domain.Media, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	propertyUUID, err := uuidParam(propertyID)
	if err != nil {
		return nil, err
	}

	if _, err := r.queries.GetProperty(ctx, propertyUUID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get property for media: %w", err)
	}

	rows, err := r.queries.ListMediaByProperty(ctx, propertyUUID)
	if err != nil {
		return nil, fmt.Errorf("list property media: %w", err)
	}

	media := make([]domain.Media, 0, len(rows))
	for _, row := range rows {
		media = append(media, mediaFromPropertyRow(row))
	}
	return media, nil
}

func (r *PropertyRepository) ListMediaByPropertyUnitType(ctx context.Context, propertyUnitTypeID domain.ID) ([]domain.Media, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	unitTypeUUID, err := uuidParam(propertyUnitTypeID)
	if err != nil {
		return nil, err
	}

	if _, err := r.queries.GetPropertyUnitType(ctx, unitTypeUUID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get property unit type for media: %w", err)
	}

	rows, err := r.queries.ListMediaByPropertyUnitType(ctx, unitTypeUUID)
	if err != nil {
		return nil, fmt.Errorf("list property unit type media: %w", err)
	}

	media := make([]domain.Media, 0, len(rows))
	for _, row := range rows {
		media = append(media, mediaFromPropertyUnitTypeRow(row))
	}
	return media, nil
}

func (r *PropertyRepository) ListMediaByAgentOffer(ctx context.Context, agentOfferID domain.ID) ([]domain.Media, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	agentOfferUUID, err := uuidParam(agentOfferID)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListMediaByAgentOffer(ctx, agentOfferUUID)
	if err != nil {
		return nil, fmt.Errorf("list agent offer media: %w", err)
	}

	media := make([]domain.Media, 0, len(rows))
	for _, row := range rows {
		media = append(media, mediaFromAgentOfferRow(row))
	}
	return media, nil
}

func optionalUUIDParam(id domain.ID) pgtype.UUID {
	if id == "" {
		return pgtype.UUID{}
	}
	uuid, err := uuidParam(id)
	if err != nil {
		return pgtype.UUID{}
	}
	return uuid
}

func int32Param(value int64) pgtype.Int4 {
	if value == 0 {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: int32(value), Valid: true}
}

func int64String(value pgtype.Int4) int64 {
	if !value.Valid {
		return 0
	}
	return int64(value.Int32)
}

func mediaFromCreateRow(row generateddb.CreateMediaRow) domain.Media {
	return mediaFromFields(row.ID, row.PropertyID, row.PropertyUnitTypeID, row.AgentOfferID, row.UploadedByAgentID, row.Url, row.ObjectKey, row.Kind, row.Caption, row.ContentType, row.SizeBytes, row.CreatedAt)
}

func mediaFromPropertyRow(row generateddb.ListMediaByPropertyRow) domain.Media {
	return mediaFromFields(row.ID, row.PropertyID, row.PropertyUnitTypeID, row.AgentOfferID, row.UploadedByAgentID, row.Url, row.ObjectKey, row.Kind, row.Caption, row.ContentType, row.SizeBytes, row.CreatedAt)
}

func mediaFromPropertyUnitTypeRow(row generateddb.ListMediaByPropertyUnitTypeRow) domain.Media {
	return mediaFromFields(row.ID, row.PropertyID, row.PropertyUnitTypeID, row.AgentOfferID, row.UploadedByAgentID, row.Url, row.ObjectKey, row.Kind, row.Caption, row.ContentType, row.SizeBytes, row.CreatedAt)
}

func mediaFromAgentOfferRow(row generateddb.ListMediaByAgentOfferRow) domain.Media {
	return mediaFromFields(row.ID, row.PropertyID, row.PropertyUnitTypeID, row.AgentOfferID, row.UploadedByAgentID, row.Url, row.ObjectKey, row.Kind, row.Caption, row.ContentType, row.SizeBytes, row.CreatedAt)
}

func mediaFromFields(id, propertyID, propertyUnitTypeID, agentOfferID, uploadedByAgentID pgtype.UUID, url string, objectKey pgtype.Text, kind string, caption pgtype.Text, contentType pgtype.Text, sizeBytes pgtype.Int4, createdAt pgtype.Timestamptz) domain.Media {
	return domain.Media{
		ID:                 domain.ID(uuidString(id)),
		PropertyID:         domain.ID(uuidString(propertyID)),
		PropertyUnitTypeID: domain.ID(uuidString(propertyUnitTypeID)),
		AgentOfferID:       domain.ID(uuidString(agentOfferID)),
		UploadedByAgentID:  domain.ID(uuidString(uploadedByAgentID)),
		URL:                url,
		ObjectKey:          textString(objectKey),
		Kind:               domain.MediaKind(kind),
		Caption:            textString(caption),
		ContentType:        textString(contentType),
		SizeBytes:          int64String(sizeBytes),
		CreatedAt:          createdAt.Time,
	}
}
