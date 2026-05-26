package server

import (
	"net/url"
	"testing"

	"github.com/ddddami/laivan/internal/validator"
)

func TestReadString(t *testing.T) {
	qs := url.Values{}
	qs.Set("name", "Alice")

	if got := readString(qs, "name", ""); got != "Alice" {
		t.Fatalf("readString = %q, want Alice", got)
	}
	if got := readString(qs, "missing", "default"); got != "default" {
		t.Fatalf("readString default = %q, want default", got)
	}
}

func TestReadCSV(t *testing.T) {
	qs := url.Values{}
	qs.Set("tags", "a,b,c")

	if got := readCSV(qs, "tags", nil); len(got) != 3 || got[0] != "a" {
		t.Fatalf("readCSV = %v, want [a b c]", got)
	}
	if got := readCSV(qs, "missing", []string{"x"}); len(got) != 1 || got[0] != "x" {
		t.Fatalf("readCSV default = %v, want [x]", got)
	}
}

func TestReadInt(t *testing.T) {
	qs := url.Values{}
	qs.Set("limit", "25")

	v := validator.New()
	if got := readInt(qs, "limit", 10, v); got != 25 {
		t.Fatalf("readInt = %d, want 25", got)
	}
	if !v.Valid() {
		t.Fatal("validator should be valid for good integer")
	}

	v2 := validator.New()
	if got := readInt(qs, "missing", 10, v2); got != 10 {
		t.Fatalf("readInt default = %d, want 10", got)
	}

	qs.Set("bad", "not-a-number")
	v3 := validator.New()
	if got := readInt(qs, "bad", 10, v3); got != 10 {
		t.Fatalf("readInt bad default = %d, want 10", got)
	}
	if v3.Valid() {
		t.Fatal("validator should be invalid for bad integer")
	}
}

func TestReadBool(t *testing.T) {
	qs := url.Values{}
	qs.Set("active", "true")

	v := validator.New()
	if got := readBool(qs, "active", nil, v); got == nil || !*got {
		t.Fatalf("readBool = %v, want true", got)
	}
	if !v.Valid() {
		t.Fatal("validator should be valid for true")
	}

	qs.Set("active", "1")
	v2 := validator.New()
	if got := readBool(qs, "active", nil, v2); got == nil || !*got {
		t.Fatalf("readBool numeric = %v, want true", got)
	}

	qs.Set("active", "false")
	v3 := validator.New()
	if got := readBool(qs, "active", nil, v3); got == nil || *got {
		t.Fatalf("readBool false = %v, want false", got)
	}

	v4 := validator.New()
	if got := readBool(qs, "missing", nil, v4); got != nil {
		t.Fatalf("readBool missing = %v, want nil", got)
	}

	qs.Set("bad", "yes")
	v5 := validator.New()
	if got := readBool(qs, "bad", nil, v5); got != nil {
		t.Fatalf("readBool bad default = %v, want nil", got)
	}
	if v5.Valid() {
		t.Fatal("validator should be invalid for bad boolean")
	}
}
