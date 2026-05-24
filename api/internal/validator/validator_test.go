package validator

import "testing"

func TestValidatorValid(t *testing.T) {
	v := New()

	if !v.Valid() {
		t.Fatal("new validator is invalid")
	}

	v.AddFieldError("name", "Name is required")

	if v.Valid() {
		t.Fatal("validator with field errors is valid")
	}
}

func TestValidatorCheck(t *testing.T) {
	v := New()

	v.Check(false, "name", "Name is required")
	v.Check(true, "price", "Price must be positive")

	if got := v.FieldErrors["name"]; got != "Name is required" {
		t.Fatalf("name error = %q, want Name is required", got)
	}

	if _, exists := v.FieldErrors["price"]; exists {
		t.Fatal("price error exists")
	}
}

func TestValidatorKeepsFirstFieldError(t *testing.T) {
	v := New()

	v.AddFieldError("name", "Name is required")
	v.AddFieldError("name", "Name is too long")

	if got := v.FieldErrors["name"]; got != "Name is required" {
		t.Fatalf("name error = %q, want Name is required", got)
	}
}

func TestValidationHelpers(t *testing.T) {
	tests := []struct {
		name string
		ok   bool
	}{
		{name: "not blank accepts visible text", ok: NotBlank("South Gate")},
		{name: "not blank rejects whitespace", ok: !NotBlank("  ")},
		{name: "max chars accepts exact limit", ok: MaxChars("FUTA", 4)},
		{name: "max chars rejects over limit", ok: !MaxChars("FUTA", 3)},
		{name: "positive accepts positive", ok: Positive(1)},
		{name: "positive rejects zero", ok: !Positive(0)},
		{name: "positive rejects negative", ok: !Positive(-1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.ok {
				t.Fatal("validation helper returned unexpected result")
			}
		})
	}
}
