package types

import (
	"log"
	"log/slog"
	"time"
)

type Config struct {
	Auth         AuthConfig
	BaseURL      string
	Country      string
	Database     DBConfig
	Email        EmailConfig
	Environment  Environment
	Image        ImageConfig
	JWT          JWTConfig
	Logger       LoggerConfig
	MachineID    uint8
	Payment      PaymentConfig
	RateLimit    bool
	Server       ServerConfig
	Tax          TaxConfig
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
}

type EmailConfig struct {
	Enabled  bool
	Host     string
	Port     int
	Username string
	Password string
	UseTLS   bool
	From     string
	FromName string
}

type JWTConfig struct {
	PrivateKey []byte        // asymmetric key for signing access tokens
	PublicKey  []byte        // asymmetric key for verifying access tokens
	Expiry     time.Duration // duration for which the access token is valid
}

type DBConfig struct {
	Host            string        // database host
	Port            int           // database port
	User            string        // database user
	Password        string        // database password
	Name            string        // database name
	SSLMode         string        // SSL mode for the database connection (e.g., "disable
	MaxOpenConns    int           // max number of open connections to the database
	MaxIdleConns    int           // max number of connections in the idle connection pool
	ConnMaxLifetime time.Duration // max time a connection may be reused
	ConnMaxIdleTime time.Duration // max time a connection may be idle
}

type LoggerConfig struct {
	Level slog.Level
}

type TaxBehavior string

const (
	TaxInclusive TaxBehavior = "inclusive"
	TaxExclusive TaxBehavior = "exclusive"
)

type TaxConfig struct {
	Behavior     TaxBehavior
	FallbackCode string // Default tax code when item/product level tax code not provided
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
	Key              []byte // hex encoded key for imgproxy
	Salt             []byte // hex encoded salt for imgproxy
	BaseURLImgproxy  string // base URL for imgproxy, e.g. http://localhost:8002
	BaseURLRembg     string // base URL for rembg, e.g. http://localhost:7001
	ImageUploadPath  string // directory for storing images, e.g. /images
	MaxMegapixels    int    // maximum allowed resolution (width x height)
	MaxFileSizeBytes int    // maximum allowed file size (in bytes)
}

type PaymentConfig struct {
	Stripe      StripeConfig
	Tax         TaxConfig
	BaseURL     string
	Environment Environment
}
