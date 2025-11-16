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
	util "github.com/dgyurics/marketplace/utilities"
)

const (
	tolerance = time.Minute * 5 // Maximum allowed time difference between Stripe's timestamp and server time
)

type PaymentService interface {
	EventHandler(ctx context.Context, event stripe.Event) error
	SignatureVerifier(payload []byte, sigHeader string) error
	CreatePaymentIntent(ctx context.Context, refID string, amount int64) (stripe.PaymentIntent, error)
}

type paymentService struct {
	HttpClient   util.HTTPClient
	config       types.PaymentConfig
	serviceEmail EmailService
	serviceTmp   TemplateService
	serviceUser  UserService
	repo         repositories.OrderRepository
}

func NewPaymentService(
	httpClient util.HTTPClient,
	config types.PaymentConfig,
	serviceEmail EmailService,
	serviceTmp TemplateService,
	serviceUser UserService,
	repo repositories.OrderRepository) PaymentService {
	return &paymentService{
		HttpClient:   httpClient,
		config:       config,
		serviceEmail: serviceEmail,
		serviceTmp:   serviceTmp,
		serviceUser:  serviceUser,
		repo:         repo,
	}
}

func (s *paymentService) shouldSkipEvent(event stripe.Event) bool {
	if event.Data == nil || event.Data.Object.Metadata == nil {
		return false
	}

	eventEnv := event.Data.Object.Metadata["environment"]
	currentEnv := string(s.config.Environment)

	if eventEnv != "" && eventEnv != currentEnv {
		slog.Debug("Skipping event from different environment",
			"eventEnv", eventEnv, "currentEnv", currentEnv, "eventID", event.ID)
		return true
	}
	return false
}

func (s *paymentService) EventHandler(ctx context.Context, event stripe.Event) (err error) {
	if s.shouldSkipEvent(event) {
		return nil
	}
	switch stripe.EventType(event.Type) {
	case stripe.EventTypePaymentIntentCreated:
		err = s.handlePaymentIntentCreated(ctx, event)
	case stripe.EventTypePaymentIntentSucceeded:
		err = s.handlePaymentIntentSucceeded(ctx, event)
	case stripe.EventTypePaymentIntentCanceled:
		err = s.handlePaymentIntentCancelled(ctx, event)
	case stripe.EventTypePaymentIntentPaymentFailed:
		err = s.handlePaymentIntentFailed(ctx, event)
	default:
		// Unhandled events - consider implementing for advanced features:
		// - charge.updated: payment retries, fees, risk scores
		// - charge.dispute.*: chargeback/dispute handling
		// - charge.refunded: refund confirmations and partial refunds
		slog.Debug("Unhandled Stripe event type", "type", event.Type)
	}

	if err != nil {
		slog.Error("Error handling event", "type", event.Type, "error", err)
		return err
	}

	// Successfully processed
	slog.Debug("Successfully processed Stripe event", "type", event.Type, "id", event.ID)
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
	expectedSignature := ComputeSignature(ts, payload, s.config.Stripe.WebhookSigningSecret)
	for _, signature := range signatures {
		if hmac.Equal(signature, expectedSignature) {
			return nil
		}
	}

	slog.Error("Signature verification failed", "signatures", signatures)
	return errors.New("signature verification failed: no matching v1 signature found")
}

