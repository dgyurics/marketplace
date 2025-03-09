package services_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
)

type MockCategoryRepository struct {
	mock.Mock
}

func (m *MockCategoryRepository) CreateCategory(ctx context.Context, category types.Category) (string, error) {
	args := m.Called(ctx, category)
	return args.String(0), args.Error(1)
}

func (m *MockCategoryRepository) GetAllCategories(ctx context.Context) ([]types.Category, error) {
	args := m.Called(ctx)
	return args.Get(0).([]types.Category), args.Error(1)
}

func (m *MockCategoryRepository) GetCategoryByID(ctx context.Context, id string) (*types.Category, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*types.Category), args.Error(1)
}

func TestCategoryService_CreateCategory(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	service := services.NewCategoryService(mockRepo)

	category := types.Category{Name: "Test Category", Description: "A test category"}
	expectedID := "1"

	mockRepo.On("CreateCategory", mock.Anything, category).Return(expectedID, nil)

	newID, err := service.CreateCategory(context.Background(), category)

	assert.NoError(t, err)
	assert.Equal(t, expectedID, newID)

	mockRepo.AssertExpectations(t)
}

func TestCategoryService_GetAllCategories(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	service := services.NewCategoryService(mockRepo)

	expectedCategories := []types.Category{
		{ID: "1", Name: "Category 1", Description: "Description 1"},
		{ID: "2", Name: "Category 2", Description: "Description 2"},
	}

	mockRepo.On("GetAllCategories", mock.Anything).Return(expectedCategories, nil)

	categories, err := service.GetAllCategories(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedCategories, categories)

	mockRepo.AssertExpectations(t)
}

func TestCategoryService_GetCategoryByID(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	service := services.NewCategoryService(mockRepo)

	expectedID := "1"
	expectedCategory := &types.Category{ID: "1", Name: "Category 1", Description: "Description 1"}

	mockRepo.On("GetCategoryByID", mock.Anything, expectedID).Return(expectedCategory, nil)

	category, err := service.GetCategoryByID(context.Background(), expectedID)

	assert.NoError(t, err)
	assert.Equal(t, expectedCategory, category)

	mockRepo.AssertExpectations(t)
}
