package repositories

import (
	"context"
	"database/sql"
	"encoding/json"

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
	if _, err := r.db.ExecContext(ctx, inventoryQuery, product.ID); err != nil {
		return err
	}

	// Refresh the materialized view to include the new product
	_, err := r.db.ExecContext(ctx, "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_product")
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

	// Refresh the materialized view within the transaction
	if _, err := tx.ExecContext(ctx, "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_product"); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *productRepository) GetAllProducts(ctx context.Context, page, limit int) ([]types.Product, error) {
	var products []types.Product
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, price, description, created_at, updated_at, images
		FROM mv_product
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, (page-1)*limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var product types.Product
		var imagesJSON []byte

		if err = rows.Scan(
			&product.ID,
			&product.Name,
			&product.Price,
			&product.Description,
			&product.CreatedAt,
			&product.UpdatedAt,
			&imagesJSON,
		); err != nil {
			return nil, err
		}

		// Convert JSONB array to Go struct
		if err := json.Unmarshal(imagesJSON, &product.Images); err != nil {
			return nil, err
		}

		products = append(products, product)
	}
	return products, nil
}

func (r *productRepository) GetProductByID(ctx context.Context, id string) (*types.Product, error) {
	query := `
	SELECT id, name, price, description, created_at, updated_at, images
	FROM mv_product
	WHERE id = $1;
	`

	var product types.Product
	var imagesJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID, &product.Name, &product.Price, &product.Description, &product.CreatedAt, &product.UpdatedAt,
		&imagesJSON,
	)
	if err != nil {
		return nil, err
	}

	// Convert JSONB to Go struct
	if err := json.Unmarshal(imagesJSON, &product.Images); err != nil {
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

	// Refresh the materialized view to include the new product
	_, err = r.db.ExecContext(ctx, "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_product")
	return err
}

func (r *productRepository) UpdateInventory(ctx context.Context, productID string, quantity int) error {
	query := `UPDATE inventory SET quantity = $1 WHERE product_id = $2`
	_, err := r.db.ExecContext(ctx, query, quantity, productID)
	return err
}
