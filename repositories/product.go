package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dgyurics/marketplace/types"
)

type ProductRepository interface {
	CreateProduct(ctx context.Context, product *types.Product) error
	CreateProductWithCategory(ctx context.Context, product *types.Product, categorySlug string) error
	GetProducts(ctx context.Context, filter types.ProductFilter) ([]types.Product, error)
	GetProductByID(ctx context.Context, id string) (*types.ProductWithInventory, error)
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

	return nil
}

func (r *productRepository) CreateProductWithCategory(ctx context.Context, product *types.Product, categorySlug string) error {
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
		VALUES ($1, (SELECT id FROM categories WHERE slug = $2))`
	if _, err = tx.ExecContext(ctx, associationQuery, product.ID, categorySlug); err != nil {
		return err
	}

	// Create any images associated with the product
	for idx, image := range product.Images {
		imageQuery := `
			INSERT INTO images (product_id, image_url, animated, display_order, alt_text)
			VALUES ($1, $2, $3, $4, $5)`
		if _, err = tx.ExecContext(ctx, imageQuery, product.ID, image.ImageURL, image.Animated, idx, image.AltText); err != nil {
			return err
		}
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

func (r *productRepository) GetProducts(ctx context.Context, filter types.ProductFilter) ([]types.Product, error) {
	var products []types.Product
	query, args := generateGetProductsQuery(filter)

	rows, err := r.db.QueryContext(ctx, query, args...)
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

// generateGetProductsQuery generates the SQL query to get products based on the filter
// and returns the query string and arguments to be used with db.QueryContext
func generateGetProductsQuery(filter types.ProductFilter) (string, []interface{}) {
	args := []interface{}{}
	var queryBuilder strings.Builder
	if len(filter.Categories) == 0 {
		queryBuilder.WriteString(`
			SELECT p.id, p.name, p.price, p.description, p.images
			FROM v_product p
			WHERE true
		`)
	} else {
		placeholders := make([]string, 0, len(filter.Categories))
		for i, slug := range filter.Categories {
			placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
			args = append(args, slug)
		}
		queryBuilder.WriteString(fmt.Sprintf(`
			WITH RECURSIVE category_tree AS (
				SELECT id FROM categories WHERE slug IN (%s)
				UNION ALL
				SELECT c.id FROM categories c
				JOIN category_tree ct ON c.parent_id = ct.id
			)
			SELECT p.id, p.name, p.price, p.description, p.images
			FROM v_product p
			JOIN product_categories pc ON p.id = pc.product_id
			JOIN category_tree ct ON ct.id = pc.category_id
			WHERE true
		`, strings.Join(placeholders, ", ")))
	}

	argIndex := len(args) + 1

	if filter.InStock {
		queryBuilder.WriteString(" AND p.quantity > 0")
	}

	if filter.SortByPrice {
		if filter.SortAsc {
			queryBuilder.WriteString(" ORDER BY p.price ASC")
		} else {
			queryBuilder.WriteString(" ORDER BY p.price DESC")
		}
	} else {
		queryBuilder.WriteString(" ORDER BY category_slug")
	}

	queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1))
	args = append(args, filter.Limit, (filter.Page-1)*filter.Limit)

	return queryBuilder.String(), args
}

func (r *productRepository) GetProductByID(ctx context.Context, id string) (*types.ProductWithInventory, error) {
	query := `
	SELECT id, name, price, description, details, images, quantity
	FROM v_product
	WHERE id = $1;
	`

	var product types.ProductWithInventory
	var imagesJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Price,
		&product.Description,
		&product.Details,
		&imagesJSON,
		&product.Quantity,
	)
	if err != nil {
		return nil, err
	}

	// FIXME store as raw json
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
	return nil
}

func (r *productRepository) UpdateInventory(ctx context.Context, productID string, quantity int) error {
	query := `UPDATE inventory SET quantity = $1 WHERE product_id = $2`
	_, err := r.db.ExecContext(ctx, query, quantity, productID)
	return err
}
