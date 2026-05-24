package domain

type Campus struct {
	ID        ID
	Slug      string
	Name      string
	ShortName string
	IsActive  bool
	Timestamps
}
