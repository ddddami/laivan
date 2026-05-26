package repo

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	generateddb "github.com/ddddami/laivan/internal/db/generated"
	"github.com/ddddami/laivan/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

const queryTimeout = 3 * time.Second

type PropertyRepository struct {
	queries *generateddb.Queries
}

type PropertyListFilter struct {
	CampusID domain.ID
	Limit    int32
}

func NewPropertyRepository(database generateddb.DBTX) *PropertyRepository {
	return &PropertyRepository{queries: generateddb.New(database)}
}

func (r *PropertyRepository) Create(ctx context.Context, property domain.Property) (domain.Property, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	campusUUID, err := uuidParam(property.CampusID)
	if err != nil {
		return domain.Property{}, err
	}

	row, err := r.queries.CreateProperty(ctx, generateddb.CreatePropertyParams{
		CampusID:    campusUUID,
		Name:        property.Name,
		Area:        property.Location.Area,
		Landmark:    textParam(property.Location.Landmark),
		Description: textParam(property.Description),
	})
	if err != nil {
		return domain.Property{}, fmt.Errorf("create property: %w", err)
	}

	return propertyFromRow(row), nil
}

func (r *PropertyRepository) Get(ctx context.Context, id domain.ID) (domain.Property, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	uuid, err := uuidParam(id)
	if err != nil {
		return domain.Property{}, err
	}

	row, err := r.queries.GetProperty(ctx, uuid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Property{}, ErrNotFound
		}

		return domain.Property{}, fmt.Errorf("get property: %w", err)
	}

	return propertyFromRow(row), nil
}

func (r *PropertyRepository) List(ctx context.Context, filter PropertyListFilter) ([]domain.Property, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	campusUUID, err := uuidParam(filter.CampusID)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListProperties(ctx, generateddb.ListPropertiesParams{
		CampusID: campusUUID,
		Limit:    filter.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list properties: %w", err)
	}

	properties := make([]domain.Property, 0, len(rows))
	for _, row := range rows {
		properties = append(properties, propertyFromRow(row))
	}

	return properties, nil
}

func (r *PropertyRepository) CreateRoomType(ctx context.Context, roomType domain.RoomType) (domain.RoomType, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	propertyUUID, err := uuidParam(roomType.PropertyID)
	if err != nil {
		return domain.RoomType{}, err
	}

	row, err := r.queries.CreateRoomType(ctx, generateddb.CreateRoomTypeParams{
		PropertyID:  propertyUUID,
		Name:        roomType.Name,
		Description: textParam(roomType.Description),
	})
	if err != nil {
		if isForeignKeyViolation(err) {
			return domain.RoomType{}, ErrNotFound
		}

		return domain.RoomType{}, fmt.Errorf("create room type: %w", err)
	}

	return roomTypeFromRow(row), nil
}

func (r *PropertyRepository) ListRoomTypes(ctx context.Context, propertyID domain.ID) ([]domain.RoomType, error) {
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

		return nil, fmt.Errorf("get property for room types: %w", err)
	}

	rows, err := r.queries.ListRoomTypesByProperty(ctx, propertyUUID)
	if err != nil {
		return nil, fmt.Errorf("list room types: %w", err)
	}

	roomTypes := make([]domain.RoomType, 0, len(rows))
	for _, row := range rows {
		roomTypes = append(roomTypes, roomTypeFromRow(row))
	}

	return roomTypes, nil
}

func propertyFromRow(row generateddb.Property) domain.Property {
	return domain.Property{
		ID:       domain.ID(uuidString(row.ID)),
		CampusID: domain.ID(uuidString(row.CampusID)),
		Name:     row.Name,
		Location: domain.ApproxLocation{
			Area:     row.Area,
			Landmark: textString(row.Landmark),
		},
		Description: textString(row.Description),
		Timestamps: domain.Timestamps{
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		},
	}
}

func roomTypeFromRow(row generateddb.RoomType) domain.RoomType {
	return domain.RoomType{
		ID:          domain.ID(uuidString(row.ID)),
		PropertyID:  domain.ID(uuidString(row.PropertyID)),
		Name:        row.Name,
		Description: textString(row.Description),
		Timestamps: domain.Timestamps{
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		},
	}
}

func textParam(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: value != ""}
}

func textString(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}

	return value.String
}

func uuidParam(id domain.ID) (pgtype.UUID, error) {
	var uuid pgtype.UUID
	if err := uuid.Scan(string(id)); err != nil {
		return pgtype.UUID{}, fmt.Errorf("invalid id: %w", err)
	}

	return uuid, nil
}

func uuidString(uuid pgtype.UUID) string {
	if !uuid.Valid {
		return ""
	}

	buf := make([]byte, 36)
	hex.Encode(buf[0:8], uuid.Bytes[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], uuid.Bytes[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], uuid.Bytes[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], uuid.Bytes[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:36], uuid.Bytes[10:16])

	return string(buf)
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
