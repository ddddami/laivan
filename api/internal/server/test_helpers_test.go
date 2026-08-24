package server

import (
	"context"
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

func (s *stubPropertyRepo) Create(ctx context.Context, property domain.Property) (domain.Property, error) {
	property.ID = domain.ID("550e8400-e29b-41d4-a716-446655440001")
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
		}, nil
	}
	return domain.Property{}, repo.ErrNotFound
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

func (s *stubPropertyRepo) CreateMediaBatch(ctx context.Context, media []domain.Media) ([]domain.Media, error) {
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

func (s *stubPropertyRepo) CreatePropertyUnitType(ctx context.Context, unitType domain.PropertyUnitType) (domain.PropertyUnitType, error) {
	if string(unitType.PropertyID) != "550e8400-e29b-41d4-a716-446655440000" {
		return domain.PropertyUnitType{}, repo.ErrNotFound
	}

	unitType.ID = domain.ID("550e8400-e29b-41d4-a716-446655440020")
	unitType.CreatedAt = time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC)
	unitType.UpdatedAt = time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC)
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
			Timestamps: domain.Timestamps{
				CreatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
			},
		},
	}, nil
}

func testAppWithRepo() *app {
	a := testApp()
	a.propertyRepo = &stubPropertyRepo{}
	return a
}

type spyPropertyRepo struct {
	stub            *stubPropertyRepo
	createdUnitType domain.PropertyUnitType
	createdOffer    domain.AgentOffer
	createdMedia    []domain.Media
	discoveryFilter repo.DiscoveryFilter
}

func (s *spyPropertyRepo) Create(ctx context.Context, property domain.Property) (domain.Property, error) {
	return s.stub.Create(ctx, property)
}

func (s *spyPropertyRepo) GetCampusBySlug(ctx context.Context, slug string) (domain.Campus, error) {
	return s.stub.GetCampusBySlug(ctx, slug)
}

func (s *spyPropertyRepo) Get(ctx context.Context, id domain.ID) (domain.Property, error) {
	return s.stub.Get(ctx, id)
}

func (s *spyPropertyRepo) GetWithDetails(ctx context.Context, id domain.ID) (domain.PropertyDetail, error) {
	return s.stub.GetWithDetails(ctx, id)
}

func (s *spyPropertyRepo) ListWithSummary(ctx context.Context, filter repo.PropertyListFilter) ([]domain.PropertySummary, int, error) {
	return s.stub.ListWithSummary(ctx, filter)
}

func (s *spyPropertyRepo) Discover(ctx context.Context, filter repo.DiscoveryFilter) ([]domain.DiscoveryResult, int, error) {
	s.discoveryFilter = filter
	return s.stub.Discover(ctx, filter)
}

func (s *spyPropertyRepo) CreateMedia(ctx context.Context, media domain.Media) (domain.Media, error) {
	s.createdMedia = append(s.createdMedia, media)
	return s.stub.CreateMedia(ctx, media)
}

func (s *spyPropertyRepo) CreateMediaBatch(ctx context.Context, media []domain.Media) ([]domain.Media, error) {
	s.createdMedia = append(s.createdMedia, media...)
	return s.stub.CreateMediaBatch(ctx, media)
}

func (s *spyPropertyRepo) ListMediaByProperty(ctx context.Context, propertyID domain.ID) ([]domain.Media, error) {
	return s.stub.ListMediaByProperty(ctx, propertyID)
}

func (s *spyPropertyRepo) ListMediaByPropertyUnitType(ctx context.Context, propertyUnitTypeID domain.ID) ([]domain.Media, error) {
	return s.stub.ListMediaByPropertyUnitType(ctx, propertyUnitTypeID)
}

func (s *spyPropertyRepo) ListMediaByAgentOffer(ctx context.Context, agentOfferID domain.ID) ([]domain.Media, error) {
	return s.stub.ListMediaByAgentOffer(ctx, agentOfferID)
}

func (s *spyPropertyRepo) CreatePropertyUnitType(ctx context.Context, unitType domain.PropertyUnitType) (domain.PropertyUnitType, error) {
	s.createdUnitType = unitType
	return s.stub.CreatePropertyUnitType(ctx, unitType)
}

