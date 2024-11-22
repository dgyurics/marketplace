package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	mathrand "math/rand"

	"github.com/dgyurics/marketplace/models"
)

type PaymentService interface {
	SendPaymentRequest(ctx context.Context, req models.PaymentIntentRequest) (models.PaymentIntentResponse, error)
	RetrievePaymentIntent(ctx context.Context, paymentIntentID string) (models.PaymentIntent, error)
}

type paymentService struct{}

func NewPaymentService() PaymentService {
	return &paymentService{}
}

func (ps *paymentService) SendPaymentRequest(ctx context.Context, req models.PaymentIntentRequest) (models.PaymentIntentResponse, error) {
	if req.TokenID == "" {
		return models.PaymentIntentResponse{}, errors.New("missing token ID")
	}
	if req.Amount.Amount <= 0 {
		return models.PaymentIntentResponse{}, errors.New("invalid amount")
	}

	// Simulate network delay with context handling
	select {
	case <-ctx.Done():
		return models.PaymentIntentResponse{}, ctx.Err()
	case <-time.After(2 * time.Second):
	}

	// Mock transaction ID generation
	transactionID := fmt.Sprintf("%s-%d", req.Provider, mathrand.Intn(1000000))

	// Simulate payment response with random success/failure
	failure := mathrand.Intn(10) == 0 // 10% chance of failure
	if failure {
		return models.PaymentIntentResponse{
			PaymentIntentID: "",
			ClientSecret:    "",
			Amount:          req.Amount,
			Currency:        req.Currency,
			Status:          "incorrect_payment_details",
		}, nil
	}

	return models.PaymentIntentResponse{
		PaymentIntentID: transactionID,
		ClientSecret:    "client_secret_" + transactionID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		Status:          "success",
	}, nil
}

func (ps *paymentService) RetrievePaymentIntent(ctx context.Context, paymentIntentID string) (models.PaymentIntent, error) {
	// Placeholder for actual payment provider
	if paymentIntentID == "" {
		return models.PaymentIntent{}, errors.New("missing payment intent ID")
	}

	// Simulate network delay with context handling
	select {
	case <-ctx.Done():
		return models.PaymentIntent{}, ctx.Err()
	case <-time.After(1 * time.Second):
	}

	// Simulate payment response with random success/failure
	failure := mathrand.Intn(10) == 0 // 10% chance of failure
	if failure {
		return models.PaymentIntent{
			Status:         "not paid",
			AmountReceived: models.Currency{Amount: 0},
		}, nil
	}

	// FIXME will need to find a way to mock amount paid and TransactionID to simulate
	// scenarios where amount paid does not match the expected amount
	// or where the transaction ID is not found or does not match the expected ID

	return models.PaymentIntent{
		Status:         "paid",
		AmountReceived: models.NewCurrency(50, 0),
	}, nil
}
