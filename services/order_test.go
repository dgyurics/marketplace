package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/types/stripe"
	"github.com/stretchr/testify/mock"
)

// mockOrderRepo implements the OrderRepository interface for testing
type mockOrderRepo struct {
	mock.Mock
}

func TestCalculateTax_Success(t *testing.T) {
	order := &types.Order{
		ID:       "order-123",
		Currency: "usd",
		Address: &types.Address{
			Country:    "US",
			State:      "CA",
			City:       "Los Angeles",
			PostalCode: "90001",
			Line1:      "123 Main St",
		},
		Items: []types.OrderItem{
			{Product: types.Product{ID: "prod-1"}, Quantity: 2, UnitPrice: 500},
		},
	}

	respBody := `{
		"tax_amount_exclusive": 100,
		"tax_amount_inclusive": 0,
		"amount_total": 1100,
		"customer_details": {
			"address": {
				"city": "Los Angeles",
				"country": "US",
				"state": "CA"
			}
		}
	}`

	httpClient := &MockHTTPClient{}
	httpClient.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(respBody)),
	}, nil)

	svc := &orderService{
		HttpClient: httpClient,
		locConfig: types.LocaleConfig{
			Country:         "US",
			Currency:        "USD",
			FallbackTaxCode: "txcd_99999999",
			TaxBehavior:     "exclusive",
		},
		strpConfig: types.StripeConfig{
			BaseURL:   "https://api.stripe.com/v1",
			SecretKey: "sk_test_123",
		},
	}

	err := svc.calculateTax(context.Background(), order)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if order.TaxAmount != 100 || order.TotalAmount != 1100 {
		t.Fatalf("unexpected tax calculation result: %+v", order)
	}

	httpClient.AssertExpectations(t)
}

func TestVerifyStripeEventSignature_Valid(t *testing.T) {
	secret := "whsec_testsecret"
	timestamp := time.Now().UTC()
	payload := []byte(`{"id":"evt_test_webhook","object":"event"}`)
	expectedSig := ComputeSignature(timestamp, payload, secret)
	sigHeader := fmt.Sprintf("t=%d,v1=%s", timestamp.Unix(), hex.EncodeToString(expectedSig))

	svc := &orderService{
		strpConfig: types.StripeConfig{
			WebhookSigningSecret: secret,
		},
	}

	err := svc.VerifyStripeEventSignature(payload, sigHeader)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestVerifyStripeEventSignature_InvalidSignature(t *testing.T) {
	secret := "whsec_testsecret"
	timestamp := time.Now().UTC()
	payload := []byte(`{"id":"evt_test_webhook","object":"event"}`)
	sigHeader := fmt.Sprintf("t=%d,v1=deadbeef", timestamp.Unix())

	svc := &orderService{
		strpConfig: types.StripeConfig{
			WebhookSigningSecret: secret,
		},
	}

	err := svc.VerifyStripeEventSignature(payload, sigHeader)
	if err == nil || !strings.Contains(err.Error(), "no matching v1 signature") {
		t.Fatalf("expected signature verification error, got %v", err)
	}
}

// MockHTTPClient implements the utilities.HTTPClient interface for testing
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	resp, _ := args.Get(0).(*http.Response)
	return resp, args.Error(1)
}

func TestCreateOrderPaymentIntent_Success(t *testing.T) {
	httpClient := &MockHTTPClient{}
	orderID := "order-123"
	pi := &stripe.PaymentIntent{
		Amount:   2000,
		Currency: "usd",
	}

	// Prepare a fake response from Stripe
	stripeResp := `{"id":"pi_1", "amount":2000, "currency":"usd", "client_secret":"secret_123"}`
	httpClient.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(stripeResp)),
	}, nil)

	svc := &orderService{
		strpConfig: types.StripeConfig{
			BaseURL:   "https://api.stripe.com/v1",
			SecretKey: "sk_test_123",
		},
		HttpClient: httpClient,
	}

	err := svc.createOrderPaymentIntent(context.Background(), orderID, pi)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if pi.ID != "pi_1" || pi.ClientSecret != "secret_123" {
		t.Fatalf("unexpected response: %+v", pi)
	}

	httpClient.AssertExpectations(t)
}

func TestCreateOrderPaymentIntent_ErrorStatus(t *testing.T) {
	httpClient := &MockHTTPClient{}
	orderID := "order-123"
	pi := &stripe.PaymentIntent{
		Amount:   2000,
		Currency: "usd",
	}

	httpClient.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(strings.NewReader(`{}`)),
	}, nil)

	svc := &orderService{
		strpConfig: types.StripeConfig{
			BaseURL:   "https://api.stripe.com/v1",
			SecretKey: "sk_test_123",
		},
		HttpClient: httpClient,
	}

	err := svc.createOrderPaymentIntent(context.Background(), orderID, pi)
	if err == nil {
		t.Fatal("expected error due to bad status code, got nil")
	}
}

func contextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserKey, &types.User{ID: userID})
}

