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

	"github.com/dgyurics/marketplace/types"
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
func ParseDuration(key string) time.Duration {
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
func LoadDBConfig() types.DBConfig {
	return types.DBConfig{
		URL:             GetEnv("DATABASE_URL"),
		MaxOpenConns:    parseInt("DATABASE_MAX_CONNECTIONS"),
		MaxIdleConns:    parseInt("DATABASE_MAX_IDLE_CONNECTIONS"),
		ConnMaxLifetime: ParseDuration("DATABASE_CONNECTION_MAX_LIFETIME"),
		ConnMaxIdleTime: ParseDuration("DATABASE_CONNECTION_MAX_IDLE_TIME"),
	}
}

// LoadJWTConfig loads configuration necessary for the JWT service
func LoadJWTConfig() types.JWTConfig {
	return types.JWTConfig{
		PrivateKey: GetKey("private.pem"),
		PublicKey:  GetKey("public.pem"),
		Expiry:     ParseDuration("JWT_EXPIRY"),
	}
}

// LoadOrderConfig loads configuration necessary for the payment service
func LoadOrderConfig() types.OrderConfig {
	env := GetEnv("ENVIRONMENT")
	var environment types.Environment
	switch env {
	case string(types.Development):
		environment = types.Development
	case string(types.Production):
		environment = types.Production
	default:
		slog.Warn("Invalid environment", "env", env)
		environment = types.Development
	}
	return types.OrderConfig{
		Envirnment:                 environment,
		StripeBaseURL:              GetEnv("STRIPE_BASE_URL"),
		StripeSecretKey:            GetEnv("STRIPE_SECRET_KEY"),
		StripeWebhookSigningSecret: GetEnv("STRIPE_WEBHOOK_SIGNING_SECRET"),
	}
}

func LoadMailjetConfig() types.MailjetConfig {
	return types.MailjetConfig{
		Enabled:   IsFeatureEnabled("MAILJET_ENABLED"),
		APIKey:    GetEnv("MAILJET_API_KEY"),
		APISecret: GetEnv("MAILJET_API_SECRET"),
		FromEmail: GetEnv("MAILJET_FROM_EMAIL"),
		FromName:  GetEnv("MAILJET_FROM_NAME"),
	}
}

// IsFeatureEnabled checks if a feature flag is enabled via environment variables (case-insensitive).
func IsFeatureEnabled(feature string) bool {
	return strings.EqualFold(os.Getenv(feature), "true")
}
