package types

import "time"

type CartItem struct {
	Product   Product   `json:"product"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
