package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/types/stripe"
	"github.com/dgyurics/marketplace/utilities"
)

type PaymentService interface {
	EventHandler(ctx context.Context, event stripe.Event) error
	SignatureVerifier(payload []byte, sigHeader string) error
	CreatePaymentIntent(ctx context.Context, refID string, amount int64) (stripe.PaymentIntent, error)
	CancelPaymentIntent(ctx context.Context, paymentIntentID string) error
}

type paymentService struct {
	HttpClient   utilities.HTTPClient
	stripeConfig types.StripeConfig
	localeConfig types.LocaleConfig
	repo         repositories.PaymentRepository
}

func NewPaymentService(
	httpClient utilities.HTTPClient,
	stripeConfig types.StripeConfig,
	localeConfig types.LocaleConfig,
	repo repositories.PaymentRepository) PaymentService {
	return &paymentService{
		HttpClient:   httpClient,
		stripeConfig: stripeConfig,
		localeConfig: localeConfig,
		repo:         repo,
	}
}

func (s *paymentService) EventHandler(ctx context.Context, event stripe.Event) error {
	// save raw event
	if err := s.repo.SaveEvent(ctx, event); err != nil {
		slog.Error("Failed to save event", "error", err)
		return err
	}

	// process event based on type
	switch stripe.EventType(event.Type) {
	case stripe.EventTypePaymentIntentCreated:
		return s.handlePaymentIntentCreated(ctx, event)
	case stripe.EventTypePaymentIntentSucceeded:
		return s.handlePaymentIntentSucceeded(ctx, event)
	case stripe.EventTypePaymentIntentCanceled:
		slog.Debug("Payment intent canceled", "id", event.Data.Object.ID)
	case stripe.EventTypePaymentIntentPaymentFailed:
		slog.Debug("Payment intent payment failed", "id", event.Data.Object.ID)
	default:
		slog.Debug("Unhandled Stripe event type", "type", event.Type)
	}
	return nil
}

// SignatureVerifier verifies the signature of a Stripe webhook event.
// It checks the signature against the payload and the Stripe-Signature header.
// [payload] raw request body
// [sigHeader] value of the Stripe-Signature header, in the format "t=timestamp,v1=signature,v1=signature,..."
func (s *paymentService) SignatureVerifier(payload []byte, sigHeader string) error {
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
	expectedSignature := ComputeSignature(ts, payload, s.stripeConfig.WebhookSigningSecret)
	for _, signature := range signatures {
		if hmac.Equal(signature, expectedSignature) {
			return nil
		}
	}

	slog.Error("Signature verification failed", "signatures", signatures)
	return errors.New("signature verification failed: no matching v1 signature found")
}

// CreatePaymentIntent creates a new Stripe Payment Intent.
// [refID] is a unique reference ID for idempotency.
// [amount] is the amount in the smallest currency unit (e.g., cents for USD).
func (s *paymentService) CreatePaymentIntent(ctx context.Context, refID string, amount int64) (pi stripe.PaymentIntent, err error) {
	// Build request
	reqURL := fmt.Sprintf("%s/payment_intents", s.stripeConfig.BaseURL)
	payload := url.Values{
		"amount":                 {fmt.Sprintf("%d", amount)},
		"currency":               {s.localeConfig.Currency},
		"payment_method_types[]": {"card"},
	}
	reqBody := strings.NewReader(payload.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, reqBody)
	if err != nil {
		return pi, err
	}

	// Set request headers
	req.SetBasicAuth(s.stripeConfig.SecretKey, "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Idempotency-Key", fmt.Sprintf("payment-intent-%s", refID))
	req.Header.Set("Stripe-Version", s.stripeConfig.Version)

	// Execute request
	res, err := s.HttpClient.Do(req)
	if err != nil {
		return pi, err
	}
	defer res.Body.Close()

	// Handle response
	if res.StatusCode != http.StatusOK {
		slog.Error("Stripe API returned non-OK status", "status", res.StatusCode, "url", s.stripeConfig.BaseURL)
		return pi, fmt.Errorf("failed to create payment intent: %s", res.Status)
	}

	// Decode response
	if err = json.NewDecoder(res.Body).Decode(&pi); err != nil {
		slog.Error("Failed to decode Stripe API response", "error", err)
		return pi, fmt.Errorf("failed to decode response: %w", err)
	}

	return
}

// handlePaymentIntentCreated processes the PaymentIntentCreated event.
// It verifies the payment intent against the order details.
// If the order is pending and the amounts match, it returns nil.
// If the order is not pending or the amounts do not match, it returns an error.
// This function is called when a PaymentIntentCreated event is received.
func (s *paymentService) handlePaymentIntentCreated(ctx context.Context, event stripe.Event) error {
	paymentIntent := event.Data.Object
	order, err := s.repo.GetOrder(ctx, paymentIntent.ID)
	if err != nil {
		return err
	}
	if order.Status != types.OrderPending {
		return nil
	}
	if order.TotalAmount != paymentIntent.Amount {
		return fmt.Errorf("payment intent amount does not match expected amount")
	}
	if !strings.EqualFold(order.Currency, paymentIntent.Currency) {
		return fmt.Errorf("payment intent currency does not match expected currency")
	}
	return nil
}

// handlePaymentIntentSucceeded processes the PaymentIntentSucceeded event.
// It verifies the payment intent against the order details.
// If the order is pending and the amounts match, it marks the order as paid.
// If the order is not pending or the amounts do not match, it returns an error.
// This function is called when a PaymentIntentSucceeded event is received.
func (s *paymentService) handlePaymentIntentSucceeded(ctx context.Context, event stripe.Event) error {
	paymentIntent := event.Data.Object

	// do some basic validation
	order, err := s.repo.GetOrder(ctx, paymentIntent.ID)
	if err != nil {
		return err
	}
	if order.Status != types.OrderPending {
		return nil
	}
	if order.TotalAmount != paymentIntent.Amount {
		slog.Error("Payment intent amount does not match expected amount", "order_id", order.ID)
		return fmt.Errorf("payment intent amount does not match expected amount")
	}
	if !strings.EqualFold(order.Currency, paymentIntent.Currency) {
		slog.Error("Payment intent currency does not match expected currency", "order_id", order.ID)
		return fmt.Errorf("payment intent currency does not match expected currency")
	}

	// mark order as paid
	err = s.repo.MarkOrderAsPaid(ctx, order.ID, paymentIntent)
	if err != nil {
		slog.Error("Error marking order as paid", "order_id", order.ID, "error", err)
		return err
	}

	slog.Info("Order marked as paid", "order_id", order.ID, "payment_intent_id", paymentIntent.ID)
	return nil
}

func (s *paymentService) CancelPaymentIntent(ctx context.Context, paymentIntentID string) error {
	reqURL := fmt.Sprintf("%s/payment_intents/%s/cancel", s.stripeConfig.BaseURL, paymentIntentID)
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create Stripe cancel request: %w", err)
	}
	req.SetBasicAuth(s.stripeConfig.SecretKey, "")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.HttpClient.Do(req)
	if err != nil {
		slog.Error("Error sending cancel request to Stripe", "url", reqURL, "error", err)
		return fmt.Errorf("failed to send cancel request to Stripe: %w", err)
	}

	defer resp.Body.Close()

	// Handle response
	if resp.StatusCode != http.StatusOK {
		slog.Error("Stripe returned error on payment intent cancel", "status", resp.StatusCode, "url", reqURL)
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

// unixTimestampToTime converts [timestamp], Unix timestamp string to a time.Time object.
func unixTimestampToTime(timestamp string) (time.Time, error) {
	seconds, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(seconds, 0), nil
}
