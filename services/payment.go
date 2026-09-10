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
	HttpClient          utilities.HTTPClient
	config              types.PaymentConfig
	notificationService NotificationService
	userService         UserService
	repo                repositories.OrderRepository
}

func NewPaymentService(
	httpClient utilities.HTTPClient,
	config types.PaymentConfig,
	notificationService NotificationService,
	userService UserService,
	repo repositories.OrderRepository) PaymentService {
	return &paymentService{
		HttpClient:          httpClient,
		config:              config,
		notificationService: notificationService,
		userService:         userService,
		repo:                repo,
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
		slog.ErrorContext(ctx, "Stripe event handler failed", "event_id", event.ID, "event_type", event.Type, "error", err)
		return err
	}

	slog.DebugContext(ctx, "Processed Stripe event", "event_type", event.Type, "event_id", event.ID)
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
		return s.handlePaymentIntentCanceled(ctx, pi)
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
		return fmt.Errorf("invalid timestamp: %w", err)
	}

	skew := time.Since(ts)
	if skew < 0 {
		skew = -skew
	}
	if skew > tolerance {
		slog.Warn("Timestamp is outside tolerance window", "timestamp", timestamp, "skew", skew)
		return errors.New("timestamp is outside tolerance window")
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

// TODO: Enforce Stripe minimum charge amount by currency
// (e.g., USD minimum is $0.50 to cover transaction cost).
//
// CreatePaymentIntent creates a Stripe PaymentIntent.
// refID is a unique idempotency reference (currently the order ID).
// amount is in the smallest currency unit (for example, cents for USD).
func (s *paymentService) CreatePaymentIntent(ctx context.Context, refID string, amount int64, email string) (stripe.PaymentIntent, error) {
	var pi stripe.PaymentIntent

	// Build request
	reqURL, err := url.JoinPath(s.config.Stripe.BaseURL, "payment_intents")
	if err != nil {
		return pi, err
	}
	data, ok := utilities.LocaleData[utilities.Locale.CountryCode]
	if !ok {
		return pi, fmt.Errorf("unsupported country code: %s", utilities.Locale.CountryCode)
	}

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

	return pi, nil
}

// handlePaymentIntentCreated processes the PaymentIntentCreated event.
func (s *paymentService) handlePaymentIntentCreated(ctx context.Context, pi *stripe.PaymentIntent) error {
	orderID := pi.Metadata["order_id"]
	if orderID == "" {
		// deterministic payload issue: retry won't fix
		slog.WarnContext(ctx, "payment_intent.created missing order_id", "pi_id", pi.ID)
		return nil
	}

	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		// transient: let caller return error so Stripe retries
		return err
	}

	// Idempotent/out-of-order no-op cases
	if *order.Status == types.OrderPaid && *order.PaymentStatus == types.PaymentStatusPaid {
		slog.DebugContext(ctx, "payment_intent.created no-op: order already paid", "order_id", order.ID, "pi_id", pi.ID)
		return nil
	}
	if *order.Status == types.OrderRefunded && *order.PaymentStatus == types.PaymentStatusRefunded {
		slog.DebugContext(ctx, "payment_intent.created no-op: order already refunded", "order_id", order.ID, "pi_id", pi.ID)
		return nil
	}

	// Deterministic validation mismatches: log and ack (no retry storm)
	if *order.Status != types.OrderPending {
		slog.WarnContext(ctx, "payment_intent.created state mismatch", "order_id", order.ID, "expected_status", types.OrderPending, "actual_status", *order.Status)
		return nil
	}
	if *order.PaymentMethod != types.PaymentMethodStripe {
		slog.WarnContext(ctx, "payment_intent.created payment method mismatch", "order_id", order.ID, "expected_payment_method", types.PaymentMethodStripe, "actual_payment_method", *order.PaymentMethod)
		return nil
	}
	if *order.PaymentStatus != types.PaymentStatusUnpaid {
		slog.WarnContext(ctx, "payment_intent.created payment status mismatch", "order_id", order.ID, "expected_payment_status", types.PaymentStatusUnpaid, "actual_payment_status", *order.PaymentStatus)
		return nil
	}
	if !strings.EqualFold(utilities.Locale.Currency, pi.Currency) {
		slog.WarnContext(ctx, "payment_intent.created currency mismatch", "order_id", order.ID, "expected_currency", utilities.Locale.Currency, "actual_currency", pi.Currency)
		return nil
	}
	if order.TotalAmount != pi.Amount {
		slog.WarnContext(ctx, "payment_intent.created amount mismatch", "order_id", order.ID, "expected_amount", order.TotalAmount, "actual_amount", pi.Amount)
		return nil
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
		// deterministic payload issue: retry won't fix
		slog.WarnContext(ctx, "payment_intent.succeeded missing order_id", "pi_id", pi.ID)
		return nil
	}

	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		// transient: allow retry
		return err
	}

	// Idempotent/out-of-order no-op cases
	if *order.Status == types.OrderPaid && *order.PaymentStatus == types.PaymentStatusPaid {
		slog.DebugContext(ctx, "payment_intent.succeeded no-op: order already paid", "order_id", order.ID, "pi_id", pi.ID)
		return nil
	}
	if *order.Status == types.OrderRefunded && *order.PaymentStatus == types.PaymentStatusRefunded {
		slog.DebugContext(ctx, "payment_intent.succeeded no-op: order already refunded", "order_id", order.ID, "pi_id", pi.ID)
		return nil
	}

	// Deterministic validation mismatches: log and ack (no retry storm)
	if *order.Status != types.OrderPending {
		slog.WarnContext(ctx, "payment_intent.succeeded state mismatch", "order_id", order.ID, "expected_status", types.OrderPending, "actual_status", *order.Status)
		return nil
	}
	if *order.PaymentMethod != types.PaymentMethodStripe {
		slog.WarnContext(ctx, "payment_intent.succeeded payment method mismatch", "order_id", order.ID, "expected_payment_method", types.PaymentMethodStripe, "actual_payment_method", *order.PaymentMethod)
		return nil
	}
	if *order.PaymentStatus != types.PaymentStatusUnpaid {
		slog.WarnContext(ctx, "payment_intent.succeeded payment status mismatch", "order_id", order.ID, "expected_payment_status", types.PaymentStatusUnpaid, "actual_payment_status", *order.PaymentStatus)
		return nil
	}
	if order.TotalAmount != pi.Amount {
		slog.WarnContext(ctx, "payment_intent.succeeded amount mismatch", "order_id", order.ID, "expected_amount", order.TotalAmount, "actual_amount", pi.Amount)
		return nil
	}
	if !strings.EqualFold(utilities.Locale.Currency, pi.Currency) {
		slog.WarnContext(ctx, "payment_intent.succeeded currency mismatch", "order_id", order.ID, "expected_currency", utilities.Locale.Currency, "actual_currency", pi.Currency)
		return nil
	}

	// mark order as paid
	*order.Status = types.OrderPaid
	*order.PaymentStatus = types.PaymentStatusPaid

	if err = s.repo.UpdateOrder(ctx, &order); err != nil {
		return fmt.Errorf("failed to mark order as paid: order_id=%s, error=%w", order.ID, err)
	}

	slog.InfoContext(ctx, "Order marked as paid", "order_id", order.ID, "payment_intent_id", pi.ID)

	go s.notificationService.NotifyOrder(order.UserID, SubjectOrderConf, NotifyOrderConf, order)

	admins, err := s.userService.GetAllAdmins(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to load admins for order notification", "order_id", order.ID, "error", err)
		return nil
	}
	for _, admin := range admins {
		go s.notificationService.NotifyOrder(admin.ID, SubjectOrderRecv, NotifyOrderRecv, order)
	}

	return nil
}

