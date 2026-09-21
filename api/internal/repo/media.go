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

type MediaRemovalTarget struct {
	MediaID            domain.ID
	ObjectKey          string
	PropertyID         domain.ID
	PropertyUnitTypeID domain.ID
	AgentOfferID       domain.ID
	TargetType         string
	CampusID           domain.ID
	AgentID            domain.ID
}

func (r *PropertyRepository) CreateMedia(ctx context.Context, media domain.Media) (domain.Media, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	params, err := mediaCreateParams(media)
	if err != nil {
		return domain.Media{}, err
	}

	row, err := r.queries.CreateMedia(ctx, params)
	if err != nil {
		return domain.Media{}, createMediaError(err)
	}

	return mediaFromCreateRow(row), nil
}

func (r *PropertyRepository) CreateMediaBatch(ctx context.Context, media []domain.Media) ([]domain.Media, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	params := make([]generateddb.CreateMediaParams, 0, len(media))
	for _, item := range media {
		param, err := mediaCreateParams(item)
		if err != nil {
			return nil, err
		}
		params = append(params, param)
	}

	beginner, ok := r.db.(interface {
		Begin(context.Context) (pgx.Tx, error)
	})
	if !ok {
		return nil, errors.New("database does not support transactions")
	}

	tx, err := beginner.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin media transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	queries := r.queries.WithTx(tx)
	created := make([]domain.Media, 0, len(params))
	for _, param := range params {
		row, err := queries.CreateMedia(ctx, param)
		if err != nil {
			return nil, createMediaError(err)
		}
		created = append(created, mediaFromCreateRow(row))
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit media transaction: %w", err)
	}

	return created, nil
}

func mediaCreateParams(media domain.Media) (generateddb.CreateMediaParams, error) {
	propertyID, err := optionalUUIDParam(media.PropertyID)
	if err != nil {
		return generateddb.CreateMediaParams{}, fmt.Errorf("invalid property id: %w", err)
	}
	propertyUnitTypeID, err := optionalUUIDParam(media.PropertyUnitTypeID)
	if err != nil {
		return generateddb.CreateMediaParams{}, fmt.Errorf("invalid property unit type id: %w", err)
	}
	agentOfferID, err := optionalUUIDParam(media.AgentOfferID)
	if err != nil {
		return generateddb.CreateMediaParams{}, fmt.Errorf("invalid agent offer id: %w", err)
	}
	uploadedByAgentID, err := optionalUUIDParam(media.UploadedByAgentID)
	if err != nil {
		return generateddb.CreateMediaParams{}, err
	}

	return generateddb.CreateMediaParams{
		PropertyID:         propertyID,
		PropertyUnitTypeID: propertyUnitTypeID,
		AgentOfferID:       agentOfferID,
		UploadedByAgentID:  uploadedByAgentID,
		Url:                media.URL,
		ObjectKey:          textParam(media.ObjectKey),
		Kind:               string(media.Kind),
		Caption:            textParam(media.Caption),
		ContentType:        textParam(media.ContentType),
		SizeBytes:          int64Param(media.SizeBytes),
	}, nil
}

