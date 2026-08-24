package config

import (
	"slices"
	"strings"
	"testing"
	"time"
)

func TestLoadUsesDefaults(t *testing.T) {
	clearConfigEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Env != EnvDevelopment {
		t.Fatalf("Env = %q, want %q", cfg.Env, EnvDevelopment)
	}

	if cfg.Port != 4000 {
		t.Fatalf("Port = %d, want 4000", cfg.Port)
	}

	if cfg.DatabaseURL == "" {
		t.Fatal("DatabaseURL is empty")
	}
	if cfg.DBMaxConns != 25 {
		t.Fatalf("DBMaxConns = %d, want 25", cfg.DBMaxConns)
	}

	if !slices.Equal(cfg.AllowedOrigins, []string{"http://localhost:5173"}) {
		t.Fatalf("AllowedOrigins = %v, want [http://localhost:5173]", cfg.AllowedOrigins)
	}

	if cfg.ReadTimeout != 5*time.Second {
		t.Fatalf("ReadTimeout = %s, want 5s", cfg.ReadTimeout)
	}

	if cfg.WriteTimeout != 10*time.Second {
		t.Fatalf("WriteTimeout = %s, want 10s", cfg.WriteTimeout)
	}

	if cfg.IdleTimeout != time.Minute {
		t.Fatalf("IdleTimeout = %s, want 1m", cfg.IdleTimeout)
	}

	if cfg.ShutdownTimeout != 10*time.Second {
		t.Fatalf("ShutdownTimeout = %s, want 10s", cfg.ShutdownTimeout)
	}

	if cfg.Media.Enabled {
		t.Fatal("Media.Enabled = true, want false")
	}
	if cfg.Auth.GoogleAuthEnabled {
		t.Fatal("Auth.GoogleAuthEnabled = true, want false")
	}

	if cfg.Media.S3Region != "us-east-1" {
		t.Fatalf("Media.S3Region = %q, want us-east-1", cfg.Media.S3Region)
	}

	if cfg.Media.MaxUploadBytes != 10<<20 {
		t.Fatalf("Media.MaxUploadBytes = %d, want %d", cfg.Media.MaxUploadBytes, 10<<20)
	}
}

