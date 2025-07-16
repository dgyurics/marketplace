package types

import (
	"log"
	"log/slog"
	"time"
)

// FIXME use pointers
type Config struct {
	Auth         AuthConfig
	BaseURL      string
	CORS         CORSConfig
	Database     DBConfig
	Email        EmailConfig
	Environment  Environment
	JWT          JWTConfig
	Locale       LocaleConfig
	Logger       LoggerConfig
	MachineID    uint8
	Server       ServerConfig
	Stripe       StripeConfig
	Image        ImageConfig
	TemplatesDir string // path to directory containing email templates
}

// ServerConfig is based on net/http.Server.
// See https://pkg.go.dev/net/http#Server for documentation.
type ServerConfig struct {
	Addr           string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	MaxHeaderBytes int
	ErrorLog       *log.Logger
}

type AuthConfig struct {
	HMACSecret    []byte
	RefreshExpiry time.Duration // duration for which the refresh token is valid
	InviteReq     bool          // flag for requiring an invite to register
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

type EmailConfig struct {
	Enabled   bool // toggle for enabling/disabling the email service
	APIKey    string
	APISecret string
	FromEmail string
	FromName  string
}

type JWTConfig struct {
	PrivateKey []byte        // asymmetric key for signing access tokens
	PublicKey  []byte        // asymmetric key for verifying access tokens
	Expiry     time.Duration // duration for which the access token is valid
}

type DBConfig struct {
	URL             string
	MaxOpenConns    int           // max number of open connections to the database
	MaxIdleConns    int           // max number of connections in the idle connection pool
	ConnMaxLifetime time.Duration // max time a connection may be reused
	ConnMaxIdleTime time.Duration // max time a connection may be idle
}

type LoggerConfig struct {
	LogFilePath string // path to the log file
	AppID       string // unique identifier for the application
	Level       slog.Level
}

type TaxBehavior string

const (
	TaxInclusive TaxBehavior = "inclusive"
	TaxExclusive TaxBehavior = "exclusive"
)

type LocaleConfig struct {
	Country         string // ISO 3166-1 alpha-2
	Currency        string // ISO 4217 currency
	FallbackTaxCode string // Default tax code when item/product level tax code not provided
	TaxBehavior     TaxBehavior
}

type Environment string

const (
	Development Environment = "development"
	Production  Environment = "production"
)

type StripeConfig struct {
	BaseURL              string // https://api.stripe.com
	SecretKey            string // sk_test_xxxxxxxxxxxxxxxxxxxxxxxx
	WebhookSigningSecret string // whsec_xxxxxxxxxxxxxxxxxxxxxxxx
	Version              string // 2025-04-30.basil
}

type ImageConfig struct {
	Key            []byte // hex encoded key for imgproxy
	Salt           []byte // hex encoded salt for imgproxy
	BaseURLImgPrxy string // base URL for imgproxy, e.g. http://localhost:8002
	BaseURLRemBg   string // base URL for rembg, e.g. http://localhost:7001
}
