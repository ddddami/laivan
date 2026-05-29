package storage

import (
	"strings"
	"testing"
)

func TestMediaObjectKeySanitization(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		targetID string
		mediaID  string
		filename string
		want     string
	}{
		{
			name:     "normal inputs",
			target:   "property",
			targetID: "550e8400-e29b-41d4-a716-446655440000",
			mediaID:  "abc123",
			filename: "room.jpg",
			want:     "media/property/550e8400-e29b-41d4-a716-446655440000/abc123/room.jpg",
		},
		{
			name:     "filename with spaces",
			target:   "property",
			targetID: "abc",
			mediaID:  "def",
			filename: "my room photo.jpg",
			want:     "media/property/abc/def/myroomphoto.jpg",
		},
		{
			name:     "path traversal attempt",
			target:   "property",
			targetID: "abc",
			mediaID:  "def",
			filename: "../../../etc/passwd",
			want:     "media/property/abc/def/passwd",
		},
		{
			name:     "special characters in target",
			target:   "agent_offer",
			targetID: "id-with_underscore",
			mediaID:  "media-123",
			filename: "kitchen.png",
			want:     "media/agent_offer/id-with_underscore/media-123/kitchen.png",
		},
		{
			name:     "empty filename",
			target:   "property",
			targetID: "abc",
			mediaID:  "def",
			filename: "",
			want:     "media/property/abc/def/upload",
		},
		{
			name:     "unicode in filename",
			target:   "property",
			targetID: "abc",
			mediaID:  "def",
			filename: "casa.jpg",
			want:     "media/property/abc/def/casa.jpg",
		},
		{
			name:     "all special chars stripped",
			target:   "!@#$%",
			targetID: "!@#$%",
			mediaID:  "!@#$%",
			filename: "!@#$%.jpg",
			want:     "media/unknown/unknown/unknown/.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MediaObjectKey(tt.target, tt.targetID, tt.mediaID, tt.filename)
			if got != tt.want {
				t.Fatalf("MediaObjectKey() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSafePathPartStripsUnsafeCharacters(t *testing.T) {
	if got := safePathPart("hello world"); got != "helloworld" {
		t.Fatalf("safePathPart = %q, want helloworld", got)
	}
	if got := safePathPart("../../etc"); got != "etc" {
		t.Fatalf("safePathPart = %q, want etc", got)
	}
	if got := safePathPart(""); got != "unknown" {
		t.Fatalf("safePathPart = %q, want unknown", got)
	}
}

func TestSafeFilenameStripsUnsafeCharacters(t *testing.T) {
	if got := safeFilename("room.jpg"); got != "room.jpg" {
		t.Fatalf("safeFilename = %q, want room.jpg", got)
	}
	if got := safeFilename("../../../etc/passwd"); got != "passwd" {
		t.Fatalf("safeFilename = %q, want passwd", got)
	}
	if got := safeFilename(""); got != "upload" {
		t.Fatalf("safeFilename = %q, want upload", got)
	}
	if got := safeFilename("!@#$%"); strings.Contains(got, "@") {
		t.Fatalf("safeFilename = %q, should not contain @", got)
	}
}
