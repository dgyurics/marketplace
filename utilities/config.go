package utilities

import (
	"encoding/hex"
	"fmt"
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
	environment := loadEnvironment()
	payment := loadPaymentConfig(environment)
	baseURL := loadBaseURL()

	// resolve the active locale before anything reads utilities.Locale
	loadLocale()

	return types.Config{
		AppMetadata:       loadAppMetadata(payment),
		BaseURL:           baseURL,
		Environment:       environment,
		Server:            loadServerConfig(),
		Auth:              loadAuthConfig(),
		Database:          loadDBConfig(),
		Email:             loadEmailConfig(),
		Logger:            loadLoggerConfig(),
		MachineID:         loadMachineID(),
		Payment:           payment,
		JWT:               loadJWTConfig(),
		HTTPClientTimeout: loadHttpClientTimeout(),
		Image:             loadImageConfig(baseURL),
		RateLimit:         loadRateLimit(),
	}
}

// loadAppMetadata builds the metadata exposed to clients via GET /config.
func loadAppMetadata(payment types.PaymentConfig) types.AppMetadata {
	return types.AppMetadata{
		Locale: Locale,
		PaymentOptions: types.PaymentOptions{
			Stripe:        payment.Stripe.Enabled,
			PayOnDelivery: payment.OnDelivery,
		},
	}
}

func loadServerConfig() types.ServerConfig {
	return types.ServerConfig{
		Addr:           mustLookupEnv("SERVER_ADDR"),
		ReadTimeout:    mustParseDuration("SERVER_READ_TIMEOUT"),
		WriteTimeout:   mustParseDuration("SERVER_WRITE_TIMEOUT"),
		IdleTimeout:    mustParseDuration("SERVER_IDLE_TIMEOUT"),
		MaxHeaderBytes: mustAtoI("SERVER_MAX_HEADER_BYTES"),
		ErrorLog:       log.New(&ErrorLog{}, "", 0),
	}
}

func loadAuthConfig() types.AuthConfig {
	return types.AuthConfig{
		HMACSecret:    []byte(mustLookupEnv("HMAC_SECRET")),
		RefreshExpiry: mustParseDuration("REFRESH_EXPIRY"),
	}
}

func loadDBConfig() types.DBConfig {
	return types.DBConfig{
		Host:            mustLookupEnv("POSTGRES_HOST"),
		Port:            mustAtoI("POSTGRES_PORT"),
		User:            mustLookupEnv("POSTGRES_USER"),
		Password:        mustLookupEnv("POSTGRES_PASSWORD"),
		Name:            mustLookupEnv("POSTGRES_DB"),
		SSLMode:         mustLookupEnv("POSTGRES_SSLMODE"),
		MaxOpenConns:    mustAtoI("POSTGRES_MAX_CONNECTIONS"),
		MaxIdleConns:    mustAtoI("POSTGRES_MAX_IDLE_CONNECTIONS"),
		ConnMaxLifetime: mustParseDuration("POSTGRES_CONNECTION_MAX_LIFETIME"),
		ConnMaxIdleTime: mustParseDuration("POSTGRES_CONNECTION_MAX_IDLE_TIME"),
	}
}

func loadTaxConfig() types.TaxConfig {
	config := types.TaxConfig{
		Behavior:     types.TaxExclusive,
		FallbackCode: mustLookupEnv("TAX_FALLBACK_CODE"),
	}

	if mustLookupEnv("TAX_BEHAVIOR") == string(types.TaxInclusive) {
		config.Behavior = types.TaxInclusive
	}

	return config
}

func loadJWTConfig() types.JWTConfig {
	privKeyPath := getEnvOrDefault("PRIVATE_KEY_PATH", "private.pem")
	pubKeyPath := getEnvOrDefault("PUBLIC_KEY_PATH", "public.pem")

	privateKey := mustReadFile(privKeyPath)
	publicKey := mustReadFile(pubKeyPath)

	// validate files aren't empty
	if len(privateKey) == 0 {
		slog.Error("Private key file empty", "file path", privKeyPath)
		os.Exit(1)
	}
	if len(publicKey) == 0 {
		slog.Error("Public key file empty", "file path", privKeyPath)
		os.Exit(1)
	}

	return types.JWTConfig{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Expiry:     mustParseDuration("JWT_EXPIRY"),
	}
}

func loadMachineID() uint8 {
	envVar := mustLookupEnv("MACHINE_ID")
	key, err := strconv.ParseUint(envVar, 10, 8)

	if err != nil {
		slog.Error("Error parsing unsigned integer", "key", key, "error", err)
		os.Exit(1)
	}
	return uint8(key)
}

