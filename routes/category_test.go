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

type MockCategoryService struct {
	mock.Mock
}

func (m *MockCategoryService) CreateCategory(ctx context.Context, category *types.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockCategoryService) UpdateCategory(ctx context.Context, category types.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockCategoryService) GetAllCategories(ctx context.Context) ([]types.Category, error) {
	args := m.Called(ctx)
	return args.Get(0).([]types.Category), args.Error(1)
}

func (m *MockCategoryService) GetCategoryByID(ctx context.Context, id string) (*types.Category, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*types.Category), args.Error(1)
}

func (m *MockCategoryService) RemoveCategory(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(1)
}

func TestCreateCategory(t *testing.T) {
	// Create a mock service
	mockService := new(MockCategoryService)

	// Set up the routes with the mock service
	routes := &CategoryRoutes{
		categoryService: mockService,
		router: router{
			muxRouter:      mux.NewRouter(),
			authMiddleware: nil,
		},
	}

	// Set up the expected behavior of the mock service
	mockService.On("CreateCategory", mock.Anything, mock.Anything).Return(nil)

	// Create a new category as the request payload
	category := types.Category{
		Name: "Test Category",
	}
	payload, _ := json.Marshal(category)

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(payload))
	require.NoError(t, err)

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Call the routes CreateCategory method directly
	routes.CreateCategory(rr, req)

	// Check the status code is what you expect
	require.Equal(t, http.StatusCreated, rr.Code)

	// Check the response body is what you expect
	var responseCategory types.Category
	err = json.NewDecoder(rr.Body).Decode(&responseCategory)
	require.NoError(t, err)
	require.Equal(t, category.Name, responseCategory.Name)

	// Assert that the mock's expectations were met
	mockService.AssertExpectations(t)
}

func TestGetCategories(t *testing.T) {
	// Create a mock service
	mockService := new(MockCategoryService)

	// Set up the routes with the mock service
	routes := &CategoryRoutes{
		categoryService: mockService,
		router: router{
			muxRouter:      mux.NewRouter(),
			authMiddleware: nil,
		},
	}
	// Create a sample list of categories that will be returned by the mock service
	categories := []types.Category{
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

	// Call the routes GetCategories method via the router
	routes.muxRouter.HandleFunc("/categories", routes.GetCategories).Methods(http.MethodGet)
	routes.muxRouter.ServeHTTP(rr, req)

	// Check the status code is what you expect
	require.Equal(t, http.StatusOK, rr.Code)

	// Check the response body is what you expect
	var responseCategories []types.Category
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

	// Set up the routes with the mock service
	routes := &CategoryRoutes{
		categoryService: mockService,
		router: router{
			muxRouter:      mux.NewRouter(),
			authMiddleware: nil,
		},
	}

	// Create a sample category that will be returned by the mock service
	category := types.Category{
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

	// Call the routes GetCategory method via the router
	routes.muxRouter.HandleFunc("/categories/{id}", routes.GetCategory).Methods(http.MethodGet)
	routes.muxRouter.ServeHTTP(rr, req)

	// Check the status code is what you expect
	require.Equal(t, http.StatusOK, rr.Code)

	// Check the response body is what you expect
	var responseCategory types.Category
	err = json.NewDecoder(rr.Body).Decode(&responseCategory)
	require.NoError(t, err)

	require.Equal(t, category.ID, responseCategory.ID)
	require.Equal(t, category.Name, responseCategory.Name)

	// Assert that the mock's expectations were met
	mockService.AssertExpectations(t)
}
