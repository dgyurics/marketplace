package repositories

import (
	"context"
	"database/sql"

	"github.com/dgyurics/marketplace/models"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, userID string, amount models.Currency) (*models.Order, error)
	UpdateOrder(ctx context.Context, orderID string, status string) error
	FetchPendingOrders(ctx context.Context, userID string) ([]*models.Order, error)
}

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) CreateOrder(ctx context.Context, userID string, amount models.Currency) (*models.Order, error) {
	totalAsFloat := float64(amount.Amount) / 100

	query := `
		INSERT INTO orders (user_id, total_amount, order_status)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, total_amount, tax_amount, order_status, created_at, updated_at
	`
	var order models.Order
	var totalAmount, taxAmount float64

	err := r.db.QueryRowContext(ctx, query, userID, totalAsFloat, "created").Scan(
		&order.ID,
		&order.UserID,
		&totalAmount,
		&taxAmount,
		&order.OrderStatus,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	order.TotalAmount = models.Currency{Amount: int64(totalAmount * 100)}
	order.TaxAmount = models.Currency{Amount: int64(taxAmount * 100)}

	return &order, nil
}

// TODO change behavior of orders table to be event-driven architecture (insert only)
func (r *orderRepository) UpdateOrder(ctx context.Context, orderID string, status string) error {
	query := `
		UPDATE orders
		SET order_status = $1
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, status, orderID)
	return err
}

func (r *orderRepository) FetchPendingOrders(ctx context.Context, userID string) ([]*models.Order, error) {
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

	var orders []*models.Order
	for rows.Next() {
		var order models.Order
		var totalAmount, taxAmount float64
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&totalAmount,
			&taxAmount,
			&order.OrderStatus,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		order.TotalAmount = models.Currency{Amount: int64(totalAmount * 100)}
		order.TaxAmount = models.Currency{Amount: int64(taxAmount * 100)}

		orders = append(orders, &order)
	}

	return orders, nil
}
