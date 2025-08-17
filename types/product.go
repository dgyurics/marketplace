package types

import (
	"encoding/json"
	"fmt"
	"strconv"
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
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

type ProductWithInventory struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Price       int64           `json:"price"`
	Summary     string          `json:"summary"`
	Description string          `json:"description"`
	Details     json.RawMessage `json:"details"`
	Images      json.RawMessage `json:"images"`
	Category    *Category       `json:"category"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
	Quantity    int             `json:"quantity"`
}

type ImageType string

const (
	Thumbnail ImageType = "thumbnail"
	Gallery   ImageType = "gallery"
	Hero      ImageType = "hero"
	Original  ImageType = "original"
)

// ParseImageType parses a string into ImageType
func ParseImageType(s string) (ImageType, error) {
	switch strings.ToLower(s) {
	case "original":
		return Original, nil
	case "hero":
		return Hero, nil
	case "thumbnail":
		return Thumbnail, nil
	case "gallery":
		return Gallery, nil
	default:
		return "", fmt.Errorf("invalid image type: %s", s)
	}
}

// ParseImageOrder parses a string into an integer for display order
func ParseImageOrder(s string) int {
	if s == "" {
		return 0 // Default display order if not provided
	}
	r, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return r
}

type Image struct {
	ID           string    `json:"id"`
	ProductID    string    `json:"product_id"`
	URL          string    `json:"url"`
	Type         ImageType `json:"type"`
	DisplayOrder int       `json:"display_order"`      // FIXME remove this + references
	AltText      *string   `json:"alt_text,omitempty"` // FIXME convert pointer to string
	Source       string    `json:"source"`
}

type ProductFilter struct {
	SortByPrice bool
	SortAsc     bool
	InStock     bool
	Page        int
	Limit       int
	Categories  []string // category slugs
}
