package storage

import "testing"

func TestMediaObjectKeySanitizesParts(t *testing.T) {
	got := MediaObjectKey("property", "550e8400-e29b-41d4-a716-446655440000", "550e8400-e29b-41d4-a716-446655440001", "../room one.jpg")
	want := "media/property/550e8400-e29b-41d4-a716-446655440000/550e8400-e29b-41d4-a716-446655440001/roomone.jpg"

	if got != want {
		t.Fatalf("MediaObjectKey() = %q, want %q", got, want)
	}
}
