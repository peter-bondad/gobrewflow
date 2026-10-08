package config

import (
	"fmt"
	"os"
	"time"
)

type PayMongoConfig struct {
	BaseURL       string
	SecretKey     string
	WebhookSecret string
	SuccessURL    string
	CancelURL     string
	RetryConfig   RetryConfig
}

type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

// Global configuration struct for the application
type Config struct {
	App      AppConfig
	Database DatabaseConfig

	JWT        JWTConfig
	Invitation InvitationConfig
	PayMongo   PayMongoConfig
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
			Port:            getEnv("PORT", "8080"),
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

		PayMongo: PayMongoConfig{
			BaseURL:       getEnv("PAYMONGO_BASE_URL", "https://api.paymongo.com"),
			SecretKey:     getEnv("PAYMONGO_TEST_SECRET_KEY", ""),
			WebhookSecret: getEnv("PAYMONGO_WEBHOOK_SECRET", ""),
			SuccessURL:    getEnv("PAYMONGO_SUCCESS_URL", "http://localhost:3000/payment/success"),
			CancelURL:     getEnv("PAYMONGO_CANCEL_URL", "http://localhost:3000/payment/cancel"),
			RetryConfig: RetryConfig{
				MaxAttempts: getEnvInt("PAYMONGO_MAX_ATTEMPTS", 3),
				BaseDelay:   getEnvDuration("PAYMONGO_RETRY_BASE_DELAY", 500*time.Millisecond),
				MaxDelay:    getEnvDuration("PAYMONGO_RETRY_MAX_DELAY", 2*time.Second),
			},
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

	if c.PayMongo.BaseURL == "" {
		return fmt.Errorf("PAYMONGO_BASE_URL is required")
	}

	if c.PayMongo.SecretKey == "" {
		return fmt.Errorf("PAYMONGO_TEST_SECRET_KEY is required")
	}

	if c.PayMongo.WebhookSecret == "" {
		return fmt.Errorf("PAYMONGO_WEBHOOK_SECRET is required")
	}

	if c.PayMongo.SuccessURL == "" {
		return fmt.Errorf("PAYMONGO_SUCCESS_URL is required")
	}

	if c.PayMongo.CancelURL == "" {
		return fmt.Errorf("PAYMONGO_CANCEL_URL is required")
	}

	if c.PayMongo.RetryConfig.MaxAttempts <= 0 {
		return fmt.Errorf("PAYMONGO_MAX_ATTEMPTS must be greater than 0")
	}

	if c.PayMongo.RetryConfig.BaseDelay <= 0 {
		return fmt.Errorf("PAYMONGO_RETRY_BASE_DELAY must be greater than 0")
	}

	if c.PayMongo.RetryConfig.MaxDelay <= 0 {
		return fmt.Errorf("PAYMONGO_RETRY_MAX_DELAY must be greater than 0")
	}

	if c.PayMongo.RetryConfig.BaseDelay > c.PayMongo.RetryConfig.MaxDelay {
		return fmt.Errorf("PAYMONGO_RETRY_BASE_DELAY cannot be greater than PAYMONGO_RETRY_MAX_DELAY")
	}

	return nil
}

func getEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		var result int

		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
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
