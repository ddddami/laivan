//go:build integration

package repo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ddddami/laivan/internal/domain"
)

func TestPropertyRepositoryCreateAndListPropertyUnitTypes(t *testing.T) {
	ctx := context.Background()
	pool := openIntegrationDB(t, ctx)
	t.Cleanup(pool.Close)

	truncateProperties(t, ctx, pool)
	t.Cleanup(func() { truncateProperties(t, ctx, pool) })

	campusID := testCampusID(t, ctx, pool)
	actorUserID := insertUser(t, ctx, pool, "Unit Type Operator")
	grantCampusOperator(t, ctx, pool, actorUserID, campusID)
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
	if created.Version != 1 {
		t.Fatalf("created version = %d, want 1", created.Version)
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

	updatedDescription := "Corrected unit description."
	updated, err := repository.UpdatePropertyUnitType(ctx, created.ID, created.Version, domain.PropertyUnitTypePatch{Description: &updatedDescription}, actorUserID)
	if err != nil {
		t.Fatalf("update property unit type: %v", err)
	}
	if updated.Description != updatedDescription || updated.Version != 2 {
		t.Fatalf("updated unit type = %#v, want corrected description at version 2", updated)
	}

	_, err = repository.UpdatePropertyUnitType(ctx, created.ID, created.Version, domain.PropertyUnitTypePatch{Description: &updatedDescription}, actorUserID)
	if !errors.Is(err, ErrStaleUpdate) {
		t.Fatalf("stale unit type update error = %v, want %v", err, ErrStaleUpdate)
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

func intPtr(value int) *int {
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}
