package media

import (
	"strings"
	"testing"

	"github.com/ddddami/laivan/internal/config"
)

func TestURLBuilderBuildsSignedThumbnailURL(t *testing.T) {
	builder, err := NewURLBuilder(config.MediaConfig{
		Imgproxy: config.ImgproxyConfig{
			BaseURL: "https://images.example.test/",
			Key:     "00112233445566778899aabbccddeeff",
			Salt:    "ffeeddccbbaa99887766554433221100",
		},
	})
	if err != nil {
		t.Fatalf("NewURLBuilder returned error: %v", err)
	}

	got := builder.ThumbnailURL("https://media.example.test/alice.jpg")

	if !strings.HasPrefix(got, "https://images.example.test/") {
		t.Fatalf("thumbnail URL = %q, want imgproxy base URL", got)
	}
	if !strings.Contains(got, "/rs:fill:400:300/plain/") {
		t.Fatalf("thumbnail URL = %q, want thumbnail resize options", got)
	}
	if !strings.HasSuffix(got, "@webp") {
		t.Fatalf("thumbnail URL = %q, want webp extension", got)
	}
}

func TestURLBuilderUsesInternalSourceBaseURL(t *testing.T) {
	builder, err := NewURLBuilder(config.MediaConfig{
		PublicBaseURL: "http://localhost:9000/laivan-dev",
		Imgproxy: config.ImgproxyConfig{
			BaseURL:       "http://localhost:8080",
			SourceBaseURL: "http://minio:9000/laivan-dev",
			Key:           "00112233445566778899aabbccddeeff",
			Salt:          "ffeeddccbbaa99887766554433221100",
		},
	})
	if err != nil {
		t.Fatalf("NewURLBuilder returned error: %v", err)
	}

	got := builder.ThumbnailURL("http://localhost:9000/laivan-dev/seed/alice.jpg")

	if !strings.Contains(got, "http:%2F%2Fminio:9000%2Flaivan-dev%2Fseed%2Falice.jpg") {
		t.Fatalf("thumbnail URL = %q, want internal MinIO source URL", got)
	}
}

func TestURLBuilderRejectsInvalidSigningConfig(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.MediaConfig
	}{
		{name: "missing base URL", cfg: config.MediaConfig{Imgproxy: config.ImgproxyConfig{Key: "abcd", Salt: "1234"}}},
		{name: "invalid key", cfg: config.MediaConfig{Imgproxy: config.ImgproxyConfig{BaseURL: "https://images.example.test", Key: "not-hex", Salt: "1234"}}},
		{name: "invalid salt", cfg: config.MediaConfig{Imgproxy: config.ImgproxyConfig{BaseURL: "https://images.example.test", Key: "abcd", Salt: "not-hex"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewURLBuilder(tt.cfg)
			if err == nil {
				t.Fatal("NewURLBuilder returned nil error")
			}
		})
	}
}
