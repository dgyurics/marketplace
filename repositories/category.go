package repositories

import (
	"context"
	"database/sql"

	"github.com/dgyurics/marketplace/types"
)

type CategoryRepository interface {
	CreateCategory(ctx context.Context, category *types.Category) error
	GetAllCategories(ctx context.Context) ([]types.Category, error)
	GetCategoryByID(ctx context.Context, id string) (*types.Category, error)
	RemoveCategory(ctx context.Context, id string) error
}

type categoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) CreateCategory(ctx context.Context, category *types.Category) error {
	query := `
		INSERT INTO categories (id, name, slug, description, parent_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		category.ID,
		category.Name,
		category.Slug,
		category.Description,
		category.ParentID,
	).Scan(&category.CreatedAt, &category.UpdatedAt)
}

func (r *categoryRepository) GetAllCategories(ctx context.Context) ([]types.Category, error) {
	categories := []types.Category{}
	query := `
		SELECT
			id,
			name,
			description,
			slug,
			parent_id,
			created_at,
			updated_at
		FROM categories`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var category types.Category
		if err = rows.Scan(
			&category.ID,
			&category.Name,
			&category.Description,
			&category.Slug,
			&category.ParentID,
			&category.CreatedAt,
			&category.UpdatedAt,
		); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	// Check for errors from iterating over rows.
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *categoryRepository) GetCategoryByID(ctx context.Context, id string) (*types.Category, error) {
	var category types.Category
	query := `
		SELECT
			id,
			name,
			description,
			created_at,
			updated_at
		FROM categories
		WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) RemoveCategory(ctx context.Context, id string) error {
	// delete will propogade and delete entries from product_categories table
	query := `
		DELETE FROM categories
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return types.ErrNotFound
	}
	return nil
}
