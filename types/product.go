package types

import (
	"encoding/json"
	"strings"
	"time"
)

type ProductStatus string

const (
	ProductDraft     ProductStatus = "draft"
	ProductPublished ProductStatus = "published"
	ProductArchived  ProductStatus = "archived"
)

type Product struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Price       *int64          `json:"price"`
	MemberPrice *int64          `json:"member_price,omitempty"`
	Status      *ProductStatus  `json:"status"`
	Details     json.RawMessage `json:"details"`
	Summary     *string         `json:"summary,omitempty"`
	Description *string         `json:"description,omitempty"`
	Images      []Image         `json:"images"`
	Category    *Category       `json:"category,omitempty"`
	TaxCode     *string         `json:"tax_code,omitempty"`
	Inventory   int             `json:"inventory"`
	Featured    bool            `json:"featured"`
	SortOrder   int             `json:"sort_order"`
	Negotiable  bool            `json:"negotiable"`
	CartLimit   *int            `json:"cart_limit,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

func (p *Product) HideMemberPrice(user *User) {
	if user == nil || !user.HasMinimumRole(RoleMember) {
		p.MemberPrice = nil
	}
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
	Source    string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProductFilter struct {
	SortBy     SortBy
	InStock    bool
	Featured   bool
	Page       int
	Limit      int
	Categories []string // category slugs
}

type SortBy string

const (
	SortByPrice      SortBy = "price"
	SortByPopularity SortBy = "total_sold"
	SortByNewest     SortBy = "created_at"
	SortBySortOrder  SortBy = "sort_order"
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
		return SortBySortOrder
	}
}
