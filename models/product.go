package models

type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Price       int64  `json:"price"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
