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
	CompleteOrder(ctx context.Context, userID, paymentIntentID string) error
	StartOrder(ctx context.Context, userID, paymentIntentID string, amount models.Currency) error
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
		if err := rows.Scan(&item.ProductID, &item.Quantity, &item.UnitPrice); err != nil {
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

	// Fetch unit_price from the product table
	if err := r.db.QueryRowContext(ctx, "SELECT price FROM products WHERE id = $1", item.ProductID).Scan(&item.UnitPrice); err != nil {
		return err
	}

	// Add item to cart using the fetched unit_price
	unitPriceAsFloat := float64(item.UnitPrice.Amount) / 100
	query := `
		INSERT INTO cart_items (user_id, product_id, quantity, unit_price)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, product_id) DO UPDATE
		SET quantity = EXCLUDED.quantity,
		    unit_price = EXCLUDED.unit_price`
	_, err := r.db.ExecContext(ctx, query, userID, item.ProductID, item.Quantity, unitPriceAsFloat)
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
		SELECT COALESCE(CAST(SUM(quantity * unit_price) AS DECIMAL(10,2)), 0.00)
		FROM cart_items
		WHERE user_id = $1`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&total)
	return total, err
}

func (r *cartRepository) StartOrder(ctx context.Context, userID, paymentIntentID string, amount models.Currency) error {
	totalAsFloat := float64(amount.Amount) / 100
	query := `
		INSERT INTO orders (user_id, total_amount, order_status, payment_intent_id)
		VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, userID, totalAsFloat, "created", paymentIntentID)
	return err
}

func (r *cartRepository) CompleteOrder(ctx context.Context, userID, paymentIntentID string) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // Roll back the transaction in case of an error

	// Update the order status to "paid"
	updateOrderQuery := `
		UPDATE orders
		SET order_status = 'paid'
		WHERE payment_intent_id = $1 AND order_status = 'created'`
	res, err := tx.ExecContext(ctx, updateOrderQuery, paymentIntentID)
	if err != nil {
		return err
	}

	// Check if exactly one row was affected
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to fetch rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no order found for payment intent ID %s", paymentIntentID)
	}
	if rowsAffected > 1 {
		return fmt.Errorf("multiple orders found for payment intent ID %s", paymentIntentID)
	}

	// clear reserved inventory
	clearInventoryQuery := `
		DELETE FROM inventory_reservations
		WHERE user_id = $1`
	if _, err := tx.ExecContext(ctx, clearInventoryQuery, userID); err != nil {
		return err
	}

	// clear cart
	clearCartQuery := `
		DELETE FROM cart_items
		WHERE user_id = $1`
	if _, err := tx.ExecContext(ctx, clearCartQuery, userID); err != nil {
		return err
	}

	return tx.Commit()
}