func (s *spyPropertyRepo) ListPropertyUnitTypes(ctx context.Context, propertyID domain.ID) ([]domain.PropertyUnitType, error) {
	return s.stub.ListPropertyUnitTypes(ctx, propertyID)
}

func (s *spyPropertyRepo) CreateAgentOffer(ctx context.Context, offer domain.AgentOffer) (domain.AgentOffer, error) {
	s.createdOffer = offer
	return s.stub.CreateAgentOffer(ctx, offer)
}

func (s *spyPropertyRepo) ListAgentOffers(ctx context.Context, unitTypeID domain.ID) ([]domain.AgentOffer, error) {
	return s.stub.ListAgentOffers(ctx, unitTypeID)
}

type duplicateAgentOfferRepo struct {
	stub *stubPropertyRepo
}

func (s *duplicateAgentOfferRepo) GetCampusBySlug(ctx context.Context, slug string) (domain.Campus, error) {
	return s.stub.GetCampusBySlug(ctx, slug)
}

func (s *duplicateAgentOfferRepo) Create(ctx context.Context, property domain.Property) (domain.Property, error) {
	return s.stub.Create(ctx, property)
}
func (s *duplicateAgentOfferRepo) Get(ctx context.Context, id domain.ID) (domain.Property, error) {
	return s.stub.Get(ctx, id)
}
func (s *duplicateAgentOfferRepo) GetWithDetails(ctx context.Context, id domain.ID) (domain.PropertyDetail, error) {
	return s.stub.GetWithDetails(ctx, id)
}
func (s *duplicateAgentOfferRepo) ListWithSummary(ctx context.Context, filter repo.PropertyListFilter) ([]domain.PropertySummary, int, error) {
	return s.stub.ListWithSummary(ctx, filter)
}
func (s *duplicateAgentOfferRepo) Discover(ctx context.Context, filter repo.DiscoveryFilter) ([]domain.DiscoveryResult, int, error) {
	return s.stub.Discover(ctx, filter)
}
func (s *duplicateAgentOfferRepo) CreateMedia(ctx context.Context, media domain.Media) (domain.Media, error) {
	return s.stub.CreateMedia(ctx, media)
}
func (s *duplicateAgentOfferRepo) CreateMediaBatch(ctx context.Context, media []domain.Media) ([]domain.Media, error) {
	return s.stub.CreateMediaBatch(ctx, media)
}
func (s *duplicateAgentOfferRepo) ListMediaByProperty(ctx context.Context, propertyID domain.ID) ([]domain.Media, error) {
	return s.stub.ListMediaByProperty(ctx, propertyID)
}
func (s *duplicateAgentOfferRepo) ListMediaByPropertyUnitType(ctx context.Context, propertyUnitTypeID domain.ID) ([]domain.Media, error) {
	return s.stub.ListMediaByPropertyUnitType(ctx, propertyUnitTypeID)
}
func (s *duplicateAgentOfferRepo) ListMediaByAgentOffer(ctx context.Context, agentOfferID domain.ID) ([]domain.Media, error) {
	return s.stub.ListMediaByAgentOffer(ctx, agentOfferID)
}
func (s *duplicateAgentOfferRepo) CreatePropertyUnitType(ctx context.Context, unitType domain.PropertyUnitType) (domain.PropertyUnitType, error) {
	return s.stub.CreatePropertyUnitType(ctx, unitType)
}
func (s *duplicateAgentOfferRepo) ListPropertyUnitTypes(ctx context.Context, propertyID domain.ID) ([]domain.PropertyUnitType, error) {
	return s.stub.ListPropertyUnitTypes(ctx, propertyID)
}
func (s *duplicateAgentOfferRepo) CreateAgentOffer(ctx context.Context, offer domain.AgentOffer) (domain.AgentOffer, error) {
	return domain.AgentOffer{}, repo.ErrDuplicate
}
func (s *duplicateAgentOfferRepo) ListAgentOffers(ctx context.Context, unitTypeID domain.ID) ([]domain.AgentOffer, error) {
	return s.stub.ListAgentOffers(ctx, unitTypeID)
}

func intPointer(value int) *int {
	return &value
}

func boolPointer(value bool) *bool {
	return &value
}
