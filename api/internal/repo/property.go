package repo

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/ddddami/laivan/internal/data"
	generateddb "github.com/ddddami/laivan/internal/db/generated"
	"github.com/ddddami/laivan/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

const queryTimeout = 3 * time.Second

type PropertyRepository struct {
	queries *generateddb.Queries
	db      generateddb.DBTX
}

type PropertyListFilter struct {
	CampusID domain.ID
	Area     string
	Filters  data.Filters
}

func NewPropertyRepository(database generateddb.DBTX) *PropertyRepository {
	return &PropertyRepository{queries: generateddb.New(database), db: database}
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

func (r *PropertyRepository) GetWithDetails(ctx context.Context, id domain.ID) (domain.PropertyDetail, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	propertyUUID, err := uuidParam(id)
	if err != nil {
		return domain.PropertyDetail{}, err
	}

	propertyRow, err := r.queries.GetProperty(ctx, propertyUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.PropertyDetail{}, ErrNotFound
		}

		return domain.PropertyDetail{}, fmt.Errorf("get property: %w", err)
	}

	unitTypeRows, err := r.queries.ListPropertyUnitTypesByProperty(ctx, propertyUUID)
	if err != nil {
		return domain.PropertyDetail{}, fmt.Errorf("list property unit types: %w", err)
	}

	unitTypes := make([]domain.PropertyUnitTypeDetail, 0, len(unitTypeRows))
	for _, utRow := range unitTypeRows {
		offerRows, err := r.queries.ListAgentOffersByPropertyUnitType(ctx, utRow.ID)
		if err != nil {
			return domain.PropertyDetail{}, fmt.Errorf("list agent offers: %w", err)
		}

		offers := make([]domain.AgentOffer, 0, len(offerRows))
		for _, offerRow := range offerRows {
			offers = append(offers, agentOfferFromRow(offerRow))
		}

		unitTypes = append(unitTypes, domain.PropertyUnitTypeDetail{
			PropertyUnitType: propertyUnitTypeFromRow(utRow),
			AgentOffers:      offers,
		})
	}

	return domain.PropertyDetail{
		Property:  propertyFromRow(propertyRow),
		UnitTypes: unitTypes,
	}, nil
}

// ListWithSummary uses a raw query here, sqlc cannot safely generate dynamic ORDER BY clauses.
func (r *PropertyRepository) ListWithSummary(ctx context.Context, filter PropertyListFilter) ([]domain.PropertySummary, int, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	campusUUID, err := uuidParam(filter.CampusID)
	if err != nil {
		return nil, 0, err
	}

	whereClause := "p.campus_id = $1"
	args := []any{campusUUID}
	argPos := 2

	if filter.Area != "" {
		whereClause += fmt.Sprintf(" AND p.area ILIKE $%d", argPos)
		args = append(args, "%"+filter.Area+"%")
		argPos++
	}

	query := fmt.Sprintf(`
		SELECT
		  p.id, p.campus_id, p.name, p.area, p.landmark, p.description, p.created_at, p.updated_at,
		  COALESCE((SELECT COUNT(*) FROM property_unit_types WHERE property_id = p.id), 0)::integer AS unit_type_count,
		  COALESCE((SELECT COUNT(*) FROM agent_offers ao JOIN property_unit_types put ON ao.property_unit_type_id = put.id WHERE put.property_id = p.id AND ao.status = 'available'), 0)::integer AS available_offer_count,
		  COALESCE((SELECT MIN(ao.price_kobo) FROM agent_offers ao JOIN property_unit_types put ON ao.property_unit_type_id = put.id WHERE put.property_id = p.id AND ao.status = 'available'), 0)::integer AS lowest_price_kobo,
		  COUNT(*) OVER() AS total_count
		FROM properties p
		WHERE %s
		ORDER BY %s %s, p.id ASC
		LIMIT $%d OFFSET $%d`, whereClause, filter.Filters.SortColumn(), filter.Filters.SortDirection(), argPos, argPos+1)

	args = append(args, filter.Filters.Limit(), filter.Filters.Offset())

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list properties with summary: %w", err)
	}
	defer rows.Close()

	var totalCount int64
	var summaries []domain.PropertySummary
	for rows.Next() {
		var row generateddb.ListPropertiesWithSummaryRow
		if err := rows.Scan(
			&row.ID,
			&row.CampusID,
			&row.Name,
			&row.Area,
			&row.Landmark,
			&row.Description,
			&row.CreatedAt,
			&row.UpdatedAt,
			&row.UnitTypeCount,
			&row.AvailableOfferCount,
			&row.LowestPriceKobo,
			&row.TotalCount,
		); err != nil {
			return nil, 0, fmt.Errorf("scan property summary: %w", err)
		}
		totalCount = row.TotalCount
		summaries = append(summaries, propertySummaryFromRow(row))
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate property summaries: %w", err)
	}

	return summaries, int(totalCount), nil
}

