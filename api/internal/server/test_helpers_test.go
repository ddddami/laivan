package server

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ddddami/laivan/internal/domain"
	"github.com/ddddami/laivan/internal/repo"
)

type stubPropertyRepo struct{}

func (s *stubPropertyRepo) GetCampusBySlug(ctx context.Context, slug string) (domain.Campus, error) {
	if slug != "futa" {
		return domain.Campus{}, repo.ErrNotFound
	}

	return domain.Campus{
		ID:        domain.ID("550e8400-e29b-41d4-a716-446655440002"),
		Slug:      "futa",
		Name:      "Federal University of Technology, Akure",
		ShortName: "FUTA",
		IsActive:  true,
	}, nil
}

func (s *stubPropertyRepo) Create(ctx context.Context, property domain.Property, _ domain.ID) (domain.Property, error) {
	property.ID = domain.ID("550e8400-e29b-41d4-a716-446655440001")
	property.Version = 1
	property.CreatedAt = time.Now()
	property.UpdatedAt = time.Now()
	return property, nil
}

func (s *stubPropertyRepo) Get(ctx context.Context, id domain.ID) (domain.Property, error) {
	if string(id) == "550e8400-e29b-41d4-a716-446655440000" {
		return domain.Property{
			ID:       id,
			CampusID: domain.ID("550e8400-e29b-41d4-a716-446655440002"),
			Name:     "Alice Lodge",
			Location: domain.ApproxLocation{Area: "Obanla"},
			Version:  1,
		}, nil
	}
	return domain.Property{}, repo.ErrNotFound
}

func (s *stubPropertyRepo) Update(ctx context.Context, id domain.ID, expectedVersion int, patch domain.PropertyPatch, _ domain.ID) (domain.Property, error) {
	property, err := s.Get(ctx, id)
	if err != nil {
		return domain.Property{}, err
	}
	if property.Version != expectedVersion {
		return domain.Property{}, repo.ErrStaleUpdate
	}
	if patch.Name != nil {
		property.Name = *patch.Name
	}
	if patch.Area != nil {
		property.Location.Area = *patch.Area
	}
	if patch.Landmark != nil {
		property.Location.Landmark = *patch.Landmark
	}
	if patch.Description != nil {
		property.Description = *patch.Description
	}
	property.Version++
	property.UpdatedAt = time.Now()
	return property, nil
}

