package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgyurics/marketplace/types"
	util "github.com/dgyurics/marketplace/utilities"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type dummyAuth struct{}

func (d dummyAuth) RequireRole(role types.Role) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return next
	}
}

func (d dummyAuth) OptionalAuth(next http.HandlerFunc) http.HandlerFunc {
	return next
}

type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) CreateProduct(ctx context.Context, product *types.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductService) GetProducts(ctx context.Context, filter types.ProductFilter) ([]types.Product, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]types.Product), args.Error(1)
}

func (m *MockProductService) GetProductByID(ctx context.Context, id string) (types.Product, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(types.Product), args.Error(1)
}

func (m *MockProductService) UpdateProduct(ctx context.Context, product types.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductService) RemoveProduct(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestGetProductByID(t *testing.T) {
	mockService := new(MockProductService)

	routes := &ProductRoutes{
		productService: mockService,
		router: router{
			mux:            http.NewServeMux(),
			authMiddleware: &dummyAuth{},
		},
	}

	routes.RegisterRoutes()

	product := types.Product{
		ID:          "1",
		Name:        "Test Product",
		Price:       100000,
		Summary:     "This is a test product summary",
		Description: util.Ptr("This is a test product description"),
		Inventory:   10,
	}

	mockService.On("GetProductByID", mock.Anything, "1").Return(product, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/products/1", nil)
	rr := httptest.NewRecorder()
	routes.mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	expected, err := json.Marshal(product)
	require.NoError(t, err)
	require.JSONEq(t, string(expected), rr.Body.String())
	mockService.AssertExpectations(t)
}

func TestGetProducts(t *testing.T) {
	mockService := new(MockProductService)

	routes := &ProductRoutes{
		productService: mockService,
		router: router{
			mux: http.NewServeMux(),
		},
	}

	products := []types.Product{
		{
			ID:          "1",
			Name:        "Test Product 1",
			Price:       100000,
			Summary:     "This is the first test product summary",
			Description: util.Ptr("This is the first test product description"),
		},
		{
			ID:          "2",
			Name:        "Test Product 2",
			Price:       200000,
			Summary:     "This is the second test product summary",
			Description: util.Ptr("This is the second test product description"),
		},
	}

	mockService.On("GetProducts", mock.Anything, mock.Anything).Return(products, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rr := httptest.NewRecorder()
	routes.mux.HandleFunc("GET /products", routes.GetProducts)
	routes.mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	expected, err := json.Marshal(products)
	require.NoError(t, err)
	require.JSONEq(t, string(expected), rr.Body.String())

	mockService.AssertExpectations(t)
}
