package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/dgyurics/marketplace/models"
)

type PaymentRepository interface {
	SavePayment(ctx context.Context, paymentResponse *models.PaymentIntentResponse, orderID string) error
}

type paymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) SavePayment(ctx context.Context, paymentResponse *models.PaymentIntentResponse, orderID string) error {
	if paymentResponse == nil {
		return errors.New("paymentResponse cannot be nil")
	}

	query := `
		INSERT INTO payments (
			id,
			payment_intent_id,
			client_secret,
			amount,
			currency,
			status,
			order_id,
			created_at,
			updated_at
		)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		paymentResponse.ID,           // payment_intent_id
		paymentResponse.ClientSecret, // client_secret
		paymentResponse.Amount,       // amount
		paymentResponse.Currency,     // currency
		paymentResponse.Status,       // status
		orderID,                      // order_id
		time.Now(),                   // created_at
		time.Now(),                   // updated_at
	)
	return err
}
