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
	return args.Error(0)
}

func TestCreateCategory(t *testing.T) {
	mockService := new(MockCategoryService)

	routes := &CategoryRoutes{
		categoryService: mockService,
	}

	category := types.Category{
		Name: "Test Category",
	}
	mockService.On("CreateCategory", mock.Anything, &category).Return(nil).Once()

	payload, err := json.Marshal(category)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(payload))
	rr := httptest.NewRecorder()

	routes.CreateCategory(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)

	var responseCategory types.Category
	err = json.NewDecoder(rr.Body).Decode(&responseCategory)
	require.NoError(t, err)
	require.Equal(t, category, responseCategory)
	mockService.AssertExpectations(t)
}

func TestGetCategories(t *testing.T) {
	mockService := new(MockCategoryService)

	routes := &CategoryRoutes{
		categoryService: mockService,
		router: router{
			mux: http.NewServeMux(),
		},
	}

	categories := []types.Category{
		{ID: "1", Name: "Category 1"},
		{ID: "2", Name: "Category 2"},
	}

	mockService.On("GetAllCategories", mock.Anything).Return(categories, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	rr := httptest.NewRecorder()
	routes.mux.HandleFunc("GET /categories", routes.GetCategories)
	routes.mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var responseCategories []types.Category
	err := json.NewDecoder(rr.Body).Decode(&responseCategories)
	require.NoError(t, err)

	require.Equal(t, categories, responseCategories)
	mockService.AssertExpectations(t)
}

func TestGetCategory(t *testing.T) {
	mockService := new(MockCategoryService)

	routes := &CategoryRoutes{
		categoryService: mockService,
		router: router{
			mux: http.NewServeMux(),
		},
	}

	category := types.Category{
		ID:   "1",
		Name: "Test Category",
	}

	mockService.On("GetCategoryByID", mock.Anything, "1").Return(&category, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/categories/1", nil)
	rr := httptest.NewRecorder()
	routes.mux.HandleFunc("GET /categories/{id}", routes.GetCategory)
	routes.mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var responseCategory types.Category
	err := json.NewDecoder(rr.Body).Decode(&responseCategory)
	require.NoError(t, err)

	require.Equal(t, category, responseCategory)
	mockService.AssertExpectations(t)
}
