package models

type Cart struct {
	UserID string     `json:"user_id"`
	Items  []CartItem `json:"items"`
	Total  Currency   `json:"total"`
}

type CartItem struct {
	ProductID  string   `json:"product_id"`
	Quantity   int      `json:"quantity"`
	UnitPrice  Currency `json:"unit_price"`
	TotalPrice Currency `json:"total_price"`
}
