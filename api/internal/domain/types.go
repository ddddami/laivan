package domain

import "time"

type ID string

type Money struct {
	AmountKobo int
}

type ApproxLocation struct {
	Area     string
	Landmark string
	District string
}

type Timestamps struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}