func TestGetOrders_Success(t *testing.T) {
	mockRepo := new(mockOrderRepo)
	svc := &orderService{
		orderRepo: mockRepo,
	}

	ctx := contextWithUserID(context.Background(), "user-123")
	expectedOrders := []types.Order{
		{ID: "order-1", UserID: "user-123", Amount: 1000},
	}

	mockRepo.On("GetOrders", ctx, "user-123", 1, 10).Return(expectedOrders, nil)
	mockRepo.On("PopulateOrderItems", ctx, &expectedOrders).Return(nil)

	result, err := svc.GetOrders(ctx, 1, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 1 || result[0].ID != "order-1" {
		t.Fatalf("unexpected result: %+v", result)
	}

	mockRepo.AssertExpectations(t)
}

func TestGetOrders_RepoError(t *testing.T) {
	mockRepo := new(mockOrderRepo)
	svc := &orderService{
		orderRepo: mockRepo,
	}

	ctx := contextWithUserID(context.Background(), "user-123")
	mockRepo.On("GetOrders", ctx, "user-123", 1, 10).Return(nil, errors.New("db error"))

	result, err := svc.GetOrders(ctx, 1, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}

	mockRepo.AssertExpectations(t)
}

func (m *mockOrderRepo) GetOrder(ctx context.Context, order *types.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *mockOrderRepo) GetOrders(ctx context.Context, userID string, page, limit int) ([]types.Order, error) {
	args := m.Called(ctx, userID, page, limit)
	var orders []types.Order
	if v := args.Get(0); v != nil {
		orders = v.([]types.Order)
	}
	return orders, args.Error(1)
}

func (m *mockOrderRepo) PopulateOrderItems(ctx context.Context, orders *[]types.Order) error {
	args := m.Called(ctx, orders)
	return args.Error(0)
}

func (m *mockOrderRepo) CreateOrder(ctx context.Context, order *types.Order) error {
	return nil
}

func (m *mockOrderRepo) UpdateOrder(ctx context.Context, order *types.Order) error {
	return nil
}

func (m *mockOrderRepo) CreateStripeEvent(ctx context.Context, event stripe.Event) error {
	return nil
}

func (m *mockOrderRepo) CancelPendingOrders(ctx context.Context, interval time.Duration) ([]string, error) {
	return nil, nil
}

func TestComputeSignature(t *testing.T) {
	payload := []byte(`{"id":"evt_test_webhook","object":"event"}`)
	secret := "whsec_testsecret"
	timestamp := time.Unix(1713200000, 0) // use fixed time for deterministic output

	// Manually compute expected signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("1713200000"))
	mac.Write([]byte("."))
	mac.Write(payload)
	expected := mac.Sum(nil)

	// Call actual function
	actual := ComputeSignature(timestamp, payload, secret)

	// Compare
	if !hmac.Equal(expected, actual) {
		t.Errorf("Expected signature %s, got %s", hex.EncodeToString(expected), hex.EncodeToString(actual))
	}
}

func TestUnixTimestampToTime(t *testing.T) {
	validTimestamp := "1713200000"
	expected := time.Unix(1713200000, 0).UTC()

	parsed, err := unixTimestampToTime(validTimestamp)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !parsed.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, parsed)
	}
}

func TestUnixTimestampToTime_Invalid(t *testing.T) {
	invalidTimestamp := "not-a-valid-timestamp"

	_, err := unixTimestampToTime(invalidTimestamp)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCancelPaymentIntent_Success(t *testing.T) {
	httpClient := &MockHTTPClient{}
	orderSvc := &orderService{
		HttpClient: httpClient,
		strpConfig: types.StripeConfig{
			BaseURL:   "https://api.stripe.com/v1",
			SecretKey: "sk_test_123",
		},
	}

	intentID := "pi_123"
	response := `{"id":"pi_123","status":"canceled"}`

	httpClient.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(response)),
	}, nil)

	ctx := context.Background()
	err := orderSvc.cancelPaymentIntent(ctx, intentID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	httpClient.AssertExpectations(t)
}

func TestCancelPaymentIntent_FailedCancel(t *testing.T) {
	httpClient := &MockHTTPClient{}
	orderSvc := &orderService{
		HttpClient: httpClient,
		strpConfig: types.StripeConfig{
			BaseURL:   "https://api.stripe.com/v1",
			SecretKey: "sk_test_123",
		},
	}

	intentID := "pi_123"
	response := `{"id":"pi_123","status":"requires_payment_method"}`

	httpClient.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(response)),
	}, nil)

	ctx := context.Background()
	err := orderSvc.cancelPaymentIntent(ctx, intentID)
	if err == nil || !strings.Contains(err.Error(), "was not canceled") {
		t.Fatalf("expected error for non-canceled intent, got %v", err)
	}
	httpClient.AssertExpectations(t)
}

func TestCancelPaymentIntent_BadStatus(t *testing.T) {
	httpClient := &MockHTTPClient{}
	orderSvc := &orderService{
		HttpClient: httpClient,
		strpConfig: types.StripeConfig{
			BaseURL:   "https://api.stripe.com/v1",
			SecretKey: "sk_test_123",
		},
	}

	intentID := "pi_123"

	httpClient.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(strings.NewReader(`{}`)),
	}, nil)

	ctx := context.Background()
	err := orderSvc.cancelPaymentIntent(ctx, intentID)
	if err == nil {
		t.Fatal("expected error for bad status code, got nil")
	}
	httpClient.AssertExpectations(t)
}
