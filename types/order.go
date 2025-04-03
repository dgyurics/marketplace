package types

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
	Address         *Address    `json:"address"`
	PaymentIntentID string      `json:"-"` // stripe payment intent id
	Currency        string      `json:"currency"`
	Amount          int64       `json:"amount"`
	TaxAmount       int64       `json:"tax_amount"`
	ShippingAmount  int64       `json:"shipping_amount"`
	TotalAmount     int64       `json:"total_amount"`
	Status          OrderStatus `json:"status"`
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
