package data

import (
	"testing"

	"github.com/ddddami/laivan/internal/validator"
)

func TestFiltersSortColumn(t *testing.T) {
	f := Filters{Sort: "-created_at", SortSafelist: []string{"created_at", "-created_at", "name", "-name"}}
	if got := f.SortColumn(); got != "created_at" {
		t.Fatalf("SortColumn = %q, want created_at", got)
	}
}

func TestFiltersSortDirection(t *testing.T) {
	f := Filters{Sort: "-created_at", SortSafelist: []string{"created_at", "-created_at"}}
	if got := f.SortDirection(); got != "DESC" {
		t.Fatalf("SortDirection = %q, want DESC", got)
	}

	f.Sort = "created_at"
	if got := f.SortDirection(); got != "ASC" {
		t.Fatalf("SortDirection = %q, want ASC", got)
	}
}

func TestFiltersLimitAndOffset(t *testing.T) {
	f := Filters{Page: 3, PageSize: 20}
	if got := f.Limit(); got != 20 {
		t.Fatalf("Limit = %d, want 20", got)
	}
	if got := f.Offset(); got != 40 {
		t.Fatalf("Offset = %d, want 40", got)
	}
}

func TestValidateFilters(t *testing.T) {
	v := validator.New()
	f := Filters{Page: 1, PageSize: 20, Sort: "created_at", SortSafelist: []string{"created_at"}}
	ValidateFilters(v, f)
	if !v.Valid() {
		t.Fatal("valid filters should pass")
	}

	v2 := validator.New()
	f2 := Filters{Page: 0, PageSize: 200, Sort: "hacked", SortSafelist: []string{"created_at"}}
	ValidateFilters(v2, f2)
	if v2.Valid() {
		t.Fatal("invalid filters should fail")
	}
	if v2.FieldErrors["page"] == "" {
		t.Fatal("page error missing")
	}
	if v2.FieldErrors["page_size"] == "" {
		t.Fatal("page_size error missing")
	}
	if v2.FieldErrors["sort"] == "" {
		t.Fatal("sort error missing")
	}
}

func TestCalculateMetadata(t *testing.T) {
	m := CalculateMetadata(95, 1, 20)
	if m.CurrentPage != 1 {
		t.Fatalf("CurrentPage = %d, want 1", m.CurrentPage)
	}
	if m.LastPage != 5 {
		t.Fatalf("LastPage = %d, want 5", m.LastPage)
	}
	if m.TotalRecords != 95 {
		t.Fatalf("TotalRecords = %d, want 95", m.TotalRecords)
	}

	m2 := CalculateMetadata(0, 1, 20)
	if m2.TotalRecords != 0 {
		t.Fatalf("empty metadata TotalRecords = %d, want 0", m2.TotalRecords)
	}
}
