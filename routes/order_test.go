package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgyurics/marketplace/models"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) VerifyWebhookEventSignature(payload []byte, sigHeader string) error {
	args := m.Called(payload, sigHeader)
	return args.Error(0)
}

func (m *MockOrderService) ProcessWebhookEvent(ctx context.Context, event models.StripeWebhookEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockOrderService) CreateOrder(ctx context.Context) (models.PaymentIntent, error) {
	args := m.Called(ctx)
	return args.Get(0).(models.PaymentIntent), args.Error(1)
}

func (m *MockOrderService) GetOrders(ctx context.Context) ([]models.Order, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Order), args.Error(1)
}

func TestCreateOrder(t *testing.T) {
	mockOrderService := new(MockOrderService)
	routes := &OrderRoutes{
		orderService: mockOrderService,
		router: router{
			muxRouter:      mux.NewRouter(),
			authMiddleware: nil,
		},
	}

	// Prepare mock data for the request and expected response
	expectedResponse := models.PaymentIntent{
		Status: "pending",
	}

	// Mock the CreateOrder method to return a successful PaymentIntentResponse
	mockOrderService.On("CreateOrder", mock.Anything).Return(expectedResponse, nil)

	// Create a new HTTP POST request for CreateOrder
	req, err := http.NewRequest(http.MethodPost, "/orders", nil)
	require.NoError(t, err)

	// Set up a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Add the route to the mux router
	routes.muxRouter.HandleFunc("/orders", routes.CreateOrder).Methods(http.MethodPost)

	// Serve the request via the router
	routes.muxRouter.ServeHTTP(rr, req)

	// Check that the status code is HTTP 200 OK
	require.Equal(t, http.StatusOK, rr.Code)

	// Decode the response body
	var response models.PaymentIntent
	err = json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)

	// Verify that the response matches the expected response
	require.Equal(t, expectedResponse.Status, response.Status)

	// Assert that the mock's expectations were met
	mockOrderService.AssertExpectations(t)
}
