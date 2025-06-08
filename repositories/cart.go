package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/dgyurics/marketplace/types"
)

type CartRepository interface {
	AddItemToCart(ctx context.Context, userID string, item *types.CartItem) error
	GetCart(ctx context.Context, userID string) ([]types.CartItem, error)
	UpdateCartItem(ctx context.Context, userID string, item *types.CartItem) error
	RemoveItemFromCart(ctx context.Context, userID, productID string) error
	ClearCart(ctx context.Context, userID string) error
}

type cartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) GetCart(ctx context.Context, userID string) ([]types.CartItem, error) {
	// Fetch cart items using the materialized view (contains product and images)
	itemsQuery := `
		SELECT
			ci.product_id,
			ci.quantity,
			ci.unit_price,
		  pv.name,
			pv.price,
			pv.description,
			pv.images
		FROM cart_items ci
		JOIN v_products pv ON ci.product_id = pv.id
		WHERE ci.user_id = $1`

	rows, err := r.db.QueryContext(ctx, itemsQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]types.CartItem, 0)
	for rows.Next() {
		var item types.CartItem
		var imagesJSON []byte

		if err := rows.Scan(
			&item.Product.ID,
			&item.Quantity,
			&item.UnitPrice,
			&item.Product.Name,
			&item.Product.Price,
			&item.Product.Description,
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

func (r *cartRepository) AddItemToCart(ctx context.Context, userID string, item *types.CartItem) error {
	// Fetch the current quantity in the cart
	var existingQuantity int
	err := r.db.QueryRowContext(ctx, "SELECT quantity FROM cart_items WHERE user_id = $1 AND product_id = $2", userID, item.Product.ID).Scan(&existingQuantity)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// Check inventory availability considering existing cart quantity
	var availableQuantity int
	if err := r.db.QueryRowContext(ctx, "SELECT quantity FROM inventory WHERE product_id = $1", item.Product.ID).Scan(&availableQuantity); err != nil {
		return err
	}

	if availableQuantity < (existingQuantity + item.Quantity) {
		return fmt.Errorf("insufficient inventory for product %s", item.Product.ID)
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

func (r *cartRepository) UpdateCartItem(ctx context.Context, userID string, item *types.CartItem) error {
	// Check inventory availability
	var availableQuantity int
	if err := r.db.QueryRowContext(ctx, "SELECT quantity FROM inventory WHERE product_id = $1", item.Product.ID).Scan(&availableQuantity); err != nil {
		return err
	}

	// Calculate the quantity difference
	var oldQuantity int
	query := `
		SELECT quantity
		FROM cart_items
		WHERE user_id = $1 AND product_id = $2`
	err := r.db.QueryRowContext(ctx, query, userID, item.Product.ID).Scan(&oldQuantity)
	if err != nil {
		return err
	}

	// Check if the new quantity exceeds available inventory
	quantityDifference := item.Quantity - oldQuantity
	if availableQuantity < quantityDifference {
		return fmt.Errorf("insufficient inventory for product %s", item.Product.ID)
	}

	// Update the cart item
	updateQuery := `
		UPDATE cart_items
		SET quantity = $3
		WHERE user_id = $1 AND product_id = $2
		RETURNING product_id, quantity, unit_price`
	err = r.db.QueryRowContext(ctx, updateQuery, userID, item.Product.ID, item.Quantity).Scan(
		&item.Product.ID,
		&item.Quantity,
		&item.UnitPrice,
	)
	return err
}

func (r *cartRepository) RemoveItemFromCart(ctx context.Context, userID string, productID string) error {
	deleteQuery := `
		DELETE FROM cart_items
		WHERE user_id = $1 AND product_id = $2`
	result, err := r.db.ExecContext(ctx, deleteQuery, userID, productID)
	if err != nil {
		return err
	}
	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		return types.ErrNotFound
	}
	return nil
}

func (r *cartRepository) ClearCart(ctx context.Context, userID string) error {
	deleteQuery := `
		DELETE FROM cart_items
		WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, deleteQuery, userID)
	return err
}
