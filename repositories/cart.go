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
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Fetch the current quantity in the cart
	var existingQuantity int
	err = tx.QueryRowContext(ctx, "SELECT quantity FROM cart_items WHERE user_id = $1 AND product_id = $2", userID, item.Product.ID).Scan(&existingQuantity)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// Check inventory availability and cart limit
	var availableQuantity int
	var cartLimit *int
	if err := tx.QueryRowContext(ctx, "SELECT inventory, cart_limit FROM products WHERE id = $1 FOR UPDATE", item.Product.ID).Scan(&availableQuantity, &cartLimit); err != nil {
		return err
	}

	// If not enough inventory, return an error
	if availableQuantity == 0 {
		slog.Info("Product is no longer available", "product_id", item.Product.ID)
		return types.ErrConstraintViolation
	}

	// If cart limit reached, return an error
	if cartLimit != nil && *cartLimit == existingQuantity {
		slog.Info("Cart limit exceeded", "product_id", item.Product.ID, "cart_limit", cartLimit)
		return types.ErrConstraintViolation
	}

	// decrement product.inventory by 1
	// increment cart.quantity by 1
	// update price in cart to reflect latest price in products table
	query := `
		WITH update_inventory AS (
			UPDATE products
			SET inventory = inventory - 1
			WHERE id = $2
			RETURNING price
		)
		INSERT INTO cart_items (user_id, product_id, quantity, unit_price)
		SELECT $1, $2, $3, price FROM update_inventory
		ON CONFLICT (user_id, product_id) DO UPDATE
		SET quantity = cart_items.quantity + 1,
				unit_price = EXCLUDED.unit_price`
	_, err = tx.ExecContext(ctx, query, userID, item.Product.ID, item.Quantity)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *cartRepository) RemoveItem(ctx context.Context, userID string, productID string) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Fetch the current quantity in the cart
	var existingQuantity int
	err = tx.QueryRowContext(ctx, `
		DELETE FROM cart_items
		WHERE user_id = $1 AND product_id = $2
		RETURNING quantity`,
		userID, productID).Scan(&existingQuantity)
	if err == sql.ErrNoRows {
		return types.ErrNotFound
	}
	if err != nil {
		return err
	}

	// Restock inventory
	if _, err := tx.ExecContext(ctx, `
		UPDATE products
		SET inventory = inventory + $1
		WHERE id = $2`,
		existingQuantity, productID); err != nil {
		return err
	}

	return tx.Commit()
}
