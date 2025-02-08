// startup utilities provides functions to load environment and configuration data
// critical for starting the server. It ensures required resources are available,
// otherwise exiting the program.

package utilities

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgyurics/marketplace/models"
)

// GetKey is a wrapper for os.ReadFile
// Exits the program with os.Exit(1) if an error occurs while reading the file.
func GetKey(filename string) []byte {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		slog.Error("Error reading key file", "filename", filename, "error", err)
		os.Exit(1)
	}
	return bytes
}

// GetEnv is a wrapper for os.LookupEnv
// Exits the program with os.Exit(1) if the required environment variable is not set.
func GetEnv(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	slog.Error("Environment variable is required", "key", key)
	os.Exit(1)
	return ""
}

// wrapper for time.ParseDuration
// Exits the program with os.Exit(1) if an error occurs while parsing the duration.
func parseDuration(key string) time.Duration {
	duration, err := time.ParseDuration(GetEnv(key))
	if err != nil {
		slog.Error("Invalid duration", "key", key, "error", err)
		os.Exit(1)
	}
	return duration
}

// parseInt parses an integer from the environment variable
// Exits the program with os.Exit(1) if an error occurs while parsing the integer.
func parseInt(key string) int {
	value, err := strconv.Atoi(GetEnv(key))
	if err != nil {
		slog.Error("Invalid integer", "key", key, "error", err)
		os.Exit(1)
	}
	return value
}

// LoadDBConfig loads configuration necessary for the database connection
func LoadDBConfig() models.DBConfig {
	return models.DBConfig{
		URL:             GetEnv("DATABASE_URL"),
		MaxOpenConns:    parseInt("DATABASE_MAX_CONNECTIONS"),
		MaxIdleConns:    parseInt("DATABASE_MAX_IDLE_CONNECTIONS"),
		ConnMaxLifetime: parseDuration("DATABASE_CONNECTION_MAX_LIFETIME"),
		ConnMaxIdleTime: parseDuration("DATABASE_CONNECTION_MAX_IDLE_TIME"),
	}
}

// LoadAuthConfig loads configuration necessary for the auth service
func LoadAuthConfig() models.AuthConfig {
	return models.AuthConfig{
		PrivateKey:           GetKey("private.pem"),
		PublicKey:            GetKey("public.pem"),
		HMACSecret:           []byte(GetEnv("HMAC_SECRET")),
		DurationAccessToken:  parseDuration("DURATION_ACCESS_TOKEN"),
		DurationRefreshToken: parseDuration("DURATION_REFRESH_TOKEN"),
	}
}

// LoadOrderConfig loads configuration necessary for the payment service
func LoadOrderConfig() models.OrderConfig {
	env := GetEnv("ENVIRONMENT")
	var environment models.Environment
	switch env {
	case string(models.Development):
		environment = models.Development
	case string(models.Production):
		environment = models.Production
	default:
		slog.Warn("Invalid environment", "env", env)
		environment = models.Development
	}
	return models.OrderConfig{
		Envirnment:                 environment,
		StripeBaseURL:              GetEnv("STRIPE_BASE_URL"),
		StripeSecretKey:            GetEnv("STRIPE_SECRET_KEY"),
		StripeWebhookSigningSecret: GetEnv("STRIPE_WEBHOOK_SIGNING_SECRET"),
	}
}

// IsFeatureEnabled checks if a feature flag is enabled via environment variables (case-insensitive).
func IsFeatureEnabled(feature string) bool {
	return strings.EqualFold(os.Getenv(feature), "true")
}
