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

func TestValidUUID(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"valid lowercase", "550e8400-e29b-41d4-a716-446655440000", true},
		{"valid uppercase", "550E8400-E29B-41D4-A716-446655440000", true},
		{"valid mixed case", "550e8400-e29b-41d4-A716-446655440000", true},
		{"empty string", "", false},
		{"too short", "550e8400-e29b-41d4-a716-44665544", false},
		{"too long", "550e8400-e29b-41d4-a716-44665544000000", false},
		{"invalid hex char", "550e8400-e29b-41d4-a716-44665544000g", false},
		{"missing dashes", "550e8400e29b41d4a716446655440000", false},
		{"wrong dash position", "550e8400e-29b-41d4-a716-446655440000", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidUUID(tt.value); got != tt.want {
				t.Fatalf("ValidUUID(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestPermittedValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		permitted []string
		want     bool
	}{
		{"value found", "private", []string{"private", "shared", "none"}, true},
		{"value not found", "unknown", []string{"private", "shared"}, false},
		{"empty permitted list", "private", []string{}, false},
		{"single element match", "private", []string{"private"}, true},
		{"no match from multiple", "unknown", []string{"private", "shared", "none"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PermittedValue(tt.value, tt.permitted...); got != tt.want {
				t.Fatalf("PermittedValue(%q, %v) = %v, want %v", tt.value, tt.permitted, got, tt.want)
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
		{name: "max value accepts equal", ok: MaxValue(5, 5)},
		{name: "max value rejects over", ok: !MaxValue(6, 5)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.ok {
				t.Fatal("validation helper returned unexpected result")
			}
		})
	}
}
