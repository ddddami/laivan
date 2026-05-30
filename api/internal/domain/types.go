package domain

import "time"

// MaxNaira is the maximum price in naira that can be safely stored
// as kobo in a 32-bit signed integer column (price_kobo).
const MaxNaira = 21_474_836

type ID string

type Money struct {
	AmountKobo int
}

// Naira converts internal kobo storage to naira for API display.
func (m Money) Naira() int {
	return m.AmountKobo / 100
}

// Kobo converts naira to kobo for internal storage and database filtering.
func Kobo(naira int) int {
	return naira * 100
}

type ApproxLocation struct {
	Area     string
	Landmark string
}

type Timestamps struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}
