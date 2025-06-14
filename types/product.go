package types

import "encoding/json"

type Product struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Price       int64           `json:"price"`
	Details     json.RawMessage `json:"details"`
	Description string          `json:"description"`
	Images      []Image         `json:"images"`
	TaxCode     string          `json:"tax_code"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

type ProductWithInventory struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Price       int64           `json:"price"`
	Description string          `json:"description"`
	Details     json.RawMessage `json:"details"`
	Images      json.RawMessage `json:"images"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
	Quantity    int             `json:"quantity"`
}

type Image struct {
	ID           string  `json:"id"`
	ProductID    string  `json:"product_id"`
	URL          string  `json:"url"`
	Type         string  `json:"type"` // main, thumbnail, gallery
	DisplayOrder int     `json:"display_order"`
	AltText      *string `json:"alt_text,omitempty"`
}

type ProductFilter struct {
	SortByPrice bool
	SortAsc     bool
	InStock     bool
	Page        int
	Limit       int
	Categories  []string // category slugs
}
