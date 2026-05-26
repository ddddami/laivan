//go:build integration

package repo

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/ddddami/laivan/internal/data"
	appdb "github.com/ddddami/laivan/internal/db"
	"github.com/ddddami/laivan/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPropertyRepositoryCreateAndGet(t *testing.T) {
	ctx := context.Background()
	pool := openIntegrationDB(t, ctx)
	t.Cleanup(pool.Close)

	truncateProperties(t, ctx, pool)
	t.Cleanup(func() { truncateProperties(t, ctx, pool) })

	campusID := testCampusID(t, ctx, pool)
	repository := NewPropertyRepository(pool)

	created, err := repository.Create(ctx, domain.Property{
		CampusID: campusID,
		Name:     "Alice Lodge",
		Location: domain.ApproxLocation{
			Area:     "Obanla",
			Landmark: "Near South Gate",
		},
		Description: "Gated lodge with multiple room categories near campus.",
	})
	if err != nil {
		t.Fatalf("create property: %v", err)
	}

	if created.ID == "" {
		t.Fatal("created property ID is empty")
	}
	if created.CampusID != campusID {
		t.Fatalf("campus ID = %q, want %q", created.CampusID, campusID)
	}
	if created.Name != "Alice Lodge" {
		t.Fatalf("name = %q, want Alice Lodge", created.Name)
	}
	if created.Location.Area != "Obanla" {
		t.Fatalf("area = %q, want Obanla", created.Location.Area)
	}
	if created.Location.Landmark != "Near South Gate" {
		t.Fatalf("landmark = %q, want Near South Gate", created.Location.Landmark)
	}
	if created.Description != "Gated lodge with multiple room categories near campus." {
		t.Fatalf("description = %q, want Gated lodge with multiple room categories near campus.", created.Description)
	}
	if created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatal("created property timestamps must be set")
	}

	fetched, err := repository.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("get property: %v", err)
	}
	if fetched != created {
		t.Fatalf("fetched property = %#v, want %#v", fetched, created)
	}
}

func TestPropertyRepositoryGetNotFound(t *testing.T) {
	ctx := context.Background()
	pool := openIntegrationDB(t, ctx)
	t.Cleanup(pool.Close)

	repository := NewPropertyRepository(pool)
	_, err := repository.Get(ctx, domain.ID("550e8400-e29b-41d4-a716-446655440000"))
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want %v", err, ErrNotFound)
	}
}

func TestPropertyRepositoryGetWithDetails(t *testing.T) {
	ctx := context.Background()
	pool := openIntegrationDB(t, ctx)
	t.Cleanup(pool.Close)

	truncateProperties(t, ctx, pool)
	truncateAgents(t, ctx, pool)
	t.Cleanup(func() {
		truncateProperties(t, ctx, pool)
		truncateAgents(t, ctx, pool)
	})

	campusID := testCampusID(t, ctx, pool)
	propertyID := insertProperty(t, ctx, pool, campusID, "Alice Lodge", time.Date(2026, time.May, 1, 12, 0, 0, 0, time.UTC))
	unitType1 := insertPropertyUnitType(t, ctx, pool, propertyID, "Self-contained")
	unitType2 := insertPropertyUnitType(t, ctx, pool, propertyID, "Single room")
	agentID := insertAgent(t, ctx, pool, "Dami Agent")
	insertAgentOffer(t, ctx, pool, unitType1, agentID, "Selfcon offer", 25000000)
	insertAgentOffer(t, ctx, pool, unitType2, agentID, "Single room offer", 15000000)

	repository := NewPropertyRepository(pool)

	detail, err := repository.GetWithDetails(ctx, propertyID)
	if err != nil {
		t.Fatalf("get property with details: %v", err)
	}

	if detail.Name != "Alice Lodge" {
		t.Fatalf("name = %q, want Alice Lodge", detail.Name)
	}
	if len(detail.UnitTypes) != 2 {
		t.Fatalf("unit types length = %d, want 2", len(detail.UnitTypes))
	}
	if detail.UnitTypes[0].Name != "Self-contained" {
		t.Fatalf("unit type name = %q, want Self-contained", detail.UnitTypes[0].Name)
	}
	if len(detail.UnitTypes[0].AgentOffers) != 1 {
		t.Fatalf("agent offers length = %d, want 1", len(detail.UnitTypes[0].AgentOffers))
	}
	if detail.UnitTypes[0].AgentOffers[0].Price.AmountKobo != 25000000 {
		t.Fatalf("price = %d, want 25000000", detail.UnitTypes[0].AgentOffers[0].Price.AmountKobo)
	}

	// Test not found
	_, err = repository.GetWithDetails(ctx, domain.ID("550e8400-e29b-41d4-a716-446655440000"))
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want %v", err, ErrNotFound)
	}
}

