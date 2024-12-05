package models

import "time"

type Order struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	TotalAmount Currency  `json:"total_amount"`
	TaxAmount   Currency  `json:"tax_amount"`
	OrderStatus string    `json:"order_status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
