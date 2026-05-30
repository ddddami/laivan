package phone

import (
	"errors"
	"testing"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		// Happy path: valid Nigerian mobile formats
		{name: "local 11-digit 080", input: "08031234567", want: "+2348031234567", wantErr: nil},
		{name: "local 11-digit 070", input: "07031234567", want: "+2347031234567", wantErr: nil},
		{name: "local 11-digit 081", input: "08101234567", want: "+2348101234567", wantErr: nil},
		{name: "local 11-digit 090", input: "09031234567", want: "+2349031234567", wantErr: nil},
		{name: "local 11-digit 091", input: "09121234567", want: "+2349121234567", wantErr: nil},
		{name: "international with +", input: "+2348031234567", want: "+2348031234567", wantErr: nil},
		{name: "international without +", input: "2348031234567", want: "+2348031234567", wantErr: nil},
		{name: "already normalized", input: "+2347031234567", want: "+2347031234567", wantErr: nil},

		// With separators
		{name: "spaces local", input: "0803 123 4567", want: "+2348031234567", wantErr: nil},
		{name: "dashes local", input: "0803-123-4567", want: "+2348031234567", wantErr: nil},
		{name: "dots local", input: "0803.123.4567", want: "+2348031234567", wantErr: nil},
		{name: "spaces intl", input: "+234 803 123 4567", want: "+2348031234567", wantErr: nil},
		{name: "mixed separators", input: "+234-803.123 4567", want: "+2348031234567", wantErr: nil},

		// Edge cases around whitespace
		{name: "leading trailing space", input: "  08031234567  ", want: "+2348031234567", wantErr: nil},
		{name: "tab and newline", input: "\t08031234567\n", want: "+2348031234567", wantErr: nil},

		// Invalid: empty or non-numeric
		{name: "empty string", input: "", want: "", wantErr: ErrEmptyPhone},
		{name: "whitespace only", input: "   ", want: "", wantErr: ErrEmptyPhone},
		{name: "letters", input: "abc", want: "", wantErr: ErrInvalidPhone},
		{name: "letters mixed", input: "0803abc4567", want: "", wantErr: ErrInvalidPhone},
		{name: "plus only", input: "+", want: "", wantErr: ErrInvalidPhone},
		{name: "plus 234 no digits", input: "+234", want: "", wantErr: ErrInvalidPhone},

		// Invalid: wrong length
		{name: "too short local", input: "0803123456", want: "", wantErr: ErrInvalidPhone},
		{name: "too long local", input: "080312345678", want: "", wantErr: ErrInvalidPhone},
		{name: "too short intl", input: "+23480312345", want: "", wantErr: ErrInvalidPhone},
		{name: "too long intl", input: "+23480312345678", want: "", wantErr: ErrInvalidPhone},
		{name: "too short 234 prefix", input: "23480312345", want: "", wantErr: ErrInvalidPhone},
		{name: "too long 234 prefix", input: "23480312345678", want: "", wantErr: ErrInvalidPhone},

		// Invalid: missing prefix
		{name: "no prefix no leading zero", input: "12345678901", want: "", wantErr: ErrInvalidPhone},
		{name: "double country code", input: "+23408031234567", want: "", wantErr: ErrInvalidPhone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Normalize(tt.input)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeDoesNotValidateNetworkPrefix(t *testing.T) {
	// Any 10 digits after +234 should be accepted. We do not validate
	// Nigerian mobile network prefixes to avoid maintenance burden
	// when new prefixes are allocated.
	tests := []struct {
		input string
		want  string
	}{
		{"08001234567", "+2348001234567"},
		{"06031234567", "+2346031234567"},
		{"09231234567", "+2349231234567"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := Normalize(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
