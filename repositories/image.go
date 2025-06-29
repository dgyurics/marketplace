package repositories

import (
	"context"
	"database/sql"

	"github.com/dgyurics/marketplace/types"
)

type ImageRepository interface {
	ProductExists(ctx context.Context, productID string) (bool, error)
	CreateImage(ctx context.Context, image *types.Image) error
}

type imageRepository struct {
	db *sql.DB
}

func NewImageRepository(db *sql.DB) ImageRepository {
	return &imageRepository{db: db}
}

func (r *imageRepository) ProductExists(ctx context.Context, productID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, productID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // Product does not exist
		}
		return false, err // Other error
	}
	return exists, nil // Product exists
}

func (r *imageRepository) CreateImage(ctx context.Context, image *types.Image) error {
	query := `
        INSERT INTO images (id, product_id, url, type, display_order, alt_text)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err := r.db.ExecContext(ctx, query,
		image.ID,
		image.ProductID,
		image.URL,
		image.Type,
		image.DisplayOrder,
		image.AltText,
	)
	return err
}
