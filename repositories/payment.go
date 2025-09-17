package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/types/stripe"
)

type PaymentRepository interface {
	GetOrder(ctx context.Context, paymentIntentID string) (types.Order, error)
	SaveEvent(ctx context.Context, event stripe.Event) error
	MarkOrderAsPaid(ctx context.Context, orderID string, object stripe.PaymentIntent) error
}

type paymentRepository struct {
	*sql.DB
}

func (r *paymentRepository) SaveEvent(ctx context.Context, event stripe.Event) error {
	query := `
	  INSERT INTO stripe_events (
		id,
		event_type,
		payload,
		processed_at
	  ) VALUES ($1, $2, $3, $4)
	`
	payload, err := json.Marshal(event.Data.Object)
	if err != nil {
		return err
	}
	_, err = r.ExecContext(ctx, query,
		event.ID,
		event.Type,
		payload,
		time.Unix(event.Created, 0).UTC(),
	)
	return err
}

func (r *paymentRepository) GetOrder(ctx context.Context, paymentIntentID string) (types.Order, error) {
	query := `
		SELECT
			o.id,
			o.user_id,
			o.email,
			o.currency,
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
			o.created_at,
			o.updated_at
		FROM orders o
		JOIN addresses a ON o.address_id = a.id
		WHERE o.stripe_payment_intent->>'id' = $1
	` // TODO find a way to include order ID metadata in payment intent events -- then we can remove payment_intent entirely from order table

	// Execute the query
	var order types.Order
	order.Address = &types.Address{}
	err := r.QueryRowContext(ctx, query, paymentIntentID).Scan(
		&order.ID,
		&order.UserID,
		&order.Email,
		&order.Currency,
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
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return order, types.ErrNotFound
	}
	if err != nil {
		return order, err
	}

	return order, nil
}

func (r *paymentRepository) MarkOrderAsPaid(ctx context.Context, orderID string, object stripe.PaymentIntent) error {
	query := `
		UPDATE orders
		SET
			status = 'paid',
			stripe_payment_intent = $1,
			updated_at = NOW()
		WHERE id = $2
	`
	payload, err := json.Marshal(object)
	if err != nil {
		return fmt.Errorf("failed to marshal Stripe payment intent: %w", err)
	}
	_, err = r.ExecContext(ctx, query, payload, orderID)
	return err
}

func NewPaymentRepository(db *sql.DB) PaymentRepository {
	return &paymentRepository{db}
}
