package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"

	"github.com/dgyurics/marketplace/types"
)

type CartRepository interface {
	AddItem(ctx context.Context, userID string, item *types.CartItem) error
	GetItems(ctx context.Context, userID string) ([]types.CartItem, error)
	RemoveItem(ctx context.Context, userID, productID string) error
	Clear(ctx context.Context, userID string) error
}

type cartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) GetItems(ctx context.Context, userID string) ([]types.CartItem, error) {
	// Fetch cart items using the materialized view (contains product and images)
	itemsQuery := `
		SELECT
			ci.product_id,
			ci.quantity,
			ci.unit_price,
		  pv.name,
			pv.price,
			pv.summary,
			pv.images
		FROM cart_items ci
		JOIN v_products pv ON ci.product_id = pv.id
		WHERE ci.user_id = $1`

	rows, err := r.db.QueryContext(ctx, itemsQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []types.CartItem{}
	for rows.Next() {
		var item types.CartItem
		var imagesJSON []byte

		if err := rows.Scan(
			&item.Product.ID,
			&item.Quantity,
			&item.UnitPrice,
			&item.Product.Name,
			&item.Product.Price,
			&item.Product.Summary,
			&imagesJSON,
		); err != nil {
			return nil, err
		}

		// Convert JSON array to Go struct
		if err := json.Unmarshal(imagesJSON, &item.Product.Images); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	// Check for errors from iterating over rows.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *cartRepository) AddItem(ctx context.Context, userID string, item *types.CartItem) error {
	// Fetch the current quantity in the cart
	var existingQuantity int
	err := r.db.QueryRowContext(ctx, "SELECT quantity FROM cart_items WHERE user_id = $1 AND product_id = $2", userID, item.Product.ID).Scan(&existingQuantity)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// Check inventory availability and cart limit
	var availableQuantity int
	var cartLimit *int
	if err := r.db.QueryRowContext(ctx, "SELECT inventory, cart_limit FROM products WHERE id = $1", item.Product.ID).Scan(&availableQuantity, &cartLimit); err != nil {
		return err
	}

	// If not enough inventory, return an error
	if availableQuantity < (existingQuantity + item.Quantity) {
		slog.Info("Insufficient inventory for product", "product_id", item.Product.ID, "available", availableQuantity, "requested", item.Quantity)
		return types.ErrConstraintViolation
	}

	// If cart limit reached, return an error
	if cartLimit != nil && *cartLimit < (existingQuantity+item.Quantity) {
		slog.Info("Cart limit exceeded for product", "product_id", item.Product.ID, "cart_limit", cartLimit, "requested", item.Quantity)
		return types.ErrConstraintViolation
	}

	// Fetch unit_price from the product table
	if err := r.db.QueryRowContext(ctx, "SELECT price FROM products WHERE id = $1", item.Product.ID).Scan(&item.UnitPrice); err != nil {
		return err
	}

	// Add item to cart using the fetched unit_price
	query := `
		INSERT INTO cart_items (user_id, product_id, quantity, unit_price)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, product_id) DO UPDATE
		SET quantity = cart_items.quantity + EXCLUDED.quantity,
		    unit_price = EXCLUDED.unit_price`
	_, err = r.db.ExecContext(ctx, query, userID, item.Product.ID, item.Quantity, item.UnitPrice)
	return err
}

func (r *cartRepository) RemoveItem(ctx context.Context, userID string, productID string) error {
	deleteQuery := `
		DELETE FROM cart_items
		WHERE user_id = $1 AND product_id = $2`

	res, err := r.db.ExecContext(ctx, deleteQuery, userID, productID)
	if err != nil {
		return err
	}
	// lib/pq always returns nil error for RowsAffected()
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return types.ErrNotFound
	}
	return nil
}

func (r *cartRepository) Clear(ctx context.Context, userID string) error {
	deleteQuery := `
		DELETE FROM cart_items
		WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, deleteQuery, userID)
	return err
}
