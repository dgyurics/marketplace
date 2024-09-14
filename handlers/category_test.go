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

type MockCategoryService struct {
	mock.Mock
}

func (m *MockCategoryService) CreateCategory(ctx context.Context, category models.Category) (string, error) {
	args := m.Called(ctx, category)
	return args.String(0), args.Error(1)
}

func (m *MockCategoryService) GetAllCategories(ctx context.Context) ([]models.Category, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Category), args.Error(1)
}

func (m *MockCategoryService) GetCategoryByID(ctx context.Context, id string) (*models.Category, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Category), args.Error(1)
}

func (m *MockCategoryService) GetProductsByCategoryID(ctx context.Context, categoryId string) ([]models.Product, error) {
	args := m.Called(ctx, categoryId)
	return args.Get(0).([]models.Product), args.Error(1)
}

func TestCreateCategory(t *testing.T) {
	// Create a mock service
	mockService := new(MockCategoryService)

	// Set up the handler with the mock service
	router := mux.NewRouter()
	handler := &categoryHandler{
		categoryService: mockService,
		router:          router,
	}

	// Set up the expected behavior of the mock service
	mockService.On("CreateCategory", mock.Anything, mock.AnythingOfType("models.Category")).Return("1", nil)

	// Create a new category as the request payload
	category := models.Category{
		Name: "Test Category",
	}
	payload, _ := json.Marshal(category)

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(payload))
	require.NoError(t, err)

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Call the handler's CreateCategory method directly
	handler.CreateCategory(rr, req)

	// Check the status code is what you expect
	require.Equal(t, http.StatusCreated, rr.Code)

	// Check the response body is what you expect
	var responseCategory models.Category
	err = json.NewDecoder(rr.Body).Decode(&responseCategory)
	require.NoError(t, err)

	require.Equal(t, "1", responseCategory.ID)
	require.Equal(t, category.Name, responseCategory.Name)

	// Assert that the mock's expectations were met
	mockService.AssertExpectations(t)
}

func TestGetCategories(t *testing.T) {
	// Create a mock service
	mockService := new(MockCategoryService)

	// Set up the handler with the mock service
	router := mux.NewRouter()
	handler := &categoryHandler{
		categoryService: mockService,
		router:          router,
	}

	// Create a sample list of categories that will be returned by the mock service
	categories := []models.Category{
		{ID: "1", Name: "Category 1"},
		{ID: "2", Name: "Category 2"},
	}

	// Set up the expected behavior of the mock service
	mockService.On("GetAllCategories", mock.Anything).Return(categories, nil)

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, "/categories", nil)
	require.NoError(t, err)

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Call the handler's GetCategories method via the router
	handler.router.HandleFunc("/categories", handler.GetCategories).Methods(http.MethodGet)
	handler.router.ServeHTTP(rr, req)

	// Check the status code is what you expect
	require.Equal(t, http.StatusOK, rr.Code)

	// Check the response body is what you expect
	var responseCategories []models.Category
	err = json.NewDecoder(rr.Body).Decode(&responseCategories)
	require.NoError(t, err)

	require.Len(t, responseCategories, len(categories))
	require.Equal(t, categories[0].Name, responseCategories[0].Name)
	require.Equal(t, categories[1].Name, responseCategories[1].Name)

	// Assert that the mock's expectations were met
	mockService.AssertExpectations(t)
}

func TestGetCategory(t *testing.T) {
	// Create a mock service
	mockService := new(MockCategoryService)

	// Set up the handler with the mock service
	router := mux.NewRouter()
	handler := &categoryHandler{
		categoryService: mockService,
		router:          router,
	}

	// Create a sample category that will be returned by the mock service
	category := models.Category{
		ID:   "1",
		Name: "Test Category",
	}

	// Set up the expected behavior of the mock service
	mockService.On("GetCategoryByID", mock.Anything, "1").Return(&category, nil)

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, "/categories/1", nil)
	require.NoError(t, err)

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Call the handler's GetCategory method via the router
	handler.router.HandleFunc("/categories/{id}", handler.GetCategory).Methods(http.MethodGet)
	handler.router.ServeHTTP(rr, req)

	// Check the status code is what you expect
	require.Equal(t, http.StatusOK, rr.Code)

	// Check the response body is what you expect
	var responseCategory models.Category
	err = json.NewDecoder(rr.Body).Decode(&responseCategory)
	require.NoError(t, err)

	require.Equal(t, category.ID, responseCategory.ID)
	require.Equal(t, category.Name, responseCategory.Name)

	// Assert that the mock's expectations were met
	mockService.AssertExpectations(t)
}

func TestGetProductsByCategory(t *testing.T) {
	// Create a mock service
	mockService := new(MockCategoryService)

	// Set up the handler with the mock service
	router := mux.NewRouter()
	handler := &categoryHandler{
		categoryService: mockService,
		router:          router,
	}

	// Create a sample list of products that will be returned by the mock service
	products := []models.Product{
		{ID: "1", Name: "Product 1", Price: models.Currency{Amount: 1000}},
		{ID: "2", Name: "Product 2", Price: models.Currency{Amount: 2000}},
	}

	// Set up the expected behavior of the mock service
	mockService.On("GetProductsByCategoryID", mock.Anything, "1").Return(products, nil)

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, "/categories/1/products", nil)
	require.NoError(t, err)

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Call the handler's GetProductsByCategory method via the router
	handler.router.HandleFunc("/categories/{id}/products", handler.GetProductsByCategory).Methods(http.MethodGet)
	handler.router.ServeHTTP(rr, req)

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
