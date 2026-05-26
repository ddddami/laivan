//go:build integration

package repo

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

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
	roomType1 := insertRoomType(t, ctx, pool, propertyID, "Self-contained")
	roomType2 := insertRoomType(t, ctx, pool, propertyID, "Single room")
	agentID := insertAgent(t, ctx, pool, "Dami Agent")
	insertAgentOffer(t, ctx, pool, roomType1, agentID, "Selfcon offer", 25000000)
	insertAgentOffer(t, ctx, pool, roomType2, agentID, "Single room offer", 15000000)

	repository := NewPropertyRepository(pool)

	detail, err := repository.GetWithDetails(ctx, propertyID)
	if err != nil {
		t.Fatalf("get property with details: %v", err)
	}

	if detail.Name != "Alice Lodge" {
		t.Fatalf("name = %q, want Alice Lodge", detail.Name)
	}
	if len(detail.RoomTypes) != 2 {
		t.Fatalf("room types length = %d, want 2", len(detail.RoomTypes))
	}
	if detail.RoomTypes[0].Name != "Self-contained" {
		t.Fatalf("room type name = %q, want Self-contained", detail.RoomTypes[0].Name)
	}
	if len(detail.RoomTypes[0].AgentOffers) != 1 {
		t.Fatalf("agent offers length = %d, want 1", len(detail.RoomTypes[0].AgentOffers))
	}
	if detail.RoomTypes[0].AgentOffers[0].Price.AmountKobo != 25000000 {
		t.Fatalf("price = %d, want 25000000", detail.RoomTypes[0].AgentOffers[0].Price.AmountKobo)
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

	// Add room types and offers to newer property for summary testing
	roomType1 := insertRoomType(t, ctx, pool, newer, "Self-contained")
	roomType2 := insertRoomType(t, ctx, pool, newer, "Single room")
	agentID := insertAgent(t, ctx, pool, "Dami Agent")
	insertAgentOffer(t, ctx, pool, roomType1, agentID, "Selfcon offer", 25000000)
	insertAgentOffer(t, ctx, pool, roomType2, agentID, "Single room offer", 15000000)

	repository := NewPropertyRepository(pool)

	// Test plain List still works
	properties, err := repository.List(ctx, PropertyListFilter{CampusID: campusID, Limit: 10})
	if err != nil {
		t.Fatalf("list properties: %v", err)
	}
	if len(properties) != 2 {
		t.Fatalf("properties length = %d, want 2", len(properties))
	}

	// Test ListWithSummary
	summaries, err := repository.ListWithSummary(ctx, PropertyListFilter{CampusID: campusID, Limit: 10})
	if err != nil {
		t.Fatalf("list properties with summary: %v", err)
	}
	if len(summaries) != 2 {
		t.Fatalf("summaries length = %d, want 2", len(summaries))
	}
	// Newer property should be first (newest first)
	if summaries[0].ID != newer {
		t.Fatalf("first summary ID = %q, want %q", summaries[0].ID, newer)
	}
	if summaries[0].RoomTypeCount != 2 {
		t.Fatalf("room_type_count = %d, want 2", summaries[0].RoomTypeCount)
	}
	if summaries[0].AvailableOfferCount != 2 {
		t.Fatalf("available_offer_count = %d, want 2", summaries[0].AvailableOfferCount)
	}
	if summaries[0].LowestPriceKobo != 15000000 {
		t.Fatalf("lowest_price_kobo = %d, want 15000000", summaries[0].LowestPriceKobo)
	}

	// Older property has no room types or offers
	if summaries[1].ID != older {
		t.Fatalf("second summary ID = %q, want %q", summaries[1].ID, older)
	}
	if summaries[1].RoomTypeCount != 0 {
		t.Fatalf("room_type_count = %d, want 0", summaries[1].RoomTypeCount)
	}
	if summaries[1].LowestPriceKobo != 0 {
		t.Fatalf("lowest_price_kobo = %d, want 0", summaries[1].LowestPriceKobo)
	}
}

func TestPropertyRepositoryCreateAndListRoomTypes(t *testing.T) {
	ctx := context.Background()
	pool := openIntegrationDB(t, ctx)
	t.Cleanup(pool.Close)

	truncateProperties(t, ctx, pool)
	t.Cleanup(func() { truncateProperties(t, ctx, pool) })

	campusID := testCampusID(t, ctx, pool)
	propertyID := insertProperty(t, ctx, pool, campusID, "Alice Lodge", time.Date(2026, time.May, 1, 12, 0, 0, 0, time.UTC))
	repository := NewPropertyRepository(pool)

	created, err := repository.CreateRoomType(ctx, domain.RoomType{
		PropertyID:  propertyID,
		Name:        "Self-contained",
		Description: "Private room with bathroom and kitchenette.",
	})
	if err != nil {
		t.Fatalf("create room type: %v", err)
	}

	if created.ID == "" {
		t.Fatal("created room type ID is empty")
	}
	if created.PropertyID != propertyID {
		t.Fatalf("property ID = %q, want %q", created.PropertyID, propertyID)
	}
	if created.Name != "Self-contained" {
		t.Fatalf("name = %q, want Self-contained", created.Name)
	}
	if created.Description != "Private room with bathroom and kitchenette." {
		t.Fatalf("description = %q, want Private room with bathroom and kitchenette.", created.Description)
	}
	if created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatal("created room type timestamps must be set")
	}

	listed, err := repository.ListRoomTypes(ctx, propertyID)
	if err != nil {
		t.Fatalf("list room types: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("room types length = %d, want 1", len(listed))
	}
	if listed[0] != created {
		t.Fatalf("listed room type = %#v, want %#v", listed[0], created)
	}
}

