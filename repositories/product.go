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
	GetProducts(ctx context.Context, filter types.ProductFilter) ([]types.Product, error)
	GetProductByID(ctx context.Context, id string) (types.Product, error)
	UpdateProduct(ctx context.Context, product types.Product) error
	RemoveProduct(ctx context.Context, id string) error
}

type productRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) CreateProduct(ctx context.Context, product *types.Product) error {
	var categoryID sql.NullString
	if product.Category != nil {
		categoryID = sql.NullString{String: product.Category.ID, Valid: true}
	}
	query := `
		INSERT INTO products (id, name, price, summary, description, details, tax_code, inventory, cart_limit, category_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id
	`
	if err := r.db.QueryRowContext(ctx,
		query,
		product.ID,
		product.Name,
		product.Price,
		product.Summary,
		product.Description,
		product.Details,
		product.TaxCode,
		product.Inventory,
		product.CartLimit,
		categoryID,
	).Scan(&product.ID); err != nil {
		return err
	}
	return nil
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
			&product.TaxCode,
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
			SELECT p.id, p.name, p.price, p.tax_code, p.summary, p.details, p.images
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
			SELECT p.id, p.name, p.price, p.tax_code, p.summary, p.details, p.images
			FROM v_products p
			JOIN category_tree ct ON ct.id = p.category_id
			WHERE true
		`, strings.Join(placeholders, ", ")))
	}

	argIndex := len(args) + 1

	if filter.InStock {
		queryBuilder.WriteString(" AND p.inventory > 0")
	}

	if filter.SortBy != "" {
		queryBuilder.WriteString(fmt.Sprintf(" ORDER BY p.%s DESC", filter.SortBy))
	}

	queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1))
	args = append(args, filter.Limit, (filter.Page-1)*filter.Limit)

	return queryBuilder.String(), args
}

func (r *productRepository) GetProductByID(ctx context.Context, id string) (types.Product, error) {
	query := `
	SELECT
		p.id,
		p.name,
		p.price,
		p.summary,
		p.description,
		p.details,
		p.images,
		p.inventory,
		p.cart_limit,
		c.id,
		c.name,
		c.slug,
		c.description,
		c.parent_id
	FROM v_products p
	LEFT JOIN categories c ON p.category_id = c.id
	WHERE p.id = $1;
	`

	var product types.Product
	var imagesJSON []byte

	var categoryID, categoryParentID, categoryName, categorySlug, categoryDescription sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Price,
		&product.Summary,
		&product.Description,
		&product.Details,
		&imagesJSON,
		&product.Inventory,
		&product.CartLimit,
		&categoryID,
		&categoryName,
		&categorySlug,
		&categoryDescription,
		&categoryParentID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return product, types.ErrNotFound
		}
		return product, err
	}

	// Convert JSON array to Go struct
	// FIXME seems counterintuitive to convert images to JSON in view/database
	// and then convert back to Go struct/array
	if err := json.Unmarshal(imagesJSON, &product.Images); err != nil {
		return product, err
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

	return product, nil
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
		inventory = $8,
		cart_limit = $9,
		is_deleted = $10,
		updated_at = NOW()
		WHERE id = $11
	`
	result, err := r.db.ExecContext(ctx, query,
		product.Name,
		product.Price,
		product.Summary,
		product.Description,
		product.Details,
		product.TaxCode,
		categoryID,
		product.Inventory,
		product.CartLimit,
		false,
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

func (r *productRepository) RemoveProduct(ctx context.Context, id string) error {
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
