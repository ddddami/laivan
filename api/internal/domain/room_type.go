package domain

type RoomType struct {
	ID          ID
	PropertyID  ID
	Name        string
	Description string
	Timestamps
}
