package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	mathrand "math/rand"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/repositories"
)

type PaymentService interface {
	SendPaymentRequest(ctx context.Context, req models.PaymentIntentRequest) (models.PaymentIntentResponse, error)
	RetrievePaymentIntent(ctx context.Context, paymentIntentID string) (models.PaymentIntent, error)
	SavePayment(ctx context.Context, paymentResponse models.PaymentIntentResponse, orderID string) error
}

type paymentService struct {
	repo            repositories.PaymentRepository
	environment     string
	stripeBaseURL   string
	stripeSecretKey string
}

func NewPaymentService(
	repo repositories.PaymentRepository,
	environment,
	stripeBaseURL,
	stripeSecretKey string,
) PaymentService {
	return &paymentService{
		repo:            repo,
		environment:     environment,
		stripeBaseURL:   stripeBaseURL,
		stripeSecretKey: stripeSecretKey,
	}
}

func (ps *paymentService) SendPaymentRequest(ctx context.Context, req models.PaymentIntentRequest) (models.PaymentIntentResponse, error) {
	if req.Currency == "" {
		return models.PaymentIntentResponse{}, errors.New("missing currency")
	}
	if req.Amount.Amount <= 0 {
		return models.PaymentIntentResponse{}, errors.New("missing or invalid amount")
	}

	if ps.environment == "test" || ps.environment == "development" {
		return ps.MockPaymentRequest(ctx, req)
	}

	stripeURL := fmt.Sprintf("%s/payment_intents", ps.stripeBaseURL)
	data := fmt.Sprintf("amount=%d&currency=%s", req.Amount.Amount, req.Currency)
	reqBody := bytes.NewBufferString(data)
	client := &http.Client{}
	reqStripe, err := http.NewRequest("POST", stripeURL, reqBody)
	if err != nil {
		return models.PaymentIntentResponse{}, err
	}
	reqStripe.SetBasicAuth(ps.stripeSecretKey, "")
	reqStripe.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(reqStripe)
	if err != nil {
		return models.PaymentIntentResponse{}, err
	}
	defer resp.Body.Close()

	// Handle Stripe API response
	if resp.StatusCode != http.StatusOK {
		return models.PaymentIntentResponse{}, fmt.Errorf("stripe API returned status %d", resp.StatusCode)
	}

	var paymentIntent models.PaymentIntentResponse
	err = json.NewDecoder(resp.Body).Decode(&paymentIntent)
	return paymentIntent, err
}

func (ps *paymentService) MockPaymentRequest(ctx context.Context, req models.PaymentIntentRequest) (models.PaymentIntentResponse, error) {
	// Simulate network delay with context handling
	select {
	case <-ctx.Done():
		return models.PaymentIntentResponse{}, ctx.Err()
	case <-time.After(2 * time.Second):
	}

	// Simulate possible failure
	failure := mathrand.Intn(10) == 0 // 10% chance of failure
	if failure {
		return models.PaymentIntentResponse{
			ID:           fmt.Sprintf("fake_payment_intent_%d", mathrand.Intn(1000000)),
			Amount:       req.Amount.Amount,
			Currency:     req.Currency,
			Status:       "failed",
			ClientSecret: "",
			Error:        "incorrect_payment_details",
		}, nil
	}

	// Simulate successful payment
	return models.PaymentIntentResponse{
		ID:           fmt.Sprintf("fake_payment_intent_%d", mathrand.Intn(1000000)),
		Amount:       req.Amount.Amount,
		Currency:     req.Currency,
		Status:       "pending",
		ClientSecret: fmt.Sprintf("%s-%d", "fake_secret", mathrand.Intn(1000000)),
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

var validStatuses = map[string]bool{
	"requires_payment_method": true,
	"requires_confirmation":   true,
	"requires_action":         true,
	"processing":              true,
	"succeeded":               true,
	"canceled":                true,
}

func isValidStatus(status string) bool {
	return validStatuses[status]
}

func (ps *paymentService) SavePayment(ctx context.Context, paymentResponse models.PaymentIntentResponse, orderID string) error {
	if paymentResponse.ID == "" {
		return errors.New("payment response ID is required")
	}
	if orderID == "" {
		return errors.New("order ID is required")
	}

	if !isValidStatus(paymentResponse.Status) {
		// TODO log invalid status
		paymentResponse.Status = "unknown"
	}

	return ps.repo.SavePayment(ctx, &paymentResponse, orderID)
}
