package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/dgyurics/marketplace/models"
)

type PaymentRepository interface {
	SavePayment(ctx context.Context, payment models.Payment) error
	SavePaymentEvent(ctx context.Context, event models.StripeWebhookEvent) error
	GetPayment(ctx context.Context, paymentIntentID string) (*models.Payment, error)
}

type paymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) SavePaymentEvent(ctx context.Context, event models.StripeWebhookEvent) error {
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

func (r *paymentRepository) SavePayment(ctx context.Context, payment models.Payment) error {
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

func (r *paymentRepository) GetPayment(ctx context.Context, paymentIntentID string) (*models.Payment, error) {
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
