package models

import "time"

type Order struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	TotalAmount int64     `json:"total_amount"`
	TaxAmount   int64     `json:"tax_amount"`
	OrderStatus string    `json:"order_status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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
