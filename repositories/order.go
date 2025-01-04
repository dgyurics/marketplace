package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgyurics/marketplace/models"
)

type OrderRepository interface {
	GetOrder(ctx context.Context, order *models.Order) error
	CreateOrder(ctx context.Context, userID string) (*models.Order, error)
	UpdateOrder(ctx context.Context, order *models.Order) error
	CreateWebhookEvent(ctx context.Context, event models.StripeWebhookEvent) error
}

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepository{db: db}
}

// CreateOrder creates a new order from the user's cart
func (r *orderRepository) CreateOrder(ctx context.Context, userID string) (*models.Order, error) {
	// 1) Create or update the order from the user's cart
	query := "SELECT update_or_create_order_from_cart($1)"
	var orderID string
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&orderID)
	if err != nil {
		return nil, err
	}

	// 2) Retrieve the new or updated order
	query = `
	  SELECT
			id,
			user_id,
			currency,
			amount,
			tax_amount,
			total_amount,
			status,
			payment_intent_id,
			created_at,
			updated_at
		FROM orders
		WHERE id = $1
	`
	order := &models.Order{}
	err = r.db.QueryRowContext(ctx, query, orderID).Scan(
		&order.ID,
		&order.UserID,
		&order.Currency,
		&order.Amount,
		&order.TaxAmount,
		&order.TotalAmount,
		&order.Status,
		&order.PaymentIntentID,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve order: %w", err)
	}
	return order, nil
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

func (r *orderRepository) UpdateOrder(ctx context.Context, order *models.Order) error {
	if order.ID == "" {
		return fmt.Errorf("missing order ID")
	}

	query := `
		UPDATE orders
		SET updated_at = CURRENT_TIMESTAMP
	`
	args := []interface{}{}
	argCount := 1

	if order.Status != "" {
		query += fmt.Sprintf(", status = $%d", argCount)
		args = append(args, order.Status)
		argCount++
	}

	if order.PaymentIntentID != "" {
		query += fmt.Sprintf(", payment_intent_id = $%d", argCount)
		args = append(args, order.PaymentIntentID)
		argCount++
	}

	// Ensure there's something to update
	if len(args) == 0 {
		return fmt.Errorf("no fields to update")
	}

	// Add the WHERE clause
	query += fmt.Sprintf(" WHERE id = $%d", argCount)
	args = append(args, order.ID)

	// Execute the query
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// GetOrder retrieves an order by ID, user ID, or payment intent ID
func (r *orderRepository) GetOrder(ctx context.Context, order *models.Order) error {
	// Validate input
	if order.ID == "" && order.UserID == "" && order.PaymentIntentID == "" {
		return fmt.Errorf("at least one of order.ID, order.UserID, or order.PaymentIntentID must be provided")
	}

	query := `
		SELECT
			id,
			user_id,
			currency,
			amount,
			tax_amount,
			total_amount,
			status,
			payment_intent_id,
			created_at,
			updated_at
		FROM orders
	`
	args := []interface{}{}
	var whereClause string

	// Build the WHERE clause based on provided fields
	if order.ID != "" {
		whereClause = "WHERE id = $1"
		args = append(args, order.ID)
	} else if order.UserID != "" {
		whereClause = "WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1"
		args = append(args, order.UserID)
	} else if order.PaymentIntentID != "" {
		whereClause = "WHERE payment_intent_id = $1"
		args = append(args, order.PaymentIntentID)
	}

	// Combine query and where clause
	query += whereClause

	// Execute the query
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&order.ID,
		&order.UserID,
		&order.Currency,
		&order.Amount,
		&order.TaxAmount,
		&order.TotalAmount,
		&order.Status,
		&order.PaymentIntentID,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("order not found")
		}
		return err
	}

	return nil
}
