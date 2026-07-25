package repo

import (
	"context"
	"errors"
	"fmt"

	generateddb "github.com/ddddami/laivan/internal/db/generated"
	"github.com/ddddami/laivan/internal/domain"
	"github.com/jackc/pgx/v5"
)

func (r *PropertyRepository) GetCampusBySlug(ctx context.Context, slug string) (domain.Campus, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	row, err := r.queries.GetCampusBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Campus{}, ErrNotFound
		}
		return domain.Campus{}, fmt.Errorf("get campus by slug: %w", err)
	}

	return campusFromRow(row), nil
}

func campusFromRow(row generateddb.Campuse) domain.Campus {
	return domain.Campus{
		ID:        domain.ID(uuidString(row.ID)),
		Slug:      row.Slug,
		Name:      row.Name,
		ShortName: row.ShortName,
		IsActive:  row.IsActive,
		Timestamps: domain.Timestamps{
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		},
	}
}
