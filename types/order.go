package types

import (
	"time"
)

type OrderStatus string

const (
	OrderPending   OrderStatus = "pending"
	OrderPaid      OrderStatus = "paid"
	OrderFulfilled OrderStatus = "fulfilled"
	OrderShipped   OrderStatus = "shipped"
	OrderDelivered OrderStatus = "delivered"
	OrderRefunded  OrderStatus = "refunded"
	OrderCanceled  OrderStatus = "canceled"
)

type Order struct {
	ID             string      `json:"id"`
	UserID         string      `json:"-"`
	Address        Address     `json:"address"`
	Amount         int64       `json:"amount"`
	TaxAmount      int64       `json:"tax_amount"`
	ShippingAmount int64       `json:"shipping_amount"`
	TotalAmount    int64       `json:"total_amount"`
	Status         OrderStatus `json:"status"`
	Items          []OrderItem `json:"items"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

type OrderItem struct {
	Product   Product `json:"product"`
	Thumbnail string  `json:"thumbnail"`
	AltText   string  `json:"alt_text"`
	Quantity  int     `json:"quantity"`
	UnitPrice int64   `json:"unit_price"`
}
