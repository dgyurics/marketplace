package utilities

import (
	"encoding/hex"
	"fmt"
	"log"
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
		Auth:         loadAuthConfig(),
		Database:     loadDBConfig(),
		Email:        loadMailConfig(),
		Locale:       loadLocaleConfig(),
		Logger:       loadLoggerConfig(),
		MachineID:    loadMachineID(),
		Stripe:       loadStripeConfig(),
		JWT:          loadJWTConfig(),
		Image:        loadImageConfig(),
		TemplatesDir: loadTemplatesDir(),
		BaseURL:      loadBaseURL(),
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
		InviteReq:     isFeatureEnabled("INVITE_REQUIRED"),
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

func loadJWTConfig() types.JWTConfig {
	privKeyPath := getEnvOrDefault("PRIVATE_KEY_PATH", "private.pem")
	pubKeyPath := getEnvOrDefault("PUBLIC_KEY_PATH", "public.pem")

	privateKey := mustReadFile(privKeyPath)
	publicKey := mustReadFile(pubKeyPath)

	// validate files aren't empty
	if len(privateKey) == 0 {
		panic(fmt.Sprintf("Private key file is empty: %s", privKeyPath))
	}
	if len(publicKey) == 0 {
		panic(fmt.Sprintf("Public key file is empty: %s", pubKeyPath))
	}

	return types.JWTConfig{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Expiry:     mustParseDuration("JWT_EXPIRY"),
	}
}

func loadMachineID() uint8 {
	envVar := mustLookupEnv("MACHINE_ID")
	val, err := strconv.ParseUint(envVar, 10, 8)
	if err != nil {
		panic(fmt.Sprintf("Invalid integer for MACHINE_ID: %s", envVar))
	}
	return uint8(val)
}

func loadImageConfig() types.ImageConfig {
	keyHex := mustLookupEnv("IMGPROXY_KEY")
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		panic("invalid IMGPROXY_KEY: " + err.Error())
	}

	saltHex := mustLookupEnv("IMGPROXY_SALT")
	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		panic("invalid IMGPROXY_SALT: " + err.Error())
	}

	urlImgproxy := fmt.Sprintf("%s/images", loadBaseURL())
	urlRembg := "http://rembg"

	return types.ImageConfig{
		Key:             key,
		Salt:            salt,
		BaseURLImgproxy: urlImgproxy,
		BaseURLRembg:    urlRembg,
	}
}

func loadLocaleConfig() types.LocaleConfig {
	config := types.LocaleConfig{
		Currency:        mustLookupEnv("CURRENCY"),
		Country:         mustLookupEnv("COUNTRY"),
		TaxBehavior:     types.TaxExclusive,
		FallbackTaxCode: mustLookupEnv("FALLBACK_TAX_CODE"),
	}

	if _, ok := SupportedCountries[config.Country]; !ok {
		panic(fmt.Sprintf("Unsupported country code: %s", config.Country))
	}

	if _, ok := SupportedCurrencies[config.Currency]; !ok {
		panic(fmt.Sprintf("Unsupported currency code: %s", config.Currency))
	}

	taxBehavior := mustLookupEnv("TAX_BEHAVIOR")
	if taxBehavior == string(types.TaxInclusive) {
		config.TaxBehavior = types.TaxInclusive
	}

	return config
}

func loadStripeConfig() types.StripeConfig {
	return types.StripeConfig{
		BaseURL:              mustLookupEnv("STRIPE_BASE_URL"),
		SecretKey:            mustLookupEnv("STRIPE_SECRET_KEY"),
		WebhookSigningSecret: mustLookupEnv("STRIPE_WEBHOOK_SIGNING_SECRET"),
		Version:              mustLookupEnv("STRIPE_VERSION"),
	}
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

func loadTemplatesDir() string {
	return "./utilities/templates"
}

func loadBaseURL() string {
	return mustLookupEnv("BASE_URL")
}

func loadMailConfig() types.EmailConfig {
	return types.EmailConfig{
		Enabled:   isFeatureEnabled("MAIL_ENABLED"),
		APIKey:    mustLookupEnv("MAIL_API_KEY"),
		APISecret: mustLookupEnv("MAIL_API_SECRET"),
		FromEmail: mustLookupEnv("MAIL_FROM_EMAIL"),
		FromName:  mustLookupEnv("MAIL_FROM_NAME"),
	}
}

// mustReadFile reads the named file and returns the contents.
// It panics if there is an error reading the file.
func mustReadFile(filename string) []byte {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		panic(fmt.Sprintf("Error reading key file: %s error: %v", filename, err))
	}
	return bytes
}

// mustLookupEnv retrieves the value of the environment variable named by the key.
// It panics if the variable is not present.
func mustLookupEnv(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	panic(fmt.Sprintf("Environment variable %s is required", key))
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
// It panics if there is an error parsing the duration.
func mustParseDuration(key string) time.Duration {
	duration, err := time.ParseDuration(mustLookupEnv(key))
	if err != nil {
		panic(fmt.Sprintf("Error parsing duration for %s: %v", key, err))
	}
	return duration
}

// mustAtoI converts a string to an integer.
// It panics if there is an error converting the string.
func mustAtoI(key string) int {
	value, err := strconv.Atoi(mustLookupEnv(key))
	if err != nil {
		panic(fmt.Sprintf("Error converting %s to int: %v", key, err))
	}
	return value
}

// isFeatureEnabled returns true if the feature is enabled.
func isFeatureEnabled(feature string) bool {
	return strings.EqualFold(os.Getenv(feature), "true")
}
