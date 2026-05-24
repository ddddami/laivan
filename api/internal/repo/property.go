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
	"github.com/jackc/pgx/v5/pgtype"
)

const queryTimeout = 3 * time.Second

type PropertyRepository struct {
	queries *generateddb.Queries
}

func NewPropertyRepository(database generateddb.DBTX) *PropertyRepository {
	return &PropertyRepository{queries: generateddb.New(database)}
}

func (r *PropertyRepository) Create(ctx context.Context, property domain.Property) (domain.Property, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	row, err := r.queries.CreateProperty(ctx, generateddb.CreatePropertyParams{
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

func propertyFromRow(row generateddb.Property) domain.Property {
	return domain.Property{
		ID:   domain.ID(uuidString(row.ID)),
		Name: row.Name,
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
		return pgtype.UUID{}, fmt.Errorf("invalid property id: %w", err)
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
