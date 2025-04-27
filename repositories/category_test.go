package repositories

import (
	"context"
	"testing"

	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
	"github.com/stretchr/testify/assert"
)

func TestCreateCategory(t *testing.T) {
	repo := NewCategoryRepository(dbPool)
	ctx := context.Background()
	category := types.Category{
		ID:          utilities.MustGenerateIDString(),
		Name:        "Test Category",
		Description: "A test category description",
	}

	err := repo.CreateCategory(ctx, &category)
	assert.NoError(t, err, "Expected no error on category creation")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM categories WHERE id = $1", category.ID)
	assert.NoError(t, err, "Expected no error on category deletion")
}

func TestGetAllCategories(t *testing.T) {
	repo := NewCategoryRepository(dbPool)
	ctx := context.Background()
	category := types.Category{
		ID:          utilities.MustGenerateIDString(),
		Name:        "Test Category for GetAll",
		Description: "A test category for get all",
	}

	err := repo.CreateCategory(ctx, &category)
	assert.NoError(t, err, "Expected no error on category creation")

	// Get all categories
	categories, err := repo.GetAllCategories(ctx)
	assert.NoError(t, err, "Expected no error on getting all categories")
	assert.NotEmpty(t, categories, "Expected categories list to not be empty")

	// Verify the category exists in the list
	var found bool
	for _, c := range categories {
		if c.ID == category.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected to find the created category in the categories list")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM categories WHERE id = $1", category.ID)
	assert.NoError(t, err, "Expected no error on category deletion")
}

func TestGetCategoryByID(t *testing.T) {
	repo := NewCategoryRepository(dbPool)
	ctx := context.Background()
	category := types.Category{
		ID:          utilities.MustGenerateIDString(),
		Name:        "Test Category for GetByID",
		Description: "A test category for get by ID",
	}

	err := repo.CreateCategory(ctx, &category)
	assert.NoError(t, err, "Expected no error on category creation")

	// Get category by ID
	retrievedCategory, err := repo.GetCategoryByID(ctx, category.ID)
	assert.NoError(t, err, "Expected no error on getting category by ID")
	assert.NotNil(t, retrievedCategory, "Expected retrieved category to not be nil")
	assert.Equal(t, category.ID, retrievedCategory.ID, "Expected category ID to match")
	assert.Equal(t, category.Name, retrievedCategory.Name, "Expected category name to match")
	assert.Equal(t, category.Description, retrievedCategory.Description, "Expected category description to match")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM categories WHERE id = $1", category.ID)
	assert.NoError(t, err, "Expected no error on category deletion")
}
