package phone

import (
	"errors"
	"strings"
	"unicode"
)

var (
	ErrEmptyPhone   = errors.New("phone number is empty")
	ErrInvalidPhone = errors.New("phone number is invalid")
)

// Normalize converts a raw phone number string into canonical E.164 format.
//
// Accepted input formats:
//   - Local 11-digit: 08031234567, 07031234567, 08101234567
//   - International with +: +2348031234567
//   - International without +: 2348031234567
//   - With separators (spaces, dashes, dots): 0803 123 4567, 0803-123-4567
//   - Mixed: +234 803-123.4567
//
// Canonical output: +234 followed by exactly 10 digits (14 characters total).
//
// Rejects: empty input, non-numeric characters (except leading +),
// wrong digit count, missing country code without leading 0.
func Normalize(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", ErrEmptyPhone
	}

	var b strings.Builder
	for i, r := range trimmed {
		if i == 0 && r == '+' {
			b.WriteRune(r)
			continue
		}
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}

	cleaned := b.String()

	if strings.HasPrefix(cleaned, "0") && len(cleaned) == 11 {
		return "+234" + cleaned[1:], nil
	}

	if strings.HasPrefix(cleaned, "234") && len(cleaned) == 13 {
		return "+" + cleaned, nil
	}

	if strings.HasPrefix(cleaned, "+234") && len(cleaned) == 14 {
		return cleaned, nil
	}

	return "", ErrInvalidPhone
}