func (s *paymentService) handlePaymentIntentCanceled(ctx context.Context, pi *stripe.PaymentIntent) error {
	orderID := pi.Metadata["order_id"]
	if orderID == "" {
		// deterministic payload issue: retry won't fix
		slog.WarnContext(ctx, "payment_intent.canceled missing order_id", "pi_id", pi.ID)
		return nil
	}
	slog.DebugContext(ctx, "Payment intent canceled", "id", pi.ID, "order_id", orderID)
	return nil
}

func (s *paymentService) handlePaymentIntentFailed(ctx context.Context, pi *stripe.PaymentIntent) error {
	orderID := pi.Metadata["order_id"]
	if orderID == "" {
		// deterministic payload issue: retry won't fix
		slog.WarnContext(ctx, "payment_intent.payment_failed missing order_id", "pi_id", pi.ID)
		return nil
	}
	slog.DebugContext(ctx, "Payment intent payment failed", "id", pi.ID, "order_id", orderID)
	return nil
}

// handleRefund handles a successful refund event from Stripe.
// WARNING: partial refunds are not yet supported. The order will be marked as refunded regardless of the refund amount.
func (s *paymentService) handleRefund(ctx context.Context, charge *stripe.Charge) error {
	orderID := charge.Metadata["order_id"]
	if orderID == "" {
		// deterministic payload issue: retry won't fix
		slog.WarnContext(ctx, "charge.refunded missing order_id", "charge_id", charge.ID)
		return nil
	}

	// do some basic validation
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		// transient: allow retry
		return err
	}
	// Handle idempotency
	if *order.Status == types.OrderRefunded {
		slog.DebugContext(ctx, "charge.refunded no-op: order already refunded", "order_id", orderID, "charge_id", charge.ID)
		return nil
	}

	// Check if the order is eligible for a refund
	isEligible := *order.PaymentStatus == types.PaymentStatusPaid && *order.PaymentMethod == types.PaymentMethodStripe

	if !isEligible {
		slog.WarnContext(ctx,
			"charge.refunded refund eligibility mismatch",
			"order_id", order.ID,
			"expected_payment_status", types.PaymentStatusPaid,
			"actual_payment_status", *order.PaymentStatus,
			"expected_payment_method", types.PaymentMethodStripe,
			"actual_payment_method", *order.PaymentMethod,
		)
		return nil
	}

	if !strings.EqualFold(utilities.Locale.Currency, charge.Currency) {
		slog.WarnContext(ctx, "charge.refunded currency mismatch", "order_id", order.ID, "expected_currency", utilities.Locale.Currency, "actual_currency", charge.Currency)
		return nil
	}

	if order.TotalAmount != charge.AmountRefunded {
		slog.WarnContext(ctx, "charge.refunded partial refund not supported", "order_id", order.ID, "order_amount", order.TotalAmount, "refund_amount", charge.AmountRefunded)
		return nil
	}

	// mark order as refunded
	*order.Status = types.OrderRefunded
	*order.PaymentStatus = types.PaymentStatusRefunded
	err = s.repo.UpdateOrder(ctx, &order)
	if err != nil {
		return fmt.Errorf("failed to mark order as refunded: order_id=%s, error=%w", order.ID, err)
	}

	slog.DebugContext(ctx, "Charge refunded", "id", charge.ID, "order_id", orderID, "payment_intent_id", charge.PaymentIntent)

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
