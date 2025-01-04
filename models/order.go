package models

import "time"

type Order struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	Currency        string    `json:"currency"`
	Amount          int64     `json:"amount"`
	TaxAmount       int64     `json:"tax_amount"`
	TotalAmount     int64     `json:"total_amount"`
	Status          string    `json:"status"`
	PaymentIntentID string    `json:"payment_intent_id"` // Stripe PaymentIntent ID
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Environment string

const (
	Development Environment = "development"
	Production  Environment = "production"
)

type OrderConfig struct {
	Envirnment                 Environment
	StripeBaseURL              string // https://api.stripe.com
	StripeSecretKey            string // sk_test_xxxxxxxxxxxxxxxxxxxxxxxx
	StripeWebhookSigningSecret string // whsec_xxxxxxxxxxxxxxxxxxxxxxxx
}