func createMediaError(err error) error {
	if isForeignKeyViolation(err) {
		return ErrForeignKeyViolation
	}
	return fmt.Errorf("create media: %w", err)
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

func (r *PropertyRepository) GetMediaForRemoval(ctx context.Context, mediaID domain.ID) (MediaRemovalTarget, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	mediaUUID, err := uuidParam(mediaID)
	if err != nil {
		return MediaRemovalTarget{}, err
	}

	row, err := r.queries.GetMediaForRemoval(ctx, mediaUUID)
	if errors.Is(err, pgx.ErrNoRows) {
		return MediaRemovalTarget{}, ErrNotFound
	}
	if err != nil {
		return MediaRemovalTarget{}, fmt.Errorf("get media for removal: %w", err)
	}

	return MediaRemovalTarget{
		MediaID:            domain.ID(uuidString(row.ID)),
		ObjectKey:          textString(row.ObjectKey),
		PropertyID:         domain.ID(uuidString(row.PropertyID)),
		PropertyUnitTypeID: domain.ID(uuidString(row.PropertyUnitTypeID)),
		AgentOfferID:       domain.ID(uuidString(row.AgentOfferID)),
		TargetType:         row.TargetType,
		CampusID:           domain.ID(uuidString(row.CampusID)),
		AgentID:            domain.ID(uuidString(row.AgentID)),
	}, nil
}

func (r *PropertyRepository) RemoveMedia(ctx context.Context, mediaID, actorUserID domain.ID) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	mediaUUID, err := uuidParam(mediaID)
	if err != nil {
		return err
	}
	actorUserUUID, err := uuidParam(actorUserID)
	if err != nil {
		return err
	}

	if _, err := r.queries.RemoveMedia(ctx, generateddb.RemoveMediaParams{ID: mediaUUID, RemovedByUserID: actorUserUUID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("remove media: %w", err)
	}

	return nil
}

func optionalUUIDParam(id domain.ID) (pgtype.UUID, error) {
	if id == "" {
		return pgtype.UUID{}, nil
	}
	return uuidParam(id)
}

func int64Param(value int64) pgtype.Int8 {
	if value == 0 {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: value, Valid: true}
}

func int64String(value pgtype.Int8) int64 {
	if !value.Valid {
		return 0
	}
	return value.Int64
}

func mediaFromCreateRow(row generateddb.CreateMediaRow) domain.Media {
	return mediaFromFields(row.ID, row.PropertyID, row.PropertyUnitTypeID, row.AgentOfferID, row.UploadedByAgentID, row.Url, row.ObjectKey, row.Kind, row.Caption, row.ContentType, row.SizeBytes, row.CreatedAt, row.RemovedAt, row.RemovedByUserID)
}

func mediaFromPropertyRow(row generateddb.ListMediaByPropertyRow) domain.Media {
	return mediaFromFields(row.ID, row.PropertyID, row.PropertyUnitTypeID, row.AgentOfferID, row.UploadedByAgentID, row.Url, row.ObjectKey, row.Kind, row.Caption, row.ContentType, row.SizeBytes, row.CreatedAt, row.RemovedAt, row.RemovedByUserID)
}

func mediaFromPropertyUnitTypeRow(row generateddb.ListMediaByPropertyUnitTypeRow) domain.Media {
	return mediaFromFields(row.ID, row.PropertyID, row.PropertyUnitTypeID, row.AgentOfferID, row.UploadedByAgentID, row.Url, row.ObjectKey, row.Kind, row.Caption, row.ContentType, row.SizeBytes, row.CreatedAt, row.RemovedAt, row.RemovedByUserID)
}

func mediaFromPropertyUnitTypeIDsRow(row generateddb.ListMediaByPropertyUnitTypeIDsRow) domain.Media {
	return mediaFromFields(row.ID, row.PropertyID, row.PropertyUnitTypeID, row.AgentOfferID, row.UploadedByAgentID, row.Url, row.ObjectKey, row.Kind, row.Caption, row.ContentType, row.SizeBytes, row.CreatedAt, row.RemovedAt, row.RemovedByUserID)
}

func mediaFromAgentOfferRow(row generateddb.ListMediaByAgentOfferRow) domain.Media {
	return mediaFromFields(row.ID, row.PropertyID, row.PropertyUnitTypeID, row.AgentOfferID, row.UploadedByAgentID, row.Url, row.ObjectKey, row.Kind, row.Caption, row.ContentType, row.SizeBytes, row.CreatedAt, row.RemovedAt, row.RemovedByUserID)
}

func mediaFromAgentOfferIDsRow(row generateddb.ListMediaByAgentOfferIDsRow) domain.Media {
	return mediaFromFields(row.ID, row.PropertyID, row.PropertyUnitTypeID, row.AgentOfferID, row.UploadedByAgentID, row.Url, row.ObjectKey, row.Kind, row.Caption, row.ContentType, row.SizeBytes, row.CreatedAt, row.RemovedAt, row.RemovedByUserID)
}

func mediaFromFields(id, propertyID, propertyUnitTypeID, agentOfferID, uploadedByAgentID pgtype.UUID, url string, objectKey pgtype.Text, kind string, caption pgtype.Text, contentType pgtype.Text, sizeBytes pgtype.Int8, createdAt, removedAt pgtype.Timestamptz, removedByUserID pgtype.UUID) domain.Media {
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
		RemovedAt:          timestamptzPointer(removedAt),
		RemovedByUserID:    domain.ID(uuidString(removedByUserID)),
	}
}
