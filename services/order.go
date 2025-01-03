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
	"log/slog"
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

type OrderService interface {
	CreateOrder(ctx context.Context) (models.PaymentIntent, error)
	VerifyWebhookEventSignature(payload []byte, sigHeader string) error
	ProcessWebhookEvent(ctx context.Context, event models.StripeWebhookEvent) error
}

type orderService struct {
	orderRepo                  repositories.OrderRepository
	environment                models.Environment
	stripeBaseURL              string
	stripeSecretKey            string
	stripeWebhookSigningSecret string
}

func NewOrderService(
	orderRepo repositories.OrderRepository,
	config models.OrderConfig,
) OrderService {
	return &orderService{
		orderRepo:                  orderRepo,
		environment:                config.Envirnment,
		stripeBaseURL:              config.StripeBaseURL,
		stripeSecretKey:            config.StripeSecretKey,
		stripeWebhookSigningSecret: config.StripeWebhookSigningSecret,
	}
}

// Call Stripe API to create a payment intent
func (ps *orderService) createPaymentIntent(ctx context.Context, pi *models.PaymentIntent) error {
	if pi.Currency == "" {
		return errors.New("missing currency")
	}
	if pi.Amount <= 0 {
		return errors.New("missing or invalid amount")
	}

	if ps.environment == "test" || ps.environment == "development" {
		return ps.mockPaymentRequest(ctx, pi)
	}

	stripeURL := fmt.Sprintf("%s/payment_intents", ps.stripeBaseURL)
	data := fmt.Sprintf("amount=%d&currency=%s&payment_method_types[]=card", pi.Amount, pi.Currency)
	reqBody := bytes.NewBufferString(data)
	client := &http.Client{}
	reqStripe, err := http.NewRequest("POST", stripeURL, reqBody)
	if err != nil {
		return err
	}
	reqStripe.SetBasicAuth(ps.stripeSecretKey, "")
	reqStripe.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(reqStripe)
	if err != nil {
		slog.Error("Error sending request to Stripe API", "url", stripeURL, "error", err)
		return err
	}
	defer resp.Body.Close()

	// Handle Stripe API response
	if resp.StatusCode != http.StatusOK {
		slog.Error("Stripe API returned status", "status", resp.StatusCode, "url", stripeURL)
		return fmt.Errorf("stripe API returned status %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(pi)
}

func (ps *orderService) cancelPaymentIntent(ctx context.Context, pi *models.PaymentIntent) error {
	// TODO: implement this method
	return nil
}

func (ps *orderService) mockPaymentRequest(ctx context.Context, pi *models.PaymentIntent) error {
	// Simulate network delay with context handling
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(2 * time.Second):
	}

	// Simulate failure or success
	failure := mathrand.Intn(10) == 0 // 10% chance of failure
	if failure {
		pi.Status = "failed"
		pi.Error = "incorrect_payment_details"
	} else {
		pi.Status = "pending"
		pi.Error = ""
		pi.ID = fmt.Sprintf("fake_payment_intent_%d", mathrand.Intn(1000000))
		pi.ClientSecret = fmt.Sprintf("%s-%d", "fake_secret", mathrand.Intn(1000000))
	}

	return nil
}

// [payload] raw request body
// [sigHeader] value of the Stripe-Signature header, in the format "t=timestamp,v1=signature,v1=signature,..."
func (ps *orderService) VerifyWebhookEventSignature(payload []byte, sigHeader string) error {
	parts := strings.Split(sigHeader, ",")
	if len(parts) < 2 {
		slog.Error("Invalid signature header", "header", sigHeader)
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

	if timestamp == "" {
		slog.Error("Timestamp missing from signature header", "header", sigHeader)
		return errors.New("missing timestamp")
	}

	if len(signatures) == 0 {
		slog.Error("Signature missing from signature header", "header", sigHeader)
		return errors.New("missing signature")
	}

	ts, err := unixTimestampToTime(timestamp)
	if err != nil {
		slog.Error("Error parsing timestamp", "timestamp", timestamp)
		return errors.New("invalid timestamp")
	} else if time.Since(ts) > tolerance {
		slog.Error("Timestamp is too old", "timestamp", timestamp)
		return errors.New("timestamp is too old")
	}

	// Compare expected signature with provided signatures
	// Use a constant-time comparison function to mitigate timing attacks
	// If a matching signature is found, return nil
	expectedSignature := ComputeSignature(ts, payload, ps.stripeWebhookSigningSecret)
	for _, signature := range signatures {
		if hmac.Equal(signature, expectedSignature) {
			return nil
		}
	}

	slog.Error("Signature verification failed", "signatures", signatures)
	return errors.New("signature verification failed: no matching v1 signature found")
}

// ProcessWebhookEvent processes a Stripe webhook event. It is idompotent
func (ps *orderService) ProcessWebhookEvent(ctx context.Context, event models.StripeWebhookEvent) error {
	// save raw event data
	if err := ps.orderRepo.CreateWebhookEvent(ctx, event); err != nil {
		return err
	}
	switch event.Type {
	case "payment_intent.created":
		return ps.paymentIntentCreated(ctx, event)
	case "payment_intent.succeeded":
		return ps.paymentIntentSucceeded(ctx, event)
	case "payment_intent.canceled":
		slog.Warn("Payment canceled", "event", event.ID, "PaymentIntentID", event.Data.Object.ID)
	case "payment_intent.payment_failed":
		slog.Error("Payment failed", "event", event.ID, "PaymentIntentID", event.Data.Object.ID)
	default:
		slog.Debug("Unhandled event", "event", event.ID, "type", event.Type)
	}
	return nil
}

// paymentIntentCreated is called when a webhook event is received from Stripe for a payment intent creation.
// A matching payment entry should be found in the orders table
func (ps *orderService) paymentIntentCreated(ctx context.Context, event models.StripeWebhookEvent) error {
	// verify event has matching entry in payment table
	paymentIntent := event.Data.Object
	payment, err := ps.orderRepo.GetPayment(ctx, paymentIntent.ID)
	if err != nil {
		return err
	}
	if payment.Status != "pending" {
		return nil // do nothing if order not pending
	}
	if payment.Amount != paymentIntent.Amount {
		return fmt.Errorf("payment intent amount does not match expected amount")
	}
	if payment.Currency != paymentIntent.Currency {
		return fmt.Errorf("payment intent currency does not match expected currency")
	}
	if payment.ClientSecret != paymentIntent.ClientSecret {
		return fmt.Errorf("payment intent client secret does not match expected client secret")
	}
	return nil
}

// paymentIntentSucceeded is called when a webhook event is received from Stripe for a payment intent success
// A matching payment entry should be found in the orders table
func (ps *orderService) paymentIntentSucceeded(ctx context.Context, event models.StripeWebhookEvent) error {
	paymentIntent := event.Data.Object
	payment, err := ps.orderRepo.GetPayment(ctx, paymentIntent.ID)
	if err != nil {
		return err
	}
	if payment.Status != "pending" {
		return nil // do nothing if order not pending
	}
	if payment.Amount != paymentIntent.Amount {
		return fmt.Errorf("payment intent amount does not match expected amount")
	}
	if payment.Currency != paymentIntent.Currency {
		return fmt.Errorf("payment intent currency does not match expected currency")
	}
	if payment.ClientSecret != paymentIntent.ClientSecret {
		return fmt.Errorf("payment intent client secret does not match expected client secret")
	}

	// complete order payment flow
	return ps.orderRepo.CompleteOrderPayment(ctx, payment.OrderID)
}

// unixTimestampToTime converts [timestamp], Unix timestamp string to a time.Time object.
func unixTimestampToTime(timestamp string) (time.Time, error) {
	seconds, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(seconds, 0), nil
}

// cancelPendingOrders cancels any unpaid orders for the specified [userID].
func (s *orderService) cancelUnpaidOrders(ctx context.Context, userID string) {
	// TODO: implement this method
	// call cancelPaymentIntent
}

// ComputeSignature computes an API request signature using Stripe's v1 signing method.
// [t] timestamp of the event
// [payload] is the raw request body
// [secret] webhook signing secret.
// See https://stripe.com/docs/webhooks#signatures for more information.
func ComputeSignature(t time.Time, payload []byte, secret string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%d", t.Unix())))
	mac.Write([]byte("."))
	mac.Write(payload)
	return mac.Sum(nil)
}

func (s *orderService) CreateOrder(ctx context.Context) (models.PaymentIntent, error) {
	var userID = getUserID(ctx)

	// Cancel unpaid orders
	s.cancelUnpaidOrders(ctx, userID)

	// Create order
	order, err := s.orderRepo.CreateOrder(ctx, userID)
	if err != nil {
		slog.Error("Error creating order", "user_id", userID, "error", err)
		return models.PaymentIntent{}, err
	}

	// Create payment intent
	pi := models.PaymentIntent{
		Amount:   order.TotalAmount + order.TaxAmount,
		Currency: "usd",
	}
	if err = s.createPaymentIntent(ctx, &pi); err != nil {
		return models.PaymentIntent{}, err
	}

	// Save payment intent details
	err = s.orderRepo.CreatePayment(ctx, models.Payment{
		PaymentIntentID: pi.ID,
		ClientSecret:    pi.ClientSecret,
		Amount:          pi.Amount,
		Currency:        pi.Currency,
		Status:          "pending",
		OrderID:         order.ID,
	})

	return pi, err
}
