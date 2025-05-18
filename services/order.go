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
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/types/stripe"
	"github.com/dgyurics/marketplace/utilities"
)

const (
	tolerance = time.Minute * 5 // Maximum allowed time difference between Stripe's timestamp and server time
)

type OrderService interface {
	CreateOrder(ctx context.Context, order *types.Order) error
	GetOrders(ctx context.Context, page, limit int) ([]types.Order, error)
	VerifyStripeEventSignature(payload []byte, sigHeader string) error
	ProcessStripeEvent(ctx context.Context, event stripe.Event) error
	CancelStaleOrders(ctx context.Context)
}

type orderService struct {
	orderRepo  repositories.OrderRepository
	cartRepo   repositories.CartRepository
	HttpClient utilities.HTTPClient
	strpConfig types.StripeConfig
	locConfig  types.LocaleConfig
}

func NewOrderService(
	orderRepo repositories.OrderRepository,
	cartRepo repositories.CartRepository,
	strpConfig types.StripeConfig,
	locConfig types.LocaleConfig,
	httpClient utilities.HTTPClient,
) OrderService {
	if httpClient == nil {
		httpClient = utilities.NewDefaultHTTPClient(10 * time.Second)
	}
	return &orderService{
		orderRepo:  orderRepo,
		cartRepo:   cartRepo,
		HttpClient: httpClient,
		strpConfig: strpConfig,
		locConfig:  locConfig,
	}
}

func (os *orderService) CancelStaleOrders(ctx context.Context) {
	interval := 10 * time.Minute // Cancel orders older than 10 minutes with status "pending"
	pymtIntentIDs, err := os.orderRepo.CancelPendingOrders(ctx, interval)
	if err != nil {
		slog.Error("Error canceling stale orders", "error", err)
	}

	var wg sync.WaitGroup
	for _, id := range pymtIntentIDs {
		wg.Add(1)
		go func(pID string) {
			defer wg.Done()
			slog.Debug("Cancelling payment intent", "id", id)
			if err := os.cancelPaymentIntent(ctx, pID); err != nil {
				slog.Error("Error canceling payment intent", "id", pID)
			}
		}(id)
	}
	wg.Wait()
}