func TestLoadUsesEnvironmentOverrides(t *testing.T) {
	clearConfigEnv(t)

	t.Setenv("LAIVAN_ENV", EnvTest)
	t.Setenv("LAIVAN_PORT", "8080")
	t.Setenv("LAIVAN_DB_URL", "postgres://laivan:laivan@localhost:5432/laivan?sslmode=disable")
	t.Setenv("LAIVAN_DB_MAX_CONNS", "10")
	t.Setenv("LAIVAN_READ_TIMEOUT", "2s")
	t.Setenv("LAIVAN_WRITE_TIMEOUT", "3s")
	t.Setenv("LAIVAN_IDLE_TIMEOUT", "4s")
	t.Setenv("LAIVAN_SHUTDOWN_TIMEOUT", "5s")
	t.Setenv("LAIVAN_ALLOWED_ORIGINS", "http://localhost:5173, http://localhost:4173")
	t.Setenv("LAIVAN_GOOGLE_AUTH_ENABLED", "false")
	t.Setenv("LAIVAN_MEDIA_ENABLED", "true")
	t.Setenv("LAIVAN_S3_ENDPOINT", "http://localhost:9000")
	t.Setenv("LAIVAN_S3_BUCKET", "laivan-dev")
	t.Setenv("LAIVAN_S3_REGION", "auto")
	t.Setenv("LAIVAN_S3_ACCESS_KEY_ID", "access-key")
	t.Setenv("LAIVAN_S3_SECRET_ACCESS_KEY", "secret-key")
	t.Setenv("LAIVAN_S3_USE_SSL", "true")
	t.Setenv("LAIVAN_MEDIA_PUBLIC_BASE_URL", "https://media.example.test")
	t.Setenv("LAIVAN_MEDIA_MAX_UPLOAD_BYTES", "1024")
	t.Setenv("LAIVAN_IMGPROXY_BASE_URL", "https://images.example.test")
	t.Setenv("LAIVAN_IMGPROXY_SOURCE_BASE_URL", "https://media-internal.example.test")
	t.Setenv("LAIVAN_IMGPROXY_KEY", "abcd")
	t.Setenv("LAIVAN_IMGPROXY_SALT", "1234")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Env != EnvTest {
		t.Fatalf("Env = %q, want %q", cfg.Env, EnvTest)
	}

	if cfg.Port != 8080 {
		t.Fatalf("Port = %d, want 8080", cfg.Port)
	}

	if cfg.DatabaseURL == "" {
		t.Fatal("DatabaseURL is empty")
	}
	if cfg.DBMaxConns != 10 {
		t.Fatalf("DBMaxConns = %d, want 10", cfg.DBMaxConns)
	}

	wantOrigins := []string{"http://localhost:5173", "http://localhost:4173"}
	if !slices.Equal(cfg.AllowedOrigins, wantOrigins) {
		t.Fatalf("AllowedOrigins = %v, want %v", cfg.AllowedOrigins, wantOrigins)
	}

	if cfg.ReadTimeout != 2*time.Second {
		t.Fatalf("ReadTimeout = %s, want 2s", cfg.ReadTimeout)
	}

	if cfg.WriteTimeout != 3*time.Second {
		t.Fatalf("WriteTimeout = %s, want 3s", cfg.WriteTimeout)
	}

	if cfg.IdleTimeout != 4*time.Second {
		t.Fatalf("IdleTimeout = %s, want 4s", cfg.IdleTimeout)
	}

	if cfg.ShutdownTimeout != 5*time.Second {
		t.Fatalf("ShutdownTimeout = %s, want 5s", cfg.ShutdownTimeout)
	}

	if !cfg.Media.Enabled {
		t.Fatal("Media.Enabled = false, want true")
	}
	if cfg.Media.S3Endpoint != "http://localhost:9000" {
		t.Fatalf("Media.S3Endpoint = %q, want http://localhost:9000", cfg.Media.S3Endpoint)
	}
	if cfg.Media.S3Bucket != "laivan-dev" {
		t.Fatalf("Media.S3Bucket = %q, want laivan-dev", cfg.Media.S3Bucket)
	}
	if !cfg.Media.S3UseSSL {
		t.Fatal("Media.S3UseSSL = false, want true")
	}
	if cfg.Media.MaxUploadBytes != 1024 {
		t.Fatalf("Media.MaxUploadBytes = %d, want 1024", cfg.Media.MaxUploadBytes)
	}
	if cfg.Media.Imgproxy.BaseURL != "https://images.example.test" {
		t.Fatalf("Media.Imgproxy.BaseURL = %q, want https://images.example.test", cfg.Media.Imgproxy.BaseURL)
	}
	if cfg.Media.Imgproxy.SourceBaseURL != "https://media-internal.example.test" {
		t.Fatalf("Media.Imgproxy.SourceBaseURL = %q, want https://media-internal.example.test", cfg.Media.Imgproxy.SourceBaseURL)
	}
}

func TestLoadRejectsEnabledMediaWithoutRequiredConfig(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("LAIVAN_MEDIA_ENABLED", "true")

	_, err := Load()
	if err == nil {
		t.Fatal("Load returned nil error")
	}
}

func TestLoadRejectsInvalidMediaBooleans(t *testing.T) {
	tests := []string{"LAIVAN_GOOGLE_AUTH_ENABLED", "LAIVAN_MEDIA_ENABLED", "LAIVAN_S3_USE_SSL"}

	for _, key := range tests {
		t.Run(key, func(t *testing.T) {
			clearConfigEnv(t)
			t.Setenv(key, "maybe")

			_, err := Load()
			if err == nil {
				t.Fatal("Load returned nil error")
			}
		})
	}
}

func TestLoadRejectsInvalidMediaUploadLimit(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("LAIVAN_MEDIA_MAX_UPLOAD_BYTES", "0")

	_, err := Load()
	if err == nil {
		t.Fatal("Load returned nil error")
	}
}

