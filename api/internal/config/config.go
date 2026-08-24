package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	EnvDevelopment = "development"
	EnvTest        = "test"
	EnvProduction  = "production"
)

type Config struct {
	Env             string
	Port            int
	DatabaseURL     string
	DBMaxConns      int32
	AllowedOrigins  []string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	Auth            AuthConfig
	Media           MediaConfig
}

type AuthConfig struct {
	GoogleClientID      string
	GoogleClientSecret  string
	GoogleRedirectURL   string
	WebOrigin           string
	SessionCookieName   string
	CSRFCookieName      string
	OIDCStateSigningKey string
	OIDCStateDuration   time.Duration
	SessionDuration     time.Duration
	SecureCookies       bool
}

func (c AuthConfig) OIDCStateCookieName() string {
	return "laivan_oidc_attempt"
}

func (c AuthConfig) IsConfigured() bool {
	return c.GoogleClientID != ""
}

type MediaConfig struct {
	Enabled        bool
	S3Endpoint     string
	S3Bucket       string
	S3Region       string
	S3AccessKeyID  string
	S3SecretKey    string
	S3UseSSL       bool
	PublicBaseURL  string
	MaxUploadBytes int64
	Imgproxy       ImgproxyConfig
}

type ImgproxyConfig struct {
	BaseURL       string
	SourceBaseURL string
	Key           string
	Salt          string
}

