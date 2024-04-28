package models

type Product struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Price       Currency `json:"price"`
	Description string   `json:"description"`
}
