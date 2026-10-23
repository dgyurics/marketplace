package repositories

import (
	"context"
	"database/sql"

	"github.com/dgyurics/marketplace/types"
)

type ImageRepository interface {
	CreateImage(ctx context.Context, image *types.Image) error
	RemoveImage(ctx context.Context, id string) (ImageDeletionResult, error)
	GetImage(ctx context.Context, id string) (types.Image, error)
	GetSourceImages(ctx context.Context, productID, source string) ([]types.Image, error)
	ReplaceSource(ctx context.Context, productID, oldSource string, images []types.Image) error
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
	usageQuery := `SELECT COUNT(*) FROM images WHERE product_id = $2 AND source = $1`
	var usageCount int
	err = tx.QueryRowContext(ctx, usageQuery, result.SourceImage, result.ProductID).Scan(&usageCount)
	if err != nil {
		return result, err
	}
	result.CanDeleteSource = (usageCount == 0)

	return result, tx.Commit()
}

func (r *imageRepository) GetImage(ctx context.Context, id string) (types.Image, error) {
	var img types.Image
	query := `
		SELECT id, product_id, url, type, alt_text, source
		FROM images
		WHERE id = $1
	`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&img.ID,
		&img.ProductID,
		&img.URL,
		&img.Type,
		&img.AltText,
		&img.Source,
	)
	if err == sql.ErrNoRows {
		err = types.ErrNotFound
	}
	return img, err
}

// ReplaceSource updates variants atomically
func (r *imageRepository) ReplaceSource(ctx context.Context, productID, oldSource string, variants []types.Image) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	updateQuery := `
		UPDATE images
		SET source = $1, url = $2, updated_at = NOW()
		WHERE id = $3 AND product_id = $4 AND source = $5
	`
	for _, img := range variants {
		result, err := tx.ExecContext(ctx, updateQuery,
			img.Source,
			img.URL,
			img.ID,
			productID,
			oldSource,
		)
		if err != nil {
			return err
		}
		count, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if count == 0 {
			return types.ErrNotFound
		}
	}

	return tx.Commit()
}

func (r *imageRepository) GetSourceImages(ctx context.Context, productID, source string) ([]types.Image, error) {
	query := `
		SELECT id, product_id, url, type, alt_text, source
		FROM images
		WHERE product_id = $1 AND source = $2
		ORDER BY id
	`
	rows, err := r.db.QueryContext(ctx, query, productID, source)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	images := []types.Image{}
	for rows.Next() {
		var img types.Image
		if err := rows.Scan(
			&img.ID,
			&img.ProductID,
			&img.URL,
			&img.Type,
			&img.AltText,
			&img.Source,
		); err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	return images, rows.Err()
}