func (s *stubPropertyRepo) GetWithDetails(ctx context.Context, id domain.ID) (domain.PropertyDetail, error) {
	if string(id) != "550e8400-e29b-41d4-a716-446655440000" {
		return domain.PropertyDetail{}, repo.ErrNotFound
	}

	return domain.PropertyDetail{
		Property: domain.Property{
			ID:          id,
			CampusID:    domain.ID("550e8400-e29b-41d4-a716-446655440002"),
			Name:        "Alice Lodge",
			Location:    domain.ApproxLocation{Area: "Obanla", Landmark: "Near South Gate"},
			Description: "Gated lodge with multiple room categories near campus.",
			Version:     1,
			Timestamps: domain.Timestamps{
				CreatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
			},
		},
		UnitTypes: []domain.PropertyUnitTypeDetail{
			{
				PropertyUnitType: domain.PropertyUnitType{
					ID:          domain.ID("550e8400-e29b-41d4-a716-446655440020"),
					PropertyID:  id,
					Category:    domain.UnitCategorySelfContained,
					Name:        "Self-contained",
					Description: "Private room with bathroom and kitchenette.",
					Structure: domain.UnitStructure{
						BedroomCount: intPointer(1),
						HasParlour:   boolPointer(false),
						BathroomType: "private",
						KitchenType:  "private",
					},
					Timestamps: domain.Timestamps{
						CreatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
					},
				},
				Media: []domain.Media{{
					ID:                 domain.ID("550e8400-e29b-41d4-a716-446655440060"),
					PropertyUnitTypeID: domain.ID("550e8400-e29b-41d4-a716-446655440020"),
					UploadedByAgentID:  domain.ID("550e8400-e29b-41d4-a716-446655440040"),
					URL:                "https://media.example.test/unit.jpg",
					Kind:               domain.MediaKindImage,
					Caption:            "Unit media",
					ContentType:        "image/jpeg",
					SizeBytes:          1024,
					CreatedAt:          time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
				}},
				AgentOffers: []domain.AgentOfferDetail{
					{
						AgentOffer: domain.AgentOffer{
							ID:                 domain.ID("550e8400-e29b-41d4-a716-446655440030"),
							PropertyUnitTypeID: domain.ID("550e8400-e29b-41d4-a716-446655440020"),
							AgentID:            domain.ID("550e8400-e29b-41d4-a716-446655440040"),
							Title:              "Fresh self-contained room",
							Description:        "Recently painted room with private bathroom.",
							Price:              domain.Money{AmountKobo: 35000000},
							Status:             domain.AgentOfferStatusAvailable,
							Timestamps: domain.Timestamps{
								CreatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
								UpdatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
							},
						},
						Agent: domain.AgentSummary{
							ID:          domain.ID("550e8400-e29b-41d4-a716-446655440040"),
							DisplayName: "Bisi Housing Connect",
						},
						Media: []domain.Media{{
							ID:                domain.ID("550e8400-e29b-41d4-a716-446655440060"),
							AgentOfferID:      domain.ID("550e8400-e29b-41d4-a716-446655440030"),
							UploadedByAgentID: domain.ID("550e8400-e29b-41d4-a716-446655440040"),
							URL:               "https://media.example.test/offer.jpg",
							Kind:              domain.MediaKindImage,
							ContentType:       "image/jpeg",
							SizeBytes:         1024,
							CreatedAt:         time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
						}},
					},
				},
			},
		},
	}, nil
}

func (s *stubPropertyRepo) GetMediaTarget(ctx context.Context, targetType string, id domain.ID) (repo.MediaTarget, error) {
	switch targetType {
	case "property":
		if id == "550e8400-e29b-41d4-a716-446655440000" {
			return repo.MediaTarget{PropertyID: id, CampusID: domain.ID("550e8400-e29b-41d4-a716-446655440002")}, nil
		}
	case "property_unit_type":
		if id == "550e8400-e29b-41d4-a716-446655440020" {
			return repo.MediaTarget{
				PropertyID:         domain.ID("550e8400-e29b-41d4-a716-446655440000"),
				PropertyUnitTypeID: id,
				CampusID:           domain.ID("550e8400-e29b-41d4-a716-446655440002"),
			}, nil
		}
	case "agent_offer":
		if id == "550e8400-e29b-41d4-a716-446655440030" {
			return repo.MediaTarget{
				PropertyID:         domain.ID("550e8400-e29b-41d4-a716-446655440000"),
				PropertyUnitTypeID: domain.ID("550e8400-e29b-41d4-a716-446655440020"),
				AgentOfferID:       id,
				CampusID:           domain.ID("550e8400-e29b-41d4-a716-446655440002"),
				AgentID:            domain.ID("550e8400-e29b-41d4-a716-446655440040"),
			}, nil
		}
	}
	return repo.MediaTarget{}, repo.ErrNotFound
}

func (s *stubPropertyRepo) ListWithSummary(ctx context.Context, filter repo.PropertyListFilter) ([]domain.PropertySummary, int, error) {
	return []domain.PropertySummary{
		{
			Property: domain.Property{
				ID:          domain.ID("550e8400-e29b-41d4-a716-446655440010"),
				CampusID:    filter.CampusID,
				Name:        "Alice Lodge",
				Location:    domain.ApproxLocation{Area: "Obanla", Landmark: "Near South Gate"},
				Description: "Gated lodge with multiple room categories near campus.",
				Timestamps: domain.Timestamps{
					CreatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
				},
			},
			UnitTypeCount:       3,
			AvailableOfferCount: 5,
			LowestPrice:         domain.Money{AmountKobo: 25000000},
		},
	}, 1, nil
}

