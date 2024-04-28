package repositories

import (
	"context"

	"github.com/dgyurics/marketplace/models"
	"github.com/jackc/pgx/v4/pgxpool"
)

type CategoryRepository interface {
	CreateCategory(ctx context.Context, category models.Category) (int, error)
	GetAllCategories(ctx context.Context) ([]models.Category, error)
	GetCategoryByID(ctx context.Context, id int) (*models.Category, error)
	GetProductsByCategoryID(ctx context.Context, id string) ([]models.Product, error)
}

type categoryRepository struct {
	pool *pgxpool.Pool
}

func NewCategoryRepository(pool *pgxpool.Pool) CategoryRepository {
	return &categoryRepository{pool: pool}
}

func (r *categoryRepository) CreateCategory(ctx context.Context, category models.Category) (int, error) {
	var newID int
	query := `INSERT INTO categories (name, description) VALUES ($1, $2) RETURNING id`
	err := r.pool.QueryRow(ctx, query, category.Name, category.Description).Scan(&newID)
	if err != nil {
		return 0, err
	}
	return newID, nil
}

func (r *categoryRepository) GetAllCategories(ctx context.Context) ([]models.Category, error) {
	var categories []models.Category
	rows, err := r.pool.Query(ctx, "SELECT id, name, description FROM categories")
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

func (r *categoryRepository) GetCategoryByID(ctx context.Context, id int) (*models.Category, error) {
	var category models.Category
	if err := r.pool.QueryRow(ctx, "SELECT id, name, description FROM categories WHERE id = $1", id).Scan(&category.ID, &category.Name, &category.Description); err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) GetProductsByCategoryID(ctx context.Context, id string) ([]models.Product, error) {
	var products []models.Product
	rows, err := r.pool.Query(ctx, `
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
