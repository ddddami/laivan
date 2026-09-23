//go:build integration

package repo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ddddami/laivan/internal/domain"
)

func TestPropertyRepositoryCreateMediaBatchRollsBackOnFailure(t *testing.T) {
	ctx := context.Background()
	pool := openIntegrationDB(t, ctx)
	t.Cleanup(pool.Close)

	truncateProperties(t, ctx, pool)
	truncateAgents(t, ctx, pool)
	t.Cleanup(func() {
		truncateProperties(t, context.Background(), pool)
		truncateAgents(t, context.Background(), pool)
	})

	campusID := testCampusID(t, ctx, pool)
	propertyID := insertProperty(t, ctx, pool, campusID, "Atomic Lodge", time.Now().UTC())
	agentID := insertAgent(t, ctx, pool, "Atomic Agent")
	repository := NewPropertyRepository(pool)

	_, err := repository.CreateMediaBatch(ctx, []domain.Media{
		{
			PropertyID:        propertyID,
			UploadedByAgentID: agentID,
			URL:               "https://media.example.test/first.jpg",
			Kind:              domain.MediaKindImage,
		},
		{
			PropertyID:        propertyID,
			UploadedByAgentID: agentID,
			URL:               "https://media.example.test/second.jpg",
			Kind:              domain.MediaKind("invalid"),
		},
	})
	if err == nil {
		t.Fatal("CreateMediaBatch error = nil, want constraint error")
	}

	media, err := repository.ListMediaByProperty(ctx, propertyID)
	if err != nil {
		t.Fatalf("list media after failed batch: %v", err)
	}
	if len(media) != 0 {
		t.Fatalf("media after failed batch = %#v, want no records", media)
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

func TestPropertyRepositorySoftRemovesMedia(t *testing.T) {
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
	propertyID := insertProperty(t, ctx, pool, campusID, "Removal Lodge", time.Now().UTC())
	agentID := insertAgent(t, ctx, pool, "Uploader Agent")
	actorUserID := insertUser(t, ctx, pool, "Media Operator")
	if _, err := pool.Exec(ctx, `INSERT INTO campus_operators (user_id, campus_id) VALUES ($1, $2)`, string(actorUserID), string(campusID)); err != nil {
		t.Fatalf("grant media operator role: %v", err)
	}
	repository := NewPropertyRepository(pool)

	created, err := repository.CreateMedia(ctx, domain.Media{
		PropertyID:        propertyID,
		UploadedByAgentID: agentID,
		URL:               "https://media.example.test/removal.jpg",
		ObjectKey:         "media/property/removal.jpg",
		Kind:              domain.MediaKindImage,
	})
	if err != nil {
		t.Fatalf("create media: %v", err)
	}

	target, err := repository.GetMediaForRemoval(ctx, created.ID)
	if err != nil {
		t.Fatalf("get media removal target: %v", err)
	}
	if target.TargetType != "property" || target.CampusID != campusID || target.PropertyID != propertyID {
		t.Fatalf("removal target = %#v", target)
	}

	if err := repository.RemoveMedia(ctx, created.ID, actorUserID); err != nil {
		t.Fatalf("remove media: %v", err)
	}
	if _, err := repository.GetMediaForRemoval(ctx, created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("removed target error = %v, want %v", err, ErrNotFound)
	}
	if err := repository.RemoveMedia(ctx, created.ID, actorUserID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("repeated removal error = %v, want %v", err, ErrNotFound)
	}

	items, err := repository.ListMediaByProperty(ctx, propertyID)
	if err != nil {
		t.Fatalf("list active media: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("active media after removal = %#v, want none", items)
	}

	var removedBy, storedObjectKey, storedUploader string
	var removed bool
	if err := pool.QueryRow(ctx, `
		SELECT removed_by_user_id::text, object_key, uploaded_by_agent_id::text, removed_at IS NOT NULL
		FROM media WHERE id = $1`, string(created.ID)).Scan(&removedBy, &storedObjectKey, &storedUploader, &removed); err != nil {
		t.Fatalf("read removed media: %v", err)
	}
	if removedBy != string(actorUserID) || storedObjectKey != "media/property/removal.jpg" || storedUploader != string(agentID) || !removed {
		t.Fatalf("removed media state = removed_by=%q object_key=%q uploader=%q removed=%t", removedBy, storedObjectKey, storedUploader, removed)
	}
	var auditCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM audit_events WHERE action = 'media_removed' AND resource_id = $1`, string(created.ID)).Scan(&auditCount); err != nil {
		t.Fatalf("count media removal audit events: %v", err)
	}
	if auditCount != 1 {
		t.Fatalf("media removal audit count = %d, want 1", auditCount)
	}
}

func TestPropertyRepositoryRemoveMediaRechecksPrivileges(t *testing.T) {
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
	propertyID := insertProperty(t, ctx, pool, campusID, "Scoped Removal Lodge", time.Now().UTC())
	unitTypeID := insertPropertyUnitType(t, ctx, pool, propertyID, "Single room")
	agentID := insertAgent(t, ctx, pool, "Scoped Agent")
	associateAgentWithCampus(t, ctx, pool, agentID, campusID)
	agentUserID := agentUserID(t, ctx, pool, agentID)
	offerID := insertAgentOffer(t, ctx, pool, unitTypeID, agentID, "Available room", 25000000)
	operatorID := insertUser(t, ctx, pool, "Scoped Operator")
	if _, err := pool.Exec(ctx, `INSERT INTO campus_operators (user_id, campus_id) VALUES ($1, $2)`, string(operatorID), string(campusID)); err != nil {
		t.Fatalf("grant campus operator role: %v", err)
	}

	repository := NewPropertyRepository(pool)
	propertyMedia, err := repository.CreateMedia(ctx, domain.Media{PropertyID: propertyID, URL: "https://media.example.test/scoped-property.jpg", Kind: domain.MediaKindImage})
	if err != nil {
		t.Fatalf("create property media: %v", err)
	}
	offerMedia, err := repository.CreateMedia(ctx, domain.Media{AgentOfferID: offerID, URL: "https://media.example.test/scoped-offer.jpg", Kind: domain.MediaKindImage})
	if err != nil {
		t.Fatalf("create offer media: %v", err)
	}

	if _, err := repository.GetMediaForRemoval(ctx, propertyMedia.ID); err != nil {
		t.Fatalf("load property media before role removal: %v", err)
	}
	if _, err := pool.Exec(ctx, `DELETE FROM campus_operators WHERE user_id = $1 AND campus_id = $2`, string(operatorID), string(campusID)); err != nil {
		t.Fatalf("revoke campus operator role: %v", err)
	}
	if err := repository.RemoveMedia(ctx, propertyMedia.ID, operatorID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("remove property media after role revocation error = %v, want %v", err, ErrNotFound)
	}

	if _, err := repository.GetMediaForRemoval(ctx, offerMedia.ID); err != nil {
		t.Fatalf("load offer media before suspension: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE agents SET status = 'suspended' WHERE id = $1`, string(agentID)); err != nil {
		t.Fatalf("suspend agent: %v", err)
	}
	if err := repository.RemoveMedia(ctx, offerMedia.ID, agentUserID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("remove offer media after agent suspension error = %v, want %v", err, ErrNotFound)
	}

	for _, mediaID := range []domain.ID{propertyMedia.ID, offerMedia.ID} {
		if _, err := repository.GetMediaForRemoval(ctx, mediaID); err != nil {
			t.Fatalf("media %s was removed without current permission: %v", mediaID, err)
		}
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

func TestPropertyRepositoryGetMediaTarget(t *testing.T) {
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
	propertyID := insertProperty(t, ctx, pool, campusID, "Target Lodge", time.Now().UTC())
	unitTypeID := insertPropertyUnitType(t, ctx, pool, propertyID, "Self-contained")
	agentID := insertAgent(t, ctx, pool, "Target Agent")
	offerID := insertAgentOffer(t, ctx, pool, unitTypeID, agentID, "Target offer", 35000000)
	repository := NewPropertyRepository(pool)

	tests := []struct {
		name       string
		targetType string
		id         domain.ID
		propertyID domain.ID
		unitTypeID domain.ID
		offerID    domain.ID
		campusID   domain.ID
		agentID    domain.ID
	}{
		{name: "property", targetType: "property", id: propertyID, propertyID: propertyID, campusID: campusID},
		{name: "property unit type", targetType: "property_unit_type", id: unitTypeID, propertyID: propertyID, unitTypeID: unitTypeID, campusID: campusID},
		{name: "agent offer", targetType: "agent_offer", id: offerID, propertyID: propertyID, unitTypeID: unitTypeID, offerID: offerID, campusID: campusID, agentID: agentID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, err := repository.GetMediaTarget(ctx, tt.targetType, tt.id)
			if err != nil {
				t.Fatalf("get media target: %v", err)
			}
			if target.PropertyID != tt.propertyID || target.PropertyUnitTypeID != tt.unitTypeID || target.AgentOfferID != tt.offerID || target.CampusID != tt.campusID || target.AgentID != tt.agentID {
				t.Fatalf("target = %#v, want property=%q unit_type=%q offer=%q campus=%q agent=%q", target, tt.propertyID, tt.unitTypeID, tt.offerID, tt.campusID, tt.agentID)
			}
		})
	}

	if _, err := repository.GetMediaTarget(ctx, "property", domain.ID("11111111-1111-1111-1111-111111111111")); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing target error = %v, want %v", err, ErrNotFound)
	}
}

func TestPropertyRepositoryCreateMediaAllowsMissingAgentProvenance(t *testing.T) {
	ctx := context.Background()
	pool := openIntegrationDB(t, ctx)
	t.Cleanup(pool.Close)

	truncateProperties(t, ctx, pool)
	t.Cleanup(func() { truncateProperties(t, ctx, pool) })

	campusID := testCampusID(t, ctx, pool)
	propertyID := insertProperty(t, ctx, pool, campusID, "Operator Lodge", time.Now().UTC())
	repository := NewPropertyRepository(pool)

	created, err := repository.CreateMedia(ctx, domain.Media{
		PropertyID: propertyID,
		URL:        "https://media.example.test/operator.jpg",
		Kind:       domain.MediaKindImage,
	})
	if err != nil {
		t.Fatalf("create media: %v", err)
	}
	if created.UploadedByAgentID != "" {
		t.Fatalf("uploaded by agent ID = %q, want empty provenance", created.UploadedByAgentID)
	}
}
