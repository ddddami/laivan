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

func TestValidPhone(t *testing.T) {
	tests := []struct {
		name string
		ok   bool
	}{
		{name: "local 11-digit", ok: ValidPhone("08031234567")},
		{name: "local 070", ok: ValidPhone("07031234567")},
		{name: "local 081", ok: ValidPhone("08101234567")},
		{name: "local 090", ok: ValidPhone("09031234567")},
		{name: "local 091", ok: ValidPhone("09121234567")},
		{name: "intl with +", ok: ValidPhone("+2348031234567")},
		{name: "intl without +", ok: ValidPhone("2348031234567")},
		{name: "with spaces", ok: ValidPhone("0803 123 4567")},
		{name: "with dashes", ok: ValidPhone("0803-123-4567")},
		{name: "with dots", ok: ValidPhone("0803.123.4567")},
		{name: "mixed separators", ok: ValidPhone("+234-803.123 4567")},
		{name: "already normalized", ok: ValidPhone("+2347031234567")},
		{name: "empty rejects", ok: !ValidPhone("")},
		{name: "whitespace rejects", ok: !ValidPhone("   ")},
		{name: "letters reject", ok: !ValidPhone("abc")},
		{name: "too short local rejects", ok: !ValidPhone("0803123456")},
		{name: "too long local rejects", ok: !ValidPhone("080312345678")},
		{name: "too short intl rejects", ok: !ValidPhone("+23480312345")},
		{name: "too long intl rejects", ok: !ValidPhone("+23480312345678")},
		{name: "no prefix rejects", ok: !ValidPhone("12345678901")},
		{name: "double country code rejects", ok: !ValidPhone("+23408031234567")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.ok {
				t.Fatal("ValidPhone returned unexpected result")
			}
		})
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
