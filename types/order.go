package types

import (
	"time"

	"github.com/dgyurics/marketplace/types/stripe"
)

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

type OrderParams struct {
	ID                  string                `json:"id"`
	UserID              string                `json:"user_id"`
	AddressID           *string               `json:"address_id"`
	Email               *string               `json:"email"`
	Amount              *int64                `json:"amount"`
	TaxAmount           *int64                `json:"tax_amount"`
	ShippingAmount      *int64                `json:"shipping_amount"`
	TotalAmount         *int64                `json:"total_amount"`
	Status              *OrderStatus          `json:"status"`
	StripePaymentIntent *stripe.PaymentIntent `json:"stripe_payment_intent"`
}

type Order struct {
	ID                  string                `json:"id"`
	UserID              string                `json:"-"`
	Email               string                `json:"email"`
	Address             *Address              `json:"address"` // TODO change to shipping address
	StripePaymentIntent *stripe.PaymentIntent `json:"-"`
	Currency            string                `json:"currency"`
	Amount              int64                 `json:"amount"`
	TaxAmount           int64                 `json:"tax_amount"`
	ShippingAmount      int64                 `json:"shipping_amount"`
	TotalAmount         int64                 `json:"total_amount"`
	Status              OrderStatus           `json:"status"`
	Items               []OrderItem           `json:"items"`
	CreatedAt           time.Time             `json:"created_at"`
	UpdatedAt           time.Time             `json:"updated_at"`
}

type OrderItem struct {
	Product   Product `json:"product"`
	Thumbnail string  `json:"thumbnail"`
	AltText   string  `json:"alt_text"`
	Quantity  int     `json:"quantity"`
	UnitPrice int64   `json:"unit_price"`
}
