//go:build integration

package repo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ddddami/laivan/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

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
	associateAgentWithCampus(t, ctx, pool, agentID, campusID)
	actorUserID := agentUserID(t, ctx, pool, agentID)
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
	if created.Version != 1 {
		t.Fatalf("created version = %d, want 1", created.Version)
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

	updatedTitle := "Corrected self-contained room"
	updatedPrice := domain.Kobo(360000)
	updated, err := repository.UpdateAgentOffer(ctx, created.ID, created.Version, domain.AgentOfferPatch{
		Title:     &updatedTitle,
		PriceKobo: &updatedPrice,
	}, actorUserID)
	if err != nil {
		t.Fatalf("update agent offer: %v", err)
	}
	if updated.Title != updatedTitle || updated.Price.AmountKobo != updatedPrice || updated.Version != 2 {
		t.Fatalf("updated agent offer = %#v, want changed fields at version 2", updated)
	}

	_, err = repository.UpdateAgentOffer(ctx, created.ID, created.Version, domain.AgentOfferPatch{Title: &updatedTitle}, actorUserID)
	if !errors.Is(err, ErrStaleUpdate) {
		t.Fatalf("stale agent offer update error = %v, want %v", err, ErrStaleUpdate)
	}

	archived, err := repository.ArchiveAgentOffer(ctx, created.ID, updated.Version, actorUserID)
	if err != nil {
		t.Fatalf("archive agent offer: %v", err)
	}
	if archived.Status != domain.AgentOfferStatusUnavailable || archived.Version != 3 || archived.ArchivedAt == nil {
		t.Fatalf("archived agent offer = %#v, want unavailable at version 3", archived)
	}

	listed, err = repository.ListAgentOffers(ctx, unitTypeID)
	if err != nil {
		t.Fatalf("list archived agent offers: %v", err)
	}
	if len(listed) != 0 {
		t.Fatalf("listed archived offers = %d, want 0", len(listed))
	}

	fetched, err := repository.GetAgentOffer(ctx, created.ID)
	if err != nil {
		t.Fatalf("get archived agent offer: %v", err)
	}
	if fetched.ID != created.ID || fetched.ArchivedAt == nil {
		t.Fatalf("fetched archived offer = %#v, want retained archived offer", fetched)
	}

	detail, err := repository.GetWithDetails(ctx, propertyID)
	if err != nil {
		t.Fatalf("get property details after archive: %v", err)
	}
	if len(detail.UnitTypes) != 1 || len(detail.UnitTypes[0].AgentOffers) != 0 {
		t.Fatalf("property detail offers = %#v, want no archived offers", detail.UnitTypes)
	}
}

