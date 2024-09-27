package repositories

import (
	"context"
	"database/sql"

	"github.com/dgyurics/marketplace/models"
)

type CategoryRepository interface {
	CreateCategory(ctx context.Context, category models.Category) (string, error)
	GetAllCategories(ctx context.Context) ([]models.Category, error)
	GetCategoryByID(ctx context.Context, id string) (*models.Category, error)
	GetProductsByCategoryID(ctx context.Context, id string) ([]models.Product, error)
}

type categoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) CreateCategory(ctx context.Context, category models.Category) (string, error) {
	var newID string
	query := `INSERT INTO categories (name, description) VALUES ($1, $2) RETURNING id`
	err := r.db.QueryRowContext(ctx, query, category.Name, category.Description).Scan(&newID)
	if err != nil {
		return "", err
	}
	return newID, nil
}

func (r *categoryRepository) GetAllCategories(ctx context.Context) ([]models.Category, error) {
	var categories []models.Category
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, description FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var category models.Category
		if err = rows.Scan(&category.ID, &category.Name, &category.Description); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func (r *categoryRepository) GetCategoryByID(ctx context.Context, id string) (*models.Category, error) {
	var category models.Category
	err := r.db.QueryRowContext(ctx, "SELECT id, name, description FROM categories WHERE id = $1", id).
		Scan(&category.ID, &category.Name, &category.Description)
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) GetProductsByCategoryID(ctx context.Context, id string) ([]models.Product, error) {
	var products []models.Product
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id, p.name, p.description, p.price
		FROM products p
		JOIN product_categories pc ON p.id = pc.product_id
		WHERE category_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var product models.Product
		if err = rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price); err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}
