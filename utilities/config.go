package utilities

import (
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgyurics/marketplace/types"
)

// LoadConfig loads the configuration from environment variables
func LoadConfig() types.Config {
	return types.Config{
		Server:       loadServerConfig(),
		Environment:  loadEnvironment(),
		CORS:         loadCORSConfig(),
		Auth:         loadAuthConfig(),
		Database:     loadDBConfig(),
		Email:        loadMailConfig(),
		Logger:       LoadLoggerConfig(),
		Order:        loadOrderConfig(),
		JWT:          loadJWTConfig(),
		TemplatesDir: loadTemplatesDir(),
		BaseURL:      loadBaseURL(),
	}
}

func loadCORSConfig() types.CORSConfig {
	return types.CORSConfig{
		AllowedOrigins:   strings.Split(getEnv("CORS_ALLOWED_ORIGINS"), ","),
		AllowedMethods:   strings.Split(getEnv("CORS_ALLOWED_METHODS"), ","),
		AllowedHeaders:   strings.Split(getEnv("CORS_ALLOWED_HEADERS"), ","),
		AllowCredentials: isFeatureEnabled("CORS_ALLOW_CREDENTIALS"),
	}
}

func loadServerConfig() types.ServerConfig {
	return types.ServerConfig{
		Addr:           getEnv("SERVER_ADDR"),
		ReadTimeout:    parseDuration("SERVER_READ_TIMEOUT"),
		WriteTimeout:   parseDuration("SERVER_WRITE_TIMEOUT"),
		IdleTimeout:    parseDuration("SERVER_IDLE_TIMEOUT"),
		MaxHeaderBytes: parseInt("SERVER_MAX_HEADER_BYTES"),
		ErrorLog:       log.New(&ErrorLog{}, "", 0),
	}
}

func loadAuthConfig() types.AuthConfig {
	return types.AuthConfig{
		HMACSecret:    []byte(getEnv("HMAC_SECRET")),
		RefreshExpiry: parseDuration("REFRESH_EXPIRY"),
		InviteReq:     isFeatureEnabled("INVITE_REQUIRED"),
	}
}

func loadDBConfig() types.DBConfig {
	return types.DBConfig{
		URL:             getEnv("DATABASE_URL"),
		MaxOpenConns:    parseInt("DATABASE_MAX_CONNECTIONS"),
		MaxIdleConns:    parseInt("DATABASE_MAX_IDLE_CONNECTIONS"),
		ConnMaxLifetime: parseDuration("DATABASE_CONNECTION_MAX_LIFETIME"),
		ConnMaxIdleTime: parseDuration("DATABASE_CONNECTION_MAX_IDLE_TIME"),
	}
}

func loadJWTConfig() types.JWTConfig {
	return types.JWTConfig{
		PrivateKey: getKey("private.pem"),
		PublicKey:  getKey("public.pem"),
		Expiry:     parseDuration("JWT_EXPIRY"),
	}
}

func loadOrderConfig() types.OrderConfig {
	return types.OrderConfig{
		DefaultTaxCode:     getEnv("ORDER_DEFAULT_TAX_CODE"),
		DefaultTaxBehavior: getEnv("ORDER_DEFAULT_TAX_BEHAVIOR"),
		StripeConfig:       loadStripeConfig(),
	}
}

func loadStripeConfig() types.StripeConfig {
	return types.StripeConfig{
		BaseURL:              getEnv("STRIPE_BASE_URL"),
		SecretKey:            getEnv("STRIPE_SECRET_KEY"),
		WebhookSigningSecret: getEnv("STRIPE_WEBHOOK_SIGNING_SECRET"),
	}
}

func loadEnvironment() types.Environment {
	env := getEnv("ENVIRONMENT")
	switch env {
	case string(types.Development):
		return types.Development
	case string(types.Production):
		return types.Production
	default:
		return types.Development
	}
}

func loadTemplatesDir() string {
	return "./utilities/templates"
}

func loadBaseURL() string {
	return getEnv("BASE_URL")
}

func loadMailConfig() types.EmailConfig {
	return types.EmailConfig{
		Enabled:   isFeatureEnabled("MAIL_ENABLED"),
		APIKey:    getEnv("MAIL_API_KEY"),
		APISecret: getEnv("MAIL_API_SECRET"),
		FromEmail: getEnv("MAIL_FROM_EMAIL"),
		FromName:  getEnv("MAIL_FROM_NAME"),
	}
}

// getKey reads a key file and exits if an error occurs.
func getKey(filename string) []byte {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		slog.Error("Error reading key file", "filename", filename, "error", err)
		os.Exit(1)
	}
	return bytes
}

// getEnv retrieves an environment variable and exits if an error occurs.
func getEnv(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	slog.Error("Environment variable is required", "key", key)
	os.Exit(1)
	return ""
}

// parseDuration parses a duration from the environment variable and exits if an error occurs.
func parseDuration(key string) time.Duration {
	duration, err := time.ParseDuration(getEnv(key))
	if err != nil {
		slog.Error("Invalid duration", "key", key, "error", err)
		os.Exit(1)
	}
	return duration
}

// parseInt parses an integer from the environment variable and exits if an error occurs.
func parseInt(key string) int {
	value, err := strconv.Atoi(getEnv(key))
	if err != nil {
		slog.Error("Invalid integer", "key", key, "error", err)
		os.Exit(1)
	}
	return value
}

// isFeatureEnabled returns whether a feature is enabled (case-insensitive).
func isFeatureEnabled(feature string) bool {
	return strings.EqualFold(os.Getenv(feature), "true")
}