func TestPropertyRepositoryAgentOfferMutationsCheckCurrentAuthorization(t *testing.T) {
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
	otherCampusID := createTestCampus(t, ctx, pool, "offer-mutation-other-campus")
	propertyID := insertProperty(t, ctx, pool, campusID, "Mutation Lodge", time.Now().UTC())
	unitTypeID := insertPropertyUnitType(t, ctx, pool, propertyID, "Self-contained")
	ownerAgentID := insertAgent(t, ctx, pool, "Offer Owner")
	associateAgentWithCampus(t, ctx, pool, ownerAgentID, campusID)
	ownerUserID := agentUserID(t, ctx, pool, ownerAgentID)
	operatorUserID := insertUser(t, ctx, pool, "Campus Operator")
	grantCampusOperator(t, ctx, pool, operatorUserID, campusID)
	wrongCampusOperatorID := insertUser(t, ctx, pool, "Other Campus Operator")
	grantCampusOperator(t, ctx, pool, wrongCampusOperatorID, otherCampusID)
	globalAdminUserID := insertUser(t, ctx, pool, "Global Admin")
	grantGlobalAdmin(t, ctx, pool, globalAdminUserID)
	unauthorizedUserID := insertUser(t, ctx, pool, "Unauthorized User")
	repository := NewPropertyRepository(pool)

	offer, err := repository.CreateAgentOffer(ctx, domain.AgentOffer{
		PropertyUnitTypeID: unitTypeID,
		AgentID:            ownerAgentID,
		Title:              "Original offer",
		Price:              domain.Money{AmountKobo: 35000000},
		Status:             domain.AgentOfferStatusAvailable,
	})
	if err != nil {
		t.Fatalf("create agent offer: %v", err)
	}

	title := "Unauthorized update"
	_, err = repository.UpdateAgentOffer(ctx, offer.ID, offer.Version, domain.AgentOfferPatch{Title: &title}, unauthorizedUserID)
	if !errors.Is(err, ErrAgentForbidden) {
		t.Fatalf("unauthorized offer update error = %v, want %v", err, ErrAgentForbidden)
	}

	title = "Owner update"
	offer, err = repository.UpdateAgentOffer(ctx, offer.ID, offer.Version, domain.AgentOfferPatch{Title: &title}, ownerUserID)
	if err != nil {
		t.Fatalf("owner offer update: %v", err)
	}
	_, err = repository.UpdateAgentOffer(ctx, offer.ID, offer.Version-1, domain.AgentOfferPatch{Title: &title}, unauthorizedUserID)
	if !errors.Is(err, ErrAgentForbidden) {
		t.Fatalf("stale unauthorized update error = %v, want %v", err, ErrAgentForbidden)
	}

	ownerArchiveUnitTypeID := insertPropertyUnitType(t, ctx, pool, propertyID, "Owner archive")
	ownerArchiveOffer, err := repository.CreateAgentOffer(ctx, domain.AgentOffer{
		PropertyUnitTypeID: ownerArchiveUnitTypeID,
		AgentID:            ownerAgentID,
		Title:              "Owner archive offer",
		Price:              domain.Money{AmountKobo: 30000000},
		Status:             domain.AgentOfferStatusAvailable,
	})
	if err != nil {
		t.Fatalf("create owner archive offer: %v", err)
	}
	if _, err := repository.ArchiveAgentOffer(ctx, ownerArchiveOffer.ID, ownerArchiveOffer.Version, ownerUserID); err != nil {
		t.Fatalf("owner archive offer: %v", err)
	}

	if _, err := pool.Exec(ctx, "DELETE FROM agent_campuses WHERE agent_id = $1 AND campus_id = $2", string(ownerAgentID), string(campusID)); err != nil {
		t.Fatalf("remove owner campus: %v", err)
	}
	title = "Removed campus update"
	_, err = repository.UpdateAgentOffer(ctx, offer.ID, offer.Version, domain.AgentOfferPatch{Title: &title}, ownerUserID)
	if !errors.Is(err, ErrAgentForbidden) {
		t.Fatalf("removed-campus owner update error = %v, want %v", err, ErrAgentForbidden)
	}
	associateAgentWithCampus(t, ctx, pool, ownerAgentID, campusID)

	if _, err := pool.Exec(ctx, "UPDATE agents SET status = 'suspended' WHERE id = $1", string(ownerAgentID)); err != nil {
		t.Fatalf("suspend owner agent: %v", err)
	}
	title = "Suspended update"
	_, err = repository.UpdateAgentOffer(ctx, offer.ID, offer.Version, domain.AgentOfferPatch{Title: &title}, ownerUserID)
	if !errors.Is(err, ErrAgentForbidden) {
		t.Fatalf("suspended owner update error = %v, want %v", err, ErrAgentForbidden)
	}

	title = "Operator update"
	offer, err = repository.UpdateAgentOffer(ctx, offer.ID, offer.Version, domain.AgentOfferPatch{Title: &title}, operatorUserID)
	if err != nil {
		t.Fatalf("campus operator offer update: %v", err)
	}

	_, err = repository.UpdateAgentOffer(ctx, offer.ID, offer.Version, domain.AgentOfferPatch{Title: &title}, wrongCampusOperatorID)
	if !errors.Is(err, ErrAgentForbidden) {
		t.Fatalf("wrong-campus update error = %v, want %v", err, ErrAgentForbidden)
	}

	title = "Global admin update"
	offer, err = repository.UpdateAgentOffer(ctx, offer.ID, offer.Version, domain.AgentOfferPatch{Title: &title}, globalAdminUserID)
	if err != nil {
		t.Fatalf("global admin update: %v", err)
	}
	archived, err := repository.ArchiveAgentOffer(ctx, offer.ID, offer.Version, operatorUserID)
	if err != nil {
		t.Fatalf("campus operator archive: %v", err)
	}
	if archived.ArchivedAt == nil || archived.Version != offer.Version+1 {
		t.Fatalf("archived offer = %#v, want archived version %d", archived, offer.Version+1)
	}
}

