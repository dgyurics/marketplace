package types

import (
	"log"
	"log/slog"
	"time"
)

type Config struct {
	AppMetadata       AppMetadata
	Auth              AuthConfig
	BaseURL           string
	Database          DBConfig
	Email             EmailConfig
	Environment       Environment
	HTTPClientTimeout time.Duration
	Image             ImageConfig
	JWT               JWTConfig
	Logger            LoggerConfig
	MachineID         uint8
	Payment           PaymentConfig
	RateLimit         bool
	Server            ServerConfig
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
	SSLMode         string        // SSL mode for the connection, e.g. "disable", "require"
	MaxOpenConns    int           // max number of open connections to the database
	MaxIdleConns    int           // max number of idle connections retained in the pool
	ConnMaxLifetime time.Duration // max duration a connection may be reused
	ConnMaxIdleTime time.Duration // max duration a connection may sit idle
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
	FallbackCode string // tax code used when a product does not define one
}

type Environment string

const (
	Development Environment = "development"
	Production  Environment = "production"
)

type StripeConfig struct {
	Enabled              bool   // whether Stripe payments are offered
	BaseURL              string // Stripe API base URL, e.g. https://api.stripe.com
	SecretKey            string // Stripe API secret key, e.g. sk_test_xxxxxxxx
	WebhookSigningSecret string // secret used to verify webhook signatures, e.g. whsec_xxxxxxxx
	Version              string // pinned Stripe API version, e.g. 2025-04-30.basil
}

// TODO break this up into ImgProxyConfig & RemBgConfig
type ImageConfig struct {
	Key             []byte // hex decoded signing key for imgproxy URLs
	Salt            []byte // hex decoded salt for imgproxy URLs
	BaseURLImgproxy string // base URL for imgproxy, e.g. http://localhost:8002
	BaseURLRembg    string // base URL for rembg, e.g. http://localhost:7001
	// TODO RembgMode string // see https://github.com/danielgatis/rembg for available models
	ImageUploadPath  string // directory where uploaded images are stored, e.g. /images
	MaxMegapixels    int    // max allowed source resolution in megapixels
	MaxFileSizeBytes int    // max allowed source file size, in bytes
}

type PaymentConfig struct {
	Stripe      StripeConfig
	OnDelivery  bool
	Tax         TaxConfig
	Environment Environment
}

type PaymentOptions struct {
	Stripe        bool `json:"stripe"`
	PayOnDelivery bool `json:"pay_on_delivery"`
}

// struct exposed to UI in GET /config
type AppMetadata struct {
	PaymentOptions PaymentOptions `json:"payment_options"`
	Locale         LocaleConfig   `json:"locale"`
}

type LocaleConfig struct {
	CountryCode       string            `json:"country_code"`        // ISO 3166-1 alpha-2 code, e.g. "US", "CA", "DE"
	Country           string            `json:"country"`             // display name, e.g. "United States", "Canada"
	PostalCodeLabel   string            `json:"postal_code_label"`   // label for postal code field, e.g. "Postal Code", "Postcode"
	PostalCodePattern string            `json:"postal_code_pattern"` // regex used to validate postal codes, e.g. "^\d{5}(-\d{4})?$"
	StateLabel        string            `json:"state_label"`         // label for state field, e.g. "State", "Province"
	StateRequired     bool              `json:"state_required"`      // whether a state is required for addresses
	StateCodes        map[string]string `json:"state_codes"`         // state code to name, e.g. "CA": "California"
	Currency          string            `json:"currency"`            // ISO 4217 currency code, e.g. "USD", "CAD", "EUR"
	MinorUnits        int               `json:"minor_units"`         // currency fractional digits, e.g. 2 for USD, 0 for JPY
	Language          string            `json:"language"`            // BCP 47 language tag, e.g. "en-US", "fr-CA"
	Line2Label        string            `json:"line2_label"`         // label for address line 2, e.g. "Apt, suite, etc."
	// TODO InclusiveTax bool
}
