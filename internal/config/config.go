package config

import (
	"fmt"
	"os"
	"time"
)

// Global configuration struct for the application
type Config struct {
	App      AppConfig
	Database DatabaseConfig

	JWT        JWTConfig
	Invitation InvitationConfig
}

type Env string

const (
	EnvDevelopment Env = "development"
	EnvStaging     Env = "staging"
	EnvProduction  Env = "production"
)

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

type AppConfig struct {
	Env             Env           // Environment (e.g., development, production)
	Port            string        // Port on which the application will run
	LogLevel        LogLevel      // Logging level (e.g., debug, info, warn, error)
	ShutdownTimeout time.Duration // Timeout for graceful shutdown
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	Secret string
}
type InvitationConfig struct {
	BaseURL string
	TTL     time.Duration
}

// NewConfig creates a new Config instance with the provided environment and logger.
func Load() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Env:             Env(getEnv("APP_ENV", "development")),
			Port:            getEnv("APP_PORT", "8080"),
			LogLevel:        LogLevel(getEnv("LOG_LEVEL", "debug")),
			ShutdownTimeout: getEnvDuration("APP_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5433"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "brewflow_db"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},

		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET_KEY", ""),
		},

		Invitation: InvitationConfig{
			BaseURL: getEnv("INVITATION_BASE_URL", "http://localhost:3000"),
			TTL:     getEnvDuration("INVITATION_TTL", 24*time.Hour),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	switch c.App.Env {
	case EnvDevelopment, EnvStaging, EnvProduction:
	default:
		return fmt.Errorf("invalid APP_ENV: %q", c.App.Env)
	}

	switch c.App.LogLevel {
	case LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError:
	default:
		return fmt.Errorf("invalid LOG_LEVEL: %q", c.App.LogLevel)
	}

	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	if c.Invitation.BaseURL == "" {
		return fmt.Errorf("INVITATION_BASE_URL is required")
	}

	if c.Invitation.TTL <= 0 {
		return fmt.Errorf("INVITATION_TTL must be greater than 0")
	}

	return nil
}

func getEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}

	return fallback
}
