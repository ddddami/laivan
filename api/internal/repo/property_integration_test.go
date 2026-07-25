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
	"github.com/jackc/pgx/v5/pgconn"
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

func TestPropertyRepositoryGetCampusBySlug(t *testing.T) {
	ctx := context.Background()
	pool := openIntegrationDB(t, ctx)
	t.Cleanup(pool.Close)

	repository := NewPropertyRepository(pool)
	campus, err := repository.GetCampusBySlug(ctx, "futa")
	if err != nil {
		t.Fatalf("get FUTA campus: %v", err)
	}
	if campus.Slug != "futa" || campus.ShortName != "FUTA" {
		t.Fatalf("campus = %#v, want FUTA campus", campus)
	}

	inactiveCampusID := createTestCampus(t, ctx, pool, "inactive-campus")
	if _, err := pool.Exec(ctx, "UPDATE campuses SET is_active = false WHERE id = $1", string(inactiveCampusID)); err != nil {
		t.Fatalf("deactivate campus: %v", err)
	}
	_, err = repository.GetCampusBySlug(ctx, "inactive-campus")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("inactive campus error = %v, want %v", err, ErrNotFound)
	}

	_, err = repository.GetCampusBySlug(ctx, "missing-campus")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing campus error = %v, want %v", err, ErrNotFound)
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
	offerID := insertAgentOffer(t, ctx, pool, unitType1, agentID, "Selfcon offer", 25000000)
	insertAgentOffer(t, ctx, pool, unitType2, agentID, "Single room offer", 15000000)

	repository := NewPropertyRepository(pool)
	if _, err := repository.CreateMedia(ctx, domain.Media{
		PropertyUnitTypeID: unitType1,
		UploadedByAgentID:  agentID,
		URL:                "https://media.example.test/self-contained.jpg",
		Kind:               domain.MediaKindImage,
	}); err != nil {
		t.Fatalf("create self-contained media: %v", err)
	}
	if _, err := repository.CreateMedia(ctx, domain.Media{
		PropertyUnitTypeID: unitType2,
		UploadedByAgentID:  agentID,
		URL:                "https://media.example.test/single-room.jpg",
		Kind:               domain.MediaKindImage,
	}); err != nil {
		t.Fatalf("create single-room media: %v", err)
	}
	if _, err := repository.CreateMedia(ctx, domain.Media{
		AgentOfferID:      offerID,
		UploadedByAgentID: agentID,
		URL:               "https://media.example.test/self-contained-offer.jpg",
		Kind:              domain.MediaKindImage,
	}); err != nil {
		t.Fatalf("create agent offer media: %v", err)
	}

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
	if detail.UnitTypes[0].AgentOffers[0].Agent.DisplayName != "Dami Agent" {
		t.Fatalf("agent display name = %q, want Dami Agent", detail.UnitTypes[0].AgentOffers[0].Agent.DisplayName)
	}
	if len(detail.UnitTypes[0].AgentOffers[0].Media) != 1 ||
		detail.UnitTypes[0].AgentOffers[0].Media[0].URL != "https://media.example.test/self-contained-offer.jpg" {
		t.Fatalf("agent offer media = %#v", detail.UnitTypes[0].AgentOffers[0].Media)
	}
	if len(detail.UnitTypes[0].Media) != 1 || detail.UnitTypes[0].Media[0].URL != "https://media.example.test/self-contained.jpg" {
		t.Fatalf("self-contained media = %#v", detail.UnitTypes[0].Media)
	}
	if detail.UnitTypes[1].Name != "Single room" {
		t.Fatalf("unit type name = %q, want Single room", detail.UnitTypes[1].Name)
	}
	if len(detail.UnitTypes[1].AgentOffers) != 1 || detail.UnitTypes[1].AgentOffers[0].Price.AmountKobo != 15000000 {
		t.Fatalf("single-room offers = %#v", detail.UnitTypes[1].AgentOffers)
	}
	if len(detail.UnitTypes[1].Media) != 1 || detail.UnitTypes[1].Media[0].URL != "https://media.example.test/single-room.jpg" {
		t.Fatalf("single-room media = %#v", detail.UnitTypes[1].Media)
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

func TestPropertyRepositoryCreateAndListMedia(t *testing.T) {
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
	agentID := insertAgent(t, ctx, pool, "Dami Agent")
	repository := NewPropertyRepository(pool)

	created, err := repository.CreateMedia(ctx, domain.Media{
		PropertyID:        propertyID,
		UploadedByAgentID: agentID,
		URL:               "https://media.example.test/alice/front.jpg",
		ObjectKey:         "media/property/alice/front.jpg",
		Kind:              domain.MediaKindImage,
		Caption:           "Front view",
		ContentType:       "image/jpeg",
		SizeBytes:         2048,
	})
	if err != nil {
		t.Fatalf("create media: %v", err)
	}

	if created.ID == "" {
		t.Fatal("created media ID is empty")
	}
	if created.PropertyID != propertyID {
		t.Fatalf("property ID = %q, want %q", created.PropertyID, propertyID)
	}
	if created.Caption != "Front view" {
		t.Fatalf("caption = %q, want Front view", created.Caption)
	}
	if created.ContentType != "image/jpeg" {
		t.Fatalf("content type = %q, want image/jpeg", created.ContentType)
	}
	if created.SizeBytes != 2048 {
		t.Fatalf("size bytes = %d, want 2048", created.SizeBytes)
	}

	items, err := repository.ListMediaByProperty(ctx, propertyID)
	if err != nil {
		t.Fatalf("list media by property: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("media length = %d, want 1", len(items))
	}
	if items[0].URL != "https://media.example.test/alice/front.jpg" {
		t.Fatalf("media URL = %q, want original URL", items[0].URL)
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

	// Simulate what handler does: convert 350000 naira to kobo
	nairaInput := 350000
	koboStored := domain.Kobo(nairaInput)

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
	if !errors.Is(err, ErrUnitTypeNotFound) {
		t.Fatalf("create error = %v, want %v", err, ErrUnitTypeNotFound)
	}

	_, err = repository.ListAgentOffers(ctx, missingUnitTypeID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("list error = %v, want %v", err, ErrNotFound)
	}
}

func TestPropertyRepositoryAgentOffersAgentNotFound(t *testing.T) {
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
	propertyID := insertProperty(t, ctx, pool, campusID, "Alice Lodge", time.Now().UTC())
	unitTypeID := insertPropertyUnitType(t, ctx, pool, propertyID, "Self-contained")
	repository := NewPropertyRepository(pool)

	_, err := repository.CreateAgentOffer(ctx, domain.AgentOffer{
		PropertyUnitTypeID: unitTypeID,
		AgentID:            domain.ID("550e8400-e29b-41d4-a716-446655440040"),
		Title:              "Fresh self-contained room",
		Price:              domain.Money{AmountKobo: 35000000},
		Status:             domain.AgentOfferStatusAvailable,
	})
	if !errors.Is(err, ErrAgentNotFound) {
		t.Fatalf("create error = %v, want %v", err, ErrAgentNotFound)
	}
}

func openIntegrationDB(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("LAIVAN_TEST_DB_URL")
	if databaseURL == "" {
		t.Fatal("LAIVAN_TEST_DB_URL is required for integration tests")
	}

	pool, err := appdb.Open(ctx, databaseURL, 25)
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

func insertPropertyWithArea(t *testing.T, ctx context.Context, db *pgxpool.Pool, campusID domain.ID, name string, area string, createdAt time.Time) domain.ID {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var propertyID string
	err := db.QueryRow(ctx, `
		INSERT INTO properties (campus_id, name, area, landmark, description, created_at, updated_at)
		VALUES ($1, $2, $3, 'Near South Gate', 'Test property', $4, $4)
		RETURNING id::text
	`, string(campusID), name, area, createdAt).Scan(&propertyID)
	if err != nil {
		t.Fatalf("insert property with area: %v", err)
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

func insertPropertyUnitTypeFull(t *testing.T, ctx context.Context, db *pgxpool.Pool, propertyID domain.ID, category, name string, bedroomCount int, hasParlour bool, bathroomType, kitchenType, notes string) domain.ID {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var unitTypeID string
	err := db.QueryRow(ctx, `
		INSERT INTO property_unit_types (property_id, category, name, description, notes, bedroom_count, has_parlour, bathroom_type, kitchen_type)
		VALUES ($1, $2, $3, 'Test unit type', $4, $5, $6, $7, $8)
		RETURNING id::text
	`, string(propertyID), category, name, notes, bedroomCount, hasParlour, bathroomType, kitchenType).Scan(&unitTypeID)
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

	return insertAgentOfferWithStatus(t, ctx, db, unitTypeID, agentID, title, priceKobo, "available")
}

func insertAgentOfferWithStatus(t *testing.T, ctx context.Context, db *pgxpool.Pool, unitTypeID domain.ID, agentID domain.ID, title string, priceKobo int, status string) domain.ID {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var offerID string
	err := db.QueryRow(ctx, `
		INSERT INTO agent_offers (property_unit_type_id, agent_id, title, description, price_kobo, status)
		VALUES ($1, $2, $3, 'Test offer', $4, $5)
		RETURNING id::text
	`, string(unitTypeID), string(agentID), title, priceKobo, status).Scan(&offerID)
	if err != nil {
		t.Fatalf("insert agent offer: %v", err)
	}

	return domain.ID(offerID)
}

func TestRepositoryDiscover(t *testing.T) {
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
	agentID := insertAgent(t, ctx, pool, "Test Agent")

	// Property A: Alice Lodge, Obanla
	alice := insertPropertyWithArea(t, ctx, pool, campusID, "Alice Lodge", "Obanla", time.Date(2026, time.May, 3, 12, 0, 0, 0, time.UTC))
	aliceSelfCon := insertPropertyUnitTypeFull(t, ctx, pool, alice, "self_contained", "", 1, false, "private", "private", "Top floor")
	aliceSingle := insertPropertyUnitTypeFull(t, ctx, pool, alice, "single_room", "", 1, false, "shared", "shared", "Upstairs")
	insertAgentOffer(t, ctx, pool, aliceSelfCon, agentID, "Alice self-con", 35000000)
	insertAgentOffer(t, ctx, pool, aliceSingle, agentID, "Alice single", 18000000)

	// Property B: Blue Roof, Aule
	blueRoof := insertPropertyWithArea(t, ctx, pool, campusID, "Blue Roof", "Aule", time.Date(2026, time.May, 2, 12, 0, 0, 0, time.UTC))
	blueRoomParlour := insertPropertyUnitTypeFull(t, ctx, pool, blueRoof, "room_and_parlour", "", 1, true, "private", "private", "Gate closes by 10pm")
	blueSelfCon := insertPropertyUnitTypeFull(t, ctx, pool, blueRoof, "self_contained", "", 1, false, "private", "private", "Ground floor")
	insertAgentOffer(t, ctx, pool, blueRoomParlour, agentID, "Blue room and parlour", 50000000)
	insertAgentOffer(t, ctx, pool, blueSelfCon, agentID, "Blue self-con", 32000000)

	// Property C: Quiet Place, South Gate
	quiet := insertPropertyWithArea(t, ctx, pool, campusID, "Quiet Place", "South Gate", time.Date(2026, time.May, 1, 12, 0, 0, 0, time.UTC))
	quietSingle := insertPropertyUnitTypeFull(t, ctx, pool, quiet, "single_room", "", 1, false, "shared", "shared", "Shared kitchen")
	insertAgentOffer(t, ctx, pool, quietSingle, agentID, "Quiet single", 15000000)

	// Property D: a unit type with no agent offers yet and valid property media.
	empty := insertPropertyWithArea(t, ctx, pool, campusID, "Empty Lodge", "Obanla", time.Date(2026, time.April, 30, 12, 0, 0, 0, time.UTC))
	insertPropertyUnitTypeFull(t, ctx, pool, empty, "single_room", "", 1, false, "shared", "shared", "No offers yet")

	// Property E: only a paused offer, a video, and paused-offer media.
	pausedProperty := insertPropertyWithArea(t, ctx, pool, campusID, "Paused Lodge", "Obanla", time.Date(2026, time.April, 29, 12, 0, 0, 0, time.UTC))
	pausedUnit := insertPropertyUnitTypeFull(t, ctx, pool, pausedProperty, "single_room", "", 1, false, "shared", "shared", "Paused offer only")
	pausedOffer := insertAgentOfferWithStatus(t, ctx, pool, pausedUnit, agentID, "Paused single", 12000000, "paused")

	// An available offer on another campus must never leak into FUTA discovery.
	otherCampusID := createTestCampus(t, ctx, pool, "other-discovery-campus")
	otherProperty := insertPropertyWithArea(t, ctx, pool, otherCampusID, "Other Campus Lodge", "Obanla", time.Date(2026, time.May, 4, 12, 0, 0, 0, time.UTC))
	otherUnit := insertPropertyUnitTypeFull(t, ctx, pool, otherProperty, "single_room", "", 1, false, "shared", "shared", "Other campus")
	insertAgentOffer(t, ctx, pool, otherUnit, agentID, "Other campus single", 10000000)

	repository := NewPropertyRepository(pool)
	if _, err := repository.CreateMedia(ctx, domain.Media{
		PropertyUnitTypeID: blueRoomParlour,
		UploadedByAgentID:  agentID,
		URL:                "https://media.example.test/blue-roof.jpg",
		Kind:               domain.MediaKindImage,
	}); err != nil {
		t.Fatalf("create discovery media: %v", err)
	}
	if _, err := repository.CreateMedia(ctx, domain.Media{
		PropertyID:        empty,
		UploadedByAgentID: agentID,
		URL:               "https://media.example.test/empty-lodge.jpg",
		Kind:              domain.MediaKindImage,
	}); err != nil {
		t.Fatalf("create no-offer property media: %v", err)
	}
	if _, err := repository.CreateMedia(ctx, domain.Media{
		PropertyUnitTypeID: pausedUnit,
		UploadedByAgentID:  agentID,
		URL:                "https://media.example.test/paused-lodge.mp4",
		Kind:               domain.MediaKindVideo,
	}); err != nil {
		t.Fatalf("create paused unit video: %v", err)
	}
	if _, err := repository.CreateMedia(ctx, domain.Media{
		AgentOfferID:      pausedOffer,
		UploadedByAgentID: agentID,
		URL:               "https://media.example.test/paused-offer.jpg",
		Kind:              domain.MediaKindImage,
	}); err != nil {
		t.Fatalf("create paused offer media: %v", err)
	}
	baseFilter := data.Filters{Page: 1, PageSize: 10, Sort: "-created_at", SortSafelist: []string{"created_at", "-created_at"}}

	// Public discovery defaults to unit types with at least one available offer.
	results, total, err := repository.Discover(ctx, DiscoveryFilter{CampusID: campusID, Filters: baseFilter})
	if err != nil {
		t.Fatalf("discover available: %v", err)
	}
	if total != 5 {
		t.Fatalf("available total = %d, want 5", total)
	}
	if len(results) != 5 {
		t.Fatalf("available results length = %d, want 5", len(results))
	}

	// Callers can explicitly include unit types without available offers.
	results, total, err = repository.Discover(ctx, DiscoveryFilter{
		CampusID:     campusID,
		Availability: DiscoveryAvailabilityAll,
		Filters:      baseFilter,
	})
	if err != nil {
		t.Fatalf("discover all: %v", err)
	}
	if total != 7 {
		t.Fatalf("all total = %d, want 7", total)
	}
	for _, result := range results {
		if result.PropertyName == "Other Campus Lodge" {
			t.Fatal("other campus result leaked into FUTA discovery")
		}
		if result.PropertyName == "Paused Lodge" {
			if result.AvailableOfferCount != 0 {
				t.Fatalf("paused result available offer count = %d, want 0", result.AvailableOfferCount)
			}
			if result.ThumbnailURL != "" {
				t.Fatalf("paused result thumbnail = %q, want empty", result.ThumbnailURL)
			}
		}
	}

	// Filter by single category
	results, total, err = repository.Discover(ctx, DiscoveryFilter{CampusID: campusID, Categories: []string{"self_contained"}, Filters: baseFilter})
	if err != nil {
		t.Fatalf("discover by category: %v", err)
	}
	if total != 2 {
		t.Fatalf("self_contained total = %d, want 2", total)
	}

	// Filter by multiple categories
	results, total, err = repository.Discover(ctx, DiscoveryFilter{CampusID: campusID, Categories: []string{"self_contained", "single_room"}, Filters: baseFilter})
	if err != nil {
		t.Fatalf("discover by categories: %v", err)
	}
	if total != 4 {
		t.Fatalf("multi-category total = %d, want 4", total)
	}

	// Filter by area partial match
	results, total, err = repository.Discover(ctx, DiscoveryFilter{CampusID: campusID, Area: "Obanla", Filters: baseFilter})
	if err != nil {
		t.Fatalf("discover by area: %v", err)
	}
	if total != 2 {
		t.Fatalf("Obanla total = %d, want 2", total)
	}

	// Filter by min price (naira converted to kobo internally)
	minPrice := 200000
	results, total, err = repository.Discover(ctx, DiscoveryFilter{CampusID: campusID, MinPrice: &minPrice, Filters: baseFilter})
	if err != nil {
		t.Fatalf("discover by min price: %v", err)
	}
	if total != 3 {
		t.Fatalf("min_price=200000 total = %d, want 3", total)
	}

	// Filter by max price
	maxPrice := 200000
	results, total, err = repository.Discover(ctx, DiscoveryFilter{CampusID: campusID, MaxPrice: &maxPrice, Filters: baseFilter})
	if err != nil {
		t.Fatalf("discover by max price: %v", err)
	}
	if total != 2 {
		t.Fatalf("max_price=200000 total = %d, want 2", total)
	}

	priceSort := data.Filters{
		Page:          1,
		PageSize:      10,
		Sort:          "lowest_price_kobo",
		SortSafelist:  []string{"lowest_price_kobo", "-lowest_price_kobo"},
		SortColumnMap: map[string]string{"lowest_price_kobo": "lowest_price_kobo"},
	}
	results, _, err = repository.Discover(ctx, DiscoveryFilter{CampusID: campusID, Filters: priceSort})
	if err != nil {
		t.Fatalf("discover sorted by price: %v", err)
	}
	if results[0].PropertyName != "Quiet Place" {
		t.Fatalf("first price-sorted property = %q, want Quiet Place", results[0].PropertyName)
	}
	if results[len(results)-1].PropertyName != "Blue Roof" {
		t.Fatalf("last price-sorted property = %q, want Blue Roof", results[len(results)-1].PropertyName)
	}
	priceSort.Sort = "-lowest_price_kobo"
	results, _, err = repository.Discover(ctx, DiscoveryFilter{CampusID: campusID, Filters: priceSort})
	if err != nil {
		t.Fatalf("discover sorted by descending price: %v", err)
	}
	if results[0].PropertyName != "Blue Roof" || results[0].LowestPrice.AmountKobo != 50000000 {
		t.Fatalf("first descending price result = %#v", results[0])
	}
	if results[len(results)-1].PropertyName != "Quiet Place" {
		t.Fatalf("last descending price property = %q, want Quiet Place", results[len(results)-1].PropertyName)
	}

	// Filter by bathroom_type
	results, total, err = repository.Discover(ctx, DiscoveryFilter{CampusID: campusID, BathroomType: "private", Filters: baseFilter})
	if err != nil {
		t.Fatalf("discover by bathroom: %v", err)
	}
	if total != 3 {
		t.Fatalf("private bathroom total = %d, want 3", total)
	}

	// Filter by has_parlour
	hasParlour := true
	results, total, err = repository.Discover(ctx, DiscoveryFilter{CampusID: campusID, HasParlour: &hasParlour, Filters: baseFilter})
	if err != nil {
		t.Fatalf("discover by parlour: %v", err)
	}
	if total != 1 {
		t.Fatalf("has_parlour=true total = %d, want 1", total)
	}

	// Combined filters
	maxPrice = 330000
	results, total, err = repository.Discover(ctx, DiscoveryFilter{CampusID: campusID, Categories: []string{"self_contained"}, MaxPrice: &maxPrice, Filters: baseFilter})
	if err != nil {
		t.Fatalf("discover combined: %v", err)
	}
	if total != 1 {
		t.Fatalf("self_contained + max_price=330000 total = %d, want 1", total)
	}
	if results[0].PropertyName != "Blue Roof" {
		t.Fatalf("property name = %q, want Blue Roof", results[0].PropertyName)
	}

	recommended := data.Filters{
		Page:         1,
		PageSize:     10,
		Sort:         "recommended",
		SortSafelist: []string{"recommended"},
	}
	results, _, err = repository.Discover(ctx, DiscoveryFilter{CampusID: campusID, Filters: recommended})
	if err != nil {
		t.Fatalf("discover recommended: %v", err)
	}
	if results[0].PropertyName != "Blue Roof" || results[0].ThumbnailURL == "" {
		t.Fatalf("first recommended result = %#v, want Blue Roof result with media", results[0])
	}
	if results[0].UpdatedAt.IsZero() {
		t.Fatal("recommended result updated_at is zero")
	}

	allRecommendedFilter := DiscoveryFilter{
		CampusID:     campusID,
		Availability: DiscoveryAvailabilityAll,
		Filters:      recommended,
	}
	allRecommended, _, err := repository.Discover(ctx, allRecommendedFilter)
	if err != nil {
		t.Fatalf("discover all recommended: %v", err)
	}
	if allRecommended[0].AvailableOfferCount == 0 {
		t.Fatalf("first all-availability result = %#v, want an available opportunity", allRecommended[0])
	}

	repeated, _, err := repository.Discover(ctx, allRecommendedFilter)
	if err != nil {
		t.Fatalf("repeat all recommended: %v", err)
	}
	if len(repeated) != len(allRecommended) {
		t.Fatalf("repeated result length = %d, want %d", len(repeated), len(allRecommended))
	}
	for i := range allRecommended {
		if repeated[i].UnitTypeID != allRecommended[i].UnitTypeID {
			t.Fatalf("result %d changed from %q to %q across identical requests", i, allRecommended[i].UnitTypeID, repeated[i].UnitTypeID)
		}
	}
}

func TestRepositoryListPropertiesWithFilters(t *testing.T) {
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
	agentID := insertAgent(t, ctx, pool, "Test Agent")

	// Property with offers
	alice := insertProperty(t, ctx, pool, campusID, "Alice Lodge", time.Date(2026, time.May, 3, 12, 0, 0, 0, time.UTC))
	aliceUnit := insertPropertyUnitType(t, ctx, pool, alice, "Self-contained")
	insertAgentOffer(t, ctx, pool, aliceUnit, agentID, "Alice offer", 25000000)

	// Property without offers
	insertProperty(t, ctx, pool, campusID, "Empty Lodge", time.Date(2026, time.May, 2, 12, 0, 0, 0, time.UTC))

	// Property with higher price
	premium := insertProperty(t, ctx, pool, campusID, "Premium Lodge", time.Date(2026, time.May, 1, 12, 0, 0, 0, time.UTC))
	premiumUnit := insertPropertyUnitType(t, ctx, pool, premium, "Self-contained")
	insertAgentOffer(t, ctx, pool, premiumUnit, agentID, "Premium offer", 45000000)

	repository := NewPropertyRepository(pool)
	baseFilter := data.Filters{Page: 1, PageSize: 10, Sort: "-created_at", SortSafelist: []string{"created_at", "-created_at"}}

	// Filter by name partial match
	summaries, total, err := repository.ListWithSummary(ctx, PropertyListFilter{CampusID: campusID, Name: "Alice", Filters: baseFilter})
	if err != nil {
		t.Fatalf("list by name: %v", err)
	}
	if total != 1 {
		t.Fatalf("name='Alice' total = %d, want 1", total)
	}
	if summaries[0].Name != "Alice Lodge" {
		t.Fatalf("name = %q, want Alice Lodge", summaries[0].Name)
	}

	// Filter by has_offers=true
	hasOffers := true
	summaries, total, err = repository.ListWithSummary(ctx, PropertyListFilter{CampusID: campusID, HasOffers: &hasOffers, Filters: baseFilter})
	if err != nil {
		t.Fatalf("list has_offers=true: %v", err)
	}
	if total != 2 {
		t.Fatalf("has_offers=true total = %d, want 2", total)
	}

	// Filter by has_offers=false
	hasOffers = false
	summaries, total, err = repository.ListWithSummary(ctx, PropertyListFilter{CampusID: campusID, HasOffers: &hasOffers, Filters: baseFilter})
	if err != nil {
		t.Fatalf("list has_offers=false: %v", err)
	}
	if total != 1 {
		t.Fatalf("has_offers=false total = %d, want 1", total)
	}
	if summaries[0].Name != "Empty Lodge" {
		t.Fatalf("name = %q, want Empty Lodge", summaries[0].Name)
	}

	// Filter by min price
	minPrice := 300000
	summaries, total, err = repository.ListWithSummary(ctx, PropertyListFilter{CampusID: campusID, MinPrice: &minPrice, Filters: baseFilter})
	if err != nil {
		t.Fatalf("list min_price=300000: %v", err)
	}
	if total != 1 {
		t.Fatalf("min_price=300000 total = %d, want 1", total)
	}
	if summaries[0].Name != "Premium Lodge" {
		t.Fatalf("name = %q, want Premium Lodge", summaries[0].Name)
	}

	// Filter by max price excludes properties with no pricing
	maxPrice := 300000
	summaries, total, err = repository.ListWithSummary(ctx, PropertyListFilter{CampusID: campusID, MaxPrice: &maxPrice, Filters: baseFilter})
	if err != nil {
		t.Fatalf("list max_price=300000: %v", err)
	}
	if total != 1 {
		t.Fatalf("max_price=300000 total = %d, want 1", total)
	}
	if summaries[0].Name != "Alice Lodge" {
		t.Fatalf("name = %q, want Alice Lodge", summaries[0].Name)
	}

	priceSort := data.Filters{
		Page:          1,
		PageSize:      10,
		Sort:          "lowest_price_kobo",
		SortSafelist:  []string{"lowest_price_kobo", "-lowest_price_kobo"},
		SortColumnMap: map[string]string{"lowest_price_kobo": "lowest_price_kobo"},
	}
	summaries, _, err = repository.ListWithSummary(ctx, PropertyListFilter{CampusID: campusID, Filters: priceSort})
	if err != nil {
		t.Fatalf("list sorted by price: %v", err)
	}
	if summaries[0].Name != "Alice Lodge" {
		t.Fatalf("first price-sorted property = %q, want Alice Lodge", summaries[0].Name)
	}
	if summaries[len(summaries)-1].Name != "Empty Lodge" {
		t.Fatalf("last price-sorted property = %q, want Empty Lodge", summaries[len(summaries)-1].Name)
	}
	priceSort.Sort = "-lowest_price_kobo"
	summaries, _, err = repository.ListWithSummary(ctx, PropertyListFilter{CampusID: campusID, Filters: priceSort})
	if err != nil {
		t.Fatalf("list sorted by descending price: %v", err)
	}
	if summaries[0].Name != "Premium Lodge" {
		t.Fatalf("first descending price property = %q, want Premium Lodge", summaries[0].Name)
	}
	if summaries[len(summaries)-1].Name != "Empty Lodge" {
		t.Fatalf("last descending price property = %q, want Empty Lodge", summaries[len(summaries)-1].Name)
	}
}

func TestPropertyRepositoryAgentPhoneNumberUnique(t *testing.T) {
	ctx := context.Background()
	pool := openIntegrationDB(t, ctx)
	t.Cleanup(pool.Close)

	truncateAgents(t, ctx, pool)
	t.Cleanup(func() { truncateAgents(t, ctx, pool) })

	insertAgent(t, ctx, pool, "First Agent")

	// Attempt to insert a second agent with the same phone number.
	_, err := pool.Exec(ctx, `
		INSERT INTO agents (display_name, phone_number)
		VALUES ($1, '+2348012345678')
	`, "Second Agent")

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		t.Fatalf("expected unique violation (23505), got: %v", err)
	}
}

func TestPropertyRepositoryCreateMediaRejectsInvalidOptionalUUID(t *testing.T) {
	ctx := context.Background()
	pool := openIntegrationDB(t, ctx)
	t.Cleanup(pool.Close)

	repository := NewPropertyRepository(pool)

	_, err := repository.CreateMedia(ctx, domain.Media{
		PropertyID:        domain.ID("not-a-uuid"),
		UploadedByAgentID: domain.ID("550e8400-e29b-41d4-a716-446655440040"),
		URL:               "https://example.test/image.jpg",
		Kind:              domain.MediaKindImage,
	})
	if err == nil {
		t.Fatal("expected error for invalid optional UUID, got nil")
	}
}

func intPtr(value int) *int {
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}