func TestPropertyRepositoryConcurrentAgentOfferUpdates(t *testing.T) {
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
	propertyID := insertProperty(t, ctx, pool, campusID, "Concurrent Offer Lodge", time.Now().UTC())
	unitTypeID := insertPropertyUnitType(t, ctx, pool, propertyID, "Self-contained")
	agentID := insertAgent(t, ctx, pool, "Concurrent Offer Agent")
	associateAgentWithCampus(t, ctx, pool, agentID, campusID)
	actorUserID := agentUserID(t, ctx, pool, agentID)
	repository := NewPropertyRepository(pool)
	offer, err := repository.CreateAgentOffer(ctx, domain.AgentOffer{
		PropertyUnitTypeID: unitTypeID,
		AgentID:            agentID,
		Title:              "Original offer",
		Price:              domain.Money{AmountKobo: 35000000},
		Status:             domain.AgentOfferStatusAvailable,
	})
	if err != nil {
		t.Fatalf("create agent offer: %v", err)
	}

	type result struct {
		offer domain.AgentOffer
		err   error
	}
	ready := make(chan struct{}, 2)
	release := make(chan struct{})
	results := make(chan result, 2)

	for _, title := range []string{"First concurrent offer update", "Second concurrent offer update"} {
		go func(title string) {
			ready <- struct{}{}
			<-release
			updated, err := repository.UpdateAgentOffer(ctx, offer.ID, offer.Version, domain.AgentOfferPatch{Title: &title}, actorUserID)
			results <- result{offer: updated, err: err}
		}(title)
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
			t.Fatalf("concurrent agent offer update error = %v, want one success and one stale update", result.err)
		}
	}
	if successes != 1 || stale != 1 {
		t.Fatalf("concurrent agent offer updates = %d successes, %d stale updates; want 1, 1", successes, stale)
	}

	final, err := repository.GetAgentOffer(ctx, offer.ID)
	if err != nil {
		t.Fatalf("get final agent offer: %v", err)
	}
	if final.Version != 2 {
		t.Fatalf("final agent offer version = %d, want 2", final.Version)
	}
}

