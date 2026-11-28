package repositories

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/dgyurics/marketplace/types"
)

type CartRepository interface {
	AddItem(ctx context.Context, userID, productID string) error
	GetItems(ctx context.Context, userID string) ([]types.CartItem, error)
	RemoveItem(ctx context.Context, userID, productID string) error
	RemoveItems(ctx context.Context, userID string) error
}

type cartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) GetItems(ctx context.Context, userID string) ([]types.CartItem, error) {
	// Fetch cart items and product information.
	itemsQuery := `
		SELECT
			ci.product_id,
			ci.quantity,
			pv.name,
			pv.price,
			pv.member_price,
			pv.summary
		FROM cart_items ci
		JOIN v_products pv ON ci.product_id = pv.id
		WHERE ci.user_id = $1`

	rows, err := r.db.QueryContext(ctx, itemsQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []types.CartItem{}
	productIDs := []string{}
	for rows.Next() {
		var item types.CartItem

		if err := rows.Scan(
			&item.Product.ID,
			&item.Quantity,
			&item.Product.Name,
			&item.Product.Price,
			&item.Product.MemberPrice,
			&item.Product.Summary,
		); err != nil {
			return nil, err
		}

		items = append(items, item)
		productIDs = append(productIDs, item.Product.ID)
	}

	// Check for errors from iterating over rows.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	images, err := getProductImages(ctx, r.db, productIDs)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].Product.Images = images[items[i].Product.ID]
	}
	return items, nil
}

func (r *cartRepository) AddItem(ctx context.Context, userID, productID string) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Lock the product before reading the cart so concurrent additions are serialized.
	var availableQuantity int
	var cartLimit *int
	if err := tx.QueryRowContext(ctx, "SELECT inventory, cart_limit FROM products WHERE id = $1 FOR UPDATE", productID).Scan(&availableQuantity, &cartLimit); err != nil {
		return err
	}
	var existingQuantity int
	err = tx.QueryRowContext(ctx, "SELECT quantity FROM cart_items WHERE user_id = $1 AND product_id = $2", userID, productID).Scan(&existingQuantity)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// If out of stock, return an error
	if existingQuantity >= availableQuantity {
		slog.Info("Product out of stock", "product_id", productID)
		return types.ErrConstraintViolation
	}

	// If cart limit reached, return an error
	if cartLimit != nil && existingQuantity >= *cartLimit {
		slog.Info("Cart limit exceeded", "product_id", productID, "cart_limit", cartLimit)
		return types.ErrConstraintViolation
	}

	query := `
		INSERT INTO cart_items (user_id, product_id, quantity)
		VALUES ($1, $2, 1)
		ON CONFLICT (user_id, product_id) DO UPDATE
		SET quantity = cart_items.quantity + 1,
				updated_at = NOW()`
	_, err = tx.ExecContext(ctx, query, userID, productID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *cartRepository) RemoveItem(ctx context.Context, userID string, productID string) error {
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM cart_items
		WHERE user_id = $1 AND product_id = $2`,
		userID, productID)
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

func (r *cartRepository) RemoveItems(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM cart_items
		WHERE user_id = $1	
	`, userID)
	return err
}
