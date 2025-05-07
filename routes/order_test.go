package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/types/stripe"
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

func (m *MockOrderService) ProcessStripeEvent(ctx context.Context, event stripe.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockOrderService) CreateOrder(ctx context.Context, order *types.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderService) GetOrders(ctx context.Context, page, limit int) ([]types.Order, error) {
	args := m.Called(ctx, page, limit)
	return args.Get(0).([]types.Order), args.Error(1)
}

func (m *MockOrderService) CancelStaleOrders(ctx context.Context) {
	m.Called(ctx)
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
	expectedOrder := &types.Order{
		Address: &types.Address{ID: expectedAddressID},
		Email:   "testemail@email.com",
	}

	// Mock the CreateOrder method
	mockOrderService.On("CreateOrder", mock.Anything, expectedOrder).Return(nil)

	// Create a new HTTP POST request with a JSON body containing the addressID
	requestBody := map[string]string{
		"address_id": expectedAddressID,
		"email":      "testemail@email.com",
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
	var response types.Order
	err = json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)

	// Assert that the mock's expectations were met
	mockOrderService.AssertExpectations(t)
}