func TestPropertyRepositoryList(t *testing.T) {
	ctx := context.Background()
	pool := openIntegrationDB(t, ctx)
	t.Cleanup(pool.Close)

	truncateProperties(t, ctx, pool)
	truncateAgents(t, ctx, pool)
	t.Cleanup(func() {
		truncateProperties(t, ctx, pool)
		truncateAgents(t, ctx, pool)
	})

	campusID := testCampusID(t, ctx, pool)
	otherCampusID := createTestCampus(t, ctx, pool, "futa-north")

	insertProperty(t, ctx, pool, otherCampusID, "Other Campus Lodge", time.Date(2026, time.May, 4, 12, 0, 0, 0, time.UTC))
	older := insertProperty(t, ctx, pool, campusID, "Older Lodge", time.Date(2026, time.May, 1, 12, 0, 0, 0, time.UTC))
	newer := insertProperty(t, ctx, pool, campusID, "Newer Lodge", time.Date(2026, time.May, 3, 12, 0, 0, 0, time.UTC))

	// Add unit types and offers to newer property for summary testing
	unitType1 := insertPropertyUnitType(t, ctx, pool, newer, "Self-contained")
	unitType2 := insertPropertyUnitType(t, ctx, pool, newer, "Single room")
	agentID := insertAgent(t, ctx, pool, "Dami Agent")
	insertAgentOffer(t, ctx, pool, unitType1, agentID, "Selfcon offer", 25000000)
	insertAgentOffer(t, ctx, pool, unitType2, agentID, "Single room offer", 15000000)

	repository := NewPropertyRepository(pool)

	// Test ListWithSummary
	summaries, totalRecords, err := repository.ListWithSummary(ctx, PropertyListFilter{
		CampusID: campusID,
		Filters: data.Filters{
			Page:         1,
			PageSize:     10,
			Sort:         "-created_at",
			SortSafelist: []string{"created_at", "-created_at"},
		},
	})
	if err != nil {
		t.Fatalf("list properties with summary: %v", err)
	}
	if totalRecords != 2 {
		t.Fatalf("total_records = %d, want 2", totalRecords)
	}
	if len(summaries) != 2 {
		t.Fatalf("summaries length = %d, want 2", len(summaries))
	}
	// Newer property should be first (newest first)
	if summaries[0].ID != newer {
		t.Fatalf("first summary ID = %q, want %q", summaries[0].ID, newer)
	}
	if summaries[0].UnitTypeCount != 2 {
		t.Fatalf("unit_type_count = %d, want 2", summaries[0].UnitTypeCount)
	}
	if summaries[0].AvailableOfferCount != 2 {
		t.Fatalf("available_offer_count = %d, want 2", summaries[0].AvailableOfferCount)
	}
	if summaries[0].LowestPrice.AmountKobo != 15000000 {
		t.Fatalf("lowest_price_kobo = %d, want 15000000", summaries[0].LowestPrice.AmountKobo)
	}

	// Older property has no unit types or offers
	if summaries[1].ID != older {
		t.Fatalf("second summary ID = %q, want %q", summaries[1].ID, older)
	}
	if summaries[1].UnitTypeCount != 0 {
		t.Fatalf("unit_type_count = %d, want 0", summaries[1].UnitTypeCount)
	}
	if summaries[1].LowestPrice.AmountKobo != 0 {
		t.Fatalf("lowest_price_kobo = %d, want 0", summaries[1].LowestPrice.AmountKobo)
	}
}