func (s *stubPropertyRepo) Discover(ctx context.Context, filter repo.DiscoveryFilter) ([]domain.DiscoveryResult, int, error) {
	alice := domain.DiscoveryResult{
		PropertyID:       domain.ID("550e8400-e29b-41d4-a716-446655440010"),
		PropertyName:     "Alice Lodge",
		PropertyArea:     "Obanla",
		PropertyLandmark: "Near South Gate",
		UnitTypeID:       domain.ID("550e8400-e29b-41d4-a716-446655440020"),
		UnitTypeCategory: domain.UnitCategorySelfContained,
		UnitTypeName:     "Self-contained",
		Structure: domain.UnitStructure{
			BedroomCount: intPointer(1),
			HasParlour:   boolPointer(false),
			BathroomType: "private",
			KitchenType:  "private",
		},
		LowestPrice:         domain.Money{AmountKobo: 35000000},
		AvailableOfferCount: 2,
		ThumbnailSource:     domain.DiscoveryThumbnailSourceUnitType,
		CreatedAt:           time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
		UpdatedAt:           time.Date(2026, time.May, 2, 10, 0, 0, 0, time.UTC),
	}
	results := []domain.DiscoveryResult{alice}
	if len(filter.Categories) == 0 {
		blueRoof := alice
		blueRoof.PropertyID = domain.ID("550e8400-e29b-41d4-a716-446655440012")
		blueRoof.PropertyName = "Blue Roof"
		blueRoof.UnitTypeID = domain.ID("550e8400-e29b-41d4-a716-446655440022")
		blueRoof.UnitTypeCategory = domain.UnitCategorySingleRoom
		blueRoof.UnitTypeName = "Single room"
		blueRoof.LowestPrice = domain.Money{AmountKobo: 30000000}
		blueRoof.ThumbnailSource = domain.DiscoveryThumbnailSourceUnitType
		results = append(results, blueRoof)
	}
	if filter.Availability != repo.DiscoveryAvailabilityAvailable {
		unavailable := domain.DiscoveryResult{
			PropertyID:       domain.ID("550e8400-e29b-41d4-a716-446655440011"),
			PropertyName:     "Empty Lodge",
			PropertyArea:     "Obanla",
			PropertyLandmark: "Near North Gate",
			UnitTypeID:       domain.ID("550e8400-e29b-41d4-a716-446655440021"),
			UnitTypeCategory: domain.UnitCategorySingleRoom,
			UnitTypeName:     "Single room",
			Structure: domain.UnitStructure{
				BedroomCount: intPointer(1),
				HasParlour:   boolPointer(false),
				BathroomType: "shared",
				KitchenType:  "shared",
			},
			CreatedAt:       time.Date(2026, time.April, 30, 10, 0, 0, 0, time.UTC),
			UpdatedAt:       time.Date(2026, time.April, 30, 10, 0, 0, 0, time.UTC),
			ThumbnailSource: domain.DiscoveryThumbnailSourceProperty,
		}
		results = append(results, unavailable)
	}
	if filter.Filters.Sort != "recommended" && len(results) > 1 {
		results[0], results[1] = results[1], results[0]
	}

	return results, len(results), nil
}

func (s *stubPropertyRepo) CreateMedia(ctx context.Context, media domain.Media) (domain.Media, error) {
	media.ID = domain.ID("550e8400-e29b-41d4-a716-446655440050")
	media.CreatedAt = time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC)
	return media, nil
}

func (s *stubPropertyRepo) CreateMediaBatch(ctx context.Context, media []domain.Media, actorUserID domain.ID) ([]domain.Media, error) {
	created := make([]domain.Media, 0, len(media))
	for _, item := range media {
		item, err := s.CreateMedia(ctx, item)
		if err != nil {
			return nil, err
		}
		created = append(created, item)
	}
	return created, nil
}

