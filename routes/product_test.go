package routes

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

// Mocking the ProductService
type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) CreateProduct(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductService) CreateProductWithCategory(ctx context.Context, product *models.Product, categoryID string) error {
	args := m.Called(ctx, product, categoryID)
	return args.Error(0)
}

func (m *MockProductService) GetAllProducts(ctx context.Context) ([]models.Product, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Product), args.Error(1)
}

func (m *MockProductService) GetProductByID(ctx context.Context, id string) (*models.Product, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductService) UpdateInventory(ctx context.Context, productID string, quantity int) error {
	args := m.Called(ctx, productID, quantity)
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
	mockService.On("CreateProduct", mock.Anything, mock.AnythingOfType("*models.Product")).Return(nil)

	// Create a new product as the request payload
	product := models.Product{
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
	var responseProduct models.Product
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
	product := &models.Product{
		ID:          "1",
		Name:        "Test Product",
		Price:       100000,
		Description: "This is a test product",
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
	var responseProduct models.Product
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
	products := []models.Product{
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
	mockService.On("GetAllProducts", mock.Anything).Return(products, nil)

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
	var responseProducts []models.Product
	err = json.NewDecoder(rr.Body).Decode(&responseProducts)
	require.NoError(t, err)

	require.Len(t, responseProducts, len(products))
	require.Equal(t, products[0].Name, responseProducts[0].Name)
	require.Equal(t, products[1].Name, responseProducts[1].Name)

	// Assert that the mock's expectations were met
	mockService.AssertExpectations(t)
}