func (r *PropertyRepository) CreatePropertyUnitType(ctx context.Context, unitType domain.PropertyUnitType) (domain.PropertyUnitType, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	propertyUUID, err := uuidParam(unitType.PropertyID)
	if err != nil {
		return domain.PropertyUnitType{}, err
	}

	row, err := r.queries.CreatePropertyUnitType(ctx, generateddb.CreatePropertyUnitTypeParams{
		PropertyID:   propertyUUID,
		Category:     string(unitType.Category),
		Name:         unitType.Name,
		Description:  textParam(unitType.Description),
		BedroomCount: intParam(unitType.Structure.BedroomCount),
		HasParlour:   boolParam(unitType.Structure.HasParlour),
		BathroomType: textParam(unitType.Structure.BathroomType),
		KitchenType:  textParam(unitType.Structure.KitchenType),
	})
	if err != nil {
		if isForeignKeyViolation(err) {
			return domain.PropertyUnitType{}, ErrNotFound
		}

		return domain.PropertyUnitType{}, fmt.Errorf("create property unit type: %w", err)
	}

	return propertyUnitTypeFromRow(row), nil
}

func (r *PropertyRepository) ListPropertyUnitTypes(ctx context.Context, propertyID domain.ID) ([]domain.PropertyUnitType, error) {
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

		return nil, fmt.Errorf("get property for unit types: %w", err)
	}

	rows, err := r.queries.ListPropertyUnitTypesByProperty(ctx, propertyUUID)
	if err != nil {
		return nil, fmt.Errorf("list property unit types: %w", err)
	}

	unitTypes := make([]domain.PropertyUnitType, 0, len(rows))
	for _, row := range rows {
		unitTypes = append(unitTypes, propertyUnitTypeFromRow(row))
	}

	return unitTypes, nil
}

func (r *PropertyRepository) CreateAgentOffer(ctx context.Context, offer domain.AgentOffer) (domain.AgentOffer, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	unitTypeUUID, err := uuidParam(offer.PropertyUnitTypeID)
	if err != nil {
		return domain.AgentOffer{}, err
	}
	agentUUID, err := uuidParam(offer.AgentID)
	if err != nil {
		return domain.AgentOffer{}, err
	}

	row, err := r.queries.CreateAgentOffer(ctx, generateddb.CreateAgentOfferParams{
		PropertyUnitTypeID: unitTypeUUID,
		AgentID:            agentUUID,
		Title:              offer.Title,
		Description:        textParam(offer.Description),
		PriceKobo:          offer.Price.AmountKobo,
		Status:             string(offer.Status),
	})
	if err != nil {
		if isForeignKeyViolation(err) {
			return domain.AgentOffer{}, ErrNotFound
		}

		return domain.AgentOffer{}, fmt.Errorf("create agent offer: %w", err)
	}

	return agentOfferFromRow(row), nil
}

func (r *PropertyRepository) ListAgentOffers(ctx context.Context, unitTypeID domain.ID) ([]domain.AgentOffer, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	unitTypeUUID, err := uuidParam(unitTypeID)
	if err != nil {
		return nil, err
	}

	if _, err := r.queries.GetPropertyUnitType(ctx, unitTypeUUID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("get property unit type for agent offers: %w", err)
	}

	rows, err := r.queries.ListAgentOffersByPropertyUnitType(ctx, unitTypeUUID)
	if err != nil {
		return nil, fmt.Errorf("list agent offers: %w", err)
	}

	offers := make([]domain.AgentOffer, 0, len(rows))
	for _, row := range rows {
		offers = append(offers, agentOfferFromRow(row))
	}

	return offers, nil
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

func propertyUnitTypeFromRow(row generateddb.PropertyUnitType) domain.PropertyUnitType {
	return domain.PropertyUnitType{
		ID:          domain.ID(uuidString(row.ID)),
		PropertyID:  domain.ID(uuidString(row.PropertyID)),
		Category:    domain.UnitCategory(row.Category),
		Name:        row.Name,
		Description: textString(row.Description),
		Structure: domain.UnitStructure{
			BedroomCount: intPointer(row.BedroomCount),
			HasParlour:   boolPointer(row.HasParlour),
			BathroomType: textString(row.BathroomType),
			KitchenType:  textString(row.KitchenType),
		},
		Timestamps: domain.Timestamps{
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		},
	}
}

func agentOfferFromRow(row generateddb.AgentOffer) domain.AgentOffer {
	return domain.AgentOffer{
		ID:                 domain.ID(uuidString(row.ID)),
		PropertyUnitTypeID: domain.ID(uuidString(row.PropertyUnitTypeID)),
		AgentID:            domain.ID(uuidString(row.AgentID)),
		Title:              row.Title,
		Description:        textString(row.Description),
		Price:              domain.Money{AmountKobo: row.PriceKobo},
		Status:             domain.AgentOfferStatus(row.Status),
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

func intParam(value *int) pgtype.Int4 {
	if value == nil {
		return pgtype.Int4{}
	}

	return pgtype.Int4{Int32: int32(*value), Valid: true}
}

func intPointer(value pgtype.Int4) *int {
	if !value.Valid {
		return nil
	}

	result := int(value.Int32)
	return &result
}

func boolParam(value *bool) pgtype.Bool {
	if value == nil {
		return pgtype.Bool{}
	}

	return pgtype.Bool{Bool: *value, Valid: true}
}

func boolPointer(value pgtype.Bool) *bool {
	if !value.Valid {
		return nil
	}

	result := value.Bool
	return &result
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

func propertySummaryFromRow(row generateddb.ListPropertiesWithSummaryRow) domain.PropertySummary {
	return domain.PropertySummary{
		Property: domain.Property{
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
		},
		UnitTypeCount:       row.UnitTypeCount,
		AvailableOfferCount: row.AvailableOfferCount,
		LowestPriceKobo:     row.LowestPriceKobo,
	}
}