func TestPropertyRepositoryCreateAndListPropertyUnitTypes(t *testing.T) {
	ctx := context.Background()
	pool := openIntegrationDB(t, ctx)
	t.Cleanup(pool.Close)

	truncateProperties(t, ctx, pool)
	t.Cleanup(func() { truncateProperties(t, ctx, pool) })

	campusID := testCampusID(t, ctx, pool)
	propertyID := insertProperty(t, ctx, pool, campusID, "Alice Lodge", time.Date(2026, time.May, 1, 12, 0, 0, 0, time.UTC))
	repository := NewPropertyRepository(pool)

	created, err := repository.CreatePropertyUnitType(ctx, domain.PropertyUnitType{
		PropertyID:  propertyID,
		Category:    domain.UnitCategorySelfContained,
		Name:        "Self-contained",
		Description: "Private room with bathroom and kitchenette.",
		Notes:       "Top floor corner unit with better ventilation.",
		Structure: domain.UnitStructure{
			BedroomCount: intPtr(1),
			HasParlour:   boolPtr(false),
			BathroomType: "private",
			KitchenType:  "private",
		},
	})
	if err != nil {
		t.Fatalf("create property unit type: %v", err)
	}

	if created.ID == "" {
		t.Fatal("created property unit type ID is empty")
	}
	if created.PropertyID != propertyID {
		t.Fatalf("property ID = %q, want %q", created.PropertyID, propertyID)
	}
	if created.Name != "Self-contained" {
		t.Fatalf("name = %q, want Self-contained", created.Name)
	}
	if created.Category != domain.UnitCategorySelfContained {
		t.Fatalf("category = %q, want self_contained", created.Category)
	}
	if created.Structure.BedroomCount == nil || *created.Structure.BedroomCount != 1 {
		t.Fatalf("bedroom_count = %v, want 1", created.Structure.BedroomCount)
	}
	if created.Structure.HasParlour == nil || *created.Structure.HasParlour {
		t.Fatalf("has_parlour = %v, want false", created.Structure.HasParlour)
	}
	if created.Structure.BathroomType != "private" {
		t.Fatalf("bathroom_type = %q, want private", created.Structure.BathroomType)
	}
	if created.Structure.KitchenType != "private" {
		t.Fatalf("kitchen_type = %q, want private", created.Structure.KitchenType)
	}
	if created.Description != "Private room with bathroom and kitchenette." {
		t.Fatalf("description = %q, want Private room with bathroom and kitchenette.", created.Description)
	}
	if created.Notes != "Top floor corner unit with better ventilation." {
		t.Fatalf("notes = %q, want Top floor corner unit with better ventilation.", created.Notes)
	}
	if created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatal("created property unit type timestamps must be set")
	}

	listed, err := repository.ListPropertyUnitTypes(ctx, propertyID)
	if err != nil {
		t.Fatalf("list property unit types: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("unit types length = %d, want 1", len(listed))
	}
	if listed[0].ID != created.ID {
		t.Fatalf("listed property unit type ID = %q, want %q", listed[0].ID, created.ID)
	}
	if listed[0].Category != created.Category {
		t.Fatalf("listed category = %q, want %q", listed[0].Category, created.Category)
	}
	if listed[0].Structure.BedroomCount == nil || *listed[0].Structure.BedroomCount != 1 {
		t.Fatalf("listed bedroom_count = %v, want 1", listed[0].Structure.BedroomCount)
	}
	if listed[0].Structure.HasParlour == nil || *listed[0].Structure.HasParlour {
		t.Fatalf("listed has_parlour = %v, want false", listed[0].Structure.HasParlour)
	}
	if listed[0].Structure.BathroomType != "private" {
		t.Fatalf("listed bathroom_type = %q, want private", listed[0].Structure.BathroomType)
	}
	if listed[0].Structure.KitchenType != "private" {
		t.Fatalf("listed kitchen_type = %q, want private", listed[0].Structure.KitchenType)
	}
	if listed[0].Notes != "Top floor corner unit with better ventilation." {
		t.Fatalf("listed notes = %q, want Top floor corner unit with better ventilation.", listed[0].Notes)
	}
}

