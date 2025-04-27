package types

import (
	"log"
	"log/slog"
	"time"
)

type Config struct {
	Auth         AuthConfig
	BaseURL      string
	CORS         CORSConfig
	Database     DBConfig
	Email        EmailConfig
	Environment  Environment
	JWT          JWTConfig
	Logger       LoggerConfig
	MachineID    uint8
	Server       ServerConfig
	Order        OrderConfig
	TemplatesDir string // path to the directory containing email templates
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

type Environment string

const (
	Development Environment = "development"
	Production  Environment = "production"
)

type OrderConfig struct {
	DefaultTaxCode     string
	DefaultTaxBehavior string
	StripeConfig
}

type StripeConfig struct {
	BaseURL              string // https://api.stripe.com
	SecretKey            string // sk_test_xxxxxxxxxxxxxxxxxxxxxxxx
	WebhookSigningSecret string // whsec_xxxxxxxxxxxxxxxxxxxxxxxx
}
