package models

import "time"

type OrderStatus string

const (
	OrderPending    OrderStatus = "pending"
	OrderPaid       OrderStatus = "paid"
	OrderRefunded   OrderStatus = "refunded"
	OrderFullfilled OrderStatus = "fullfilled"
	OrderShipped    OrderStatus = "shipped"
	OrderDelivered  OrderStatus = "delivered"
	OrderCanceled   OrderStatus = "canceled"
)

type Order struct {
	ID              string      `json:"id"`
	UserID          string      `json:"-"`
	Currency        string      `json:"currency"`
	Amount          int64       `json:"amount"`
	TaxAmount       int64       `json:"tax_amount"`
	TotalAmount     int64       `json:"total_amount"`
	Status          OrderStatus `json:"status"`
	PaymentIntentID string      `json:"-"` // stripe payment intent id
	Items           []OrderItem `json:"items"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ProductID   string `json:"product_id"`
	Description string `json:"description"`
	Thumbnail   string `json:"thumbnail"`
	Quantity    int    `json:"quantity"`
	UnitPrice   int64  `json:"unit_price"`
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
