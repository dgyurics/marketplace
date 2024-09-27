package repositories

import (
	"context"

	"github.com/dgyurics/marketplace/models"
	"github.com/jackc/pgx/v4/pgxpool"
)

type ProductRepository interface {
	CreateProduct(ctx context.Context, product *models.Product) error
	CreateProductWithCategory(ctx context.Context, product *models.Product, categoryID string) error
	GetAllProducts(ctx context.Context) ([]models.Product, error)
	GetProductByID(ctx context.Context, id string) (*models.Product, error)
	DeleteProduct(ctx context.Context, id string) error
}

type productRepository struct {
	pool *pgxpool.Pool
}

func NewProductRepository(pool *pgxpool.Pool) ProductRepository {
	return &productRepository{pool: pool}
}

func (r *productRepository) CreateProduct(ctx context.Context, product *models.Product) error {
	query := `
		INSERT INTO products (name, price, description)
		VALUES ($1, $2, $3)
		RETURNING id, name, price, description`

	priceAsFloat := float64(product.Price.Amount) / 100
	if err := r.pool.QueryRow(ctx, query, product.Name, priceAsFloat, product.Description).
		Scan(&product.ID, &product.Name, &product.Price, &product.Description); err != nil {
		return err
	}

	return nil
}

func (r *productRepository) CreateProductWithCategory(ctx context.Context, product *models.Product, categoryID string) error {
	// Begin a transaction
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // Roll back the transaction in case of an error

	query := `
		INSERT INTO products (name, price, description)
		VALUES ($1, $2, $3)
		RETURNING id, name, price, description`

	priceAsFloat := float64(product.Price.Amount) / 100
	if err = tx.QueryRow(ctx, query, product.Name, priceAsFloat, product.Description).
		Scan(&product.ID, &product.Name, &product.Price, &product.Description); err != nil {
		return err
	}

	associationQuery := `
		INSERT INTO product_categories (product_id, category_id)
		VALUES ($1, $2)`
	if _, err = tx.Exec(ctx, associationQuery, product.ID, categoryID); err != nil {
		return err
	}

	// Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (r *productRepository) GetAllProducts(ctx context.Context) ([]models.Product, error) {
	var products []models.Product
	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.name, p.price, p.description
		FROM products p`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var product models.Product
		if err = rows.Scan(&product.ID, &product.Name, &product.Price, &product.Description); err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}

func (r *productRepository) GetProductByID(ctx context.Context, id string) (*models.Product, error) {
	var product models.Product
	if err := r.pool.QueryRow(ctx, `
		SELECT p.id, p.name, p.price, p.description
		FROM products p
		WHERE p.id = $1`, id).Scan(&product.ID, &product.Name, &product.Price, &product.Description); err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) DeleteProduct(ctx context.Context, id string) error {
	query := `DELETE FROM products WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}
