package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dgyurics/marketplace/models"
)

type CartRepository interface {
	AddItemToCart(ctx context.Context, userID string, item *models.CartItem) error
	GetOrCreateCart(ctx context.Context, userID string) (*models.Cart, error)
	FetchCartTotal(ctx context.Context, userID string) (models.Currency, error)
	UpdateCartItem(ctx context.Context, userID string, item *models.CartItem) error
	RemoveItemFromCart(ctx context.Context, userID, productID string) error
	ReserveCartItems(ctx context.Context, userID string) error
	ClearCart(ctx context.Context, userID string) error
}

type cartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) GetOrCreateCart(ctx context.Context, userID string) (*models.Cart, error) {
	cart := &models.Cart{
		UserID: userID,
	}

	// Use ON CONFLICT to insert a new cart if it doesn't already exist
	query := `
		INSERT INTO carts (user_id)
		VALUES ($1)
		ON CONFLICT (user_id) DO NOTHING`
	if _, err := r.db.ExecContext(ctx, query, userID); err != nil {
		return nil, err
	}

	// populate cart items
	itemsQuery := `
		SELECT product_id, quantity, unit_price
		FROM cart_items
		WHERE user_id = $1`
	rows, err := r.db.QueryContext(ctx, itemsQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.CartItem, 0)
	for rows.Next() {
		var item models.CartItem
		if err := rows.Scan(&item.ProductID, &item.Quantity, &item.UnitPrice.Amount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	cart.Items = items
	return cart, nil
}

func (r *cartRepository) AddItemToCart(ctx context.Context, userID string, item *models.CartItem) error {
	// Check inventory availability
	var availableQuantity int
	if err := r.db.QueryRowContext(ctx, "SELECT quantity FROM inventory WHERE product_id = $1", item.ProductID).Scan(&availableQuantity); err != nil {
		return err
	}
	if availableQuantity < item.Quantity {
		return fmt.Errorf("insufficient inventory for product %s", item.ProductID)
	}

	// Add item to cart without changing inventory
	query := `
		INSERT INTO cart_items (user_id, product_id, quantity, unit_price)
		VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, userID, item.ProductID, item.Quantity, item.UnitPrice.Amount)
	return err
}

func (r *cartRepository) UpdateCartItem(ctx context.Context, userID string, item *models.CartItem) error {
	// Check inventory availability
	var availableQuantity int
	if err := r.db.QueryRowContext(ctx, "SELECT quantity FROM inventory WHERE product_id = $1", item.ProductID).Scan(&availableQuantity); err != nil {
		return err
	}

	// Calculate the quantity difference
	var oldQuantity int
	query := `
		SELECT quantity
		FROM cart_items
		WHERE user_id = $1 AND product_id = $2`
	err := r.db.QueryRowContext(ctx, query, userID, item.ProductID).Scan(&oldQuantity)
	if err != nil {
		return err
	}

	// Check if the new quantity exceeds available inventory
	quantityDifference := item.Quantity - oldQuantity
	if availableQuantity < quantityDifference {
		return fmt.Errorf("insufficient inventory for product %s", item.ProductID)
	}

	// Update the cart item
	updateQuery := `
		UPDATE cart_items
		SET quantity = $3
		WHERE user_id = $1 AND product_id = $2`
	_, err = r.db.ExecContext(ctx, updateQuery, userID, item.ProductID, item.Quantity)
	return err
}

func (r *cartRepository) RemoveItemFromCart(ctx context.Context, userID string, productID string) error {
	deleteQuery := `
		DELETE FROM cart_items
		WHERE user_id = $1 AND product_id = $2`
	_, err := r.db.ExecContext(ctx, deleteQuery, userID, productID)
	return err
}

func (r *cartRepository) ClearCart(ctx context.Context, userID string) error {
	deleteQuery := `
		DELETE FROM cart_items
		WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, deleteQuery, userID)
	return err
}

func (r *cartRepository) ReserveCartItems(ctx context.Context, userID string) error {
	var result string

	// reserve_cart_items returns "success", "empty_cart", "insufficient_inventory"
	query := `SELECT reserve_cart_items($1);`
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(&result); err != nil {
		return err
	}

	if result != "success" {
		return errors.New(result)
	}

	return nil
}

func (r *cartRepository) FetchCartTotal(ctx context.Context, userID string) (models.Currency, error) {
	var total models.Currency
	query := `
		SELECT COALESCE(SUM(quantity * unit_price), 0)
		FROM cart_items
		WHERE user_id = $1`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&total.Amount)
	return total, err
}
