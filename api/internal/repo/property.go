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
	CampusID  domain.ID
	Area      string
	Name      string
	HasOffers *bool
	MinPrice  *int
	MaxPrice  *int
	Filters   data.Filters
}

type MediaTarget struct {
	PropertyID         domain.ID
	PropertyUnitTypeID domain.ID
	AgentOfferID       domain.ID
	CampusID           domain.ID
	AgentID            domain.ID
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

func (r *PropertyRepository) Update(ctx context.Context, id domain.ID, expectedVersion int, patch domain.PropertyPatch) (domain.Property, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	propertyUUID, err := uuidParam(id)
	if err != nil {
		return domain.Property{}, err
	}

	row, err := r.queries.UpdateProperty(ctx, generateddb.UpdatePropertyParams{
		ID:              propertyUUID,
		ExpectedVersion: expectedVersion,
		Name:            optionalTextParam(patch.Name),
		Area:            optionalTextParam(patch.Area),
		Landmark:        optionalTextParam(patch.Landmark),
		Description:     optionalTextParam(patch.Description),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Property{}, ErrStaleUpdate
		}

		return domain.Property{}, fmt.Errorf("update property: %w", err)
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

	propertyMediaRows, err := r.queries.ListMediaByProperty(ctx, propertyUUID)
	if err != nil {
		return domain.PropertyDetail{}, fmt.Errorf("list property media: %w", err)
	}
	propertyMedia := make([]domain.Media, 0, len(propertyMediaRows))
	for _, mediaRow := range propertyMediaRows {
		propertyMedia = append(propertyMedia, mediaFromPropertyRow(mediaRow))
	}

	mediaByUnitType := make(map[domain.ID][]domain.Media, len(unitTypeRows))
	offersByUnitType := make(map[domain.ID][]domain.AgentOfferDetail, len(unitTypeRows))
	if len(unitTypeRows) > 0 {
		unitTypeIDs := make([]pgtype.UUID, 0, len(unitTypeRows))
		for _, unitTypeRow := range unitTypeRows {
			unitTypeIDs = append(unitTypeIDs, unitTypeRow.ID)
		}

		mediaRows, err := r.queries.ListMediaByPropertyUnitTypeIDs(ctx, unitTypeIDs)
		if err != nil {
			return domain.PropertyDetail{}, fmt.Errorf("list property unit type media: %w", err)
		}
		for _, mediaRow := range mediaRows {
			media := mediaFromPropertyUnitTypeIDsRow(mediaRow)
			mediaByUnitType[media.PropertyUnitTypeID] = append(mediaByUnitType[media.PropertyUnitTypeID], media)
		}

		offerRows, err := r.queries.ListAgentOfferDetailsByPropertyUnitTypeIDs(ctx, unitTypeIDs)
		if err != nil {
			return domain.PropertyDetail{}, fmt.Errorf("list agent offers: %w", err)
		}

		mediaByOffer := make(map[domain.ID][]domain.Media, len(offerRows))
		if len(offerRows) > 0 {
			offerIDs := make([]pgtype.UUID, 0, len(offerRows))
			for _, offerRow := range offerRows {
				offerIDs = append(offerIDs, offerRow.ID)
			}

			offerMediaRows, err := r.queries.ListMediaByAgentOfferIDs(ctx, offerIDs)
			if err != nil {
				return domain.PropertyDetail{}, fmt.Errorf("list agent offer media: %w", err)
			}
			for _, mediaRow := range offerMediaRows {
				media := mediaFromAgentOfferIDsRow(mediaRow)
				mediaByOffer[media.AgentOfferID] = append(mediaByOffer[media.AgentOfferID], media)
			}
		}

		for _, offerRow := range offerRows {
			offer := agentOfferDetailFromRow(offerRow)
			offer.Media = mediaByOffer[offer.ID]
			offersByUnitType[offer.PropertyUnitTypeID] = append(offersByUnitType[offer.PropertyUnitTypeID], offer)
		}
	}

	unitTypes := make([]domain.PropertyUnitTypeDetail, 0, len(unitTypeRows))
	for _, utRow := range unitTypeRows {
		unitTypeID := domain.ID(uuidString(utRow.ID))
		unitTypes = append(unitTypes, domain.PropertyUnitTypeDetail{
			PropertyUnitType: propertyUnitTypeFromRow(utRow),
			Media:            mediaByUnitType[unitTypeID],
			AgentOffers:      offersByUnitType[unitTypeID],
		})
	}

	return domain.PropertyDetail{
		Property:  propertyFromRow(propertyRow),
		Media:     propertyMedia,
		UnitTypes: unitTypes,
	}, nil
}

// ListWithSummary uses a raw query here, sqlc cannot safely generate dynamic ORDER BY clauses.
// IMPORTANT: If you change the SELECT column list, you must update the Scan call below.
// pgx Scan errors on column count mismatch, which our integration tests catch.
func (r *PropertyRepository) ListWithSummary(ctx context.Context, filter PropertyListFilter) ([]domain.PropertySummary, int, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	campusUUID, err := uuidParam(filter.CampusID)
	if err != nil {
		return nil, 0, err
	}

	propertyWhere := "p.campus_id = $1"
	args := []any{campusUUID}
	argPos := 2

	if filter.Area != "" {
		propertyWhere += fmt.Sprintf(" AND p.area ILIKE $%d", argPos)
		args = append(args, "%"+filter.Area+"%")
		argPos++
	}
	if filter.Name != "" {
		propertyWhere += fmt.Sprintf(" AND p.name ILIKE $%d", argPos)
		args = append(args, "%"+filter.Name+"%")
		argPos++
	}

	havingConditions := ""
	if filter.HasOffers != nil {
		if *filter.HasOffers {
			havingConditions += " AND available_offer_count > 0"
		} else {
			havingConditions += " AND available_offer_count = 0"
		}
	}
	if filter.MinPrice != nil {
		havingConditions += fmt.Sprintf(" AND lowest_price_kobo >= $%d", argPos)
		args = append(args, domain.Kobo(*filter.MinPrice))
		argPos++
	}
	if filter.MaxPrice != nil {
		havingConditions += fmt.Sprintf(" AND lowest_price_kobo <= $%d", argPos)
		args = append(args, domain.Kobo(*filter.MaxPrice))
		argPos++
	}

	sortColumn := filter.Filters.SortColumn()
	nullsOrder := ""
	if sortColumn == "lowest_price_kobo" {
		nullsOrder = " NULLS LAST"
	}

	query := fmt.Sprintf(`
		SELECT *, COUNT(*) OVER() AS total_count
		FROM (
		  SELECT
		    p.id, p.campus_id, p.name, p.area, p.landmark, p.description, p.created_at, p.updated_at, p.version,
		    COALESCE((SELECT COUNT(*) FROM property_unit_types WHERE property_id = p.id), 0)::integer AS unit_type_count,
		    COALESCE((SELECT COUNT(*) FROM agent_offers ao JOIN property_unit_types put ON ao.property_unit_type_id = put.id WHERE put.property_id = p.id AND ao.status = 'available'), 0)::integer AS available_offer_count,
		    ((SELECT MIN(ao.price_kobo) FROM agent_offers ao JOIN property_unit_types put ON ao.property_unit_type_id = put.id WHERE put.property_id = p.id AND ao.status = 'available'))::integer AS lowest_price_kobo,
		    (SELECT m.url FROM media m WHERE m.property_id = p.id OR m.property_unit_type_id IN (SELECT id FROM property_unit_types WHERE property_id = p.id) OR m.agent_offer_id IN (SELECT ao.id FROM agent_offers ao JOIN property_unit_types put ON ao.property_unit_type_id = put.id WHERE put.property_id = p.id) ORDER BY m.created_at ASC, m.id ASC LIMIT 1) AS thumbnail_url
		  FROM properties p
		  WHERE %s
		) sub
		WHERE TRUE %s
		ORDER BY %s %s%s, id ASC
		LIMIT $%d OFFSET $%d`, propertyWhere, havingConditions, sortColumn, filter.Filters.SortDirection(), nullsOrder, argPos, argPos+1)

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
		var thumbnailURL pgtype.Text
		var lowestPriceKobo pgtype.Int4
		if err := rows.Scan(
			&row.ID,
			&row.CampusID,
			&row.Name,
			&row.Area,
			&row.Landmark,
			&row.Description,
			&row.CreatedAt,
			&row.UpdatedAt,
			&row.Version,
			&row.UnitTypeCount,
			&row.AvailableOfferCount,
			&lowestPriceKobo,
			&thumbnailURL,
			&row.TotalCount,
		); err != nil {
			return nil, 0, fmt.Errorf("scan property summary: %w", err)
		}
		totalCount = row.TotalCount
		if lowestPriceKobo.Valid {
			row.LowestPriceKobo = int(lowestPriceKobo.Int32)
		}
		summary := propertySummaryFromRow(row)
		summary.ThumbnailURL = textString(thumbnailURL)
		summaries = append(summaries, summary)
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
		Notes:        textParam(unitType.Notes),
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

func (r *PropertyRepository) GetMediaTarget(ctx context.Context, targetType string, id domain.ID) (MediaTarget, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	uuid, err := uuidParam(id)
	if err != nil {
		return MediaTarget{}, err
	}

	switch targetType {
	case "property":
		row, err := r.queries.GetProperty(ctx, uuid)
		if errors.Is(err, pgx.ErrNoRows) {
			return MediaTarget{}, ErrNotFound
		}
		if err != nil {
			return MediaTarget{}, fmt.Errorf("get property media target: %w", err)
		}
		return MediaTarget{
			PropertyID: domain.ID(uuidString(row.ID)),
			CampusID:   domain.ID(uuidString(row.CampusID)),
		}, nil
	case "property_unit_type":
		row, err := r.queries.GetPropertyUnitTypeMediaTarget(ctx, uuid)
		if errors.Is(err, pgx.ErrNoRows) {
			return MediaTarget{}, ErrNotFound
		}
		if err != nil {
			return MediaTarget{}, fmt.Errorf("get property unit type media target: %w", err)
		}
		return MediaTarget{
			PropertyID:         domain.ID(uuidString(row.PropertyID)),
			PropertyUnitTypeID: domain.ID(uuidString(row.PropertyUnitTypeID)),
			CampusID:           domain.ID(uuidString(row.CampusID)),
		}, nil
	case "agent_offer":
		row, err := r.queries.GetAgentOfferMediaTarget(ctx, uuid)
		if errors.Is(err, pgx.ErrNoRows) {
			return MediaTarget{}, ErrNotFound
		}
		if err != nil {
			return MediaTarget{}, fmt.Errorf("get agent offer media target: %w", err)
		}
		return MediaTarget{
			PropertyID:         domain.ID(uuidString(row.PropertyID)),
			PropertyUnitTypeID: domain.ID(uuidString(row.PropertyUnitTypeID)),
			AgentOfferID:       domain.ID(uuidString(row.AgentOfferID)),
			CampusID:           domain.ID(uuidString(row.CampusID)),
			AgentID:            domain.ID(uuidString(row.AgentID)),
		}, nil
	default:
		return MediaTarget{}, ErrNotFound
	}
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

	if _, err := r.queries.GetPropertyUnitType(ctx, unitTypeUUID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AgentOffer{}, ErrUnitTypeNotFound
		}
		return domain.AgentOffer{}, fmt.Errorf("get unit type for agent offer: %w", err)
	}

	agent, err := r.queries.GetAgent(ctx, agentUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AgentOffer{}, ErrAgentNotFound
		}
		return domain.AgentOffer{}, fmt.Errorf("get agent for agent offer: %w", err)
	}
	if agent.Status != string(domain.AgentStatusActive) {
		return domain.AgentOffer{}, ErrAgentForbidden
	}

	row, err := r.queries.CreateAuthorizedAgentOffer(ctx, generateddb.CreateAuthorizedAgentOfferParams{
		PropertyUnitTypeID: unitTypeUUID,
		AgentID:            agentUUID,
		Title:              offer.Title,
		Description:        textParam(offer.Description),
		Notes:              textParam(offer.Notes),
		PriceKobo:          offer.Price.AmountKobo,
		Status:             string(offer.Status),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AgentOffer{}, ErrAgentForbidden
		}
		if isForeignKeyViolation(err) {
			unitTypeNotFound := isConstraintViolation(err, "agent_offers_property_unit_type_id_fkey")
			agentNotFound := isConstraintViolation(err, "agent_offers_agent_id_fkey")
			switch {
			case unitTypeNotFound:
				return domain.AgentOffer{}, ErrUnitTypeNotFound
			case agentNotFound:
				return domain.AgentOffer{}, ErrAgentNotFound
			default:
				return domain.AgentOffer{}, ErrNotFound
			}
		}
		if isUniqueViolation(err) {
			return domain.AgentOffer{}, ErrDuplicate
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
		Version:     row.Version,
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
		Version:     row.Version,
		Notes:       textString(row.Notes),
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
		Notes:              textString(row.Notes),
		Price:              domain.Money{AmountKobo: row.PriceKobo},
		Status:             domain.AgentOfferStatus(row.Status),
		Version:            row.Version,
		Timestamps: domain.Timestamps{
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		},
	}
}

func agentOfferDetailFromRow(row generateddb.ListAgentOfferDetailsByPropertyUnitTypeIDsRow) domain.AgentOfferDetail {
	return domain.AgentOfferDetail{
		AgentOffer: domain.AgentOffer{
			ID:                 domain.ID(uuidString(row.ID)),
			PropertyUnitTypeID: domain.ID(uuidString(row.PropertyUnitTypeID)),
			AgentID:            domain.ID(uuidString(row.AgentID)),
			Title:              row.Title,
			Description:        textString(row.Description),
			Notes:              textString(row.Notes),
			Price:              domain.Money{AmountKobo: row.PriceKobo},
			Status:             domain.AgentOfferStatus(row.Status),
			Version:            row.Version,
			Timestamps: domain.Timestamps{
				CreatedAt: row.CreatedAt.Time,
				UpdatedAt: row.UpdatedAt.Time,
			},
		},
		Agent: domain.AgentSummary{
			ID:          domain.ID(uuidString(row.AgentID)),
			DisplayName: row.AgentDisplayName,
		},
	}
}

func textParam(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: value != ""}
}

func optionalTextParam(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
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

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func isConstraintViolation(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503" && pgErr.ConstraintName == constraintName
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
			Version:     row.Version,
			Timestamps: domain.Timestamps{
				CreatedAt: row.CreatedAt.Time,
				UpdatedAt: row.UpdatedAt.Time,
			},
		},
		UnitTypeCount:       row.UnitTypeCount,
		AvailableOfferCount: row.AvailableOfferCount,
		LowestPrice:         domain.Money{AmountKobo: row.LowestPriceKobo},
	}
}
