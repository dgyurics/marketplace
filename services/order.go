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

type OrderService interface {
	CreatePaymentIntent(ctx context.Context, pi *models.PaymentIntent) error
	CancelPaymentIntent(ctx context.Context, pi *models.PaymentIntent) error
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

func (ps *orderService) CreatePaymentIntent(ctx context.Context, pi *models.PaymentIntent) error {
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
		return err
	}
	defer resp.Body.Close()

	// Handle Stripe API response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("stripe API returned status %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(pi)
}

func (ps *orderService) CancelPaymentIntent(ctx context.Context, pi *models.PaymentIntent) error {
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

func (ps *orderService) VerifyWebhookEventSignature(payload []byte, sigHeader string) error {
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
func (ps *orderService) ProcessWebhookEvent(ctx context.Context, event models.StripeWebhookEvent) error {
	if event.Data == nil {
		return errors.New("missing event data")
	}
	if event.Data.Object.ID == "" {
		return errors.New("missing payment intent ID")
	}
	// TODO payment_intent.canceled
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
func (ps *orderService) PaymentIntentCreated(ctx context.Context, event models.StripeWebhookEvent) error {
	// save raw event
	if err := ps.orderRepo.CreateWebhookEvent(ctx, event); err != nil {
		return err
	}
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

// To be called when a webhook event is received from Stripe for a payment intent success
func (ps *orderService) PaymentIntentSucceeded(ctx context.Context, event models.StripeWebhookEvent) error {
	// save raw event
	if err := ps.orderRepo.CreateWebhookEvent(ctx, event); err != nil {
		return err
	}
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

// To be called when a webhook event is received from Stripe for a payment intent failure
func (ps *orderService) PaymentIntentPaymentFailed(ctx context.Context, event models.StripeWebhookEvent) error {
	// save raw event
	return ps.orderRepo.CreateWebhookEvent(ctx, event)
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

func (s *orderService) CreateOrder(ctx context.Context) (models.PaymentIntent, error) {
	var userID = getUserID(ctx)

	// FIXME call Stripe API and cancel the payment intent when an existing pending order/payment is found

	// Create order
	order, err := s.orderRepo.CreateOrder(ctx, userID)
	if err != nil {
		return models.PaymentIntent{}, err
	}

	// Send payment request to Stripe
	// On success, this will trigger a webhook event where type = payment_intent.created
	pi := models.PaymentIntent{
		Amount:   order.TotalAmount + order.TaxAmount,
		Currency: "usd",
	}
	if err = s.CreatePaymentIntent(ctx, &pi); err != nil {
		return models.PaymentIntent{}, err
	}

	// Save payment details
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
