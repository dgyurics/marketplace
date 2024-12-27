package repositories

import (
	"context"
	"database/sql"

	"github.com/dgyurics/marketplace/models"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, userID string) (order *models.Order, err error)
	FetchPendingOrders(ctx context.Context, userID string) ([]*models.Order, error)
	MarkOrderAsPaid(ctx context.Context, orderID string) error
}

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) CreateOrder(ctx context.Context, userID string) (*models.Order, error) {
	// 1) Call the stored procedure which creates an order from the cart
	query := "SELECT create_order_from_cart($1)"
	var orderID string
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&orderID)
	if err != nil {
		return nil, err
	}

	// 2) Retrieve the newly created order
	query = `
		SELECT id, user_id, total_amount, tax_amount, order_status, created_at, updated_at
		FROM orders
		WHERE id = $1
	`
	order := &models.Order{}
	err = r.db.QueryRowContext(ctx, query, orderID).Scan(
		&order.ID,
		&order.UserID,
		&order.TotalAmount, // does not include tax
		&order.TaxAmount,
		&order.OrderStatus,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (r *orderRepository) MarkOrderAsPaid(ctx context.Context, orderID string) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback() // Roll back the transaction in case of an error

	// Mark payment status as paid
	query := `
		UPDATE payments
		SET status = 'paid'
		WHERE order_id = $1 AND status = 'pending'
	`
	if _, err = tx.ExecContext(ctx, query, orderID); err != nil {
		return err
	}

	// Mark order as paid
	query = `
		UPDATE orders
		SET order_status = 'paid',
				updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND order_status = 'created'
		RETURNING user_id
	`
	var userID string
	if err = tx.QueryRowContext(ctx, query, orderID).Scan(&userID); err != nil {
		return err
	}

	// Empty the cart
	query = `
		DELETE FROM cart_items
		WHERE user_id = $1
	`
	if _, err = tx.ExecContext(ctx, query, userID); err != nil {
		return err
	}

	// Delete the cart
	query = `
		DELETE FROM carts
		WHERE user_id = $1
	`
	if _, err = tx.ExecContext(ctx, query, userID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *orderRepository) FetchPendingOrders(ctx context.Context, userID string) (orders []*models.Order, err error) {
	query := `
		SELECT id, user_id, total_amount, tax_amount, order_status, created_at, updated_at
		FROM orders
		WHERE user_id = $1 AND order_status = $2
	`
	rows, err := r.db.QueryContext(ctx, query, userID, "created")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.TotalAmount,
			&order.TaxAmount,
			&order.OrderStatus,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}
	return
}
