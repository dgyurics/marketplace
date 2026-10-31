package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgyurics/marketplace/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockCartService struct {
	mock.Mock
}

func (m *MockCartService) AddItem(ctx context.Context, item *types.CartItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockCartService) GetItems(ctx context.Context) ([]types.CartItem, error) {
	args := m.Called(ctx)
	return args.Get(0).([]types.CartItem), args.Error(1)
}

func (m *MockCartService) RemoveItem(ctx context.Context, productID string) error {
	args := m.Called(ctx, productID)
	return args.Error(0)
}

func (m *MockCartService) RemoveItems(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestAddItem(t *testing.T) {
	mockCartService := new(MockCartService)
	routes := &CartRoutes{
		cartService: mockCartService,
		router: router{
			mux: http.NewServeMux(),
		},
	}

	productID := "1c2d6b57-5e1b-4f29-bb38-dbb4b065e5e8"
	item := types.CartItem{
		Quantity: 2,
		Product: types.Product{
			ID:      productID,
			Details: json.RawMessage(`{"key":"value"}`),
		},
	}

	mockCartService.On("AddItem", mock.Anything, &item).Return(nil).Once()

	payload, err := json.Marshal(item)
	require.NoError(t, err)
	url := "/carts/items/" + productID
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	rr := httptest.NewRecorder()
	routes.mux.HandleFunc("POST /carts/items/{id}", routes.AddItem)
	routes.mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	mockCartService.AssertExpectations(t)
}

func TestRemoveItem(t *testing.T) {
	mockCartService := new(MockCartService)
	routes := &CartRoutes{
		cartService: mockCartService,
		router: router{
			mux: http.NewServeMux(),
		},
	}

	productID := "test-product-id"
	mockCartService.On("RemoveItem", mock.Anything, productID).Return(nil).Once()

	req := httptest.NewRequest(http.MethodDelete, "/carts/items/"+productID, nil)
	rr := httptest.NewRecorder()
	routes.mux.HandleFunc("DELETE /carts/items/{id}", routes.RemoveItem)
	routes.mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	mockCartService.AssertExpectations(t)
}

func TestGetCartItems(t *testing.T) {
	mockCartService := new(MockCartService)
	routes := &CartRoutes{
		cartService: mockCartService,
		router: router{
			mux: http.NewServeMux(),
		},
	}

	expectedCart := []types.CartItem{
		{
			Product:   types.Product{ID: "1c2d6b57-5e1b-4f29-bb38-dbb4b065e5e8"},
			Quantity:  2,
			UnitPrice: 1000,
		},
	}

	mockCartService.On("GetItems", mock.Anything).Return(expectedCart, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/carts", nil)
	rr := httptest.NewRecorder()
	routes.mux.HandleFunc("GET /carts", routes.GetItems)
	routes.mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	expected, err := json.Marshal(expectedCart)
	require.NoError(t, err)
	require.JSONEq(t, string(expected), rr.Body.String())
	mockCartService.AssertExpectations(t)
}

func TestRemoveItems(t *testing.T) {
	mockCartService := new(MockCartService)
	routes := &CartRoutes{
		cartService: mockCartService,
		router: router{
			mux: http.NewServeMux(),
		},
	}

	mockCartService.On("RemoveItems", mock.Anything).Return(nil).Once()

	req := httptest.NewRequest(http.MethodDelete, "/carts", nil)
	rr := httptest.NewRecorder()
	routes.mux.HandleFunc("DELETE /carts", routes.RemoveItems)
	routes.mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	mockCartService.AssertExpectations(t)
}
