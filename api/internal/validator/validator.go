package validator

import "strings"

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
