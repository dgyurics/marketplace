package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	mathrand "math/rand"

	"github.com/dgyurics/marketplace/models"
)

// Placeholder for actual payment provider

type PaymentProvider string

const (
	Stripe PaymentProvider = "stripe"
	PayPal PaymentProvider = "paypal"
)

type PaymentIntentRequest struct {
	Provider PaymentProvider
	Amount   models.Currency
	Currency string
	TokenID  string
}

type PaymentIntentResponse struct {
	Status        string
	TransactionID string
	ErrorMessage  string
}

type PaymentIntent struct {
	Status         string
	AmountReceived models.Currency
}

func SendPaymentRequest(req PaymentIntentRequest) (PaymentIntentResponse, error) {
	if req.TokenID == "" {
		return PaymentIntentResponse{}, errors.New("missing token ID")
	}
	if req.Amount.Amount <= 0 {
		return PaymentIntentResponse{}, errors.New("invalid amount")
	}

	// Simulate request processing time
	time.Sleep(2 * time.Second) // Mimic network delay

	// Mock transaction ID generation
	transactionID := fmt.Sprintf("%s-%d", req.Provider, mathrand.Intn(1000000))

	// Simulate payment response with random success/failure
	failure := mathrand.Intn(10) == 0 // 10% chance of failure
	if failure {
		return PaymentIntentResponse{
			Status:        "failed",
			TransactionID: "",
			ErrorMessage:  "Payment failed",
		}, nil
	}

	return PaymentIntentResponse{
		Status:        "success",
		TransactionID: transactionID,
		ErrorMessage:  "",
	}, nil
}

func RetrievePaymentIntent(ctx context.Context, paymentIntentID string) (PaymentIntent, error) {
	// Placeholder for actual payment provider
	if paymentIntentID == "" {
		return PaymentIntent{}, errors.New("missing payment intent ID")
	}

	// Simulate network delay
	time.Sleep(1 * time.Second)

	// Simulate payment response with random success/failure
	failure := mathrand.Intn(10) == 0 // 10% chance of failure
	if failure {
		return PaymentIntent{
			Status:         "not paid",
			AmountReceived: models.Currency{Amount: 0},
		}, nil
	}

	// FIXME will need to find a way to mock amount paid and TransactionID to simulate
	// scenarios where amount paid does not match the expected amount
	// or where the transaction ID is not found or does not match the expected ID

	return PaymentIntent{
		Status:         "paid",
		AmountReceived: models.NewCurrency(50, 0),
	}, nil
}
