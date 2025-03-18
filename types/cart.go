package types

type CartItem struct {
	Product   Product `json:"product"`
	Quantity  int     `json:"quantity"`
	UnitPrice int64   `json:"unit_price"`
}
