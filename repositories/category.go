package repositories

import (
	"context"
	"database/sql"

	"github.com/dgyurics/marketplace/types"
)

type CategoryRepository interface {
	CreateCategory(ctx context.Context, category types.Category) (string, error)
	GetAllCategories(ctx context.Context) ([]types.Category, error)
	GetCategoryByID(ctx context.Context, id string) (*types.Category, error)
}

type categoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) CreateCategory(ctx context.Context, category types.Category) (string, error) {
	var newID string
	query := `INSERT INTO categories (name, description) VALUES ($1, $2) RETURNING id`
	err := r.db.QueryRowContext(ctx, query, category.Name, category.Description).Scan(&newID)
	if err != nil {
		return "", err
	}
	return newID, nil
}

func (r *categoryRepository) GetAllCategories(ctx context.Context) ([]types.Category, error) {
	var categories []types.Category
	query := `
		SELECT
			id,
			name,
			description,
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
			&category.CreatedAt,
			&category.UpdatedAt,
		); err != nil {
			return nil, err
		}
		categories = append(categories, category)
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
