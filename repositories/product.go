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
	// FIXME refactor to use product.category
	CreateProduct(ctx context.Context, product *types.Product, categorySlug string) error
	GetProducts(ctx context.Context, filter types.ProductFilter) ([]types.Product, error)
	GetProductByID(ctx context.Context, id string) (*types.ProductWithInventory, error)
	UpdateProduct(ctx context.Context, product types.Product) error
	DeleteProduct(ctx context.Context, id string) error
	UpdateInventory(ctx context.Context, productID string, quantity int) error
}

type productRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) CreateProduct(ctx context.Context, product *types.Product, categorySlug string) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // Roll back the transaction in case of an error

	// Create the product
	query := `
		INSERT INTO products (id, name, price, summary, description, details, tax_code, category_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, (SELECT id FROM categories WHERE slug = $8))
		RETURNING id, name, price, summary`
	if err = tx.QueryRowContext(ctx,
		query,
		product.ID,
		product.Name,
		product.Price,
		product.Summary,
		product.Description,
		product.Details,
		sql.NullString{
			String: product.TaxCode,
			Valid:  strings.TrimSpace(product.TaxCode) != "",
		},
		categorySlug,
	).Scan(&product.ID, &product.Name, &product.Price, &product.Summary); err != nil {
		return err
	}

	// TODO remove this - not needed
	// Create any images associated with the product
	for _, image := range product.Images {
		imageQuery := `
			INSERT INTO images (id, product_id, url, type, alt_text, source)
			VALUES ($1, $2, $3, $4, $5, $6)`
		if _, err = tx.ExecContext(ctx, imageQuery,
			image.ID, product.ID, image.URL,
			image.Type, image.AltText, image.Source,
		); err != nil {
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
	products := []types.Product{}
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
			&product.Summary,
			&product.Details,
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

	// Check for errors from iterating over rows.
	if err = rows.Err(); err != nil {
		return nil, err
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
			SELECT p.id, p.name, p.price, p.summary, p.details, p.images
			FROM v_products p
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
			SELECT p.id, p.name, p.price, p.summary, p.details, p.images
			FROM v_products p
			JOIN category_tree ct ON ct.id = p.category_id
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
	SELECT
		p.id,
		p.name,
		p.price,
		p.summary,
		p.description,
		p.details,
		p.images,
		p.quantity,
		c.id,
		c.name,
		c.slug,
		c.description,
		c.parent_id
	FROM v_products p
	LEFT JOIN categories c ON p.category_id = c.id
	WHERE p.id = $1;
	`

	var product types.ProductWithInventory
	var categoryID, categoryParentID, categoryName, categorySlug, categoryDescription sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Price,
		&product.Summary,
		&product.Description,
		&product.Details,
		&product.Images,
		&product.Quantity,
		&categoryID,
		&categoryName,
		&categorySlug,
		&categoryDescription,
		&categoryParentID,
	)
	if err != nil {
		return nil, err
	}

	// Populate Category if category data exists
	if categoryID.Valid {
		product.Category = &types.Category{
			ID:          categoryID.String,
			Name:        categoryName.String,
			Slug:        categorySlug.String,
			Description: categoryDescription.String,
		}
		if categoryParentID.Valid {
			product.Category.ParentID = &categoryParentID.String
		}
	}

	return &product, nil
}

func (r *productRepository) UpdateProduct(ctx context.Context, product types.Product) error {
	var categoryID sql.NullString
	if product.Category != nil {
		categoryID = sql.NullString{String: product.Category.ID, Valid: true}
	}

	query := `UPDATE products SET
		name = $1,
		price = $2,
		summary = $3,
		description = $4,
		details = $5,
		tax_code = $6,
		category_id = $7,
		is_deleted = $8,
		updated_at = NOW()
		WHERE id = $9
	`
	result, err := r.db.ExecContext(ctx, query,
		product.Name,
		product.Price,
		product.Summary,
		product.Description,
		product.Details,
		product.TaxCode,
		categoryID,
		false, // FIXME need field product.enabled
		product.ID,
	)
	if err != nil {
		return err
	}
	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		return types.ErrNotFound
	}
	return nil
}

func (r *productRepository) DeleteProduct(ctx context.Context, id string) error {
	query := `UPDATE products SET is_deleted = true WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		return types.ErrNotFound
	}
	return nil
}

func (r *productRepository) UpdateInventory(ctx context.Context, productID string, quantity int) error {
	query := `UPDATE inventory SET quantity = $1 WHERE product_id = $2`
	_, err := r.db.ExecContext(ctx, query, quantity, productID)
	return err
}
