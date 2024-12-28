// Package utilities environment provides functions to load environment and configuration data
// critical for running starting the server. It ensures required resources are available,
// otherwise exits the program.

package utilities

import (
	"log"
	"os"
	"time"

	"github.com/dgyurics/marketplace/models"
)

// wrapper for os.ReadFile
// calls [os.Exit](1) if error occurs reading file
func GetKey(filename string) []byte {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading %s key: %v", filename, err)
	}
	return bytes
}

// wrapper for os.LookupEnv
// calls [os.Exit](1) if error occurs fetching environment variable
func GetEnv(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	log.Fatalf("%s is required", key)
	return ""
}

// wrapper for time.ParseDuration
// calls [os.Exit](1) if error occurs parsing duration
func parseDuration(key string) time.Duration {
	duration, err := time.ParseDuration(GetEnv(key))
	if err != nil {
		log.Fatalf("Invalid duration: %v", err)
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
	return models.OrderConfig{
		Envirnment:                 models.Environment(GetEnv("ENVIRONMENT")),
		StripeBaseURL:              GetEnv("STRIPE_BASE_URL"),
		StripeSecretKey:            GetEnv("STRIPE_SECRET_KEY"),
		StripeWebhookSigningSecret: GetEnv("STRIPE_WEBHOOK_SIGNING_SECRET"),
	}
}
