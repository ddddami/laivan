package domain

type UnitCategory string

const (
	UnitCategorySingleRoom       UnitCategory = "single_room"
	UnitCategorySelfContained    UnitCategory = "self_contained"
	UnitCategoryRoomAndParlour   UnitCategory = "room_and_parlour"
	UnitCategoryOneBedroomFlat   UnitCategory = "one_bedroom_flat"
	UnitCategoryTwoBedroomFlat   UnitCategory = "two_bedroom_flat"
	UnitCategoryThreeBedroomFlat UnitCategory = "three_bedroom_flat"
	UnitCategoryOther            UnitCategory = "other"
)

type PropertyUnitType struct {
	ID          ID
	PropertyID  ID
	Category    UnitCategory
	Name        string
	Description string
	Notes       string
	Structure   UnitStructure
	Version     int
	Timestamps
}

type PropertyUnitTypePatch struct {
	Category     *UnitCategory
	Name         *string
	Description  *string
	Notes        *string
	BedroomCount *int
	HasParlour   *bool
	BathroomType *string
	KitchenType  *string
}

type UnitStructure struct {
	BedroomCount *int
	HasParlour   *bool
	BathroomType string
	KitchenType  string
}