func Load() (Config, error) {
	cfg := Config{
		Env:             EnvDevelopment,
		Port:            4000,
		DBMaxConns:      25,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     time.Minute,
		ShutdownTimeout: 10 * time.Second,
		Auth: AuthConfig{
			SessionCookieName: "laivan_session",
			CSRFCookieName:    "laivan_csrf",
			OIDCStateDuration: 10 * time.Minute,
			SessionDuration:   30 * 24 * time.Hour,
		},
		Media: MediaConfig{
			S3Region:       "us-east-1",
			MaxUploadBytes: 10 << 20,
		},
	}

	var err error

	cfg.Env = stringEnv("LAIVAN_ENV", cfg.Env)
	cfg.DatabaseURL = stringEnv("LAIVAN_DB_URL", cfg.DatabaseURL)
	cfg.AllowedOrigins = stringsEnv("LAIVAN_ALLOWED_ORIGINS", []string{"http://localhost:5173"})
	cfg.Auth.GoogleClientID = stringEnv("LAIVAN_GOOGLE_CLIENT_ID", cfg.Auth.GoogleClientID)
	cfg.Auth.GoogleClientSecret = stringEnv("LAIVAN_GOOGLE_CLIENT_SECRET", cfg.Auth.GoogleClientSecret)
	cfg.Auth.GoogleRedirectURL = stringEnv("LAIVAN_GOOGLE_REDIRECT_URL", cfg.Auth.GoogleRedirectURL)
	cfg.Auth.WebOrigin = stringEnv("LAIVAN_WEB_ORIGIN", firstString(cfg.AllowedOrigins))
	cfg.Auth.SessionCookieName = stringEnv("LAIVAN_SESSION_COOKIE_NAME", cfg.Auth.SessionCookieName)
	cfg.Auth.CSRFCookieName = stringEnv("LAIVAN_CSRF_COOKIE_NAME", cfg.Auth.CSRFCookieName)
	cfg.Auth.OIDCStateSigningKey = stringEnv("LAIVAN_OIDC_STATE_SIGNING_KEY", cfg.Auth.OIDCStateSigningKey)
	cfg.Media.S3Endpoint = stringEnv("LAIVAN_S3_ENDPOINT", cfg.Media.S3Endpoint)
	cfg.Media.S3Bucket = stringEnv("LAIVAN_S3_BUCKET", cfg.Media.S3Bucket)
	cfg.Media.S3Region = stringEnv("LAIVAN_S3_REGION", cfg.Media.S3Region)
	cfg.Media.S3AccessKeyID = stringEnv("LAIVAN_S3_ACCESS_KEY_ID", cfg.Media.S3AccessKeyID)
	cfg.Media.S3SecretKey = stringEnv("LAIVAN_S3_SECRET_ACCESS_KEY", cfg.Media.S3SecretKey)
	cfg.Media.PublicBaseURL = stringEnv("LAIVAN_MEDIA_PUBLIC_BASE_URL", cfg.Media.PublicBaseURL)
	cfg.Media.Imgproxy.BaseURL = stringEnv("LAIVAN_IMGPROXY_BASE_URL", cfg.Media.Imgproxy.BaseURL)
	cfg.Media.Imgproxy.SourceBaseURL = stringEnv("LAIVAN_IMGPROXY_SOURCE_BASE_URL", cfg.Media.Imgproxy.SourceBaseURL)
	cfg.Media.Imgproxy.Key = stringEnv("LAIVAN_IMGPROXY_KEY", cfg.Media.Imgproxy.Key)
	cfg.Media.Imgproxy.Salt = stringEnv("LAIVAN_IMGPROXY_SALT", cfg.Media.Imgproxy.Salt)

	cfg.Media.Enabled, err = boolEnv("LAIVAN_MEDIA_ENABLED", cfg.Media.Enabled)
	if err != nil {
		return Config{}, err
	}

	cfg.Media.S3UseSSL, err = boolEnv("LAIVAN_S3_USE_SSL", cfg.Media.S3UseSSL)
	if err != nil {
		return Config{}, err
	}

	cfg.Port, err = intEnv("LAIVAN_PORT", cfg.Port)
	if err != nil {
		return Config{}, err
	}

	cfg.DBMaxConns, err = int32Env("LAIVAN_DB_MAX_CONNS", cfg.DBMaxConns)
	if err != nil {
		return Config{}, err
	}

	cfg.ReadTimeout, err = durationEnv("LAIVAN_READ_TIMEOUT", cfg.ReadTimeout)
	if err != nil {
		return Config{}, err
	}

	cfg.WriteTimeout, err = durationEnv("LAIVAN_WRITE_TIMEOUT", cfg.WriteTimeout)
	if err != nil {
		return Config{}, err
	}

	cfg.IdleTimeout, err = durationEnv("LAIVAN_IDLE_TIMEOUT", cfg.IdleTimeout)
	if err != nil {
		return Config{}, err
	}

	cfg.ShutdownTimeout, err = durationEnv("LAIVAN_SHUTDOWN_TIMEOUT", cfg.ShutdownTimeout)
	if err != nil {
		return Config{}, err
	}

	cfg.Auth.OIDCStateDuration, err = durationEnv("LAIVAN_OIDC_STATE_DURATION", cfg.Auth.OIDCStateDuration)
	if err != nil {
		return Config{}, err
	}

	cfg.Auth.SessionDuration, err = durationEnv("LAIVAN_SESSION_DURATION", cfg.Auth.SessionDuration)
	if err != nil {
		return Config{}, err
	}

	cfg.Auth.SecureCookies, err = boolEnv("LAIVAN_SECURE_COOKIES", cfg.Auth.SecureCookies)
	if err != nil {
		return Config{}, err
	}

	cfg.Media.MaxUploadBytes, err = int64Env("LAIVAN_MEDIA_MAX_UPLOAD_BYTES", cfg.Media.MaxUploadBytes)
	if err != nil {
		return Config{}, err
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	switch c.Env {
	case EnvDevelopment, EnvTest, EnvProduction:
	default:
		return fmt.Errorf("LAIVAN_ENV must be one of %q, %q, or %q", EnvDevelopment, EnvTest, EnvProduction)
	}

	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("LAIVAN_PORT must be between 1 and 65535")
	}

	if c.DatabaseURL == "" {
		return fmt.Errorf("LAIVAN_DB_URL is required")
	}
	if c.DBMaxConns <= 0 {
		return fmt.Errorf("LAIVAN_DB_MAX_CONNS must be greater than zero")
	}

	if len(c.AllowedOrigins) == 0 {
		return fmt.Errorf("LAIVAN_ALLOWED_ORIGINS must include at least one origin")
	}

	if c.ReadTimeout <= 0 {
		return fmt.Errorf("LAIVAN_READ_TIMEOUT must be greater than zero")
	}

	if c.WriteTimeout <= 0 {
		return fmt.Errorf("LAIVAN_WRITE_TIMEOUT must be greater than zero")
	}

	if c.IdleTimeout <= 0 {
		return fmt.Errorf("LAIVAN_IDLE_TIMEOUT must be greater than zero")
	}

	if c.ShutdownTimeout <= 0 {
		return fmt.Errorf("LAIVAN_SHUTDOWN_TIMEOUT must be greater than zero")
	}

	if c.Auth.SessionCookieName == "" {
		return fmt.Errorf("LAIVAN_SESSION_COOKIE_NAME is required")
	}
	if c.Auth.CSRFCookieName == "" {
		return fmt.Errorf("LAIVAN_CSRF_COOKIE_NAME is required")
	}
	if c.Auth.OIDCStateDuration <= 0 || c.Auth.OIDCStateDuration > 15*time.Minute {
		return fmt.Errorf("LAIVAN_OIDC_STATE_DURATION must be greater than zero and no more than 15m")
	}
	if c.Auth.SessionDuration <= 0 || c.Auth.SessionDuration > 30*24*time.Hour {
		return fmt.Errorf("LAIVAN_SESSION_DURATION must be greater than zero and no more than 720h")
	}
	if err := validateOrigin(c.Auth.WebOrigin, c.IsProduction()); err != nil {
		return fmt.Errorf("LAIVAN_WEB_ORIGIN %w", err)
	}

	googleConfigured := c.Auth.GoogleClientID != "" ||
		c.Auth.GoogleClientSecret != "" ||
		c.Auth.GoogleRedirectURL != "" ||
		c.Auth.OIDCStateSigningKey != ""
	if googleConfigured || c.IsProduction() {
		if c.Auth.GoogleClientID == "" {
			return fmt.Errorf("LAIVAN_GOOGLE_CLIENT_ID is required when Google auth is configured")
		}
		if c.Auth.GoogleClientSecret == "" {
			return fmt.Errorf("LAIVAN_GOOGLE_CLIENT_SECRET is required when Google auth is configured")
		}
		if err := validateRedirectURL(c.Auth.GoogleRedirectURL, c.IsProduction()); err != nil {
			return fmt.Errorf("LAIVAN_GOOGLE_REDIRECT_URL %w", err)
		}
		if len(c.Auth.OIDCStateSigningKey) < 32 {
			return fmt.Errorf("LAIVAN_OIDC_STATE_SIGNING_KEY must be at least 32 characters when Google auth is configured")
		}
	}
	if c.IsProduction() && !c.Auth.SecureCookies {
		return fmt.Errorf("LAIVAN_SECURE_COOKIES must be true in production")
	}

	if c.Media.MaxUploadBytes <= 0 {
		return fmt.Errorf("LAIVAN_MEDIA_MAX_UPLOAD_BYTES must be greater than zero")
	}

	if c.Media.Enabled {
		if c.Media.S3Endpoint == "" {
			return fmt.Errorf("LAIVAN_S3_ENDPOINT is required when media is enabled")
		}
		if c.Media.S3Bucket == "" {
			return fmt.Errorf("LAIVAN_S3_BUCKET is required when media is enabled")
		}
		if c.Media.S3Region == "" {
			return fmt.Errorf("LAIVAN_S3_REGION is required when media is enabled")
		}
		if c.Media.S3AccessKeyID == "" {
			return fmt.Errorf("LAIVAN_S3_ACCESS_KEY_ID is required when media is enabled")
		}
		if c.Media.S3SecretKey == "" {
			return fmt.Errorf("LAIVAN_S3_SECRET_ACCESS_KEY is required when media is enabled")
		}
		if c.Media.PublicBaseURL == "" {
			return fmt.Errorf("LAIVAN_MEDIA_PUBLIC_BASE_URL is required when media is enabled")
		}
		if c.Media.Imgproxy.BaseURL == "" {
			return fmt.Errorf("LAIVAN_IMGPROXY_BASE_URL is required when media is enabled")
		}
		if c.Media.Imgproxy.Key == "" {
			return fmt.Errorf("LAIVAN_IMGPROXY_KEY is required when media is enabled")
		}
		if c.Media.Imgproxy.Salt == "" {
			return fmt.Errorf("LAIVAN_IMGPROXY_SALT is required when media is enabled")
		}
	}

	return nil
}