func TestConfigEnvironmentHelpers(t *testing.T) {
	tests := []struct {
		name            string
		cfg             Config
		wantDevelopment bool
		wantTest        bool
		wantProduction  bool
	}{
		{name: "development", cfg: Config{Env: EnvDevelopment}, wantDevelopment: true},
		{name: "test", cfg: Config{Env: EnvTest}, wantTest: true},
		{name: "production", cfg: Config{Env: EnvProduction}, wantProduction: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cfg.IsDevelopment() != tt.wantDevelopment {
				t.Fatalf("IsDevelopment = %t, want %t", tt.cfg.IsDevelopment(), tt.wantDevelopment)
			}

			if tt.cfg.IsTest() != tt.wantTest {
				t.Fatalf("IsTest = %t, want %t", tt.cfg.IsTest(), tt.wantTest)
			}

			if tt.cfg.IsProduction() != tt.wantProduction {
				t.Fatalf("IsProduction = %t, want %t", tt.cfg.IsProduction(), tt.wantProduction)
			}
		})
	}
}

func TestLoadRequiresDatabaseURLInAllEnvironments(t *testing.T) {
	for _, environment := range []string{EnvDevelopment, EnvTest, EnvProduction} {
		t.Run(environment, func(t *testing.T) {
			clearConfigEnv(t)
			t.Setenv("LAIVAN_ENV", environment)
			t.Setenv("LAIVAN_DB_URL", "")

			_, err := Load()
			if err == nil {
				t.Fatal("Load returned nil error")
			}
		})
	}
}

func TestLoadAllowsProductionWithDatabaseURL(t *testing.T) {
	clearConfigEnv(t)

	t.Setenv("LAIVAN_ENV", EnvProduction)
	t.Setenv("LAIVAN_DB_URL", "postgres://laivan:laivan@localhost:5432/laivan?sslmode=disable")
	t.Setenv("LAIVAN_GOOGLE_AUTH_ENABLED", "true")
	t.Setenv("LAIVAN_GOOGLE_CLIENT_ID", "client-id")
	t.Setenv("LAIVAN_GOOGLE_CLIENT_SECRET", "client-secret")
	t.Setenv("LAIVAN_GOOGLE_REDIRECT_URL", "https://api.example.test/v1/auth/google/callback")
	t.Setenv("LAIVAN_WEB_ORIGIN", "https://app.example.test")
	t.Setenv("LAIVAN_OIDC_STATE_SIGNING_KEY", strings.Repeat("a", 32))
	t.Setenv("LAIVAN_SECURE_COOKIES", "true")

	_, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
}

func TestLoadRejectsGoogleValuesWhenAuthIsDisabled(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("LAIVAN_GOOGLE_CLIENT_ID", "client-id")

	_, err := Load()
	if err == nil {
		t.Fatal("Load returned nil error")
	}
}

func TestLoadIgnoresEmptyAllowedOriginParts(t *testing.T) {
	clearConfigEnv(t)

	t.Setenv("LAIVAN_ALLOWED_ORIGINS", " http://localhost:5173, , http://localhost:4173 ")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	wantOrigins := []string{"http://localhost:5173", "http://localhost:4173"}
	if !slices.Equal(cfg.AllowedOrigins, wantOrigins) {
		t.Fatalf("AllowedOrigins = %v, want %v", cfg.AllowedOrigins, wantOrigins)
	}
}

func TestLoadRejectsEmptyAllowedOrigins(t *testing.T) {
	clearConfigEnv(t)

	t.Setenv("LAIVAN_ALLOWED_ORIGINS", ",,")

	_, err := Load()
	if err == nil {
		t.Fatal("Load returned nil error")
	}
}

func TestLoadRejectsInvalidEnvironment(t *testing.T) {
	clearConfigEnv(t)

	t.Setenv("LAIVAN_ENV", "local")

	_, err := Load()
	if err == nil {
		t.Fatal("Load returned nil error")
	}
}

func TestLoadRejectsInvalidPort(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{name: "not integer", value: "abc"},
		{name: "zero", value: "0"},
		{name: "too high", value: "65536"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearConfigEnv(t)

			t.Setenv("LAIVAN_PORT", tt.value)

			_, err := Load()
			if err == nil {
				t.Fatal("Load returned nil error")
			}
		})
	}
}