func (s *stubPropertyRepo) ListMediaByProperty(ctx context.Context, propertyID domain.ID) ([]domain.Media, error) {
	return []domain.Media{{
		ID:                domain.ID("550e8400-e29b-41d4-a716-446655440050"),
		PropertyID:        propertyID,
		UploadedByAgentID: domain.ID("550e8400-e29b-41d4-a716-446655440040"),
		URL:               "https://media.example.test/alice.jpg",
		Kind:              domain.MediaKindImage,
		ContentType:       "image/jpeg",
		SizeBytes:         512,
		CreatedAt:         time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
	}}, nil
}

func (s *stubPropertyRepo) ListMediaByPropertyUnitType(ctx context.Context, propertyUnitTypeID domain.ID) ([]domain.Media, error) {
	return nil, nil
}

func (s *stubPropertyRepo) ListMediaByAgentOffer(ctx context.Context, agentOfferID domain.ID) ([]domain.Media, error) {
	return nil, nil
}

func (s *stubPropertyRepo) GetMediaForRemoval(ctx context.Context, mediaID domain.ID) (repo.MediaRemovalTarget, error) {
	if mediaID != "550e8400-e29b-41d4-a716-446655440050" {
		return repo.MediaRemovalTarget{}, repo.ErrNotFound
	}
	return repo.MediaRemovalTarget{
		MediaID:    mediaID,
		ObjectKey:  "media/property/alice.jpg",
		TargetType: "property",
		CampusID:   domain.ID("550e8400-e29b-41d4-a716-446655440002"),
	}, nil
}

func (s *stubPropertyRepo) RemoveMedia(ctx context.Context, mediaID, actorUserID domain.ID) error {
	if mediaID != "550e8400-e29b-41d4-a716-446655440050" {
		return repo.ErrNotFound
	}
	return nil
}

func (s *stubPropertyRepo) CreatePropertyUnitType(ctx context.Context, unitType domain.PropertyUnitType, _ domain.ID) (domain.PropertyUnitType, error) {
	if string(unitType.PropertyID) != "550e8400-e29b-41d4-a716-446655440000" {
		return domain.PropertyUnitType{}, repo.ErrNotFound
	}

	unitType.ID = domain.ID("550e8400-e29b-41d4-a716-446655440020")
	unitType.Version = 1
	unitType.CreatedAt = time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC)
	unitType.UpdatedAt = time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC)
	return unitType, nil
}

func (s *stubPropertyRepo) GetPropertyUnitType(ctx context.Context, id domain.ID) (domain.PropertyUnitType, error) {
	unitTypes, err := s.ListPropertyUnitTypes(ctx, domain.ID("550e8400-e29b-41d4-a716-446655440000"))
	if err != nil {
		return domain.PropertyUnitType{}, err
	}
	for _, unitType := range unitTypes {
		if unitType.ID == id {
			return unitType, nil
		}
	}
	return domain.PropertyUnitType{}, repo.ErrNotFound
}

func (s *stubPropertyRepo) UpdatePropertyUnitType(ctx context.Context, id domain.ID, expectedVersion int, patch domain.PropertyUnitTypePatch, _ domain.ID) (domain.PropertyUnitType, error) {
	unitType, err := s.GetPropertyUnitType(ctx, id)
	if err != nil {
		return domain.PropertyUnitType{}, err
	}
	if unitType.Version != expectedVersion {
		return domain.PropertyUnitType{}, repo.ErrStaleUpdate
	}
	if patch.Category != nil {
		unitType.Category = *patch.Category
	}
	if patch.Name != nil {
		unitType.Name = *patch.Name
	}
	if patch.Description != nil {
		unitType.Description = *patch.Description
	}
	if patch.Notes != nil {
		unitType.Notes = *patch.Notes
	}
	if patch.BedroomCount != nil {
		unitType.Structure.BedroomCount = patch.BedroomCount
	}
	if patch.HasParlour != nil {
		unitType.Structure.HasParlour = patch.HasParlour
	}
	if patch.BathroomType != nil {
		unitType.Structure.BathroomType = *patch.BathroomType
	}
	if patch.KitchenType != nil {
		unitType.Structure.KitchenType = *patch.KitchenType
	}
	unitType.Version++
	unitType.UpdatedAt = time.Now()
	return unitType, nil
}

