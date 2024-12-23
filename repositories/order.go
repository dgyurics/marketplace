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
	query := "SELECT mark_order_as_paid($1)"
	_, err := r.db.ExecContext(ctx, query, orderID)
	return err
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
