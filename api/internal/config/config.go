package config

import (
	"fmt"
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
	AllowedOrigins  []string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Env:             EnvDevelopment,
		Port:            4000,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     time.Minute,
		ShutdownTimeout: 10 * time.Second,
	}

	var err error

	cfg.Env = stringEnv("LAIVAN_ENV", cfg.Env)
	cfg.DatabaseURL = stringEnv("LAIVAN_DB_URL", cfg.DatabaseURL)
	cfg.AllowedOrigins = stringsEnv("LAIVAN_ALLOWED_ORIGINS", []string{"http://localhost:5173"})

	cfg.Port, err = intEnv("LAIVAN_PORT", cfg.Port)
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

	if c.IsProduction() && c.DatabaseURL == "" {
		return fmt.Errorf("LAIVAN_DB_URL is required in production")
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
