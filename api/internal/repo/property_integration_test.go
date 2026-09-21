//go:build integration

package repo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ddddami/laivan/internal/data"
	"github.com/ddddami/laivan/internal/domain"
)

func TestPropertyRepositoryCreateAndGet(t *testing.T) {
	ctx := context.Background()
	pool := openIntegrationDB(t, ctx)
	t.Cleanup(pool.Close)

	truncateProperties(t, ctx, pool)
	t.Cleanup(func() { truncateProperties(t, ctx, pool) })

	campusID := testCampusID(t, ctx, pool)
	actorUserID := insertUser(t, ctx, pool, "Property Contributor")
	repository := NewPropertyRepository(pool)

	created, err := repository.Create(ctx, domain.Property{
		CampusID: campusID,
		Name:     "Alice Lodge",
		Location: domain.ApproxLocation{
			Area:     "Obanla",
			Landmark: "Near South Gate",
		},
		Description: "Gated lodge with multiple room categories near campus.",
	}, actorUserID)
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
	var storedActorID string
	if err := pool.QueryRow(ctx, `SELECT created_by_user_id::text FROM properties WHERE id = $1`, string(created.ID)).Scan(&storedActorID); err != nil {
		t.Fatalf("read property provenance: %v", err)
	}
	if storedActorID != string(actorUserID) {
		t.Fatalf("property provenance = %q, want %q", storedActorID, actorUserID)
	}

	fetched, err := repository.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("get property: %v", err)
	}
	if fetched != created {
		t.Fatalf("fetched property = %#v, want %#v", fetched, created)
	}
}

