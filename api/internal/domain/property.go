package domain

type Property struct {
	ID          ID
	CampusID    ID
	Name        string
	Location    ApproxLocation
	Description string
	Timestamps
}