func (s *stubPropertyRepo) ListPropertyUnitTypes(ctx context.Context, propertyID domain.ID) ([]domain.PropertyUnitType, error) {
	if string(propertyID) != "550e8400-e29b-41d4-a716-446655440000" {
		return nil, repo.ErrNotFound
	}

	return []domain.PropertyUnitType{
		{
			ID:          domain.ID("550e8400-e29b-41d4-a716-446655440020"),
			PropertyID:  propertyID,
			Category:    domain.UnitCategorySelfContained,
			Name:        "Self-contained",
			Description: "Private room with bathroom and kitchenette.",
			Version:     1,
			Structure: domain.UnitStructure{
				BedroomCount: intPointer(1),
				HasParlour:   boolPointer(false),
				BathroomType: "private",
				KitchenType:  "private",
			},
			Timestamps: domain.Timestamps{
				CreatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
			},
		},
	}, nil
}

func (s *stubPropertyRepo) CreateAgentOffer(ctx context.Context, offer domain.AgentOffer) (domain.AgentOffer, error) {
	if string(offer.PropertyUnitTypeID) != "550e8400-e29b-41d4-a716-446655440020" {
		return domain.AgentOffer{}, repo.ErrUnitTypeNotFound
	}
	if string(offer.AgentID) != "550e8400-e29b-41d4-a716-446655440040" {
		return domain.AgentOffer{}, repo.ErrAgentNotFound
	}

	offer.ID = domain.ID("550e8400-e29b-41d4-a716-446655440030")
	offer.Version = 1
	offer.CreatedAt = time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC)
	offer.UpdatedAt = time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC)
	return offer, nil
}

func (s *stubPropertyRepo) ListAgentOffers(ctx context.Context, unitTypeID domain.ID) ([]domain.AgentOffer, error) {
	if string(unitTypeID) != "550e8400-e29b-41d4-a716-446655440020" {
		return nil, repo.ErrNotFound
	}

	return []domain.AgentOffer{
		{
			ID:                 domain.ID("550e8400-e29b-41d4-a716-446655440030"),
			PropertyUnitTypeID: domain.ID(unitTypeID),
			AgentID:            domain.ID("550e8400-e29b-41d4-a716-446655440040"),
			Title:              "Fresh self-contained room",
			Description:        "Recently painted room with private bathroom.",
			Price:              domain.Money{AmountKobo: 35000000},
			Status:             domain.AgentOfferStatusAvailable,
			Version:            1,
			Timestamps: domain.Timestamps{
				CreatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
			},
		},
	}, nil
}

func (s *stubPropertyRepo) GetAgentOffer(ctx context.Context, id domain.ID) (domain.AgentOffer, error) {
	offers, err := s.ListAgentOffers(ctx, domain.ID("550e8400-e29b-41d4-a716-446655440020"))
	if err != nil {
		return domain.AgentOffer{}, err
	}
	for _, offer := range offers {
		if offer.ID == id {
			return offer, nil
		}
	}
	return domain.AgentOffer{}, repo.ErrNotFound
}

func (s *stubPropertyRepo) UpdateAgentOffer(ctx context.Context, id domain.ID, expectedVersion int, patch domain.AgentOfferPatch, _ domain.ID) (domain.AgentOffer, error) {
	offer, err := s.GetAgentOffer(ctx, id)
	if err != nil {
		return domain.AgentOffer{}, err
	}
	if offer.Version != expectedVersion {
		return domain.AgentOffer{}, repo.ErrStaleUpdate
	}
	if patch.Title != nil {
		offer.Title = *patch.Title
	}
	if patch.Description != nil {
		offer.Description = *patch.Description
	}
	if patch.Notes != nil {
		offer.Notes = *patch.Notes
	}
	if patch.PriceKobo != nil {
		offer.Price.AmountKobo = *patch.PriceKobo
	}
	if patch.Status != nil {
		offer.Status = *patch.Status
	}
	offer.Version++
	offer.UpdatedAt = time.Now()
	return offer, nil
}