func (os *orderService) GetOrders(ctx context.Context, page, limit int) ([]types.Order, error) {
	var userID = getUserID(ctx)
	orders, err := os.orderRepo.GetOrders(ctx, userID, page, limit)
	if err != nil {
		return nil, err
	}
	if err = os.orderRepo.PopulateOrderItems(ctx, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}

// Call Stripe API to create a payment intent for a given order
func (os *orderService) createOrderPaymentIntent(ctx context.Context, orderID string, pi *stripe.PaymentIntent) error {
	// Validate input
	if pi.Amount <= 0 {
		return errors.New("invalid amount")
	}

	// Prepare request
	stripeURL := fmt.Sprintf("%s/payment_intents", os.strpConfig.BaseURL)
	payload := url.Values{
		"amount":                 {strconv.FormatInt(pi.Amount, 10)},
		"currency":               {pi.Currency},
		"payment_method_types[]": {"card"},
	}
	reqBody := strings.NewReader(payload.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, stripeURL, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.SetBasicAuth(os.strpConfig.SecretKey, "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Idempotency-Key", fmt.Sprintf("pi-%s", orderID))

	// Execute request
	resp, err := os.HttpClient.Do(req)
	if err != nil {
		slog.Error("Failed to send request to Stripe API", "url", stripeURL, "error", err)
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Handle response
	if resp.StatusCode != http.StatusOK {
		slog.Error("Stripe API returned non-OK status", "status", resp.StatusCode, "url", stripeURL)
		return fmt.Errorf("stripe API returned status: %d", resp.StatusCode)
	}

	// Decode response
	if err := json.NewDecoder(resp.Body).Decode(pi); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// [payload] raw request body
// [sigHeader] value of the Stripe-Signature header, in the format "t=timestamp,v1=signature,v1=signature,..."
func (os *orderService) VerifyStripeEventSignature(payload []byte, sigHeader string) error {
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
	expectedSignature := ComputeSignature(ts, payload, os.strpConfig.WebhookSigningSecret)
	for _, signature := range signatures {
		if hmac.Equal(signature, expectedSignature) {
			return nil
		}
	}

	slog.Error("Signature verification failed", "signatures", signatures)
	return errors.New("signature verification failed: no matching v1 signature found")
}

// ProcessStripeEvent processes a Stripe event. It is idompotent
func (os *orderService) ProcessStripeEvent(ctx context.Context, event stripe.Event) error {
	// save raw event data
	if err := os.orderRepo.CreateStripeEvent(ctx, event); err != nil {
		return err
	}
	switch event.Type {
	case "payment_intent.created":
		return os.paymentIntentCreated(ctx, event)
	case "payment_intent.succeeded":
		return os.paymentIntentSucceeded(ctx, event)
	case "payment_intent.canceled":
		slog.Debug("Payment canceled", "event", event.ID, "PaymentIntentID", event.Data.Object.ID)
	case "payment_intent.payment_failed":
		slog.Info("Payment failed", "event", event.ID, "PaymentIntentID", event.Data.Object.ID)
	default:
		slog.Debug("Unhandled event", "event", event.ID, "type", event.Type)
	}
	return nil
}

// paymentIntentCreated is called when a Stripe event is received for a payment intent creation.
// A matching payment entry should be found in the orders table
func (os *orderService) paymentIntentCreated(ctx context.Context, event stripe.Event) error {
	paymentIntent := event.Data.Object
	// verify event has matching entry in payment table
	order := &types.Order{
		StripePaymentIntent: &paymentIntent,
	}
	err := os.orderRepo.GetOrder(ctx, order)
	if err != nil {
		return err
	}
	if order.Status != types.OrderPending {
		return nil // do nothing if order not pending
	}
	if order.TotalAmount != paymentIntent.Amount {
		return fmt.Errorf("payment intent amount does not match expected amount")
	}
	if !strings.EqualFold(order.Currency, paymentIntent.Currency) {
		return fmt.Errorf("payment intent currency does not match expected currency")
	}
	return nil
}

// paymentIntentSucceeded is called when a Stripe event is received for a payment intent success
// A matching payment entry should be found in the orders table
func (os *orderService) paymentIntentSucceeded(ctx context.Context, event stripe.Event) error {
	paymentIntent := event.Data.Object
	order := &types.Order{
		StripePaymentIntent: &paymentIntent,
	}
	err := os.orderRepo.GetOrder(ctx, order)
	if err != nil {
		return err
	}
	if order.Status != types.OrderPending {
		return nil // do nothing if order not pending
	}
	if order.TotalAmount != paymentIntent.Amount {
		slog.Error("Payment intent amount does not match expected amount", "order_id", order.ID)
		return fmt.Errorf("payment intent amount does not match expected amount")
	}
	if !strings.EqualFold(order.Currency, paymentIntent.Currency) {
		slog.Error("Payment intent currency does not match expected currency", "order_id", order.ID)
		return fmt.Errorf("payment intent currency does not match expected currency")
	}

	// mark order as paid & update payment intent
	order.Status = types.OrderPaid
	order.StripePaymentIntent = &paymentIntent
	if err = os.orderRepo.UpdateOrder(ctx, order); err != nil {
		slog.Error("Error updating order", "order_id", order.ID, "error", err)
		return err
	}

	// clear cart
	if err = os.cartRepo.ClearCart(ctx, order.UserID); err != nil {
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

func (os *orderService) cancelPaymentIntent(ctx context.Context, paymentIntentID string) error {
	if paymentIntentID == "" {
		return errors.New("missing payment intent ID")
	}

	stripeURL := fmt.Sprintf("%s/payment_intents/%s/cancel",
		os.strpConfig.BaseURL,
		paymentIntentID,
	)

	reqStripe, err := http.NewRequestWithContext(ctx, "POST", stripeURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create Stripe cancel request: %w", err)
	}
	reqStripe.SetBasicAuth(os.strpConfig.SecretKey, "")
	reqStripe.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := os.HttpClient.Do(reqStripe)
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

func (os *orderService) CreateOrder(ctx context.Context, order *types.Order) error {
	userID := getUserID(ctx)
	order.UserID = userID

	if err := os.createAndLogOrder(ctx, order); err != nil {
		return err
	}

	if err := os.calculateTax(ctx, order); err != nil {
		slog.Error("Error calculating taxes", "order_id", order.ID, "error", err)
		return err
	}
	order.TotalAmount = order.Amount + order.ShippingAmount + order.TaxAmount

	clientSecret, err := os.setupStripePayment(ctx, order)
	if err != nil {
		return err
	}

	order.StripePaymentIntent.ClientSecret = clientSecret
	return nil
}

func (os *orderService) createAndLogOrder(ctx context.Context, order *types.Order) error {
	orderID, err := utilities.GenerateIDString()
	if err != nil {
		return err
	}
	order.ID = orderID
	order.Currency = os.locConfig.Currency
	if err := os.orderRepo.CreateOrder(ctx, order); err != nil {
		slog.Error("Error creating order", "user_id", order.UserID, "error", err)
		return err
	}

	slog.Info("Order created",
		"order_id", order.ID,
		"user_id", order.UserID,
		"currency", order.Currency,
		"amount", order.Amount,
		"shipping_amount", order.ShippingAmount,
		"tax_amount", order.TaxAmount,
		"total_amount", order.TotalAmount,
	)

	return nil
}

func (os *orderService) setupStripePayment(ctx context.Context, order *types.Order) (string, error) {
	pi := &stripe.PaymentIntent{
		Amount:   order.TotalAmount,
		Currency: order.Currency,
	}
	if err := os.createOrderPaymentIntent(ctx, order.ID, pi); err != nil {
		return "", err
	}

	clientSecret := pi.ClientSecret
	pi.ClientSecret = ""

	order.StripePaymentIntent = pi
	if err := os.orderRepo.UpdateOrder(ctx, order); err != nil {
		slog.Error("Error updating order",
			"order_id", order.ID,
			"status", order.Status,
			"stripe_payment_intent_id", pi.ID,
			"error", err,
		)
		return "", err
	}

	return clientSecret, nil
}

func (os *orderService) calculateTax(ctx context.Context, order *types.Order) error {
	form := url.Values{}
	form.Set("currency", order.Currency)

	if len(order.Items) == 0 {
		return errors.New("order has no items")
	}
	if order.Address.Country == "" {
		return errors.New("missing country code")
	}

	// Line Items
	for i, item := range order.Items {
		itmQty := int64(item.Quantity)
		form.Set(fmt.Sprintf("line_items[%d][amount]", i), strconv.FormatInt(item.UnitPrice*itmQty, 10))
		form.Set(fmt.Sprintf("line_items[%d][quantity]", i), strconv.FormatInt(itmQty, 10))
		form.Set(fmt.Sprintf("line_items[%d][reference]", i), fmt.Sprintf("%s-%s", order.ID, item.Product.ID))
		form.Set(fmt.Sprintf("line_items[%d][tax_behavior]", i), string(os.locConfig.TaxBehavior))
		if item.Product.TaxCode == "" {
			form.Set(fmt.Sprintf("line_items[%d][tax_code]", i), os.locConfig.FallbackTaxCode)
		} else {
			form.Set(fmt.Sprintf("line_items[%d][tax_code]", i), item.Product.TaxCode)
		}
	}

	// Customer Address
	form.Set("customer_details[address_source]", "shipping") // FIXME make configurable (need way to retrieve billing address)
	form.Set("customer_details[address][country]", order.Address.Country)
	form.Set("customer_details[address][city]", order.Address.City)
	form.Set("customer_details[address][line1]", order.Address.Line1)
	if line2 := order.Address.Line2; line2 != nil && *line2 != "" {
		form.Set("customer_details[address][line2]", *line2)
	}
	form.Set("customer_details[address][state]", order.Address.State)
	form.Set("customer_details[address][postal_code]", order.Address.PostalCode)

	url := fmt.Sprintf("%s/tax/calculations", os.strpConfig.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBufferString(form.Encode()))
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return fmt.Errorf("tax calculation canceled: %w", err)
		}
		return err
	}

	req.Header.Set("Authorization", "Bearer "+os.strpConfig.SecretKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Idempotency-Key", order.ID)

	resp, err := os.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("stripe returned status %d: %s", resp.StatusCode, string(body))
	}

	var tax stripe.TaxCalculationResponse
	if err := json.NewDecoder(resp.Body).Decode(&tax); err != nil {
		return err
	}

	slog.DebugContext(ctx,
		"Tax estimate retrieved",
		"order_id", order.ID,
		"tax_amount_exclusive", tax.TaxAmountExclusive,
		"tax_amount_inclusive", tax.TaxAmountInclusive,
		"total_amount", tax.AmountTotal,
		"country", tax.CustomerDetails.Address.Country,
		"state", tax.CustomerDetails.Address.State,
		"breakdown", tax.TaxBreakdown,
	)

	order.TaxAmount = tax.TaxAmountInclusive + tax.TaxAmountExclusive
	order.TotalAmount = tax.AmountTotal
	return nil
}
