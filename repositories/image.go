package repositories

import (
	"context"
	"database/sql"

	"github.com/dgyurics/marketplace/types"
)

type ImageRepository interface {
	ProductExists(ctx context.Context, productID string) (bool, error)
	CreateImage(ctx context.Context, image *types.Image) error
	RemoveImage(ctx context.Context, id string) (ImageDeletionResult, error)
	PromoteImage(ctx context.Context, id string) error
}

type imageRepository struct {
	db *sql.DB
}

func NewImageRepository(db *sql.DB) ImageRepository {
	return &imageRepository{db: db}
}

func (r *imageRepository) ProductExists(ctx context.Context, productID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)`
	err := r.db.QueryRowContext(ctx, query, productID).Scan(&exists)
	return exists, err
}

func (r *imageRepository) CreateImage(ctx context.Context, image *types.Image) error {
	query := `
        INSERT INTO images (id, product_id, url, type, alt_text, source)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err := r.db.ExecContext(ctx, query,
		image.ID,
		image.ProductID,
		image.URL,
		image.Type,
		image.AltText,
		image.Source,
	)
	return err
}

func (r *imageRepository) PromoteImage(ctx context.Context, id string) error {
	query := `
		UPDATE images
		SET updated_at = NOW()
		WHERE id = $1
	`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	// lib/pq always returns nil error for RowsAffected()
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return types.ErrNotFound
	}
	return nil
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
		WHERE id = $1
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
	usageQuery := `SELECT COUNT(*) FROM images WHERE source = $1`
	var usageCount int
	err = tx.QueryRowContext(ctx, usageQuery, result.SourceImage).Scan(&usageCount)
	if err != nil {
		return result, err
	}
	result.CanDeleteSource = (usageCount == 0)

	return result, tx.Commit()
}
