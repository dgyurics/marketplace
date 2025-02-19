package types

type Cart struct {
	UserID string     `json:"user_id"`
	Items  []CartItem `json:"items"`
}

type CartItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	UnitPrice int64  `json:"unit_price"`
}
