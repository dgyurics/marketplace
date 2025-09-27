package types

import (
	"encoding/json"
	"strings"
)

type Product struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Price       int64           `json:"price"`
	Details     json.RawMessage `json:"details"`
	Summary     string          `json:"summary"`
	Description string          `json:"description"`
	Images      []Image         `json:"images"`
	Category    *Category       `json:"category"`
	TaxCode     string          `json:"tax_code"`
	Inventory   int             `json:"inventory"`
	CartLimit   int             `json:"cart_limit"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

type ImageType string

const (
	Thumbnail ImageType = "thumbnail"
	Gallery   ImageType = "gallery"
	Hero      ImageType = "hero"
)

func ParseImageType(s string) ImageType {
	switch strings.ToLower(s) {
	case "hero":
		return Hero
	case "thumbnail":
		return Thumbnail
	case "gallery":
		return Gallery
	default:
		return Gallery
	}
}

type Image struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	URL       string    `json:"url"`
	Type      ImageType `json:"type"`
	AltText   *string   `json:"alt_text,omitempty"` // FIXME convert pointer to string
	Source    string    `json:"source"`
}

type ProductFilter struct {
	SortBy     SortBy
	InStock    bool
	Page       int
	Limit      int
	Categories []string // category slugs
}

type SortBy string

const (
	SortByPrice      SortBy = "price"
	SortByPopularity SortBy = "total_sold"
	SortByNewest     SortBy = "created_at"
	SortByDefault    SortBy = "created_at"
)

func ParseSortBy(sortBy string) SortBy {
	switch sortBy {
	case "price":
		return SortByPrice
	case "popularity":
		return SortByPopularity
	case "newest":
		return SortByNewest
	default:
		return SortByDefault
	}
}
