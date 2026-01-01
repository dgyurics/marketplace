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

const (
	tolerance = time.Minute * 5 // Maximum allowed time difference between Stripe's timestamp and server time
)

type PaymentService interface {
	EventHandler(ctx context.Context, event stripe.Event) error
	SupportedEvent(ctx context.Context, event stripe.Event) bool
	SignatureVerifier(payload []byte, sigHeader string) error
	CreatePaymentIntent(ctx context.Context, refID string, amount int64, email string) (stripe.PaymentIntent, error)
}

type paymentService struct {
	HttpClient   utilities.HTTPClient
	config       types.PaymentConfig
	serviceEmail EmailService
	serviceTmp   TemplateService
	serviceUser  UserService
	repo         repositories.OrderRepository
}

func NewPaymentService(
	httpClient utilities.HTTPClient,
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

// EventHandler handles incoming Stripe events.
// It routes the event to the appropriate handler based on its type.
func (s *paymentService) EventHandler(ctx context.Context, event stripe.Event) error {
	var err error
	switch event.Type {
	// Payment intent group
	case
		stripe.EventTypePaymentIntentSucceeded,
		stripe.EventTypePaymentIntentCanceled,
		stripe.EventTypePaymentIntentCreated,
		stripe.EventTypePaymentIntentPaymentFailed:
		err = s.handlePIEvent(ctx, event)

	// Charge group
	case stripe.EventTypeChargeRefunded:
		err = s.handleChargeRefund(ctx, event)

	default:
		slog.DebugContext(ctx, "Unhandled Stripe event type", "type", event.Type)
		return nil
	}

	if err != nil {
		slog.Error("Handler failed", "type", event.Type, "error", err)
		return err
	}

	slog.Debug("Processed event", "type", event.Type, "id", event.ID)
	return nil
}

// handleChargeRefund handles Stripe Refund events.
func (s *paymentService) handleChargeRefund(ctx context.Context, event stripe.Event) error {
	charge, err := stripe.UnmarshalEventObject[stripe.Charge](&event)
	if err != nil {
		return err
	}
	return s.handleRefund(ctx, charge)
}

// handlePIEvent handles Stripe Payment Intent events.
func (s *paymentService) handlePIEvent(ctx context.Context, event stripe.Event) error {
	pi, err := stripe.UnmarshalEventObject[stripe.PaymentIntent](&event)
	if err != nil {
		return err
	}

	switch event.Type {
	case stripe.EventTypePaymentIntentCreated:
		return s.handlePaymentIntentCreated(ctx, pi)
	case stripe.EventTypePaymentIntentSucceeded:
		return s.handlePaymentIntentSucceeded(ctx, pi)
	case stripe.EventTypePaymentIntentCanceled:
		return s.handlePaymentIntentCancelled(ctx, pi)
	case stripe.EventTypePaymentIntentPaymentFailed:
		return s.handlePaymentIntentFailed(ctx, pi)
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
		slog.Warn("Invalid signature header", "header", sigHeader)
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
		slog.Warn("Timestamp missing from signature header", "header", sigHeader)
		return errors.New("missing timestamp")
	}

	if len(signatures) == 0 {
		slog.Warn("Signature missing from signature header", "header", sigHeader)
		return errors.New("missing signature")
	}

	ts, err := unixTimestampToTime(timestamp)
	if err != nil {
		slog.Warn("Error parsing timestamp", "timestamp", timestamp)
		return errors.New("invalid timestamp")
	} else if time.Since(ts) > tolerance {
		slog.Warn("Timestamp is too old", "timestamp", timestamp)
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

	slog.Warn("Signature verification failed", "signatures", signatures)
	return errors.New("signature verification failed: no matching v1 signature found")
}

// FIXME impose a minimum cart-total when using Stripe checkout
// For example, USD is $0.50 (equivalent to the cost of executing the transaction)
// CreatePaymentIntent creates a new Stripe Payment Intent.
// [refID] is a unique reference ID for idempotency. Currently this is the order ID
// [amount] is the amount in the smallest currency unit (e.g., cents for USD).
func (s *paymentService) CreatePaymentIntent(ctx context.Context, refID string, amount int64, email string) (pi stripe.PaymentIntent, err error) {
	// Build request
	reqURL := fmt.Sprintf("%s/payment_intents", s.config.Stripe.BaseURL)
	data, ok := utilities.LocaleData[utilities.Locale.CountryCode]
	if !ok {
		return pi, fmt.Errorf("unsupported country code: %s", utilities.Locale.CountryCode)
	}
	// https://selfco.io/checkout/confirmation
	payload := url.Values{
		"amount":                {fmt.Sprintf("%d", amount)},
		"currency":              {data.Currency},
		"receipt_email":         {email},
		"metadata[order_id]":    {refID},
		"metadata[environment]": {string(s.config.Environment)},
		// "payment_method_types[]": {"card"}, // omit to have automatic payment options displayed to user
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
// It verifies the payment intent against the order details stored in database.
// If the order is pending and the amounts match, it returns nil.
// If the order is not pending or the amounts do not match, it returns an error.
// This function is called when a PaymentIntentCreated event is received.
func (s *paymentService) handlePaymentIntentCreated(ctx context.Context, pi *stripe.PaymentIntent) error {
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
func (s *paymentService) handlePaymentIntentSucceeded(ctx context.Context, pi *stripe.PaymentIntent) error {
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

func (s *paymentService) handlePaymentIntentCancelled(_ context.Context, pi *stripe.PaymentIntent) error {
	orderID := pi.Metadata["order_id"]
	if orderID == "" {
		return fmt.Errorf("order_id not found in payment intent metadata")
	}
	slog.Debug("Payment intent canceled", "id", pi.ID, "order_id", orderID)
	return nil
}

func (s *paymentService) handlePaymentIntentFailed(_ context.Context, pi *stripe.PaymentIntent) error {
	orderID := pi.Metadata["order_id"]
	if orderID == "" {
		return fmt.Errorf("order_id not found in payment intent metadata")
	}
	slog.Debug("Payment intent payment failed", "id", pi.ID, "order_id", orderID)
	return nil
}

// handleRefund handles a successful refund event from Stripe.
// WARNING: partial refunds are not yet supported. The order will be marked as refunded regardless of the refund amount.
func (s *paymentService) handleRefund(ctx context.Context, charge *stripe.Charge) error {
	orderID := charge.Metadata["order_id"]
	if orderID == "" {
		return fmt.Errorf("order_id not found in charge metadata")
	}

	// do some basic validation
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}
	// Handle idempotency
	if order.Status == types.OrderRefunded {
		slog.Debug("Order already marked as refunded", "order_id", orderID)
		return nil
	}

	// Check if the order is eligible for a refund
	isEligible := order.Status == types.OrderPaid ||
		order.Status == types.OrderFulfilled ||
		order.Status == types.OrderShipped ||
		order.Status == types.OrderDelivered

	if !isEligible {
		return fmt.Errorf("refund received for non-eligible order: %s, status=%s", order.ID, order.Status)
	}

	if !strings.EqualFold(utilities.Locale.Currency, charge.Currency) {
		return fmt.Errorf("currency mismatch: expected %s, got %s, order_id=%s", utilities.Locale.Currency, charge.Currency, order.ID)
	}

	if order.TotalAmount != charge.AmountRefunded {
		slog.Warn("partial refund received", "order_id", order.ID, "order_amount", order.TotalAmount, "refund_amount", charge.AmountRefunded)
	}

	// mark order as refunded
	err = s.repo.RefundOrder(ctx, order.ID)
	if err != nil {
		return fmt.Errorf("failed to mark order as refunded: order_id=%s, error=%w", order.ID, err)
	}

	slog.Debug("Charge refunded", "id", charge.ID, "order_id", orderID, "payment_intent_id", charge.PaymentIntent)

	return nil
}

// SupportedEvent checks if the given Stripe event is supported by the payment service.
func (s *paymentService) SupportedEvent(ctx context.Context, event stripe.Event) bool {
	if !event.IsSupported() {
		slog.DebugContext(ctx, "Skipping unsupported event", "type", event.Type)
		return false
	}

	metadata := event.GetMetadata()
	if env, exists := metadata["environment"]; exists {
		return strings.EqualFold(env, string(s.config.Environment))
	}

	return true // Process events without environment metadata
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