func TestPropertyRepositoryUpdateIsVersionChecked(t *testing.T) {
	ctx := t.Context()
	pool := openIntegrationDB(t, ctx)
	t.Cleanup(pool.Close)

	truncateProperties(t, ctx, pool)
	t.Cleanup(func() { truncateProperties(t, context.Background(), pool) })

	campusID := testCampusID(t, ctx, pool)
	actorUserID := insertUser(t, ctx, pool, "Property Operator")
	grantCampusOperator(t, ctx, pool, actorUserID, campusID)
	repository := NewPropertyRepository(pool)
	created, err := repository.Create(ctx, domain.Property{
		CampusID: campusID,
		Name:     "Original Lodge",
		Location: domain.ApproxLocation{Area: "Obanla", Landmark: "South Gate"},
	}, actorUserID)
	if err != nil {
		t.Fatalf("create property: %v", err)
	}
	if created.Version != 1 {
		t.Fatalf("created version = %d, want 1", created.Version)
	}

	updatedName := "Updated Lodge"
	updatedDescription := ""
	updated, err := repository.Update(ctx, created.ID, created.Version, domain.PropertyPatch{
		Name:        &updatedName,
		Description: &updatedDescription,
	}, actorUserID)
	if err != nil {
		t.Fatalf("update property: %v", err)
	}
	if updated.Name != "Updated Lodge" || updated.Description != "" {
		t.Fatalf("updated property = %#v, want changed fields", updated)
	}
	if updated.Version != 2 {
		t.Fatalf("updated version = %d, want 2", updated.Version)
	}
	var auditCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM audit_events WHERE action = 'property_corrected' AND resource_id = $1`, string(created.ID)).Scan(&auditCount); err != nil {
		t.Fatalf("count property correction audit events: %v", err)
	}
	if auditCount != 1 {
		t.Fatalf("property correction audit count = %d, want 1", auditCount)
	}

	lostUpdateName := "Lost update"
	_, err = repository.Update(ctx, created.ID, created.Version, domain.PropertyPatch{Name: &lostUpdateName}, actorUserID)
	if !errors.Is(err, ErrStaleUpdate) {
		t.Fatalf("stale update error = %v, want %v", err, ErrStaleUpdate)
	}
}

func TestPropertyRepositoryConcurrentVersionedUpdates(t *testing.T) {
	ctx := t.Context()
	pool := openIntegrationDB(t, ctx)
	t.Cleanup(pool.Close)

	truncateProperties(t, ctx, pool)
	t.Cleanup(func() { truncateProperties(t, context.Background(), pool) })

	campusID := testCampusID(t, ctx, pool)
	operatorID := insertUser(t, ctx, pool, "Concurrent Property Operator")
	grantCampusOperator(t, ctx, pool, operatorID, campusID)
	repository := NewPropertyRepository(pool)
	property, err := repository.Create(ctx, domain.Property{
		CampusID: campusID,
		Name:     "Concurrent Lodge",
		Location: domain.ApproxLocation{Area: "Obanla"},
	}, operatorID)
	if err != nil {
		t.Fatalf("create property: %v", err)
	}

	type result struct {
		property domain.Property
		err      error
	}
	ready := make(chan struct{}, 2)
	release := make(chan struct{})
	results := make(chan result, 2)

	for _, name := range []string{"First concurrent update", "Second concurrent update"} {
		go func(name string) {
			ready <- struct{}{}
			<-release
			updated, err := repository.Update(ctx, property.ID, property.Version, domain.PropertyPatch{Name: &name}, operatorID)
			results <- result{property: updated, err: err}
		}(name)
	}

	<-ready
	<-ready
	close(release)

	var successes, stale int
	for range 2 {
		result := <-results
		switch {
		case result.err == nil:
			successes++
		case errors.Is(result.err, ErrStaleUpdate):
			stale++
		default:
			t.Fatalf("concurrent property update error = %v, want one success and one stale update", result.err)
		}
	}
	if successes != 1 || stale != 1 {
		t.Fatalf("concurrent property updates = %d successes, %d stale updates; want 1, 1", successes, stale)
	}

	final, err := repository.Get(ctx, property.ID)
	if err != nil {
		t.Fatalf("get final property: %v", err)
	}
	if final.Version != 2 {
		t.Fatalf("final property version = %d, want 2", final.Version)
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

func TestPropertyRepositoryListWithSummaryExcludesArchivedOffers(t *testing.T) {
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
	propertyID := insertProperty(t, ctx, pool, campusID, "Archived Offer Lodge", time.Now().UTC())
	unitTypeID := insertPropertyUnitType(t, ctx, pool, propertyID, "Self-contained")
	agentID := insertAgent(t, ctx, pool, "Archived Offer Agent")
	associateAgentWithCampus(t, ctx, pool, agentID, campusID)
	actorUserID := agentUserID(t, ctx, pool, agentID)
	repository := NewPropertyRepository(pool)

	offer, err := repository.CreateAgentOffer(ctx, domain.AgentOffer{
		PropertyUnitTypeID: unitTypeID,
		AgentID:            agentID,
		Title:              "Archived offer",
		Price:              domain.Money{AmountKobo: 25000000},
		Status:             domain.AgentOfferStatusAvailable,
	})
	if err != nil {
		t.Fatalf("create agent offer: %v", err)
	}
	if _, err := repository.CreateMedia(ctx, domain.Media{
		AgentOfferID:      offer.ID,
		UploadedByAgentID: agentID,
		URL:               "https://media.example.test/archived-offer.jpg",
		Kind:              domain.MediaKindImage,
	}); err != nil {
		t.Fatalf("create agent offer media: %v", err)
	}
	if _, err := repository.ArchiveAgentOffer(ctx, offer.ID, offer.Version, actorUserID); err != nil {
		t.Fatalf("archive agent offer: %v", err)
	}

	filters := data.Filters{Page: 1, PageSize: 10, Sort: "-created_at", SortSafelist: []string{"created_at", "-created_at"}}
	summaries, total, err := repository.ListWithSummary(ctx, PropertyListFilter{CampusID: campusID, Filters: filters})
	if err != nil {
		t.Fatalf("list property summaries: %v", err)
	}
	if total != 1 || len(summaries) != 1 {
		t.Fatalf("summaries = %d of %d, want 1 of 1", len(summaries), total)
	}
	if summaries[0].AvailableOfferCount != 0 || summaries[0].LowestPrice.AmountKobo != 0 || summaries[0].ThumbnailURL != "" {
		t.Fatalf("archived offer summary = %#v, want no public offer data", summaries[0])
	}

	hasOffers := true
	summaries, total, err = repository.ListWithSummary(ctx, PropertyListFilter{CampusID: campusID, HasOffers: &hasOffers, Filters: filters})
	if err != nil {
		t.Fatalf("list summaries with offers: %v", err)
	}
	if total != 0 || len(summaries) != 0 {
		t.Fatalf("summaries with offers = %d of %d, want none", len(summaries), total)
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