func (s *stubPropertyRepo) ArchiveAgentOffer(ctx context.Context, id domain.ID, expectedVersion int, _ domain.ID) (domain.AgentOffer, error) {
	offer, err := s.GetAgentOffer(ctx, id)
	if err != nil {
		return domain.AgentOffer{}, err
	}
	if offer.Version != expectedVersion {
		return domain.AgentOffer{}, repo.ErrStaleUpdate
	}
	now := time.Now()
	offer.Status = domain.AgentOfferStatusUnavailable
	offer.ArchivedAt = &now
	offer.Version++
	offer.UpdatedAt = now
	return offer, nil
}

func testAppWithRepo() *app {
	a := testApp()
	a.propertyRepo = &stubPropertyRepo{}
	return a
}

func testAppWithActiveAgentRepo() *app {
	userID := domain.ID("550e8400-e29b-41d4-a716-446655440001")
	app := authenticatedTestApp(userID, &fakeAgentApplicationStore{access: domain.EffectiveAccess{
		Agent:          &domain.LinkedAgent{ID: domain.ID("550e8400-e29b-41d4-a716-446655440040"), Status: domain.AgentStatusActive},
		AgentCampusIDs: []domain.ID{"550e8400-e29b-41d4-a716-446655440002"},
	}})
	app.propertyRepo = &stubPropertyRepo{}
	return app
}

type spyPropertyRepo struct {
	*stubPropertyRepo
	createdUnitType        domain.PropertyUnitType
	createdOffer           domain.AgentOffer
	createdMedia           []domain.Media
	mediaActorUserID       domain.ID
	discoveryFilter        repo.DiscoveryFilter
	updatePropertyCalls    int
	updatePropertyErr      error
	updateUnitTypeCalls    int
	updateUnitTypeErr      error
	updateAgentOfferCalls  int
	updateAgentOfferErr    error
	archiveAgentOfferCalls int
	archiveAgentOfferErr   error
}

func (s *spyPropertyRepo) Discover(ctx context.Context, filter repo.DiscoveryFilter) ([]domain.DiscoveryResult, int, error) {
	s.discoveryFilter = filter
	return s.stubPropertyRepo.Discover(ctx, filter)
}

func (s *spyPropertyRepo) CreateMedia(ctx context.Context, media domain.Media) (domain.Media, error) {
	s.createdMedia = append(s.createdMedia, media)
	return s.stubPropertyRepo.CreateMedia(ctx, media)
}

func (s *spyPropertyRepo) CreateMediaBatch(ctx context.Context, media []domain.Media, actorUserID domain.ID) ([]domain.Media, error) {
	s.createdMedia = append(s.createdMedia, media...)
	s.mediaActorUserID = actorUserID
	return s.stubPropertyRepo.CreateMediaBatch(ctx, media, actorUserID)
}

func (s *spyPropertyRepo) Update(ctx context.Context, id domain.ID, expectedVersion int, patch domain.PropertyPatch, actorUserID domain.ID) (domain.Property, error) {
	s.updatePropertyCalls++
	if s.updatePropertyErr != nil {
		return domain.Property{}, s.updatePropertyErr
	}
	return s.stubPropertyRepo.Update(ctx, id, expectedVersion, patch, actorUserID)
}

func (s *spyPropertyRepo) ListMediaByProperty(ctx context.Context, propertyID domain.ID) ([]domain.Media, error) {
	return s.stubPropertyRepo.ListMediaByProperty(ctx, propertyID)
}

func (s *spyPropertyRepo) ListMediaByPropertyUnitType(ctx context.Context, propertyUnitTypeID domain.ID) ([]domain.Media, error) {
	return s.stubPropertyRepo.ListMediaByPropertyUnitType(ctx, propertyUnitTypeID)
}

func (s *spyPropertyRepo) ListMediaByAgentOffer(ctx context.Context, agentOfferID domain.ID) ([]domain.Media, error) {
	return s.stubPropertyRepo.ListMediaByAgentOffer(ctx, agentOfferID)
}

