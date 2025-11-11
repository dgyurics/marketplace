package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dgyurics/marketplace/types"
)

type OrderRepository interface {
	/* Modify order(s) */
	CreateOrder(ctx context.Context, order *types.Order) error
	ConfirmOrderPayment(ctx context.Context, orderID string) error
	/* GET order(s) */
	GetOrderByIDAndUser(ctx context.Context, orderID, userID string) (types.Order, error)
	GetOrderByID(ctx context.Context, orderID string) (types.Order, error)
	GetOrderByIDPublic(ctx context.Context, orderID string) (types.Order, error)
	GetOrders(ctx context.Context, page, limit int) ([]types.Order, error)
}

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) CreateOrder(ctx context.Context, order *types.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert order
	query := `
		INSERT INTO orders (id, user_id, address_id, amount, tax_amount, shipping_amount, total_amount, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending')`
	if _, err := tx.ExecContext(ctx, query, order.ID, order.UserID, order.Address.ID, order.Amount,
		order.TaxAmount, order.ShippingAmount, order.TotalAmount); err != nil {
		return err
	}

	// Insert order items
	for _, item := range order.Items {
		itemQuery := `
			INSERT INTO order_items (order_id, product_id, quantity, unit_price)
			VALUES ($1, $2, $3, $4)`
		if _, err := tx.ExecContext(ctx, itemQuery, order.ID, item.Product.ID, item.Quantity, item.UnitPrice); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetOrders retrieves all orders in descending order
func (r *orderRepository) GetOrders(ctx context.Context, page, limit int) ([]types.Order, error) {
	query := `
		SELECT
			o.id,
			o.user_id,
			o.amount,
			o.tax_amount,
			o.total_amount,
			o.status,
			a.id AS address_id,
			a.addressee,
			a.line1,
			a.line2,
			a.city,
			a.state,
			a.postal_code,
			a.country,
			a.email,
			o.created_at,
			o.updated_at
		FROM orders o
		JOIN addresses a ON o.address_id = a.id
		ORDER BY o.created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, (page-1)*limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []types.Order{}
	for rows.Next() {
		order := types.Order{
			Address: types.Address{},
		}

		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Amount,
			&order.TaxAmount,
			&order.TotalAmount,
			&order.Status,
			&order.Address.ID,
			&order.Address.Addressee,
			&order.Address.Line1,
			&order.Address.Line2,
			&order.Address.City,
			&order.Address.State,
			&order.Address.PostalCode,
			&order.Address.Country,
			&order.Address.Email,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, order)
	}

	// Check for errors from iterating over rows.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// populateOrderItems populates the order items for a list of orders
func (r *orderRepository) populateOrderItems(ctx context.Context, orderID string) ([]types.OrderItem, error) {
	query := `
		SELECT
			product_id,
			name,
			summary,
			thumbnail,
			alt_text,
			quantity,
			unit_price
		FROM v_order_items
		WHERE order_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Process query results
	items := []types.OrderItem{}
	for rows.Next() {
		item := types.OrderItem{}
		if err := rows.Scan(
			&item.Product.ID,
			&item.Product.Name,
			&item.Product.Summary,
			&item.Thumbnail,
			&item.AltText,
			&item.Quantity,
			&item.UnitPrice,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *orderRepository) GetOrderByIDAndUser(ctx context.Context, orderID, userID string) (types.Order, error) {
	var order types.Order
	if orderID == "" {
		return order, errors.New("order ID is required")
	}
	if userID == "" {
		return order, errors.New("user ID is required")
	}
	query := `
		SELECT
			o.id,
			o.user_id,
			o.amount,
			o.tax_amount,
			o.total_amount,
			o.status,
			o.address_id,
			a.addressee,
			a.line1,
			a.line2,
			a.city,
			a.state,
			a.postal_code,
			a.country,
			a.email,
			o.created_at,
			o.updated_at
		FROM orders o
		LEFT JOIN addresses a ON o.address_id = a.id
		WHERE
			o.id = $1 AND
			o.user_id = $2
	`
	order.Address = types.Address{}
	err := r.db.QueryRowContext(ctx, query, orderID, userID).Scan(
		&order.ID,
		&order.UserID,
		&order.Amount,
		&order.TaxAmount,
		&order.TotalAmount,
		&order.Status,
		&order.Address.ID,
		&order.Address.Addressee,
		&order.Address.Line1,
		&order.Address.Line2,
		&order.Address.City,
		&order.Address.State,
		&order.Address.PostalCode,
		&order.Address.Country,
		&order.Address.Email,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return order, types.ErrNotFound
	}
	if err != nil {
		return order, err
	}

	// Populate order items
	if order.Items, err = r.populateOrderItems(ctx, order.ID); err != nil {
		return order, fmt.Errorf("failed to populate order items: %w", err)
	}

	return order, nil
}

func (r *orderRepository) GetOrderByIDPublic(ctx context.Context, orderID string) (types.Order, error) {
	var order types.Order
	query := `
		SELECT
			o.id,
			o.amount,
			o.tax_amount,
			o.total_amount,
			o.status,
			o.created_at,
			o.updated_at
		FROM orders o
		WHERE o.id = $1
	`
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&order.ID,
		&order.Amount,
		&order.TaxAmount,
		&order.TotalAmount,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return order, types.ErrNotFound
	}
	if err != nil {
		return order, err
	}

	// Populate order items for this order
	if order.Items, err = r.populateOrderItems(ctx, order.ID); err != nil {
		return order, fmt.Errorf("failed to populate order items: %w", err)
	}

	return order, nil
}

func (r *orderRepository) GetOrderByID(ctx context.Context, orderID string) (types.Order, error) {
	var order types.Order
	order.Address = types.Address{}
	query := `
		SELECT
			o.id,
			o.user_id,
			o.amount,
			o.tax_amount,
			o.total_amount,
			o.status,
			o.address_id,
			a.addressee,
			a.line1,
			a.line2,
			a.city,
			a.state,
			a.postal_code,
			a.country,
			a.email,
			o.created_at,
			o.updated_at
		FROM orders o
		LEFT JOIN addresses a ON o.address_id = a.id
		WHERE o.id = $1
	`
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&order.ID,
		&order.UserID,
		&order.Amount,
		&order.TaxAmount,
		&order.TotalAmount,
		&order.Status,
		&order.Address.ID,
		&order.Address.Addressee,
		&order.Address.Line1,
		&order.Address.Line2,
		&order.Address.City,
		&order.Address.State,
		&order.Address.PostalCode,
		&order.Address.Country,
		&order.Address.Email,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return order, types.ErrNotFound
	}
	if err != nil {
		return order, err
	}

	// Populate order items for this order
	if order.Items, err = r.populateOrderItems(ctx, order.ID); err != nil {
		return order, fmt.Errorf("failed to populate order items: %w", err)
	}

	return order, nil
}

func (r *orderRepository) ConfirmOrderPayment(ctx context.Context, orderID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update order status and get user_id
	var userID string
	updateQuery := `
        UPDATE orders
        SET status = 'paid', updated_at = NOW()
        WHERE id = $1
        RETURNING user_id`

	if err := tx.QueryRowContext(ctx, updateQuery, orderID).Scan(&userID); err != nil {
		return err
	}

	// Clear cart for that user
	deleteQuery := `DELETE FROM cart_items WHERE user_id = $1`
	if _, err := tx.ExecContext(ctx, deleteQuery, userID); err != nil {
		return err
	}

	return tx.Commit()
}