func TestPropertyRepositoryRoomTypesPropertyNotFound(t *testing.T) {
	ctx := context.Background()
	pool := openIntegrationDB(t, ctx)
	t.Cleanup(pool.Close)

	repository := NewPropertyRepository(pool)
	missingPropertyID := domain.ID("550e8400-e29b-41d4-a716-446655440000")

	_, err := repository.CreateRoomType(ctx, domain.RoomType{
		PropertyID: missingPropertyID,
		Name:       "Self-contained",
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("create error = %v, want %v", err, ErrNotFound)
	}

	_, err = repository.ListRoomTypes(ctx, missingPropertyID)
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
	roomTypeID := insertRoomType(t, ctx, pool, propertyID, "Self-contained")
	agentID := insertAgent(t, ctx, pool, "Dami Agent")
	repository := NewPropertyRepository(pool)

	created, err := repository.CreateAgentOffer(ctx, domain.AgentOffer{
		RoomTypeID:  roomTypeID,
		AgentID:     agentID,
		Title:       "Fresh self-contained room",
		Description: "Recently painted room with private bathroom.",
		Price:       domain.Money{AmountKobo: 35000000},
		Status:      domain.AgentOfferStatusAvailable,
	})
	if err != nil {
		t.Fatalf("create agent offer: %v", err)
	}

	if created.ID == "" {
		t.Fatal("created agent offer ID is empty")
	}
	if created.RoomTypeID != roomTypeID {
		t.Fatalf("room type ID = %q, want %q", created.RoomTypeID, roomTypeID)
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

	listed, err := repository.ListAgentOffers(ctx, roomTypeID)
	if err != nil {
		t.Fatalf("list agent offers: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("agent offers length = %d, want 1", len(listed))
	}
	if listed[0] != created {
		t.Fatalf("listed agent offer = %#v, want %#v", listed[0], created)
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
	roomTypeID := insertRoomType(t, ctx, pool, propertyID, "Self-contained")
	agentID := insertAgent(t, ctx, pool, "Dami Agent")
	repository := NewPropertyRepository(pool)

	// Simulate what handler does: convert 350000 naira to 35000000 kobo
	nairaInput := int32(350000)
	koboStored := nairaInput * 100

	created, err := repository.CreateAgentOffer(ctx, domain.AgentOffer{
		RoomTypeID:  roomTypeID,
		AgentID:     agentID,
		Title:       "Fresh self-contained room",
		Description: "Recently painted room with private bathroom.",
		Price:       domain.Money{AmountKobo: koboStored},
		Status:      domain.AgentOfferStatusAvailable,
	})
	if err != nil {
		t.Fatalf("create agent offer: %v", err)
	}

	// Verify stored as kobo
	if created.Price.AmountKobo != 35000000 {
		t.Fatalf("stored price = %d kobo, want 35000000", created.Price.AmountKobo)
	}

	// Verify raw DB value is kobo
	var rawPriceKobo int32
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
	listed, err := repository.ListAgentOffers(ctx, roomTypeID)
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

func TestPropertyRepositoryAgentOffersRoomTypeNotFound(t *testing.T) {
	ctx := context.Background()
	pool := openIntegrationDB(t, ctx)
	t.Cleanup(pool.Close)

	truncateAgents(t, ctx, pool)
	t.Cleanup(func() { truncateAgents(t, ctx, pool) })

	agentID := insertAgent(t, ctx, pool, "Dami Agent")
	repository := NewPropertyRepository(pool)
	missingRoomTypeID := domain.ID("550e8400-e29b-41d4-a716-446655440000")

	_, err := repository.CreateAgentOffer(ctx, domain.AgentOffer{
		RoomTypeID: missingRoomTypeID,
		AgentID:    agentID,
		Title:      "Fresh self-contained room",
		Price:      domain.Money{AmountKobo: 35000000},
		Status:     domain.AgentOfferStatusAvailable,
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("create error = %v, want %v", err, ErrNotFound)
	}

	_, err = repository.ListAgentOffers(ctx, missingRoomTypeID)
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

func insertRoomType(t *testing.T, ctx context.Context, db *pgxpool.Pool, propertyID domain.ID, name string) domain.ID {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var roomTypeID string
	err := db.QueryRow(ctx, `
		INSERT INTO room_types (property_id, name, description)
		VALUES ($1, $2, 'Private room with bathroom and kitchenette.')
		RETURNING id::text
	`, string(propertyID), name).Scan(&roomTypeID)
	if err != nil {
		t.Fatalf("insert room type: %v", err)
	}

	return domain.ID(roomTypeID)
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

func insertAgentOffer(t *testing.T, ctx context.Context, db *pgxpool.Pool, roomTypeID domain.ID, agentID domain.ID, title string, priceKobo int32) domain.ID {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var offerID string
	err := db.QueryRow(ctx, `
		INSERT INTO agent_offers (room_type_id, agent_id, title, description, price_kobo, status)
		VALUES ($1, $2, $3, 'Test offer', $4, 'available')
		RETURNING id::text
	`, string(roomTypeID), string(agentID), title, priceKobo).Scan(&offerID)
	if err != nil {
		t.Fatalf("insert agent offer: %v", err)
	}

	return domain.ID(offerID)
}
