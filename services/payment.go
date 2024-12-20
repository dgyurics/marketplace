package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	mathrand "math/rand"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/repositories"
)

const (
	tolerance = time.Minute * 5 // Maximum allowed time difference between Stripe's timestamp and server time
)

type PaymentService interface {
	SendPaymentRequest(ctx context.Context, req models.PaymentIntentRequest) (models.PaymentIntentResponse, error)
	RetrievePaymentIntent(ctx context.Context, paymentIntentID string) (models.PaymentIntent, error)
	SavePayment(ctx context.Context, payment models.Payment) error
	VerifyWebhookSignature(payload []byte, sigHeader string) error
	ProcessWebhookEvent(ctx context.Context, event models.StripeWebhookEvent) error
}

type paymentService struct {
	paymentRepo                repositories.PaymentRepository
	orderRepo                  repositories.OrderRepository
	environment                string
	stripeBaseURL              string
	stripeSecretKey            string
	stripeWebhookSigningSecret string
}

func NewPaymentService(
	paymentRepo repositories.PaymentRepository,
	orderRepo repositories.OrderRepository,
	environment,
	stripeBaseURL,
	stripeSecretKey,
	stripeWebhookSigningSecret string,
) PaymentService {
	return &paymentService{
		paymentRepo:                paymentRepo,
		orderRepo:                  orderRepo,
		environment:                environment,
		stripeBaseURL:              stripeBaseURL,
		stripeSecretKey:            stripeSecretKey,
		stripeWebhookSigningSecret: stripeWebhookSigningSecret,
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
	"pending":   true,
	"paid":      true,
	"cancelled": true,
	"refunded":  true,
}

func isValidStatus(status string) bool {
	return validStatuses[status]
}

func (ps *paymentService) SavePayment(ctx context.Context, payment models.Payment) error {
	if payment.PaymentIntentID == "" {
		return errors.New("payment intent ID is required")
	}
	if payment.OrderID == "" {
		return errors.New("order ID is required")
	}
	if !isValidStatus(payment.Status) {
		return errors.New("invalid payment status")
	}
	return ps.paymentRepo.SavePayment(ctx, payment)
}

func (ps *paymentService) VerifyWebhookSignature(payload []byte, sigHeader string) error {
	// Split the signature header into components (e.g. "t=timestamp,v1=signature,v0=signature")
	parts := strings.Split(sigHeader, ",")
	if len(parts) < 2 {
		return errors.New("invalid signature header")
	}

	var timestamp string
	var signatures [][]byte
	for _, part := range parts {
		if strings.HasPrefix(part, "t=") {
			timestamp = part[2:]
		} else if strings.HasPrefix(part, "v1=") {
			decodedSignature, err := hex.DecodeString(part[3:])
			if err == nil {
				signatures = append(signatures, decodedSignature)
			}
		}
	}

	if timestamp == "" || len(signatures) == 0 {
		return errors.New("missing timestamp or signature")
	}

	ts, err := unixTimestampToTime(timestamp)
	if err != nil {
		return errors.New("invalid timestamp")
	} else if time.Since(ts) > tolerance {
		return errors.New("timestamp is too old")
	}

	expectedSignature := ComputeSignature(ts, payload, ps.stripeWebhookSigningSecret)

	for _, signature := range signatures {
		// use a constant-time comparison function to mitigate timing attacks
		if hmac.Equal(signature, expectedSignature) {
			return nil
		}
	}

	return errors.New("signature verification failed: no matching v1 signature found")
}

// Stripe events can be triggered out of order, as well as be duplicated. This function should be idempotent.
func (ps *paymentService) ProcessWebhookEvent(ctx context.Context, event models.StripeWebhookEvent) error {
	if event.Data == nil {
		return errors.New("missing event data")
	}
	if event.Data.Object.ID == "" {
		return errors.New("missing payment intent ID")
	}
	switch event.Type {
	case "payment_intent.created":
		return ps.PaymentIntentCreated(ctx, event)
	case "payment_intent.succeeded":
		return ps.PaymentIntentSucceeded(ctx, event)
	case "payment_intent.payment_failed":
		return ps.PaymentIntentPaymentFailed(ctx, event)
	default:
		// Placeholder for logging other events
	}
	return nil
}

// To be called when a webhook event is received from Stripe for a payment intent that has been created
func (ps *paymentService) PaymentIntentCreated(ctx context.Context, event models.StripeWebhookEvent) error {
	// save raw event
	if err := ps.paymentRepo.SavePaymentEvent(ctx, event); err != nil {
		return err
	}
	// verify event has matching entry in payment table
	paymentIntent := event.Data.Object
	payments, err := ps.paymentRepo.GetPaymentsByPaymentIntentID(ctx, paymentIntent.ID)
	if err != nil {
		return err
	}
	for _, payment := range payments {
		if payment.Status == "pending" &&
			payment.Amount == paymentIntent.Amount &&
			payment.Currency == paymentIntent.Currency &&
			payment.ClientSecret == paymentIntent.ClientSecret {
			return nil
		}
	}
	return fmt.Errorf("no pending payment found for intent %s", paymentIntent.ID)
}

// To be called when a webhook event is received from Stripe for a payment intent success
func (ps *paymentService) PaymentIntentSucceeded(ctx context.Context, event models.StripeWebhookEvent) error {
	// save raw event
	if err := ps.paymentRepo.SavePaymentEvent(ctx, event); err != nil {
		return err
	}
	paymentIntent := event.Data.Object
	payments, err := ps.paymentRepo.GetPaymentsByPaymentIntentID(ctx, paymentIntent.ID)
	if err != nil {
		return err
	}
	paid := false
	orderID := ""
	// verify entry in payment table exists, with status pending
	for _, payment := range payments {
		if payment.Status == "paid" {
			return nil // do nothing if order is already marked as paid
		}
		if payment.Status == "pending" {
			if payment.Amount != paymentIntent.Amount {
				return fmt.Errorf("payment intent amount does not match expected amount")
			}
			if payment.Currency != paymentIntent.Currency {
				return fmt.Errorf("payment intent currency does not match expected currency")
			}
			if payment.ClientSecret != paymentIntent.ClientSecret {
				return fmt.Errorf("payment intent client secret does not match expected client secret")
			}
			paid = true
			orderID = payment.OrderID
		}
	}
	// if no pending payment found, log and return error
	if !paid {
		return fmt.Errorf("no pending payment found for intent %s", paymentIntent.ID)
	}
	// update payment status to paid
	if err := ps.paymentRepo.SavePayment(ctx, models.Payment{
		PaymentIntentID: paymentIntent.ID,
		ClientSecret:    paymentIntent.ClientSecret,
		Amount:          paymentIntent.Amount,
		Currency:        paymentIntent.Currency,
		Status:          "paid",
		OrderID:         orderID,
	}); err != nil {
		return err
	}
	// update order status to paid
	return ps.orderRepo.UpdateOrder(ctx, orderID, "paid")
}

// To be called when a webhook event is received from Stripe for a payment intent failure
func (ps *paymentService) PaymentIntentPaymentFailed(ctx context.Context, event models.StripeWebhookEvent) error {
	// save raw event
	return ps.paymentRepo.SavePaymentEvent(ctx, event)
}

// unixTimestampToTime converts a Unix timestamp string to a time.Time object.
// TODO: move this to a common utility package
func unixTimestampToTime(timestamp string) (time.Time, error) {
	seconds, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(seconds, 0), nil
}

// ComputeSignature computes a webhook signature using Stripe's v1 signing
// method.
//
// See https://stripe.com/docs/webhooks#signatures for more information.
func ComputeSignature(t time.Time, payload []byte, secret string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%d", t.Unix())))
	mac.Write([]byte("."))
	mac.Write(payload)
	return mac.Sum(nil)
}

// {
//   "id": "evt_3QWuXnAyEBQV0eUX29XrGRsW",
//   "object": "event",
//   "api_version": "2024-11-20.acacia",
//   "created": 1734417972,
//   "data": {
//     "object": {
//       "id": "pi_3QWuXnAyEBQV0eUX2OHUOpug",
//       "object": "payment_intent",
//       "amount": 2000,
//       "amount_capturable": 0,
//       "amount_details": {
//         "tip": {
//         }
//       },
//       "amount_received": 2000,
//       "application": null,
//       "application_fee_amount": null,
//       "automatic_payment_methods": null,
//       "canceled_at": null,
//       "cancellation_reason": null,
//       "capture_method": "automatic_async",
//       "client_secret": "pi_3QWuXnAyEBQV0eUX2OHUOpug_secret_DPHJ1Df0Ps9EG1MNgJj8K6OC8",
//       "confirmation_method": "automatic",
//       "created": 1734417971,
//       "currency": "usd",
//       "customer": null,
//       "description": "(created by Stripe CLI)",
//       "invoice": null,
//       "last_payment_error": null,
//       "latest_charge": "ch_3QWuXnAyEBQV0eUX2OwuNCYr",
//       "livemode": false,
//       "metadata": {
//       },
//       "next_action": null,
//       "on_behalf_of": null,
//       "payment_method": "pm_1QWuXnAyEBQV0eUX1Eeu6yu2",
//       "payment_method_configuration_details": null,
//       "payment_method_options": {
//         "card": {
//           "installments": null,
//           "mandate_options": null,
//           "network": null,
//           "request_three_d_secure": "automatic"
//         }
//       },
//       "payment_method_types": [
//         "card"
//       ],
//       "processing": null,
//       "receipt_email": null,
//       "review": null,
//       "setup_future_usage": null,
//       "shipping": {
//         "address": {
//           "city": "San Francisco",
//           "country": "US",
//           "line1": "510 Townsend St",
//           "line2": null,
//           "postal_code": "94103",
//           "state": "CA"
//         },
//         "carrier": null,
//         "name": "Jenny Rosen",
//         "phone": null,
//         "tracking_number": null
//       },
//       "source": null,
//       "statement_descriptor": null,
//       "statement_descriptor_suffix": null,
//       "status": "succeeded",
//       "transfer_data": null,
//       "transfer_group": null
//     }
//   },
//   "livemode": false,
//   "pending_webhooks": 2,
//   "request": {
//     "id": "req_WxCltYVDry7wPS",
//     "idempotency_key": "f124f4f1-6cc4-4fb0-a934-3832e46f382f"
//   },
//   "type": "payment_intent.succeeded"
// }
