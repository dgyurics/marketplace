package types

type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Price       int64   `json:"price"`
	Description string  `json:"description"`
	Images      []Image `json:"images"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type Image struct {
	ID           string  `json:"id"`
	ProductID    string  `json:"product_id"`
	ImageURL     string  `json:"image_url"`
	ImageType    string  `json:"image_type"`
	Format       string  `json:"format"`
	Animated     bool    `json:"animated"`
	DisplayOrder int     `json:"display_order"`
	AltText      *string `json:"alt_text,omitempty"`
}
