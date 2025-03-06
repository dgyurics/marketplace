package types

type Cart struct {
	UserID string     `json:"user_id"`
	Items  []CartItem `json:"items"`
}

type CartItem struct {
	Product   Product `json:"product"`
	Quantity  int     `json:"quantity"`
	UnitPrice int64   `json:"unit_price"`
}