func (c Config) IsDevelopment() bool {
	return c.Env == EnvDevelopment
}

func (c Config) IsTest() bool {
	return c.Env == EnvTest
}

func (c Config) IsProduction() bool {
	return c.Env == EnvProduction
}

func stringEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}

	return strings.TrimSpace(value)
}

func stringsEnv(key string, fallback []string) []string {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

func intEnv(key string, fallback int) (int, error) {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", key)
	}

	return parsed, nil
}

func int32Env(key string, fallback int32) (int32, error) {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", key)
	}

	return int32(parsed), nil
}

func int64Env(key string, fallback int64) (int64, error) {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", key)
	}

	return parsed, nil
}

func boolEnv(key string, fallback bool) (bool, error) {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean", key)
	}

	return parsed, nil
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration", key)
	}

	return parsed, nil
}

func firstString(values []string) string {
	if len(values) == 0 {
		return ""
	}

	return values[0]
}

func validateOrigin(value string, requireHTTPS bool) error {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("must be a valid HTTP(S) origin")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("must use HTTP or HTTPS")
	}
	if requireHTTPS && parsed.Scheme != "https" {
		return fmt.Errorf("must use HTTPS outside local development")
	}
	if parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("must contain only scheme and host")
	}

	return nil
}

func validateRedirectURL(value string, requireHTTPS bool) error {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("must be a valid HTTP(S) redirect URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("must use HTTP or HTTPS")
	}
	if requireHTTPS && parsed.Scheme != "https" {
		return fmt.Errorf("must use HTTPS outside local development")
	}
	if parsed.User != nil || parsed.Fragment != "" {
		return fmt.Errorf("must not contain credentials or a fragment")
	}

	return nil
}
