package validator

import (
	"strings"

	"github.com/ddddami/laivan/internal/phone"
)

type Validator struct {
	FieldErrors map[string]string
}

func New() *Validator {
	return &Validator{FieldErrors: map[string]string{}}
}

func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0
}

func (v *Validator) AddFieldError(field, message string) {
	if v.FieldErrors == nil {
		v.FieldErrors = map[string]string{}
	}

	if _, exists := v.FieldErrors[field]; !exists {
		v.FieldErrors[field] = message
	}
}

func (v *Validator) Check(ok bool, field, message string) {
	if !ok {
		v.AddFieldError(field, message)
	}
}

func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

func MaxChars(value string, limit int) bool {
	return len([]rune(value)) <= limit
}

func Positive(value int) bool {
	return value > 0
}

func MaxValue(value, max int) bool {
	return value <= max
}

func ValidUUID(value string) bool {
	if len(value) != 36 {
		return false
	}
	for i, c := range value {
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
				return false
			}
		}
	}
	return true
}

func ValidPhone(value string) bool {
	_, err := phone.Normalize(value)
	return err == nil
}

func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	for i := range permittedValues {
		if value == permittedValues[i] {
			return true
		}
	}
	return false
}
