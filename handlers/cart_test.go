package handlers

import (
	"bytes"
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

type MockCartService struct {
	mock.Mock
}

func (m *MockCartService) AddItemToCart(ctx context.Context, item *models.CartItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockCartService) GetCart(ctx context.Context) (*models.Cart, error) {
	args := m.Called(ctx)
	return args.Get(0).(*models.Cart), args.Error(1)
}

func (m *MockCartService) UpdateCartItem(ctx context.Context, item *models.CartItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockCartService) RemoveItemFromCart(ctx context.Context, productID string) error {
	args := m.Called(ctx, productID)
	return args.Error(0)
}

func (m *MockCartService) ClearCart(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCartService) CheckOut(ctx context.Context) (models.PaymentIntentResponse, error) {
	args := m.Called(ctx)
	return args.Get(0).(models.PaymentIntentResponse), args.Error(1)
}

func (m *MockCartService) ConfirmPayment(ctx context.Context, paymentIntentID string) error {
	args := m.Called(ctx, paymentIntentID)
	return args.Error(0)
}

func TestAddItemToCart(t *testing.T) {
	mockCartService := new(MockCartService)
	router := mux.NewRouter()
	handler := &cartHandler{
		cartService: mockCartService,
		router:      router,
	}

	item := models.CartItem{
		ProductID: "1c2d6b57-5e1b-4f29-bb38-dbb4b065e5e8",
		Quantity:  2,
	}

	mockCartService.On("AddItemToCart", mock.Anything, &item).Return(nil)

	// Create a new HTTP POST request with the cart ID in the URL and the item as the payload
	payload, _ := json.Marshal(item)
	req, err := http.NewRequest(http.MethodPost, "/carts/items", bytes.NewBuffer(payload))
	require.NoError(t, err)

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Add the route to the mux router
	handler.router.HandleFunc("/carts/items", handler.AddItemToCart).Methods(http.MethodPost)

	// Serve the request via the router
	handler.router.ServeHTTP(rr, req)

	// Check the status code is what you expect
	require.Equal(t, http.StatusCreated, rr.Code)

	// Assert that the mock's expectations were met
	mockCartService.AssertExpectations(t)
}

func TestRemoveItemFromCart(t *testing.T) {
	mockCartService := new(MockCartService)
	router := mux.NewRouter()
	handler := &cartHandler{
		cartService: mockCartService,
		router:      router,
	}

	productID := "test-product-id"
	mockCartService.On("RemoveItemFromCart", mock.Anything, productID).Return(nil)

	// Create a new HTTP DELETE request with cartID and productID in the URL
	req, err := http.NewRequest(http.MethodDelete, "/carts/items/"+productID, nil)
	require.NoError(t, err)

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Add the route to the mux router
	handler.router.HandleFunc("/carts/items/{product_id}", handler.RemoveItemFromCart).Methods(http.MethodDelete)

	// Serve the request via the router
	handler.router.ServeHTTP(rr, req)

	// Check the status code is what you expect
	require.Equal(t, http.StatusOK, rr.Code)

	// Assert that the mock's expectations were met
	mockCartService.AssertExpectations(t)
}

func TestGetCart(t *testing.T) {
	mockCartService := new(MockCartService)
	router := mux.NewRouter()
	handler := &cartHandler{
		cartService: mockCartService,
		router:      router,
	}

	expectedCart := &models.Cart{
		UserID: "test-user-id",
		Items:  []models.CartItem{{ProductID: "1c2d6b57-5e1b-4f29-bb38-dbb4b065e5e8", Quantity: 2}},
	}

	mockCartService.On("GetCart", mock.Anything).Return(expectedCart, nil)

	// Create a new HTTP GET request with cartID in the URL
	req, err := http.NewRequest(http.MethodGet, "/carts", nil)
	require.NoError(t, err)

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Add the route to the mux router
	handler.router.HandleFunc("/carts", handler.GetCart).Methods(http.MethodGet)

	// Serve the request via the router
	handler.router.ServeHTTP(rr, req)

	// Check the status code is what you expect
	require.Equal(t, http.StatusOK, rr.Code)

	// Decode the response body
	var responseCart models.Cart
	err = json.NewDecoder(rr.Body).Decode(&responseCart)
	require.NoError(t, err)

	// Verify the cart details
	require.Equal(t, expectedCart.UserID, responseCart.UserID)

	// Assert that the mock's expectations were met
	mockCartService.AssertExpectations(t)
}

func TestCheckout(t *testing.T) {
	mockCartService := new(MockCartService)
	router := mux.NewRouter()
	handler := &cartHandler{
		cartService: mockCartService,
		router:      router,
	}

	// Prepare mock data for the request and expected response
	expectedResponse := models.PaymentIntentResponse{
		Status: "pending",
	}

	// Mock the CheckOut method to return a successful PaymentIntentResponse
	mockCartService.On("CheckOut", mock.Anything).Return(expectedResponse, nil)

	// Create a new HTTP POST request for checkout
	req, err := http.NewRequest(http.MethodPost, "/carts/checkout", nil)
	require.NoError(t, err)

	// Set up a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Add the route to the mux router
	handler.router.HandleFunc("/carts/checkout", handler.Checkout).Methods(http.MethodPost)

	// Serve the request via the router
	handler.router.ServeHTTP(rr, req)

	// Check that the status code is HTTP 200 OK
	require.Equal(t, http.StatusOK, rr.Code)

	// Decode the response body
	var response models.PaymentIntentResponse
	err = json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)

	// Verify that the response matches the expected response
	require.Equal(t, expectedResponse.Status, response.Status)

	// Assert that the mock's expectations were met
	mockCartService.AssertExpectations(t)
}
