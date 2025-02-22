package repositories

import (
	"context"
	"database/sql"

	"github.com/dgyurics/marketplace/types"
)

type ProductRepository interface {
	CreateProduct(ctx context.Context, product *types.Product) error
	CreateProductWithCategory(ctx context.Context, product *types.Product, categoryID string) error
	GetAllProducts(ctx context.Context, page, limit int) ([]types.Product, error)
	GetProductByID(ctx context.Context, id string) (*types.Product, error)
	DeleteProduct(ctx context.Context, id string) error
	UpdateInventory(ctx context.Context, productID string, quantity int) error
}

type productRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) CreateProduct(ctx context.Context, product *types.Product) error {
	query := `
		INSERT INTO products (name, price, description)
		VALUES ($1, $2, $3)
		RETURNING id, name, price, description`

	if err := r.db.QueryRowContext(ctx, query, product.Name, product.Price, product.Description).
		Scan(&product.ID, &product.Name, &product.Price, &product.Description); err != nil {
		return err
	}

	// Create an inventory record for the product
	inventoryQuery := `
		INSERT INTO inventory (product_id, quantity)
		VALUES ($1, 0)`
	_, err := r.db.ExecContext(ctx, inventoryQuery, product.ID)
	return err
}

func (r *productRepository) CreateProductWithCategory(ctx context.Context, product *types.Product, categoryID string) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // Roll back the transaction in case of an error

	// Create the product
	query := `
		INSERT INTO products (name, price, description)
		VALUES ($1, $2, $3)
		RETURNING id, name, price, description`
	if err = tx.QueryRowContext(ctx, query, product.Name, product.Price, product.Description).
		Scan(&product.ID, &product.Name, &product.Price, &product.Description); err != nil {
		return err
	}

	// Associate the product with the category
	associationQuery := `
		INSERT INTO product_categories (product_id, category_id)
		VALUES ($1, $2)`
	if _, err = tx.ExecContext(ctx, associationQuery, product.ID, categoryID); err != nil {
		return err
	}

	// Create an inventory record for the product
	inventoryQuery := `
	INSERT INTO inventory (product_id, quantity)
	VALUES ($1, 0)`
	if _, err := tx.ExecContext(ctx, inventoryQuery, product.ID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *productRepository) GetAllProducts(ctx context.Context, page, limit int) ([]types.Product, error) {
	var products []types.Product
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			p.id,
			p.name,
			p.price,
			p.description,
			p.created_at,
			p.updated_at
		FROM products p
		WHERE is_deleted = false
		ORDER BY p.created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, (page-1)*limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var product types.Product
		if err = rows.Scan(
			&product.ID,
			&product.Name,
			&product.Price,
			&product.Description,
			&product.CreatedAt,
			&product.UpdatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}

func (r *productRepository) GetProductByID(ctx context.Context, id string) (*types.Product, error) {
	var product types.Product
	if err := r.db.QueryRowContext(ctx, `
		SELECT
			p.id,
			p.name,
			p.price,
			p.description,
			p.created_at,
			p.updated_at
		FROM products p
		WHERE p.id = $1
			AND is_deleted = false`, id).Scan(
		&product.ID,
		&product.Name,
		&product.Price,
		&product.Description,
		&product.CreatedAt,
		&product.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) DeleteProduct(ctx context.Context, id string) error {
	query := `UPDATE products SET is_deleted = true WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *productRepository) UpdateInventory(ctx context.Context, productID string, quantity int) error {
	query := `UPDATE inventory SET quantity = $1 WHERE product_id = $2`
	_, err := r.db.ExecContext(ctx, query, quantity, productID)
	return err
}