func TestLoadRejectsInvalidDatabaseMaxConnections(t *testing.T) {
	for _, value := range []string{"invalid", "0", "-1"} {
		t.Run(value, func(t *testing.T) {
			clearConfigEnv(t)
			t.Setenv("LAIVAN_DB_MAX_CONNS", value)

			_, err := Load()
			if err == nil {
				t.Fatal("Load returned nil error")
			}
		})
	}
}

func TestLoadRejectsInvalidTimeouts(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{name: "read timeout", key: "LAIVAN_READ_TIMEOUT"},
		{name: "write timeout", key: "LAIVAN_WRITE_TIMEOUT"},
		{name: "idle timeout", key: "LAIVAN_IDLE_TIMEOUT"},
		{name: "shutdown timeout", key: "LAIVAN_SHUTDOWN_TIMEOUT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearConfigEnv(t)

			t.Setenv(tt.key, "0s")

			_, err := Load()
			if err == nil {
				t.Fatal("Load returned nil error")
			}
		})
	}
}

func TestLoadRejectsMalformedDuration(t *testing.T) {
	clearConfigEnv(t)

	t.Setenv("LAIVAN_READ_TIMEOUT", "slow")

	_, err := Load()
	if err == nil {
		t.Fatal("Load returned nil error")
	}
}

func clearConfigEnv(t *testing.T) {
	t.Helper()

	t.Setenv("LAIVAN_ENV", "")
	t.Setenv("LAIVAN_PORT", "")
	t.Setenv("LAIVAN_DB_URL", "postgres://laivan:laivan@localhost:5432/laivan?sslmode=disable")
	t.Setenv("LAIVAN_DB_MAX_CONNS", "")
	t.Setenv("LAIVAN_READ_TIMEOUT", "")
	t.Setenv("LAIVAN_WRITE_TIMEOUT", "")
	t.Setenv("LAIVAN_IDLE_TIMEOUT", "")
	t.Setenv("LAIVAN_SHUTDOWN_TIMEOUT", "")
	t.Setenv("LAIVAN_ALLOWED_ORIGINS", "")
	t.Setenv("LAIVAN_GOOGLE_AUTH_ENABLED", "")
	t.Setenv("LAIVAN_GOOGLE_CLIENT_ID", "")
	t.Setenv("LAIVAN_GOOGLE_CLIENT_SECRET", "")
	t.Setenv("LAIVAN_GOOGLE_REDIRECT_URL", "")
	t.Setenv("LAIVAN_WEB_ORIGIN", "")
	t.Setenv("LAIVAN_SESSION_COOKIE_NAME", "")
	t.Setenv("LAIVAN_CSRF_COOKIE_NAME", "")
	t.Setenv("LAIVAN_OIDC_STATE_SIGNING_KEY", "")
	t.Setenv("LAIVAN_OIDC_STATE_DURATION", "")
	t.Setenv("LAIVAN_SESSION_DURATION", "")
	t.Setenv("LAIVAN_SECURE_COOKIES", "")
	t.Setenv("LAIVAN_MEDIA_ENABLED", "")
	t.Setenv("LAIVAN_S3_ENDPOINT", "")
	t.Setenv("LAIVAN_S3_BUCKET", "")
	t.Setenv("LAIVAN_S3_REGION", "")
	t.Setenv("LAIVAN_S3_ACCESS_KEY_ID", "")
	t.Setenv("LAIVAN_S3_SECRET_ACCESS_KEY", "")
	t.Setenv("LAIVAN_S3_USE_SSL", "")
	t.Setenv("LAIVAN_MEDIA_PUBLIC_BASE_URL", "")
	t.Setenv("LAIVAN_MEDIA_MAX_UPLOAD_BYTES", "")
	t.Setenv("LAIVAN_IMGPROXY_BASE_URL", "")
	t.Setenv("LAIVAN_IMGPROXY_SOURCE_BASE_URL", "")
	t.Setenv("LAIVAN_IMGPROXY_KEY", "")
	t.Setenv("LAIVAN_IMGPROXY_SALT", "")
}
