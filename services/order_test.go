package services_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/services"
)

type MockOrderRepository struct {
	mock.Mock
}

type MockCartRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) GetOrder(ctx context.Context, order *models.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) GetOrders(ctx context.Context, userID string, page, limit int) ([]models.Order, error) {
	args := m.Called(ctx, userID, page, limit)
	return args.Get(0).([]models.Order), args.Error(1)
}

func (m *MockOrderRepository) PopulateOrderItems(ctx context.Context, orders *[]models.Order) error {
	args := m.Called(ctx, orders)
	return args.Error(0)
}

func (m *MockOrderRepository) CreateWebhookEvent(ctx context.Context, event models.StripeWebhookEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockOrderRepository) CreateOrder(ctx context.Context, userID, addressID string) (*models.Order, error) {
	args := m.Called(ctx, userID, addressID)
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockOrderRepository) UpdateOrder(ctx context.Context, order *models.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockCartRepository) AddItemToCart(ctx context.Context, userID string, item *models.CartItem) error {
	args := m.Called(ctx, userID, item)
	return args.Error(0)
}

func (m *MockCartRepository) GetOrCreateCart(ctx context.Context, userID string) (*models.Cart, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*models.Cart), args.Error(1)
}

func (m *MockCartRepository) UpdateCartItem(ctx context.Context, userID string, item *models.CartItem) error {
	args := m.Called(ctx, userID, item)
	return args.Error(0)
}

func (m *MockCartRepository) RemoveItemFromCart(ctx context.Context, userID, productID string) error {
	args := m.Called(ctx, userID, productID)
	return args.Error(0)
}

func (m *MockCartRepository) ClearCart(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestOrderService_CreateOrder(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockCartRepo := new(MockCartRepository)
	mockHTTPClient := new(MockHTTPClient)

	orderService := services.NewOrderService(mockOrderRepo, mockCartRepo, models.OrderConfig{
		StripeSecretKey: "test_secret_key",
		StripeBaseURL:   "https://api.stripe.com/v1",
		Envirnment:      models.Production,
	}, mockHTTPClient)

	user := &models.User{
		ID:    "user123",
		Email: "user@example.com",
	}
	addressID := "address123"
	ctx := context.WithValue(context.Background(), services.UserKey, user)

	// Mock existing order to simulate no pending orders
	mockOrderRepo.On("GetOrder", mock.Anything, mock.Anything).Return(nil)

	// Mock order creation
	newOrder := &models.Order{
		ID:          "order123",
		UserID:      user.ID,
		TotalAmount: 2000,
		Status:      models.OrderPending,
	}
	mockOrderRepo.On("CreateOrder", mock.Anything, user.ID, addressID).Return(newOrder, nil)

	// Mock Stripe API call for payment intent creation
	stripeResponse := `{
		"id": "pi_mock_123",
		"status": "requires_payment_method",
		"amount": 2000,
		"currency": "usd",
		"client_secret": "test_client_secret"
	}`
	mockHTTPClient.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(stripeResponse)),
	}, nil)

	// Mock order update to handle the `UpdateOrder` call
	mockOrderRepo.On("UpdateOrder", mock.Anything, mock.MatchedBy(func(order *models.Order) bool {
		return order.ID == "order123" && order.PaymentIntentID == "pi_mock_123"
	})).Return(nil)

	// Call CreateOrder
	paymentIntent, err := orderService.CreateOrder(ctx, addressID)

	// Assertions
	assert.NoError(t, err, "CreateOrder should not return an error")
	assert.NotEmpty(t, paymentIntent.ID, "PaymentIntent ID should be set")
	assert.Equal(t, newOrder.TotalAmount, paymentIntent.Amount, "PaymentIntent amount should match order total amount")

	mockOrderRepo.AssertExpectations(t)
	mockCartRepo.AssertExpectations(t)
	mockHTTPClient.AssertExpectations(t)
}

func TestOrderService_CreateOrder_OrderCreationFails(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockCartRepo := new(MockCartRepository)
	mockHTTPClient := new(MockHTTPClient)

	orderService := services.NewOrderService(mockOrderRepo, mockCartRepo, models.OrderConfig{
		StripeSecretKey: "test_secret_key",
		StripeBaseURL:   "https://api.stripe.com/v1",
		Envirnment:      models.Production,
	}, mockHTTPClient)

	user := &models.User{
		ID:    "user123",
		Email: "user@example.com",
	}
	addressID := "address123"
	ctx := context.WithValue(context.Background(), services.UserKey, user)

	// Mock GetOrder to simulate no existing pending orders
	mockOrderRepo.On("GetOrder", mock.Anything, mock.Anything).Return(nil)

	// Mock CreateOrder to simulate a failure
	// Return a placeholder `*models.Order` along with an error to avoid nil dereference
	mockOrder := &models.Order{}
	mockOrderRepo.On("CreateOrder", mock.Anything, user.ID, addressID).
		Return(mockOrder, errors.New("failed to create order"))

	// Call CreateOrder
	_, err := orderService.CreateOrder(ctx, addressID)

	// Assertions
	assert.Error(t, err, "CreateOrder should return an error when order creation fails")
	assert.Contains(t, err.Error(), "failed to create order", "Error message should indicate failure reason")

	mockOrderRepo.AssertExpectations(t)
	mockCartRepo.AssertExpectations(t)
	mockHTTPClient.AssertExpectations(t)
}

