package domain

type PropertyUnitType struct {
	ID          ID
	PropertyID  ID
	Name        string
	Description string
	Timestamps
}
