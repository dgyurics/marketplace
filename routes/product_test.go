package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgyurics/marketplace/types"
	util "github.com/dgyurics/marketplace/utilities"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type dummyAuth struct{}

func (d dummyAuth) AuthenticateUser(next http.HandlerFunc) http.HandlerFunc {
	return next
}
func (d dummyAuth) AuthenticateAdmin(next http.HandlerFunc) http.HandlerFunc {
	return next
}

func (d dummyAuth) RequireRole(role types.Role) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return next
	}
}

// Mocking the ProductService
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

// TODO rewrite tests to use actual endpoint

func TestGetProductByID(t *testing.T) {
	// Create a mock service
	mockService := new(MockProductService)

	// Set up the routes with the mock service
	routes := &ProductRoutes{
		productService: mockService,
		router: router{
			muxRouter:      mux.NewRouter(),
			authMiddleware: &dummyAuth{},
		},
	}

	// Register all routes
	routes.RegisterRoutes()

	// Create a sample product that will be returned by the mock service
	product := &types.Product{
		ID:          "1",
		Name:        "Test Product",
		Price:       100000,
		Summary:     "This is a test product summary",
		Description: util.String("This is a test product description"),
		Inventory:   10,
	}

	// Set up the expected behavior of the mock service
	mockService.On("GetProductByID", mock.Anything, "1").Return(*product, nil)

	// Create a new HTTP request with the product ID in the URL
	req, err := http.NewRequest(http.MethodGet, "/products/1", nil)
	require.NoError(t, err)

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Serve the request via the router
	routes.muxRouter.ServeHTTP(rr, req)

	// Check the status code is what you expect
	require.Equal(t, http.StatusOK, rr.Code)

	// Check the response body is what you expect
	var responseProduct types.Product
	err = json.NewDecoder(rr.Body).Decode(&responseProduct)
	require.NoError(t, err)

	require.Equal(t, product.ID, responseProduct.ID)
	require.Equal(t, product.Name, responseProduct.Name)
	require.Equal(t, product.Price, responseProduct.Price)
	require.Equal(t, product.Summary, responseProduct.Summary)
	require.Equal(t, product.Description, responseProduct.Description)

	// Assert that the mock's expectations were met
	mockService.AssertExpectations(t)
}

func TestGetProducts(t *testing.T) {
	// Create a mock service
	mockService := new(MockProductService)

	// Set up the routes with the mock service
	routes := &ProductRoutes{
		productService: mockService,
		router: router{
			muxRouter:      mux.NewRouter(),
			authMiddleware: nil,
		},
	}

	// Create a sample list of products that will be returned by the mock service
	products := []types.Product{
		{
			ID:          "1",
			Name:        "Test Product 1",
			Price:       100000,
			Summary:     "This is the first test product summary",
			Description: util.String("This is the first test product description"),
		},
		{
			ID:          "2",
			Name:        "Test Product 2",
			Price:       200000,
			Summary:     "This is the second test product summary",
			Description: util.String("This is the second test product description"),
		},
	}

	// Set up the expected behavior of the mock service
	mockService.On("GetProducts", mock.Anything, mock.Anything).Return(products, nil)

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, "/products", nil)
	require.NoError(t, err)

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Add the route to the mux router
	routes.muxRouter.HandleFunc("/products", routes.GetProducts).Methods(http.MethodGet)

	// Serve the request via the router
	routes.muxRouter.ServeHTTP(rr, req)

	// Check the status code is what you expect
	require.Equal(t, http.StatusOK, rr.Code)

	// Check the response body is what you expect
	var responseProducts []types.Product
	err = json.NewDecoder(rr.Body).Decode(&responseProducts)
	require.NoError(t, err)

	require.Len(t, responseProducts, len(products))
	require.Equal(t, products[0].Name, responseProducts[0].Name)
	require.Equal(t, products[1].Name, responseProducts[1].Name)

	// Assert that the mock's expectations were met
	mockService.AssertExpectations(t)
}