func TestOrderService_GetOrders_Success(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockCartRepo := new(MockCartRepository)
	mockHTTPClient := new(MockHTTPClient)

	orderService := services.NewOrderService(mockOrderRepo, mockCartRepo, models.OrderConfig{
		StripeSecretKey: "test_secret_key",
		StripeBaseURL:   "https://api.stripe.com/v1",
		Envirnment:      models.Production,
	}, mockHTTPClient)

	user := &models.User{
		ID:    "user123",
		Email: "user@example.com",
	}
	ctx := context.WithValue(context.Background(), services.UserKey, user)

	// Mock orders returned by the repository
	mockOrders := []models.Order{
		{
			ID:          "order1",
			UserID:      user.ID,
			AddressID:   "address123",
			Currency:    "usd",
			TotalAmount: 1500,
			Status:      models.OrderPaid,
			Items:       nil, // Items will be populated separately
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "order2",
			UserID:      user.ID,
			AddressID:   "address456",
			Currency:    "usd",
			TotalAmount: 2500,
			Status:      models.OrderShipped,
			Items:       nil, // Items will be populated separately
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	// Mock GetOrders to return mock orders
	mockOrderRepo.On("GetOrders", mock.Anything, user.ID, 1, 10).Return(mockOrders, nil)

	// Mock PopulateOrderItems to populate items into orders
	mockOrderRepo.On("PopulateOrderItems", mock.Anything, mock.AnythingOfType("*[]models.Order")).Return(nil)

	// Call GetOrders
	orders, err := orderService.GetOrders(ctx, 1, 10)

	// Assertions
	assert.NoError(t, err, "GetOrders should not return an error")
	assert.Len(t, orders, 2, "GetOrders should return two orders")
	assert.Equal(t, "order1", orders[0].ID, "First order ID should match")
	assert.Equal(t, "order2", orders[1].ID, "Second order ID should match")

	// Verify that mocks were called
	mockOrderRepo.AssertExpectations(t)
	mockCartRepo.AssertExpectations(t)
	mockHTTPClient.AssertExpectations(t)
}

func TestOrderService_VerifyWebhookEventSignature_Valid(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockCartRepo := new(MockCartRepository)
	mockHTTPClient := new(MockHTTPClient)

	orderService := services.NewOrderService(mockOrderRepo, mockCartRepo, models.OrderConfig{
		StripeWebhookSigningSecret: "test_signing_secret",
	}, mockHTTPClient)

	payload := []byte(`{"id":"evt_test_123","type":"payment_intent.succeeded"}`)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// Compute the expected signature with correct timestamp and payload
	expectedSignature := services.ComputeSignature(time.Unix(time.Now().Unix(), 0), payload, "test_signing_secret")
	signature := fmt.Sprintf("t=%s,v1=%s", timestamp, hex.EncodeToString(expectedSignature))

	// Call VerifyWebhookEventSignature
	err := orderService.VerifyWebhookEventSignature(payload, signature)

	// Assertions
	assert.NoError(t, err, "VerifyWebhookEventSignature should not return an error for a valid signature")
}

func TestOrderService_VerifyWebhookEventSignature_Invalid(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockCartRepo := new(MockCartRepository)
	mockHTTPClient := new(MockHTTPClient)

	orderService := services.NewOrderService(mockOrderRepo, mockCartRepo, models.OrderConfig{
		StripeWebhookSigningSecret: "test_signing_secret",
	}, mockHTTPClient)

	payload := []byte(`{"id":"evt_test_123","type":"payment_intent.succeeded"}`)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// Compute an invalid signature
	invalidSignature := fmt.Sprintf("t=%s,v1=%s", timestamp, hex.EncodeToString([]byte("invalid_signature")))

	// Call VerifyWebhookEventSignature
	err := orderService.VerifyWebhookEventSignature(payload, invalidSignature)

	// Assertions
	assert.Error(t, err, "VerifyWebhookEventSignature should return an error for an invalid signature")
	assert.Contains(t, err.Error(), "signature verification failed", "Error message should indicate signature verification failure")
}

func TestOrderService_ProcessWebhookEvent(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockCartRepo := new(MockCartRepository)
	mockHTTPClient := new(MockHTTPClient)

	orderService := services.NewOrderService(mockOrderRepo, mockCartRepo, models.OrderConfig{}, mockHTTPClient)

	event := models.StripeWebhookEvent{
		ID:   "evt_test_123",
		Type: "payment_intent.succeeded",
		Data: &models.StripeWebhookData{
			Object: models.StripeWebhookPaymentIntent{
				ID:       "pi_test_123",
				Amount:   2000,
				Currency: "usd",
			},
		},
		Livemode: false,
		Created:  time.Now().Unix(),
	}

	// Mock CreateWebhookEvent to succeed
	mockOrderRepo.On("CreateWebhookEvent", mock.Anything, event).Return(nil)

	// Mock GetOrder to return an order matching the payment intent
	mockOrder := &models.Order{
		ID:              "order123",
		UserID:          "user123",
		PaymentIntentID: "pi_test_123",
		TotalAmount:     2000,
		Currency:        "usd",
		Status:          models.OrderPending,
	}
	mockOrderRepo.On("GetOrder", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(1).(*models.Order)
		*arg = *mockOrder
	})

	// Mock UpdateOrder to succeed
	mockOrderRepo.On("UpdateOrder", mock.Anything, mock.Anything).Return(nil)

	// Mock ClearCart to succeed
	mockCartRepo.On("ClearCart", mock.Anything, "user123").Return(nil)

	// Call ProcessWebhookEvent
	err := orderService.ProcessWebhookEvent(context.Background(), event)

	// Assertions
	assert.NoError(t, err, "ProcessWebhookEvent should not return an error for a valid event")
	mockOrderRepo.AssertExpectations(t)
	mockCartRepo.AssertExpectations(t)
}
