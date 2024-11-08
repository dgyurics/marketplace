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

func (r *cartRepository) CreateCart(ctx context.Context, userID string) error {
	query := `
		INSERT INTO carts (user_id)
		VALUES ($1)`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *cartRepository) GetOrCreateCart(ctx context.Context, userID string) (*models.Cart, error) {
	cart := &models.Cart{}

	// Use ON CONFLICT to insert a new cart if it doesn't already exist
	query := `
		INSERT INTO carts (user_id, total)
		VALUES ($1, 0)
		ON CONFLICT (user_id) DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	// Fetch the cart after the potential insertion
	query = `
		SELECT user_id, total
		FROM carts
		WHERE user_id = $1`
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(&cart.UserID, &cart.Total.Amount); err != nil {
		return nil, err
	}

	// populate cart items
	itemsQuery := `
		SELECT product_id, quantity, unit_price, total_price
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
		if err := rows.Scan(&item.ProductID, &item.Quantity, &item.UnitPrice.Amount, &item.TotalPrice.Amount); err != nil {
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
		INSERT INTO cart_items (user_id, product_id, quantity, unit_price, total_price)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, userID, item.ProductID, item.Quantity, item.UnitPrice.Amount, item.TotalPrice.Amount)
	if err != nil {
		return err
	}

	return r.updateCartTotal(ctx, userID, item.TotalPrice)
}

func (r *cartRepository) UpdateCartItem(ctx context.Context, userID string, item *models.CartItem) error {
	// Check inventory availability
	var availableQuantity int
	if err := r.db.QueryRowContext(ctx, "SELECT quantity FROM inventory WHERE product_id = $1", item.ProductID).Scan(&availableQuantity); err != nil {
		return err
	}

	// Calculate the quantity difference
	var oldItem models.CartItem
	query := `
		SELECT quantity, total_price
		FROM cart_items
		WHERE user_id = $1 AND product_id = $2`
	err := r.db.QueryRowContext(ctx, query, userID, item.ProductID).Scan(&oldItem.Quantity, &oldItem.TotalPrice.Amount)
	if err != nil {
		return err
	}

	// Check if the new quantity exceeds available inventory
	quantityDifference := item.Quantity - oldItem.Quantity
	if availableQuantity < quantityDifference {
		return fmt.Errorf("insufficient inventory for product %s", item.ProductID)
	}

	// Update the cart item
	updateQuery := `
		UPDATE cart_items
		SET quantity = $3, total_price = $4
		WHERE user_id = $1 AND product_id = $2`
	_, err = r.db.ExecContext(ctx, updateQuery, userID, item.ProductID, item.Quantity, item.TotalPrice.Amount)
	if err != nil {
		return err
	}

	// Update cart total
	priceDifference := item.TotalPrice.Amount - oldItem.TotalPrice.Amount
	return r.updateCartTotal(ctx, userID, models.NewCurrency(0, priceDifference))
}

func (r *cartRepository) RemoveItemFromCart(ctx context.Context, userID string, productID string) error {
	// Get total price of item to subtract from cart total
	var itemTotalPrice int64
	query := `
		SELECT total_price
		FROM cart_items
		WHERE user_id = $1 AND product_id = $2`
	if err := r.db.QueryRowContext(ctx, query, userID, productID).Scan(&itemTotalPrice); err != nil {
		return err
	}

	// Delete item from cart
	deleteQuery := `
		DELETE FROM cart_items
		WHERE user_id = $1 AND product_id = $2`
	_, err := r.db.ExecContext(ctx, deleteQuery, userID, productID)
	if err != nil {
		return err
	}

	// Update cart total
	return r.updateCartTotal(ctx, userID, models.Currency{Amount: -itemTotalPrice})
}

func (r *cartRepository) ClearCart(ctx context.Context, userID string) error {
	deleteQuery := `
		DELETE FROM cart_items
		WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, deleteQuery, userID)
	if err != nil {
		return err
	}

	return r.updateCartTotal(ctx, userID, models.Currency{Amount: 0})
}

func (r *cartRepository) updateCartTotal(ctx context.Context, userID string, priceChange models.Currency) error {
	query := `
		UPDATE carts
		SET total = total + $2
		WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID, priceChange.Amount)
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
