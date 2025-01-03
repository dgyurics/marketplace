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
	cartRepo                   repositories.CartRepository
	environment                models.Environment
	stripeBaseURL              string
	stripeSecretKey            string
	stripeWebhookSigningSecret string
}

func NewOrderService(
	orderRepo repositories.OrderRepository,
	cartRepo repositories.CartRepository,
	config models.OrderConfig,
) OrderService {
	return &orderService{
		orderRepo:                  orderRepo,
		cartRepo:                   cartRepo,
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

	if ps.environment != "production" {
		return ps.mockPaymentRequest(ctx, pi)
	}

	stripeURL := fmt.Sprintf("%s/payment_intents", ps.stripeBaseURL)
	data := fmt.Sprintf("amount=%d&currency=%s&payment_method_types[]=card", pi.Amount, pi.Currency)
	reqBody := bytes.NewBufferString(data)
	client := &http.Client{}
	reqStripe, err := http.NewRequestWithContext(ctx, "POST", stripeURL, reqBody)
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
		slog.Debug("Payment canceled", "event", event.ID, "PaymentIntentID", event.Data.Object.ID)
	case "payment_intent.payment_failed":
		slog.Info("Payment failed", "event", event.ID, "PaymentIntentID", event.Data.Object.ID)
	default:
		slog.Debug("Unhandled event", "event", event.ID, "type", event.Type)
	}
	return nil
}

// paymentIntentCreated is called when a webhook event is received from Stripe for a payment intent creation.
// A matching payment entry should be found in the orders table
func (ps *orderService) paymentIntentCreated(ctx context.Context, event models.StripeWebhookEvent) error {
	// verify event has matching entry in payment table
	order := &models.Order{PaymentIntentID: event.Data.Object.ID}
	err := ps.orderRepo.GetOrder(ctx, order)
	if err != nil {
		return err
	}
	if order.Status != "pending" {
		return nil // do nothing if order not pending
	}
	if order.TotalAmount != event.Data.Object.Amount {
		return fmt.Errorf("payment intent amount does not match expected amount")
	}
	if order.Currency != event.Data.Object.Currency {
		return fmt.Errorf("payment intent currency does not match expected currency")
	}
	return nil
}

// paymentIntentSucceeded is called when a webhook event is received from Stripe for a payment intent success
// A matching payment entry should be found in the orders table
func (ps *orderService) paymentIntentSucceeded(ctx context.Context, event models.StripeWebhookEvent) error {
	paymentIntent := event.Data.Object
	order := &models.Order{PaymentIntentID: paymentIntent.ID}
	err := ps.orderRepo.GetOrder(ctx, order)
	if err != nil {
		return err
	}
	if order.Status != "pending" {
		return nil // do nothing if order not pending
	}
	if order.TotalAmount != paymentIntent.Amount {
		slog.Error("Payment intent amount does not match expected amount", "order_id", order.ID)
		return fmt.Errorf("payment intent amount does not match expected amount")
	}
	if order.Currency != paymentIntent.Currency {
		slog.Error("Payment intent currency does not match expected currency", "order_id", order.ID)
		return fmt.Errorf("payment intent currency does not match expected currency")
	}

	// mark order as paid
	order.Status = "paid"
	if err = ps.orderRepo.UpdateOrder(ctx, order); err != nil {
		slog.Error("Error updating order", "order_id", order.ID, "error", err)
		return err
	}

	// clear cart
	if err = ps.cartRepo.ClearCart(ctx, order.UserID); err != nil {
		slog.Error("Error clearing cart", "user_id", order.UserID, "error", err)
		return err
	}
	return nil
}

// unixTimestampToTime converts [timestamp], Unix timestamp string to a time.Time object.
func unixTimestampToTime(timestamp string) (time.Time, error) {
	seconds, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(seconds, 0), nil
}

func (ps *orderService) cancelPaymentIntent(ctx context.Context, paymentIntentID string) error {
	if paymentIntentID == "" {
		return errors.New("missing payment intent ID")
	}

	if ps.environment != "production" {
		return nil // no-op in development
	}

	stripeURL := fmt.Sprintf("%s/payment_intents/%s/cancel", ps.stripeBaseURL, paymentIntentID)
	client := &http.Client{}

	reqStripe, err := http.NewRequestWithContext(ctx, "POST", stripeURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create Stripe cancel request: %w", err)
	}
	reqStripe.SetBasicAuth(ps.stripeSecretKey, "")
	reqStripe.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(reqStripe)
	if err != nil {
		slog.Error("Error sending cancel request to Stripe API", "url", stripeURL, "error", err)
		return fmt.Errorf("failed to send Stripe cancel request: %w", err)
	}
	defer resp.Body.Close()

	// Handle response
	if resp.StatusCode != http.StatusOK {
		slog.Error("Stripe API returned error on payment intent cancel", "status", resp.StatusCode, "url", stripeURL)
		return fmt.Errorf("cancel payment intent request failed with status %d", resp.StatusCode)
	}

	// Parse response
	var result struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		slog.Error("Failed to parse Stripe cancel response", "error", err)
		return fmt.Errorf("failed to decode Stripe cancel response: %w", err)
	}

	// Ensure the intent was canceled
	if result.Status != "canceled" {
		slog.Error("Payment intent not canceled", "id", result.ID, "status", result.Status)
		return fmt.Errorf("payment intent %s was not canceled, current status: %s", result.ID, result.Status)
	}

	slog.Debug("Payment intent canceled successfully", "id", result.ID)
	return nil
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

	// Cancel open/unpaid payment intent, if any
	order := &models.Order{UserID: userID}
	s.orderRepo.GetOrder(ctx, order)
	if order.Status == "pending" && order.PaymentIntentID != "" {
		go func() {
			s.cancelPaymentIntent(ctx, order.PaymentIntentID)
		}()
	}

	// Create order
	order, err := s.orderRepo.CreateOrder(ctx, userID)
	if err != nil {
		slog.Error("Error creating order", "user_id", userID, "error", err)
		return models.PaymentIntent{}, err
	}

	slog.Info("Order created",
		"order_id", order.ID,
		"user_id", order.UserID,
		"amount", order.Amount,
		"tax_amount", order.TaxAmount,
		"total_amount", order.TotalAmount,
	)

	// Create payment intent
	pi := &models.PaymentIntent{
		Amount:   order.TotalAmount,
		Currency: "usd",
	}
	if err = s.createPaymentIntent(ctx, pi); err != nil {
		return models.PaymentIntent{}, err
	}

	// Update order payment intent ID
	order.Status = "" // reset status
	order.PaymentIntentID = pi.ID
	if err = s.orderRepo.UpdateOrder(ctx, order); err != nil {
		slog.Error(
			"Error updating order",
			"order_id", order.ID,
			"status", order.Status,
			"payment_intent_id", pi.ID,
			"error", err,
		)
		return models.PaymentIntent{}, err
	}

	return *pi, err
}