func TestPropertyRepositoryCanonicalMutationsCheckCampusAuthorization(t *testing.T) {
	ctx := context.Background()
	pool := openIntegrationDB(t, ctx)
	t.Cleanup(pool.Close)

	truncateProperties(t, ctx, pool)
	t.Cleanup(func() { truncateProperties(t, ctx, pool) })

	campusID := testCampusID(t, ctx, pool)
	propertyID := insertProperty(t, ctx, pool, campusID, "Canonical Lodge", time.Now().UTC())
	unitTypeID := insertPropertyUnitType(t, ctx, pool, propertyID, "Self-contained")
	operatorUserID := insertUser(t, ctx, pool, "Canonical Operator")
	grantCampusOperator(t, ctx, pool, operatorUserID, campusID)
	unauthorizedUserID := insertUser(t, ctx, pool, "Canonical Unauthorized")
	globalAdminUserID := insertUser(t, ctx, pool, "Canonical Global Admin")
	grantGlobalAdmin(t, ctx, pool, globalAdminUserID)
	repository := NewPropertyRepository(pool)

	propertyName := "Unauthorized property update"
	_, err := repository.Update(ctx, propertyID, 1, domain.PropertyPatch{Name: &propertyName}, unauthorizedUserID)
	if !errors.Is(err, ErrCampusForbidden) {
		t.Fatalf("unauthorized property update error = %v, want %v", err, ErrCampusForbidden)
	}

	propertyName = "Authorized property update"
	property, err := repository.Update(ctx, propertyID, 1, domain.PropertyPatch{Name: &propertyName}, operatorUserID)
	if err != nil {
		t.Fatalf("authorized property update: %v", err)
	}
	_, err = repository.Update(ctx, propertyID, 1, domain.PropertyPatch{Name: &propertyName}, unauthorizedUserID)
	if !errors.Is(err, ErrCampusForbidden) {
		t.Fatalf("stale unauthorized property update error = %v, want %v", err, ErrCampusForbidden)
	}
	propertyName = "Global admin property update"
	property, err = repository.Update(ctx, propertyID, property.Version, domain.PropertyPatch{Name: &propertyName}, globalAdminUserID)
	if err != nil {
		t.Fatalf("global admin property update: %v", err)
	}

	unitName := "Unauthorized unit update"
	_, err = repository.UpdatePropertyUnitType(ctx, unitTypeID, 1, domain.PropertyUnitTypePatch{Name: &unitName}, unauthorizedUserID)
	if !errors.Is(err, ErrCampusForbidden) {
		t.Fatalf("unauthorized unit type update error = %v, want %v", err, ErrCampusForbidden)
	}

	unitName = "Authorized unit update"
	unitType, err := repository.UpdatePropertyUnitType(ctx, unitTypeID, 1, domain.PropertyUnitTypePatch{Name: &unitName}, operatorUserID)
	if err != nil {
		t.Fatalf("authorized unit type update: %v", err)
	}
	unitName = "Global admin unit update"
	unitType, err = repository.UpdatePropertyUnitType(ctx, unitTypeID, unitType.Version, domain.PropertyUnitTypePatch{Name: &unitName}, globalAdminUserID)
	if err != nil {
		t.Fatalf("global admin unit type update: %v", err)
	}
	if property.Version != 3 || unitType.Version != 3 {
		t.Fatalf("updated versions = property %d, unit type %d; want 3, 3", property.Version, unitType.Version)
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
	associateAgentWithCampus(t, ctx, pool, agentID, campusID)
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

func TestPropertyRepositoryCreateAgentOfferRequiresAgentCampusAndActiveStatus(t *testing.T) {
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
	otherCampusID := createTestCampus(t, ctx, pool, "offer-authorization-other-campus")
	propertyID := insertProperty(t, ctx, pool, campusID, "Authorized Lodge", time.Now().UTC())
	otherPropertyID := insertProperty(t, ctx, pool, otherCampusID, "Other Lodge", time.Now().UTC())
	unitTypeID := insertPropertyUnitType(t, ctx, pool, propertyID, "Self-contained")
	otherUnitTypeID := insertPropertyUnitType(t, ctx, pool, otherPropertyID, "Self-contained")
	agentID := insertAgent(t, ctx, pool, "Scoped Agent")
	associateAgentWithCampus(t, ctx, pool, agentID, campusID)
	repository := NewPropertyRepository(pool)
	offer := domain.AgentOffer{
		Title:   "Scoped offer",
		Price:   domain.Money{AmountKobo: 35000000},
		Status:  domain.AgentOfferStatusAvailable,
		AgentID: agentID,
	}

	offer.PropertyUnitTypeID = otherUnitTypeID
	if _, err := repository.CreateAgentOffer(ctx, offer); !errors.Is(err, ErrAgentForbidden) {
		t.Fatalf("cross-campus create error = %v, want %v", err, ErrAgentForbidden)
	}

	if _, err := pool.Exec(ctx, "UPDATE agents SET status = 'suspended' WHERE id = $1", string(agentID)); err != nil {
		t.Fatalf("suspend scoped agent: %v", err)
	}
	offer.PropertyUnitTypeID = unitTypeID
	if _, err := repository.CreateAgentOffer(ctx, offer); !errors.Is(err, ErrAgentForbidden) {
		t.Fatalf("suspended-agent create error = %v, want %v", err, ErrAgentForbidden)
	}
}

func TestPropertyRepositoryAgentPhoneNumberUnique(t *testing.T) {
	ctx := context.Background()
	pool := openIntegrationDB(t, ctx)
	t.Cleanup(pool.Close)

	truncateAgents(t, ctx, pool)
	t.Cleanup(func() { truncateAgents(t, ctx, pool) })

	const sharedPhone = "+2348099000001"

	var userID1 string
	if err := pool.QueryRow(ctx, `INSERT INTO users (email, display_name) VALUES ('agent1@test.com', 'First Agent') RETURNING id::text`).Scan(&userID1); err != nil {
		t.Fatalf("insert first user: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO agents (user_id, display_name, phone_number) VALUES ($1, 'First Agent', $2)`, userID1, sharedPhone); err != nil {
		t.Fatalf("insert first agent: %v", err)
	}

	// A second agent attempting the same phone number must fail.
	var userID2 string
	if err := pool.QueryRow(ctx, `INSERT INTO users (email, display_name) VALUES ('agent2@test.com', 'Second Agent') RETURNING id::text`).Scan(&userID2); err != nil {
		t.Fatalf("insert second user: %v", err)
	}
	_, err := pool.Exec(ctx, `INSERT INTO agents (user_id, display_name, phone_number) VALUES ($1, 'Second Agent', $2)`, userID2, sharedPhone)

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		t.Fatalf("expected unique violation (23505), got: %v", err)
	}
}
