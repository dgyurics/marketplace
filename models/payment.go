package models

import "time"

// TODO refactor this with stripe.go models considering these
// are specific to stripe http requests
type PaymentIntentRequest struct {
	Amount   int64
	Currency string
}

type PaymentIntentResponse struct {
	ID           string `json:"id"`
	Amount       int64  `json:"amount"`
	Currency     string `json:"currency"`
	Status       string `json:"status"`
	ClientSecret string `json:"client_secret"`
	Error        string `json:"error,omitempty"`
}

type PaymentIntent struct {
	Status         string
	AmountReceived int64
}

type Payment struct {
	OrderID         string    `json:"order_id"`
	PaymentIntentID string    `json:"payment_intent_id"`
	ClientSecret    string    `json:"client_secret"`
	Amount          int64     `json:"amount"`
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Environment string

const (
	Development Environment = "development"
	Production  Environment = "production"
)

type PaymentConfig struct {
	Envirnment                 Environment
	StripeBaseURL              string // https://api.stripe.com
	StripeSecretKey            string // sk_test_xxxxxxxxxxxxxxxxxxxxxxxx
	StripeWebhookSigningSecret string // whsec_xxxxxxxxxxxxxxxxxxxxxxxx
}
