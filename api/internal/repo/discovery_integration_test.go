//go:build integration

package repo

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/ddddami/laivan/internal/data"
	"github.com/ddddami/laivan/internal/domain"
)

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
	emptyUnit := insertPropertyUnitTypeFull(t, ctx, pool, empty, "single_room", "", 1, false, "shared", "shared", "No offers yet")

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
		PropertyID:        alice,
		UploadedByAgentID: agentID,
		URL:               "https://media.example.test/alice-property.jpg",
		Kind:              domain.MediaKindImage,
	}); err != nil {
		t.Fatalf("create Alice property media: %v", err)
	}
	if _, err := repository.CreateMedia(ctx, domain.Media{
		PropertyUnitTypeID: aliceSelfCon,
		UploadedByAgentID:  agentID,
		URL:                "https://media.example.test/alice-self-contained.jpg",
		Kind:               domain.MediaKindImage,
	}); err != nil {
		t.Fatalf("create Alice self-contained media: %v", err)
	}
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
	assertDiscoveryUnitTypes(t, results, []domain.ID{aliceSelfCon, aliceSingle, blueRoomParlour, blueSelfCon, quietSingle})
	for _, result := range results {
		if result.UnitTypeID == aliceSelfCon &&
			result.ThumbnailURL != "https://media.example.test/alice-self-contained.jpg" {
			t.Fatalf(
				"Alice self-contained thumbnail = %q, want unit media",
				result.ThumbnailURL,
			)
		}
		if result.UnitTypeID == aliceSelfCon && result.ThumbnailSource != domain.DiscoveryThumbnailSourceUnitType {
			t.Fatalf("Alice self-contained thumbnail source = %q, want unit_type", result.ThumbnailSource)
		}
		if result.UnitTypeID == aliceSingle &&
			result.ThumbnailURL != "https://media.example.test/alice-property.jpg" {
			t.Fatalf(
				"Alice single-room thumbnail = %q, want property fallback",
				result.ThumbnailURL,
			)
		}
		if result.UnitTypeID == aliceSingle && result.ThumbnailSource != domain.DiscoveryThumbnailSourceProperty {
			t.Fatalf("Alice single-room thumbnail source = %q, want property", result.ThumbnailSource)
		}
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
	if len(results) != total {
		t.Fatalf("all results length = %d, want total %d", len(results), total)
	}
	assertDiscoveryUnitTypes(t, results, []domain.ID{aliceSelfCon, aliceSingle, blueRoomParlour, blueSelfCon, quietSingle, emptyUnit, pausedUnit})
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
	assertDiscoveryUnitTypes(t, results, []domain.ID{aliceSelfCon, blueSelfCon})

	// Filter by multiple categories
	results, total, err = repository.Discover(ctx, DiscoveryFilter{CampusID: campusID, Categories: []string{"self_contained", "single_room"}, Filters: baseFilter})
	if err != nil {
		t.Fatalf("discover by categories: %v", err)
	}
	if total != 4 {
		t.Fatalf("multi-category total = %d, want 4", total)
	}
	assertDiscoveryUnitTypes(t, results, []domain.ID{aliceSelfCon, aliceSingle, blueSelfCon, quietSingle})

	// Filter by area partial match
	results, total, err = repository.Discover(ctx, DiscoveryFilter{CampusID: campusID, Area: "Obanla", Filters: baseFilter})
	if err != nil {
		t.Fatalf("discover by area: %v", err)
	}
	if total != 2 {
		t.Fatalf("Obanla total = %d, want 2", total)
	}
	assertDiscoveryUnitTypes(t, results, []domain.ID{aliceSelfCon, aliceSingle})

	// Search by property name returns each matching accommodation type.
	results, total, err = repository.Discover(ctx, DiscoveryFilter{CampusID: campusID, Search: "alice", Filters: baseFilter})
	if err != nil {
		t.Fatalf("discover by property name: %v", err)
	}
	if total != 2 {
		t.Fatalf("Alice Lodge total = %d, want 2", total)
	}
	for _, result := range results {
		if result.PropertyName != "Alice Lodge" {
			t.Fatalf("search result property = %q, want Alice Lodge", result.PropertyName)
		}
	}

	// Search composes with structural filters instead of clearing them.
	results, total, err = repository.Discover(ctx, DiscoveryFilter{
		CampusID:   campusID,
		Search:     "Alice Lodge",
		Categories: []string{"self_contained"},
		Filters:    baseFilter,
	})
	if err != nil {
		t.Fatalf("discover by property name and category: %v", err)
	}
	if total != 1 || results[0].UnitTypeID != aliceSelfCon {
		t.Fatalf("Alice self-contained search returned %d results, want 1", total)
	}

	// Canonical category language is searchable when a unit has no custom name.
	results, total, err = repository.Discover(ctx, DiscoveryFilter{CampusID: campusID, Search: "room and parlour", Filters: baseFilter})
	if err != nil {
		t.Fatalf("discover by category text: %v", err)
	}
	if total != 1 || results[0].UnitTypeID != blueRoomParlour {
		t.Fatalf("room and parlour search returned %d results, want 1", total)
	}

	// Category search accepts the same hyphenated language shown to students.
	results, total, err = repository.Discover(ctx, DiscoveryFilter{CampusID: campusID, Search: "Self-contained", Filters: baseFilter})
	if err != nil {
		t.Fatalf("discover by hyphenated category text: %v", err)
	}
	if total != 2 {
		t.Fatalf("Self-contained search returned %d results, want 2", total)
	}
	for _, result := range results {
		if result.UnitTypeCategory != domain.UnitCategorySelfContained {
			t.Fatalf("Self-contained search returned category %q", result.UnitTypeCategory)
		}
	}

	// SQL pattern characters are treated as literal search text.
	results, total, err = repository.Discover(ctx, DiscoveryFilter{CampusID: campusID, Search: "%", Filters: baseFilter})
	if err != nil {
		t.Fatalf("discover by literal pattern character: %v", err)
	}
	if total != 0 || len(results) != 0 {
		t.Fatalf("literal pattern search returned %d results, want none", total)
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
	for _, result := range results {
		if result.LowestPrice.Naira() < minPrice {
			t.Fatalf("minimum-price result %q has price %d, below %d", result.UnitTypeID, result.LowestPrice.Naira(), minPrice)
		}
	}
	assertDiscoveryUnitTypes(t, results, []domain.ID{aliceSelfCon, blueRoomParlour, blueSelfCon})

	// Filter by max price
	maxPrice := 200000
	results, total, err = repository.Discover(ctx, DiscoveryFilter{CampusID: campusID, MaxPrice: &maxPrice, Filters: baseFilter})
	if err != nil {
		t.Fatalf("discover by max price: %v", err)
	}
	if total != 2 {
		t.Fatalf("max_price=200000 total = %d, want 2", total)
	}
	for _, result := range results {
		if result.LowestPrice.Naira() > maxPrice {
			t.Fatalf("maximum-price result %q has price %d, above %d", result.UnitTypeID, result.LowestPrice.Naira(), maxPrice)
		}
	}
	assertDiscoveryUnitTypes(t, results, []domain.ID{aliceSingle, quietSingle})

	zeroPrice := 0
	results, total, err = repository.Discover(ctx, DiscoveryFilter{CampusID: campusID, MaxPrice: &zeroPrice, Filters: baseFilter})
	if err != nil {
		t.Fatalf("discover with zero maximum price: %v", err)
	}
	if total != 0 || len(results) != 0 {
		t.Fatalf("max_price=0 returned %d of %d results, want none", len(results), total)
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
	assertDiscoveryUnitTypes(t, results, []domain.ID{aliceSelfCon, blueRoomParlour, blueSelfCon})

	// Filter by has_parlour
	hasParlour := true
	results, total, err = repository.Discover(ctx, DiscoveryFilter{CampusID: campusID, HasParlour: &hasParlour, Filters: baseFilter})
	if err != nil {
		t.Fatalf("discover by parlour: %v", err)
	}
	if total != 1 {
		t.Fatalf("has_parlour=true total = %d, want 1", total)
	}
	assertDiscoveryUnitTypes(t, results, []domain.ID{blueRoomParlour})

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
	assertDiscoveryUnitTypes(t, results, []domain.ID{blueSelfCon})

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

func TestRepositoryDiscoverPaginationHasStableCompleteMembership(t *testing.T) {
	ctx := t.Context()
	pool := openIntegrationDB(t, ctx)
	t.Cleanup(pool.Close)

	truncateProperties(t, ctx, pool)
	truncateAgents(t, ctx, pool)
	t.Cleanup(func() {
		truncateProperties(t, context.Background(), pool)
		truncateAgents(t, context.Background(), pool)
	})

	campusID := testCampusID(t, ctx, pool)
	agentID := insertAgent(t, ctx, pool, "Pagination Agent")
	propertyID := insertProperty(t, ctx, pool, campusID, "Pagination Lodge", time.Date(2026, time.May, 1, 12, 0, 0, 0, time.UTC))
	expected := make([]domain.ID, 0, 5)
	for range 5 {
		unitTypeID := insertPropertyUnitTypeFull(t, ctx, pool, propertyID, "self_contained", "", 1, false, "private", "private", "Pagination unit")
		insertAgentOffer(t, ctx, pool, unitTypeID, agentID, "Pagination offer", 35000000)
		expected = append(expected, unitTypeID)
	}

	stableTime := time.Date(2026, time.May, 5, 12, 0, 0, 0, time.UTC)
	if _, err := pool.Exec(ctx, "UPDATE properties SET created_at = $1, updated_at = $1 WHERE id = $2", stableTime, string(propertyID)); err != nil {
		t.Fatalf("set property timestamps: %v", err)
	}
	if _, err := pool.Exec(ctx, "UPDATE property_unit_types SET updated_at = $1 WHERE property_id = $2", stableTime, string(propertyID)); err != nil {
		t.Fatalf("set unit type timestamps: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		UPDATE agent_offers
		SET updated_at = $1
		WHERE property_unit_type_id IN (SELECT id FROM property_unit_types WHERE property_id = $2)
	`, stableTime, string(propertyID)); err != nil {
		t.Fatalf("set offer timestamps: %v", err)
	}

	slices.Sort(expected)
	repository := NewPropertyRepository(pool)
	allResults := make([]domain.ID, 0, len(expected))
	for page := 1; page <= 3; page++ {
		results, total, err := repository.Discover(ctx, DiscoveryFilter{
			CampusID: campusID,
			Filters: data.Filters{
				Page:         page,
				PageSize:     2,
				Sort:         "-created_at",
				SortSafelist: []string{"created_at", "-created_at"},
			},
		})
		if err != nil {
			t.Fatalf("discover page %d: %v", page, err)
		}
		if total != len(expected) {
			t.Fatalf("page %d total = %d, want %d", page, total, len(expected))
		}
		for _, result := range results {
			if slices.Contains(allResults, result.UnitTypeID) {
				t.Fatalf("page %d repeated unit type %q", page, result.UnitTypeID)
			}
			allResults = append(allResults, result.UnitTypeID)
		}
	}

	if !slices.Equal(allResults, expected) {
		t.Fatalf("paginated unit types = %v, want %v", allResults, expected)
	}
}

func assertDiscoveryUnitTypes(t *testing.T, results []domain.DiscoveryResult, expected []domain.ID) {
	t.Helper()

	actual := make([]domain.ID, 0, len(results))
	for _, result := range results {
		actual = append(actual, result.UnitTypeID)
	}
	slices.Sort(actual)
	want := slices.Clone(expected)
	slices.Sort(want)
	if !slices.Equal(actual, want) {
		t.Fatalf("discovery unit types = %v, want %v", actual, want)
	}
}

func TestRepositoryDiscoverRanking(t *testing.T) {
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
	agentID := insertAgent(t, ctx, pool, "Ranking Agent")
	repository := NewPropertyRepository(pool)

	// Property A: Available, No Thumbnail, Low completeness
	propA := insertPropertyWithArea(t, ctx, pool, campusID, "Property A", "Area", time.Now().Add(-10*time.Hour))
	unitA := insertPropertyUnitTypeFull(t, ctx, pool, propA, "single_room", "", 1, false, "shared", "shared", "")
	insertAgentOffer(t, ctx, pool, unitA, agentID, "Offer A", 10000000)

	// Property B: Available, Has Thumbnail, High completeness
	propB := insertPropertyWithArea(t, ctx, pool, campusID, "Property B", "Area", time.Now().Add(-10*time.Hour))
	unitB := insertPropertyUnitTypeFull(t, ctx, pool, propB, "self_contained", "Nice Unit", 1, false, "private", "private", "Great notes")
	insertAgentOffer(t, ctx, pool, unitB, agentID, "Offer B", 20000000)
	if _, err := repository.CreateMedia(ctx, domain.Media{
		PropertyUnitTypeID: unitB,
		UploadedByAgentID:  agentID,
		URL:                "https://media.example.test/b.jpg",
		Kind:               domain.MediaKindImage,
	}); err != nil {
		t.Fatalf("create B media: %v", err)
	}

	// Property C: Unavailable (paused offer), Has Thumbnail, High completeness
	propC := insertPropertyWithArea(t, ctx, pool, campusID, "Property C", "Area", time.Now().Add(-10*time.Hour))
	unitC := insertPropertyUnitTypeFull(t, ctx, pool, propC, "self_contained", "Nice Unit", 1, false, "private", "private", "Great notes")
	insertAgentOfferWithStatus(t, ctx, pool, unitC, agentID, "Offer C", 30000000, "paused")
	if _, err := repository.CreateMedia(ctx, domain.Media{
		PropertyUnitTypeID: unitC,
		UploadedByAgentID:  agentID,
		URL:                "https://media.example.test/c.jpg",
		Kind:               domain.MediaKindImage,
	}); err != nil {
		t.Fatalf("create C media: %v", err)
	}

	// Property D: Available, No Thumbnail, High completeness
	propD := insertPropertyWithArea(t, ctx, pool, campusID, "Property D", "Area", time.Now().Add(-10*time.Hour))
	unitD := insertPropertyUnitTypeFull(t, ctx, pool, propD, "self_contained", "Detailed Unit", 1, false, "private", "private", "Very descriptive notes about the unit")
	insertAgentOffer(t, ctx, pool, unitD, agentID, "Offer D", 40000000)

	// Push A, B, C, D's updated_at back in time to avoid flakiness with E
	pastTime := time.Now().Add(-5 * time.Hour)
	_, err := pool.Exec(ctx, "UPDATE properties SET updated_at = $1", pastTime)
	if err != nil {
		t.Fatalf("update properties time: %v", err)
	}
	_, err = pool.Exec(ctx, "UPDATE property_unit_types SET updated_at = $1", pastTime)
	if err != nil {
		t.Fatalf("update property_unit_types time: %v", err)
	}
	_, err = pool.Exec(ctx, "UPDATE agent_offers SET updated_at = $1", pastTime)
	if err != nil {
		t.Fatalf("update agent_offers time: %v", err)
	}

	// Property E: Available, Has Thumbnail, High completeness, more recently updated
	propE := insertPropertyWithArea(t, ctx, pool, campusID, "Property E", "Area", time.Now())
	unitE := insertPropertyUnitTypeFull(t, ctx, pool, propE, "self_contained", "Recent Unit", 1, false, "private", "private", "Great notes")
	insertAgentOffer(t, ctx, pool, unitE, agentID, "Offer E", 50000000)
	if _, err := repository.CreateMedia(ctx, domain.Media{
		PropertyUnitTypeID: unitE,
		UploadedByAgentID:  agentID,
		URL:                "https://media.example.test/e.jpg",
		Kind:               domain.MediaKindImage,
	}); err != nil {
		t.Fatalf("create E media: %v", err)
	}

	recommended := data.Filters{
		Page:         1,
		PageSize:     10,
		Sort:         "recommended",
		SortSafelist: []string{"recommended"},
	}

	results, _, err := repository.Discover(ctx, DiscoveryFilter{
		CampusID:     campusID,
		Availability: DiscoveryAvailabilityAll,
		Filters:      recommended,
	})
	if err != nil {
		t.Fatalf("discover recommended: %v", err)
	}

	if len(results) != 5 {
		t.Fatalf("expected 5 results, got %d", len(results))
	}

	expectedOrder := []string{"Property E", "Property B", "Property D", "Property A", "Property C"}
	for i, expected := range expectedOrder {
		if results[i].PropertyName != expected {
			t.Errorf("rank %d: got %q, want %q", i+1, results[i].PropertyName, expected)
		}
	}
}
