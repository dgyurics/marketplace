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

// Mocking the ProductService
type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) CreateProduct(ctx context.Context, product *types.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductService) CreateProductWithCategory(ctx context.Context, product *types.Product, categoryID string) error {
	args := m.Called(ctx, product, categoryID)
	return args.Error(0)
}

func (m *MockProductService) GetProducts(ctx context.Context, filter types.ProductFilter) ([]types.Product, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]types.Product), args.Error(1)
}

func (m *MockProductService) GetProductsByCategory(ctx context.Context, categorySlug string, filter types.ProductFilter) ([]types.Product, error) {
	args := m.Called(ctx, categorySlug, filter)
	return args.Get(0).([]types.Product), args.Error(1)
}

func (m *MockProductService) GetProductByID(ctx context.Context, id string) (*types.ProductWithInventory, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*types.ProductWithInventory), args.Error(1)
}

func (m *MockProductService) UpdateInventory(ctx context.Context, productID string, quantity int) error {
	args := m.Called(ctx, productID, quantity)
	return args.Error(0)
}

func (m *MockProductService) RemoveProduct(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestCreateProduct(t *testing.T) {
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

	// Set up the expected behavior of the mock service
	mockService.On("CreateProduct", mock.Anything, mock.AnythingOfType("*types.Product")).Return(nil)

	// Create a new product as the request payload
	product := types.Product{
		Name:        "Test Product",
		Price:       100000,
		Description: "This is a test product",
	}
	payload, _ := json.Marshal(product)

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(payload))
	require.NoError(t, err)

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Call the router's CreateProduct method directly
	routes.CreateProduct(rr, req)

	// Check the status code is what you expect
	require.Equal(t, http.StatusCreated, rr.Code)

	// Check the response body is what you expect
	var responseProduct types.Product
	err = json.NewDecoder(rr.Body).Decode(&responseProduct)
	require.NoError(t, err)

	require.Equal(t, product.Name, responseProduct.Name)
	require.Equal(t, product.Price, responseProduct.Price)
	require.Equal(t, product.Description, responseProduct.Description)

	// Assert that the mock's expectations were met
	mockService.AssertExpectations(t)
}

func TestGetProductByID(t *testing.T) {
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

	// Create a sample product that will be returned by the mock service
	product := &types.ProductWithInventory{
		ID:          "1",
		Name:        "Test Product",
		Price:       100000,
		Description: "This is a test product",
		Quantity:    10,
	}

	// Set up the expected behavior of the mock service
	mockService.On("GetProductByID", mock.Anything, "1").Return(product, nil)

	// Create a new HTTP request with the product ID in the URL
	req, err := http.NewRequest(http.MethodGet, "/products/1", nil)
	require.NoError(t, err)

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Add the route to the mux router
	routes.muxRouter.HandleFunc("/products/{id}", routes.GetProduct).Methods(http.MethodGet)

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
			Description: "This is the first test product",
		},
		{
			ID:          "2",
			Name:        "Test Product 2",
			Price:       200000,
			Description: "This is the second test product",
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

func TestGetProductsUsingCategory(t *testing.T) {
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

	// Expected products filtered by category "Electronics"
	expectedProducts := []types.Product{
		{
			ID:          "3",
			Name:        "Smartphone",
			Price:       50000,
			Description: "Latest smartphone",
		},
	}

	// Set up the expected behavior of the mock service with filter check
	mockService.On("GetProductsByCategory", mock.Anything, "electronics", mock.Anything).Return(expectedProducts, nil)

	// Create a new HTTP request with category filter
	req, err := http.NewRequest(http.MethodGet, "/products/categories/electronics", nil)
	require.NoError(t, err)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Add the route to the mux router
	routes.muxRouter.HandleFunc("/products/categories/{category}", routes.GetProductsByCategory).Methods(http.MethodGet)

	// Serve the request via the router
	routes.muxRouter.ServeHTTP(rr, req)

	// Check the status code
	require.Equal(t, http.StatusOK, rr.Code)

	// Check the response body
	var responseProducts []types.Product
	err = json.NewDecoder(rr.Body).Decode(&responseProducts)
	require.NoError(t, err)

	require.Len(t, responseProducts, 1)
	require.Equal(t, expectedProducts[0].Name, responseProducts[0].Name)

	// Assert that the mock's expectations were met
	mockService.AssertExpectations(t)
}