func loadImageConfig(baseURL string) types.ImageConfig {
	keyHex := mustLookupEnv("IMGPROXY_KEY")
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		slog.Error("Error decoding IMGPROXY_KEY", "error", err)
		os.Exit(1)
	}

	saltHex := mustLookupEnv("IMGPROXY_SALT")
	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		slog.Error("Error decoding IMGPROXY_SALT", "error", err)
		os.Exit(1)
	}

	urlImgproxy := fmt.Sprintf("%s/images", baseURL)

	return types.ImageConfig{
		Key:              key,
		Salt:             salt,
		BaseURLImgproxy:  urlImgproxy,
		BaseURLRembg:     getEnvOrDefault("REMBG_BASE_URL", "http://rembg"),
		ImageUploadPath:  getEnvOrDefault("IMG_DIR_OVERRIDE", "images"), // need to override when doing local development
		MaxMegapixels:    mustAtoI("IMGPROXY_MAX_SRC_RESOLUTION"),
		MaxFileSizeBytes: mustAtoI("IMGPROXY_MAX_SRC_FILE_SIZE"),
	}
}

func loadPaymentConfig(env types.Environment) types.PaymentConfig {
	return types.PaymentConfig{
		Stripe:      loadStripeConfig(),
		OnDelivery:  isFeatureEnabled("PAYMENT_ON_DELIVERY_ENABLED"),
		Tax:         loadTaxConfig(),
		Environment: env,
	}
}

func loadStripeConfig() types.StripeConfig {
	if !isFeatureEnabled("STRIPE_ENABLED") {
		return types.StripeConfig{Enabled: false}
	}
	return types.StripeConfig{
		Enabled:              true,
		BaseURL:              mustLookupEnv("STRIPE_BASE_URL"),
		SecretKey:            mustLookupEnv("STRIPE_SECRET_KEY"),
		WebhookSigningSecret: mustLookupEnv("STRIPE_WEBHOOK_SIGNING_SECRET"),
		Version:              mustLookupEnv("STRIPE_VERSION"),
	}
}

// loadLocale resolves the active locale from the COUNTRY environment variable
// and assigns it to the package-level Locale.
// LocaleData is the source of truth for which countries are supported.
func loadLocale() {
	country := strings.ToUpper(mustLookupEnv("COUNTRY"))
	locale, ok := LocaleData[country]
	if !ok {
		slog.Error("country not supported", "country", country)
		os.Exit(1)
	}
	Locale = locale
}

func loadEnvironment() types.Environment {
	env := mustLookupEnv("ENVIRONMENT")
	switch env {
	case string(types.Development):
		return types.Development
	case string(types.Production):
		return types.Production
	default:
		return types.Development
	}
}

func loadHttpClientTimeout() time.Duration {
	duration, err := time.ParseDuration(getEnvOrDefault("HTTP_CLIENT_TIMEOUT", "30s"))
	if err != nil {
		slog.Error("Error parsing duration", "duration", duration, "error", err)
		os.Exit(1)
	}
	return duration
}

func loadBaseURL() string {
	return mustLookupEnv("BASE_URL")
}

func loadRateLimit() bool {
	return isFeatureEnabled("RATE_LIMIT_ENABLED")
}

func loadEmailConfig() types.EmailConfig {
	if isFeatureEnabled("MAIL_ENABLED") {
		return types.EmailConfig{
			Enabled:  true,
			Host:     mustLookupEnv("MAIL_SMTP_HOST"),
			Port:     mustAtoI("MAIL_SMTP_PORT"),
			Username: os.Getenv("MAIL_SMTP_USERNAME"),
			Password: os.Getenv("MAIL_SMTP_PASSWORD"),
			UseTLS:   isFeatureEnabled("MAIL_SMTP_USE_TLS"),
			From:     mustLookupEnv("MAIL_FROM_EMAIL"),
			FromName: mustLookupEnv("MAIL_FROM_NAME"),
		}
	}
	return types.EmailConfig{
		Enabled: false,
	}
}

// mustReadFile reads the named file and returns its contents.
func mustReadFile(filename string) []byte {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		slog.Error("Error reading file", "filename", filename, "error", err)
		os.Exit(1)
	}
	return bytes
}

// mustLookupEnv retrieves the value of the environment variable named by the key.
func mustLookupEnv(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	slog.Error("Environment variable required", "variable", key)
	os.Exit(1)
	return ""
}

// getEnvOrDefault retrieves the value of the environment variable named by the key.
// It returns the default value if the variable is not set.
func getEnvOrDefault(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

// mustParseDuration parses a duration string from the environment variable.
func mustParseDuration(key string) time.Duration {
	duration, err := time.ParseDuration(mustLookupEnv(key))
	if err != nil {
		slog.Error("Error parsing duration", "key", key, "error", err)
		os.Exit(1)
	}
	return duration
}

// mustAtoI converts a string to an integer.
func mustAtoI(key string) int {
	value, err := strconv.Atoi(mustLookupEnv(key))
	if err != nil {
		slog.Error("Error converting string to int", "key", key, "error", err)
		os.Exit(1)
	}
	return value
}

// isFeatureEnabled returns true if the feature is enabled.
func isFeatureEnabled(feature string) bool {
	return strings.EqualFold(os.Getenv(feature), "true")
}
