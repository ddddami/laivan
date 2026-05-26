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

	roomTypeRows, err := r.queries.ListRoomTypesByProperty(ctx, propertyUUID)
	if err != nil {
		return domain.PropertyDetail{}, fmt.Errorf("list room types: %w", err)
	}

	roomTypes := make([]domain.RoomTypeDetail, 0, len(roomTypeRows))
	for _, rtRow := range roomTypeRows {
		offerRows, err := r.queries.ListAgentOffersByRoomType(ctx, rtRow.ID)
		if err != nil {
			return domain.PropertyDetail{}, fmt.Errorf("list agent offers: %w", err)
		}

		offers := make([]domain.AgentOffer, 0, len(offerRows))
		for _, offerRow := range offerRows {
			offers = append(offers, agentOfferFromRow(offerRow))
		}

		roomTypes = append(roomTypes, domain.RoomTypeDetail{
			RoomType:    roomTypeFromRow(rtRow),
			AgentOffers: offers,
		})
	}

	return domain.PropertyDetail{
		Property:  propertyFromRow(propertyRow),
		RoomTypes: roomTypes,
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
		  COALESCE((SELECT COUNT(*) FROM room_types WHERE property_id = p.id), 0)::integer AS room_type_count,
		  COALESCE((SELECT COUNT(*) FROM agent_offers ao JOIN room_types rt ON ao.room_type_id = rt.id WHERE rt.property_id = p.id AND ao.status = 'available'), 0)::integer AS available_offer_count,
		  COALESCE((SELECT MIN(ao.price_kobo) FROM agent_offers ao JOIN room_types rt ON ao.room_type_id = rt.id WHERE rt.property_id = p.id AND ao.status = 'available'), 0)::integer AS lowest_price_kobo,
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
			&row.RoomTypeCount,
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

func (r *PropertyRepository) CreateAgentOffer(ctx context.Context, offer domain.AgentOffer) (domain.AgentOffer, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	roomTypeUUID, err := uuidParam(offer.RoomTypeID)
	if err != nil {
		return domain.AgentOffer{}, err
	}
	agentUUID, err := uuidParam(offer.AgentID)
	if err != nil {
		return domain.AgentOffer{}, err
	}

	row, err := r.queries.CreateAgentOffer(ctx, generateddb.CreateAgentOfferParams{
		RoomTypeID:  roomTypeUUID,
		AgentID:     agentUUID,
		Title:       offer.Title,
		Description: textParam(offer.Description),
		PriceKobo:   offer.Price.AmountKobo,
		Status:      string(offer.Status),
	})
	if err != nil {
		if isForeignKeyViolation(err) {
			return domain.AgentOffer{}, ErrNotFound
		}

		return domain.AgentOffer{}, fmt.Errorf("create agent offer: %w", err)
	}

	return agentOfferFromRow(row), nil
}

func (r *PropertyRepository) ListAgentOffers(ctx context.Context, roomTypeID domain.ID) ([]domain.AgentOffer, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	roomTypeUUID, err := uuidParam(roomTypeID)
	if err != nil {
		return nil, err
	}

	if _, err := r.queries.GetRoomType(ctx, roomTypeUUID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("get room type for agent offers: %w", err)
	}

	rows, err := r.queries.ListAgentOffersByRoomType(ctx, roomTypeUUID)
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

func agentOfferFromRow(row generateddb.AgentOffer) domain.AgentOffer {
	return domain.AgentOffer{
		ID:          domain.ID(uuidString(row.ID)),
		RoomTypeID:  domain.ID(uuidString(row.RoomTypeID)),
		AgentID:     domain.ID(uuidString(row.AgentID)),
		Title:       row.Title,
		Description: textString(row.Description),
		Price:       domain.Money{AmountKobo: row.PriceKobo},
		Status:      domain.AgentOfferStatus(row.Status),
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
		RoomTypeCount:       row.RoomTypeCount,
		AvailableOfferCount: row.AvailableOfferCount,
		LowestPriceKobo:     row.LowestPriceKobo,
	}
}
