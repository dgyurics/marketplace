package models

type Inventory struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}