// FIXME impose a minimum cart-total when using Stripe checkout
// For example, USD is $0.50 (equivalent to the cost of executing the transaction)
// CreatePaymentIntent creates a new Stripe Payment Intent.
// [refID] is a unique reference ID for idempotency. Currently this is the order ID
// [amount] is the amount in the smallest currency unit (e.g., cents for USD).
func (s *paymentService) CreatePaymentIntent(ctx context.Context, refID string, amount int64) (pi stripe.PaymentIntent, err error) {
	// Build request
	reqURL := fmt.Sprintf("%s/payment_intents", s.config.Stripe.BaseURL)
	data, ok := utilities.LocaleData[utilities.Locale.CountryCode]
	if !ok {
		return pi, fmt.Errorf("unsupported country code: %s", utilities.Locale.CountryCode)
	}
	payload := url.Values{
		"amount":                 {fmt.Sprintf("%d", amount)},
		"currency":               {data.Currency},
		"payment_method_types[]": {"card"},
		"metadata[order_id]":     {refID},
		"metadata[environment]":  {string(s.config.Environment)},
		// TODO send receipt to customer "receipt_email": {order.Address.Email}
	}
	reqBody := strings.NewReader(payload.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, reqBody)
	if err != nil {
		return pi, err
	}

	// Set request headers
	req.SetBasicAuth(s.config.Stripe.SecretKey, "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Idempotency-Key", fmt.Sprintf("payment-intent-%s", refID))
	req.Header.Set("Stripe-Version", s.config.Stripe.Version)

	// Execute request
	res, err := s.HttpClient.Do(req)
	if err != nil {
		return pi, err
	}
	defer res.Body.Close()

	// Handle response
	if res.StatusCode != http.StatusOK {
		slog.Error("Stripe API returned non-OK status", "status", res.StatusCode, "url", s.config.Stripe.BaseURL)
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
	pi := event.Data.Object

	// Get order ID from metadata instead of querying by payment intent ID
	orderID := pi.Metadata["order_id"]
	if orderID == "" {
		return fmt.Errorf("order_id not found in payment intent metadata")
	}

	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order.Status != types.OrderPending {
		return nil
	}
	if order.TotalAmount != pi.Amount {
		return fmt.Errorf("amount mismatch: expected %d, got %d, order_id=%s", order.TotalAmount, pi.Amount, order.ID)
	}
	if !strings.EqualFold(utilities.Locale.Currency, pi.Currency) {
		return fmt.Errorf("currency mismatch: expected %s, got %s, order_id=%s", utilities.Locale.Currency, pi.Currency, order.ID)
	}
	return nil
}

// handlePaymentIntentSucceeded processes the PaymentIntentSucceeded event.
// It verifies the payment intent against the order details.
// If the order is pending and the amounts match, it marks the order as paid.
// If the order is not pending or the amounts do not match, it returns an error.
// This function is called when a PaymentIntentSucceeded event is received.
func (s *paymentService) handlePaymentIntentSucceeded(ctx context.Context, event stripe.Event) error {
	pi := event.Data.Object

	// Get order ID from metadata instead of querying by payment intent ID
	orderID := pi.Metadata["order_id"]
	if orderID == "" {
		return fmt.Errorf("order_id not found in payment intent metadata")
	}

	// do some basic validation
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order.Status != types.OrderPending {
		slog.Debug("Payment intent succeeded for non-pending order", "order_id", order.ID, "status", order.Status)
		return nil
	}
	if order.TotalAmount != pi.Amount {
		return fmt.Errorf("amount mismatch: expected %d, got %d, order_id=%s", order.TotalAmount, pi.Amount, order.ID)
	}
	if !strings.EqualFold(utilities.Locale.Currency, pi.Currency) {
		return fmt.Errorf("currency mismatch: expected %s, got %s, order_id=%s", utilities.Locale.Currency, pi.Currency, order.ID)
	}

	// mark order as paid
	err = s.repo.ConfirmOrderPayment(ctx, order.ID)
	if err != nil {
		return fmt.Errorf("failed to mark order as paid: order_id=%s, error=%w", order.ID, err)
	}

	slog.Info("Order marked as paid", "order_id", order.ID, "payment_intent_id", pi.ID)

	// Send payment success email to customer
	go func(recEmail, orderID string) {
		detailsLink := fmt.Sprintf("%s/orders/%s", s.config.BaseURL, orderID)
		data := map[string]string{
			"DetailsLink": detailsLink,
		}
		body, err := s.serviceTmp.RenderToString(OrderConfirmation, data)
		if err != nil {
			slog.Error("Error loading email template: ", "error", err)
			return
		}
		email := &types.Email{
			To:      []string{recEmail},
			Subject: "Order Confirmation",
			Body:    body,
			IsHTML:  true,
		}
		if err := s.serviceEmail.Send(email); err != nil {
			slog.Error("Error sending order confirmation email: ", "order_id", order.ID, "error", err)
		}
	}(order.Address.Email, order.ID)

	// Send order notification email to admins
	go func(order types.Order) {
		admins, err := s.serviceUser.GetAllAdmins(context.Background())
		if err != nil {
			slog.Error("Error fetching admin users: ", "error", err)
			return
		}

		// Extract emails
		adminEmails := make([]string, len(admins))
		for i, admin := range admins {
			adminEmails[i] = admin.Email
		}

		detailsLink := fmt.Sprintf("%s/admin/orders/%s", s.config.BaseURL, orderID)
		data := map[string]string{
			"OrderID":       order.ID,
			"CustomerEmail": order.Address.Email,
			"DetailsLink":   detailsLink,
		}
		body, err := s.serviceTmp.RenderToString(OrderNotification, data)
		if err != nil {
			slog.Error("Error loading email template: ", "error", err)
			return
		}
		email := &types.Email{
			To:      adminEmails,
			Subject: "Order Notification",
			Body:    body,
			IsHTML:  true,
		}
		if err := s.serviceEmail.Send(email); err != nil {
			slog.Error("Error sending order notification email: ", "order_id", order.ID, "error", err)
		}
	}(order)

	return nil
}

func (s *paymentService) handlePaymentIntentCancelled(_ context.Context, event stripe.Event) error {
	orderID := event.Data.Object.Metadata["order_id"]
	if orderID == "" {
		return fmt.Errorf("order_id not found in payment intent metadata")
	}
	slog.Debug("Payment intent canceled", "id", event.Data.Object.ID, "order_id", orderID)
	return nil
}

func (s *paymentService) handlePaymentIntentFailed(_ context.Context, event stripe.Event) error {
	orderID := event.Data.Object.Metadata["order_id"]
	if orderID == "" {
		return fmt.Errorf("order_id not found in payment intent metadata")
	}
	slog.Debug("Payment intent payment failed", "id", event.Data.Object.ID, "order_id", orderID)
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
