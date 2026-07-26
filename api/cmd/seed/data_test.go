package main

import (
	"net/http"
	"testing"
)

func TestSeedMediaAssetsAreEmbeddedJPEGs(t *testing.T) {
	t.Parallel()

	ids := make(map[string]struct{}, len(mediaItems))
	objectKeys := make(map[string]struct{}, len(mediaItems))
	for _, media := range mediaItems {
		if _, exists := ids[media.ID]; exists {
			t.Fatalf("duplicate media ID %q", media.ID)
		}
		ids[media.ID] = struct{}{}

		if _, exists := objectKeys[media.ObjectKey]; exists {
			t.Fatalf("duplicate media object key %q", media.ObjectKey)
		}
		objectKeys[media.ObjectKey] = struct{}{}

		data, err := seedMediaAssets.ReadFile(media.AssetPath)
		if err != nil {
			t.Fatalf("read %s: %v", media.AssetPath, err)
		}
		if got := http.DetectContentType(data); got != "image/jpeg" {
			t.Fatalf("%s content type = %q, want image/jpeg", media.AssetPath, got)
		}
	}
}
