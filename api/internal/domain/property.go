package domain

type Property struct {
	ID          ID
	Name        string
	Location    ApproxLocation
	Description string
	Timestamps
}
