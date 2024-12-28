package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/dgyurics/marketplace/models"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, userID string) (order *models.Order, err error)
	CompleteOrderPayment(ctx context.Context, orderID string) error
	CreatePayment(ctx context.Context, payment models.Payment) error
	GetPayment(ctx context.Context, paymentIntentID string) (*models.Payment, error)
	CreateWebhookEvent(ctx context.Context, event models.StripeWebhookEvent) error
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

func (r *orderRepository) CompleteOrderPayment(ctx context.Context, orderID string) error {
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

func (r *orderRepository) CreateWebhookEvent(ctx context.Context, event models.StripeWebhookEvent) error {
	query := `
		INSERT INTO webhook_events (
			id,
			event_type,
			payload,
			processed_at
		)
		VALUES ($1, $2, $3, $4)
	`
	payload, err := json.Marshal(event.Data.Object)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, query,
		event.ID,
		event.Type,
		payload,
		time.Unix(event.Created, 0),
	)
	return err
}

func (r *orderRepository) CreatePayment(ctx context.Context, payment models.Payment) error {
	query := `
		INSERT INTO payments (
			payment_intent_id,
			client_secret,
			amount,
			currency,
			status,
			order_id
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		payment.PaymentIntentID,
		payment.ClientSecret,
		payment.Amount,
		payment.Currency,
		payment.Status,
		payment.OrderID,
	)
	return err
}

func (r *orderRepository) GetPayment(ctx context.Context, paymentIntentID string) (*models.Payment, error) {
	var payment models.Payment
	if err := r.db.QueryRowContext(ctx, `
		SELECT
			payment_intent_id,
			client_secret,
			amount,
			currency,
			status,
			order_id,
			created_at,
			updated_at
		FROM payments
		WHERE payment_intent_id = $1
	`, paymentIntentID).Scan(
		&payment.PaymentIntentID,
		&payment.ClientSecret,
		&payment.Amount,
		&payment.Currency,
		&payment.Status,
		&payment.OrderID,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &payment, nil
}
