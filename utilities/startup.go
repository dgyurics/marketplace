// startup utilities provides functions to load environment and configuration data
// critical for starting the server. It ensures required resources are available,
// otherwise exits the program.

package utilities

import (
	"log/slog"
	"os"
	"time"

	"github.com/dgyurics/marketplace/models"
)

// wrapper for os.ReadFile
// Exits the program with os.Exit(1) if an error occurs while reading the file.
func GetKey(filename string) []byte {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		slog.Error("Error reading key file", "filename", filename, "error", err)
		os.Exit(1)
	}
	return bytes
}

// wrapper for os.LookupEnv
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

// loads configuration necessary for the auth service
func LoadAuthConfig() models.AuthConfig {
	return models.AuthConfig{
		PrivateKey:           GetKey("private.pem"),
		PublicKey:            GetKey("public.pem"),
		HMACSecret:           []byte(GetEnv("HMAC_SECRET")),
		DurationAccessToken:  parseDuration("DURATION_ACCESS_TOKEN"),
		DurationRefreshToken: parseDuration("DURATION_REFRESH_TOKEN"),
	}
}

// loads configuration necessary for the payment service
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