func (s *spyPropertyRepo) CreatePropertyUnitType(ctx context.Context, unitType domain.PropertyUnitType, actorUserID domain.ID) (domain.PropertyUnitType, error) {
	s.createdUnitType = unitType
	return s.stubPropertyRepo.CreatePropertyUnitType(ctx, unitType, actorUserID)
}

func (s *spyPropertyRepo) UpdatePropertyUnitType(ctx context.Context, id domain.ID, expectedVersion int, patch domain.PropertyUnitTypePatch, actorUserID domain.ID) (domain.PropertyUnitType, error) {
	s.updateUnitTypeCalls++
	if s.updateUnitTypeErr != nil {
		return domain.PropertyUnitType{}, s.updateUnitTypeErr
	}
	return s.stubPropertyRepo.UpdatePropertyUnitType(ctx, id, expectedVersion, patch, actorUserID)
}

func (s *spyPropertyRepo) ListPropertyUnitTypes(ctx context.Context, propertyID domain.ID) ([]domain.PropertyUnitType, error) {
	return s.stubPropertyRepo.ListPropertyUnitTypes(ctx, propertyID)
}

func (s *spyPropertyRepo) CreateAgentOffer(ctx context.Context, offer domain.AgentOffer) (domain.AgentOffer, error) {
	s.createdOffer = offer
	return s.stubPropertyRepo.CreateAgentOffer(ctx, offer)
}

func (s *spyPropertyRepo) UpdateAgentOffer(ctx context.Context, id domain.ID, expectedVersion int, patch domain.AgentOfferPatch, actorUserID domain.ID) (domain.AgentOffer, error) {
	s.updateAgentOfferCalls++
	if s.updateAgentOfferErr != nil {
		return domain.AgentOffer{}, s.updateAgentOfferErr
	}
	return s.stubPropertyRepo.UpdateAgentOffer(ctx, id, expectedVersion, patch, actorUserID)
}

func (s *spyPropertyRepo) ListAgentOffers(ctx context.Context, unitTypeID domain.ID) ([]domain.AgentOffer, error) {
	return s.stubPropertyRepo.ListAgentOffers(ctx, unitTypeID)
}

func (s *spyPropertyRepo) ArchiveAgentOffer(ctx context.Context, id domain.ID, expectedVersion int, actorUserID domain.ID) (domain.AgentOffer, error) {
	s.archiveAgentOfferCalls++
	if s.archiveAgentOfferErr != nil {
		return domain.AgentOffer{}, s.archiveAgentOfferErr
	}
	return s.stubPropertyRepo.ArchiveAgentOffer(ctx, id, expectedVersion, actorUserID)
}

type duplicateAgentOfferRepo struct {
	*stubPropertyRepo
}

func (s *duplicateAgentOfferRepo) CreateAgentOffer(ctx context.Context, offer domain.AgentOffer) (domain.AgentOffer, error) {
	return domain.AgentOffer{}, repo.ErrDuplicate
}

func testAppWithCampusOperatorRepo(propertyRepo PropertyStore) *app {
	app := authenticatedTestApp(domain.ID("550e8400-e29b-41d4-a716-446655440001"), &fakeAgentApplicationStore{access: domain.EffectiveAccess{
		Roles:             []string{"campus_operator"},
		CampusOperatorIDs: []domain.ID{"550e8400-e29b-41d4-a716-446655440002"},
	}})
	app.propertyRepo = propertyRepo
	return app
}

func assertErrorCodeResponse(t *testing.T, rr *httptest.ResponseRecorder, status int, wantCode string) {
	t.Helper()

	if rr.Code != status {
		t.Fatalf("status code = %d, want %d", rr.Code, status)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if body.Error.Code != wantCode {
		t.Fatalf("code = %q, want %q", body.Error.Code, wantCode)
	}
}

func intPointer(value int) *int {
	return &value
}

func boolPointer(value bool) *bool {
	return &value
}
