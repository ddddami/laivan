package config

import (
	"slices"
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

	if cfg.DatabaseURL != "" {
		t.Fatalf("DatabaseURL = %q, want empty", cfg.DatabaseURL)
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
}

func TestLoadUsesEnvironmentOverrides(t *testing.T) {
	clearConfigEnv(t)

	t.Setenv("LAIVAN_ENV", EnvTest)
	t.Setenv("LAIVAN_PORT", "8080")
	t.Setenv("LAIVAN_DB_URL", "postgres://laivan:laivan@localhost:5432/laivan?sslmode=disable")
	t.Setenv("LAIVAN_READ_TIMEOUT", "2s")
	t.Setenv("LAIVAN_WRITE_TIMEOUT", "3s")
	t.Setenv("LAIVAN_IDLE_TIMEOUT", "4s")
	t.Setenv("LAIVAN_SHUTDOWN_TIMEOUT", "5s")
	t.Setenv("LAIVAN_ALLOWED_ORIGINS", "http://localhost:5173, http://localhost:4173")

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

func TestLoadRequiresDatabaseURLInProduction(t *testing.T) {
	clearConfigEnv(t)

	t.Setenv("LAIVAN_ENV", EnvProduction)

	_, err := Load()
	if err == nil {
		t.Fatal("Load returned nil error")
	}
}

func TestLoadAllowsProductionWithDatabaseURL(t *testing.T) {
	clearConfigEnv(t)

	t.Setenv("LAIVAN_ENV", EnvProduction)
	t.Setenv("LAIVAN_DB_URL", "postgres://laivan:laivan@localhost:5432/laivan?sslmode=disable")

	_, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
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
	t.Setenv("LAIVAN_DB_URL", "")
	t.Setenv("LAIVAN_READ_TIMEOUT", "")
	t.Setenv("LAIVAN_WRITE_TIMEOUT", "")
	t.Setenv("LAIVAN_IDLE_TIMEOUT", "")
	t.Setenv("LAIVAN_SHUTDOWN_TIMEOUT", "")
	t.Setenv("LAIVAN_ALLOWED_ORIGINS", "")
}
