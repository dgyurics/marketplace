package repositories

import (
	"context"
	"database/sql"

	"github.com/dgyurics/marketplace/types"
	"github.com/lib/pq"
)

type ImageRepository interface {
	CreateImage(ctx context.Context, image *types.Image) error
	RemoveImage(ctx context.Context, id string) (ImageDeletionResult, error)
}

type imageRepository struct {
	db *sql.DB
}

func NewImageRepository(db *sql.DB) ImageRepository {
	return &imageRepository{db: db}
}

func (r *imageRepository) CreateImage(ctx context.Context, image *types.Image) error {
	query := `
        INSERT INTO images (id, product_id, url, type, alt_text, source)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING created_at, updated_at
    `
	return r.db.QueryRowContext(ctx, query,
		image.ID,
		image.ProductID,
		image.URL,
		image.Type,
		image.AltText,
		image.Source,
	).Scan(&image.CreatedAt, &image.UpdatedAt)
}

type ImageDeletionResult struct {
	ProductID       string
	SourceImage     string
	CanDeleteSource bool // if true, delete source image
}

func (r *imageRepository) RemoveImage(ctx context.Context, id string) (ImageDeletionResult, error) {
	var result ImageDeletionResult

	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return result, err
	}
	defer tx.Rollback() // Roll back the transaction in case of an error

	// 1. fetch the productID and source image
	srcQuery := `
		SELECT product_id, source
		FROM images
		WHERE id = $1 FOR UPDATE
	`
	err = tx.QueryRowContext(ctx, srcQuery, id).Scan(&result.ProductID, &result.SourceImage)
	if err == sql.ErrNoRows {
		return result, types.ErrNotFound
	}
	if err != nil {
		return result, err
	}

	// 2. remove the specified image
	deleteQuery := `DELETE FROM images WHERE id = $1`
	res, err := tx.ExecContext(ctx, deleteQuery, id)
	if err != nil {
		return result, err
	}
	// lib/pq always returns nil error for RowsAffected()
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return result, types.ErrNotFound
	}

	// 3. check if source image still in-use
	usageQuery := `SELECT NOT EXISTS (SELECT 1 FROM images WHERE source = $1)`
	err = tx.QueryRowContext(ctx, usageQuery, result.SourceImage).Scan(&result.CanDeleteSource)
	if err != nil {
		return result, err
	}

	return result, tx.Commit()
}

// getProductImages loads images for all requested products in one query.
func getProductImages(ctx context.Context, db *sql.DB, productIDs []string) (map[string][]types.Image, error) {
	images := make(map[string][]types.Image, len(productIDs))
	if len(productIDs) == 0 {
		return images, nil
	}
	for _, id := range productIDs {
		images[id] = []types.Image{}
	}

	query := `
		SELECT id, product_id, url, type, alt_text, source, created_at, updated_at
		FROM images
		WHERE product_id = ANY($1::BIGINT[])
		ORDER BY id
	`
	rows, err := db.QueryContext(ctx, query, pq.Array(productIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var image types.Image
		if err := rows.Scan(
			&image.ID,
			&image.ProductID,
			&image.URL,
			&image.Type,
			&image.AltText,
			&image.Source,
			&image.CreatedAt,
			&image.UpdatedAt,
		); err != nil {
			return nil, err
		}
		images[image.ProductID] = append(images[image.ProductID], image)
	}
	return images, rows.Err()
}
