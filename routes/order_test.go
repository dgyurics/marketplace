package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgyurics/marketplace/types"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) VerifyStripeEventSignature(payload []byte, sigHeader string) error {
	args := m.Called(payload, sigHeader)
	return args.Error(0)
}

func (m *MockOrderService) ProcessStripeEvent(ctx context.Context, event types.StripeEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockOrderService) CreateOrder(ctx context.Context, addressID string) (types.PaymentIntent, error) {
	args := m.Called(ctx, addressID)
	return args.Get(0).(types.PaymentIntent), args.Error(1)
}

func (m *MockOrderService) GetOrders(ctx context.Context, page, limit int) ([]types.Order, error) {
	args := m.Called(ctx, page, limit)
	return args.Get(0).([]types.Order), args.Error(1)
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
	expectedAddressID := "address123"
	expectedResponse := types.PaymentIntent{
		Status: "pending",
	}

	// Mock the CreateOrder method to return a successful PaymentIntentResponse
	mockOrderService.On("CreateOrder", mock.Anything, expectedAddressID).Return(expectedResponse, nil)

	// Create a new HTTP POST request with a JSON body containing the addressID
	requestBody := map[string]string{
		"address_id": expectedAddressID,
	}
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Set up a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Add the route to the mux router
	routes.muxRouter.HandleFunc("/orders", routes.CreateOrder).Methods(http.MethodPost)

	// Serve the request via the router
	routes.muxRouter.ServeHTTP(rr, req)

	// Check that the status code is HTTP 200 OK
	require.Equal(t, http.StatusOK, rr.Code)

	// Decode the response body
	var response types.PaymentIntent
	err = json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)

	// Verify that the response matches the expected response
	require.Equal(t, expectedResponse.Status, response.Status)

	// Assert that the mock's expectations were met
	mockOrderService.AssertExpectations(t)
}
