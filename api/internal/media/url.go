package media

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/ddddami/laivan/internal/config"
)

type URLBuilder struct {
	baseURL       string
	publicBaseURL string
	sourceBaseURL string
	key           []byte
	salt          []byte
}

func NewURLBuilder(cfg config.MediaConfig) (*URLBuilder, error) {
	if cfg.Imgproxy == nil {
		return nil, errors.New("imgproxy config is required")
	}
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.Imgproxy.BaseURL), "/")
	if baseURL == "" {
		return nil, errors.New("imgproxy base URL is required")
	}

	key, err := hex.DecodeString(strings.TrimSpace(cfg.Imgproxy.Key))
	if err != nil || len(key) == 0 {
		return nil, errors.New("imgproxy key must be non-empty hex")
	}

	salt, err := hex.DecodeString(strings.TrimSpace(cfg.Imgproxy.Salt))
	if err != nil || len(salt) == 0 {
		return nil, errors.New("imgproxy salt must be non-empty hex")
	}

	return &URLBuilder{
		baseURL:       baseURL,
		publicBaseURL: strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/"),
		sourceBaseURL: strings.TrimRight(strings.TrimSpace(cfg.Imgproxy.SourceBaseURL), "/"),
		key:           key,
		salt:          salt,
	}, nil
}

func (b *URLBuilder) ThumbnailURL(sourceURL string) string {
	return b.signedURL("rs:fill:400:300/plain", sourceURL, "webp")
}

func (b *URLBuilder) MediumURL(sourceURL string) string {
	return b.signedURL("rs:fit:900:700/plain", sourceURL, "webp")
}

func (b *URLBuilder) signedURL(options, sourceURL, extension string) string {
	sourceURL = strings.TrimSpace(sourceURL)
	if sourceURL == "" {
		return ""
	}
	if b.publicBaseURL != "" && b.sourceBaseURL != "" && strings.HasPrefix(sourceURL, b.publicBaseURL) {
		sourceURL = b.sourceBaseURL + strings.TrimPrefix(sourceURL, b.publicBaseURL)
	}

	path := fmt.Sprintf("/%s/%s@%s", strings.Trim(options, "/"), url.PathEscape(sourceURL), extension)
	mac := hmac.New(sha256.New, b.key)
	mac.Write(b.salt)
	mac.Write([]byte(path))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return b.baseURL + "/" + signature + path
}