func TestPropertyRepositoryPropertyUnitTypesPropertyNotFound(t *testing.T) {
	ctx := context.Background()
	pool := openIntegrationDB(t, ctx)
	t.Cleanup(pool.Close)

	repository := NewPropertyRepository(pool)
	missingPropertyID := domain.ID("550e8400-e29b-41d4-a716-446655440000")

	_, err := repository.CreatePropertyUnitType(ctx, domain.PropertyUnitType{
		PropertyID: missingPropertyID,
		Category:   domain.UnitCategorySelfContained,
		Name:       "Self-contained",
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("create error = %v, want %v", err, ErrNotFound)
	}

	_, err = repository.ListPropertyUnitTypes(ctx, missingPropertyID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("list error = %v, want %v", err, ErrNotFound)
	}
}

func TestPropertyRepositoryCreateAndListAgentOffers(t *testing.T) {
	ctx := context.Background()
	pool := openIntegrationDB(t, ctx)
	t.Cleanup(pool.Close)

	truncateProperties(t, ctx, pool)
	truncateAgents(t, ctx, pool)
	t.Cleanup(func() {
		truncateProperties(t, ctx, pool)
		truncateAgents(t, ctx, pool)
	})

	campusID := testCampusID(t, ctx, pool)
	propertyID := insertProperty(t, ctx, pool, campusID, "Alice Lodge", time.Date(2026, time.May, 1, 12, 0, 0, 0, time.UTC))
	unitTypeID := insertPropertyUnitType(t, ctx, pool, propertyID, "Self-contained")
	agentID := insertAgent(t, ctx, pool, "Dami Agent")
	repository := NewPropertyRepository(pool)

	created, err := repository.CreateAgentOffer(ctx, domain.AgentOffer{
		PropertyUnitTypeID: unitTypeID,
		AgentID:            agentID,
		Title:              "Fresh self-contained room",
		Description:        "Recently painted room with private bathroom.",
		Notes:              "2 left. Inspection tomorrow only.",
		Price:              domain.Money{AmountKobo: 35000000},
		Status:             domain.AgentOfferStatusAvailable,
	})
	if err != nil {
		t.Fatalf("create agent offer: %v", err)
	}

	if created.ID == "" {
		t.Fatal("created agent offer ID is empty")
	}
	if created.PropertyUnitTypeID != unitTypeID {
		t.Fatalf("property unit type ID = %q, want %q", created.PropertyUnitTypeID, unitTypeID)
	}
	if created.AgentID != agentID {
		t.Fatalf("agent ID = %q, want %q", created.AgentID, agentID)
	}
	if created.Price.AmountKobo != 35000000 {
		t.Fatalf("price = %d, want 35000000", created.Price.AmountKobo)
	}
	if created.Status != domain.AgentOfferStatusAvailable {
		t.Fatalf("status = %q, want available", created.Status)
	}
	if created.Notes != "2 left. Inspection tomorrow only." {
		t.Fatalf("notes = %q, want 2 left. Inspection tomorrow only.", created.Notes)
	}

	listed, err := repository.ListAgentOffers(ctx, unitTypeID)
	if err != nil {
		t.Fatalf("list agent offers: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("agent offers length = %d, want 1", len(listed))
	}
	if listed[0].ID != created.ID {
		t.Fatalf("listed agent offer ID = %q, want %q", listed[0].ID, created.ID)
	}
	if listed[0].Notes != "2 left. Inspection tomorrow only." {
		t.Fatalf("listed notes = %q, want 2 left. Inspection tomorrow only.", listed[0].Notes)
	}
}

func TestPropertyRepositoryAgentOfferStoresKoboAndReturnsNaira(t *testing.T) {
	ctx := context.Background()
	pool := openIntegrationDB(t, ctx)
	t.Cleanup(pool.Close)

	truncateProperties(t, ctx, pool)
	truncateAgents(t, ctx, pool)
	t.Cleanup(func() {
		truncateProperties(t, ctx, pool)
		truncateAgents(t, ctx, pool)
	})

	campusID := testCampusID(t, ctx, pool)
	propertyID := insertProperty(t, ctx, pool, campusID, "Alice Lodge", time.Date(2026, time.May, 1, 12, 0, 0, 0, time.UTC))
	unitTypeID := insertPropertyUnitType(t, ctx, pool, propertyID, "Self-contained")
	agentID := insertAgent(t, ctx, pool, "Dami Agent")
	repository := NewPropertyRepository(pool)

	// Simulate what handler does: convert 350000 naira to 35000000 kobo
	nairaInput := 350000
	koboStored := nairaInput * 100

	created, err := repository.CreateAgentOffer(ctx, domain.AgentOffer{
		PropertyUnitTypeID: unitTypeID,
		AgentID:            agentID,
		Title:              "Fresh self-contained room",
		Description:        "Recently painted room with private bathroom.",
		Price:              domain.Money{AmountKobo: koboStored},
		Status:             domain.AgentOfferStatusAvailable,
	})
	if err != nil {
		t.Fatalf("create agent offer: %v", err)
	}

	// Verify stored as kobo
	if created.Price.AmountKobo != 35000000 {
		t.Fatalf("stored price = %d kobo, want 35000000", created.Price.AmountKobo)
	}

	// Verify raw DB value is kobo
	var rawPriceKobo int
	queryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err = pool.QueryRow(queryCtx, `
		SELECT price_kobo FROM agent_offers WHERE id = $1
	`, string(created.ID)).Scan(&rawPriceKobo)
	if err != nil {
		t.Fatalf("query raw price: %v", err)
	}
	if rawPriceKobo != 35000000 {
		t.Fatalf("raw DB price = %d kobo, want 35000000", rawPriceKobo)
	}

	// Verify fetched back as kobo (handler converts to naira for display)
	listed, err := repository.ListAgentOffers(ctx, unitTypeID)
	if err != nil {
		t.Fatalf("list agent offers: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("agent offers length = %d, want 1", len(listed))
	}
	if listed[0].Price.AmountKobo != 35000000 {
		t.Fatalf("fetched price = %d kobo, want 35000000", listed[0].Price.AmountKobo)
	}
}

func TestPropertyRepositoryAgentOffersUnitTypeNotFound(t *testing.T) {
	ctx := context.Background()
	pool := openIntegrationDB(t, ctx)
	t.Cleanup(pool.Close)

	truncateAgents(t, ctx, pool)
	t.Cleanup(func() { truncateAgents(t, ctx, pool) })

	agentID := insertAgent(t, ctx, pool, "Dami Agent")
	repository := NewPropertyRepository(pool)
	missingUnitTypeID := domain.ID("550e8400-e29b-41d4-a716-446655440000")

	_, err := repository.CreateAgentOffer(ctx, domain.AgentOffer{
		PropertyUnitTypeID: missingUnitTypeID,
		AgentID:            agentID,
		Title:              "Fresh self-contained room",
		Price:              domain.Money{AmountKobo: 35000000},
		Status:             domain.AgentOfferStatusAvailable,
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("create error = %v, want %v", err, ErrNotFound)
	}

	_, err = repository.ListAgentOffers(ctx, missingUnitTypeID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("list error = %v, want %v", err, ErrNotFound)
	}
}

func openIntegrationDB(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("LAIVAN_TEST_DB_URL")
	if databaseURL == "" {
		t.Fatal("LAIVAN_TEST_DB_URL is required for integration tests")
	}

	pool, err := appdb.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open integration database: %v", err)
	}

	return pool
}

func truncateProperties(t *testing.T, ctx context.Context, db *pgxpool.Pool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if _, err := db.Exec(ctx, "TRUNCATE properties CASCADE"); err != nil {
		t.Fatalf("truncate properties: %v", err)
	}
}

func truncateAgents(t *testing.T, ctx context.Context, db *pgxpool.Pool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if _, err := db.Exec(ctx, "TRUNCATE agents CASCADE"); err != nil {
		t.Fatalf("truncate agents: %v", err)
	}
}

func testCampusID(t *testing.T, ctx context.Context, db *pgxpool.Pool) domain.ID {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var campusID string
	if err := db.QueryRow(ctx, "SELECT id::text FROM campuses WHERE slug = 'futa'").Scan(&campusID); err != nil {
		t.Fatalf("select FUTA campus: %v", err)
	}

	return domain.ID(campusID)
}

func createTestCampus(t *testing.T, ctx context.Context, db *pgxpool.Pool, slug string) domain.ID {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var campusID string
	err := db.QueryRow(ctx, `
		INSERT INTO campuses (slug, name, short_name)
		VALUES ($1, 'FUTA North', 'FUTA North')
		ON CONFLICT (slug) DO UPDATE SET slug = EXCLUDED.slug
		RETURNING id::text
	`, slug).Scan(&campusID)
	if err != nil {
		t.Fatalf("create test campus: %v", err)
	}

	return domain.ID(campusID)
}

func insertProperty(t *testing.T, ctx context.Context, db *pgxpool.Pool, campusID domain.ID, name string, createdAt time.Time) domain.ID {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var propertyID string
	err := db.QueryRow(ctx, `
		INSERT INTO properties (campus_id, name, area, landmark, description, created_at, updated_at)
		VALUES ($1, $2, 'Obanla', 'Near South Gate', 'Test property', $3, $3)
		RETURNING id::text
	`, string(campusID), name, createdAt).Scan(&propertyID)
	if err != nil {
		t.Fatalf("insert property: %v", err)
	}

	return domain.ID(propertyID)
}

func insertPropertyUnitType(t *testing.T, ctx context.Context, db *pgxpool.Pool, propertyID domain.ID, name string) domain.ID {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var unitTypeID string
	err := db.QueryRow(ctx, `
		INSERT INTO property_unit_types (property_id, name, description)
		VALUES ($1, $2, 'Private room with bathroom and kitchenette.')
		RETURNING id::text
	`, string(propertyID), name).Scan(&unitTypeID)
	if err != nil {
		t.Fatalf("insert property unit type: %v", err)
	}

	return domain.ID(unitTypeID)
}

func insertAgent(t *testing.T, ctx context.Context, db *pgxpool.Pool, displayName string) domain.ID {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var agentID string
	err := db.QueryRow(ctx, `
		INSERT INTO agents (display_name, phone_number)
		VALUES ($1, '+2348012345678')
		RETURNING id::text
	`, displayName).Scan(&agentID)
	if err != nil {
		t.Fatalf("insert agent: %v", err)
	}

	return domain.ID(agentID)
}

func insertAgentOffer(t *testing.T, ctx context.Context, db *pgxpool.Pool, unitTypeID domain.ID, agentID domain.ID, title string, priceKobo int) domain.ID {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var offerID string
	err := db.QueryRow(ctx, `
		INSERT INTO agent_offers (property_unit_type_id, agent_id, title, description, price_kobo, status)
		VALUES ($1, $2, $3, 'Test offer', $4, 'available')
		RETURNING id::text
	`, string(unitTypeID), string(agentID), title, priceKobo).Scan(&offerID)
	if err != nil {
		t.Fatalf("insert agent offer: %v", err)
	}

	return domain.ID(offerID)
}

func intPtr(value int) *int {
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}
