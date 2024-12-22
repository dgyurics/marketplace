package repositories

import (
	"context"
	"testing"

	"github.com/dgyurics/marketplace/models"
	"github.com/stretchr/testify/assert"
)

func TestCreateCategory(t *testing.T) {
	repo := NewCategoryRepository(dbPool)
	ctx := context.Background()

	category := models.Category{
		Name:        "Test Category",
		Description: "A test category description",
	}

	newID, err := repo.CreateCategory(ctx, category)
	assert.NoError(t, err, "Expected no error on category creation")
	assert.NotEmpty(t, newID, "Expected category ID to be set")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM categories WHERE id = $1", newID)
	assert.NoError(t, err, "Expected no error on category deletion")
}

func TestGetAllCategories(t *testing.T) {
	repo := NewCategoryRepository(dbPool)
	ctx := context.Background()

	// Add a test category for retrieval
	category := models.Category{
		Name:        "Test Category for GetAll",
		Description: "A test category for get all",
	}

	categoryID, err := repo.CreateCategory(ctx, category)
	assert.NoError(t, err, "Expected no error on category creation")

	// Get all categories
	categories, err := repo.GetAllCategories(ctx)
	assert.NoError(t, err, "Expected no error on getting all categories")
	assert.NotEmpty(t, categories, "Expected categories list to not be empty")

	// Verify the category exists in the list
	var found bool
	for _, c := range categories {
		if c.ID == categoryID {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected to find the created category in the categories list")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM categories WHERE id = $1", categoryID)
	assert.NoError(t, err, "Expected no error on category deletion")
}

func TestGetCategoryByID(t *testing.T) {
	repo := NewCategoryRepository(dbPool)
	ctx := context.Background()

	// Add a test category for retrieval
	category := models.Category{
		Name:        "Test Category for GetByID",
		Description: "A test category for get by ID",
	}

	categoryID, err := repo.CreateCategory(ctx, category)
	assert.NoError(t, err, "Expected no error on category creation")

	// Get category by ID
	retrievedCategory, err := repo.GetCategoryByID(ctx, categoryID)
	assert.NoError(t, err, "Expected no error on getting category by ID")
	assert.NotNil(t, retrievedCategory, "Expected retrieved category to not be nil")
	assert.Equal(t, categoryID, retrievedCategory.ID, "Expected category ID to match")
	assert.Equal(t, category.Name, retrievedCategory.Name, "Expected category name to match")
	assert.Equal(t, category.Description, retrievedCategory.Description, "Expected category description to match")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM categories WHERE id = $1", categoryID)
	assert.NoError(t, err, "Expected no error on category deletion")
}

func TestGetProductsByCategoryID(t *testing.T) {
	repo := NewCategoryRepository(dbPool)
	ctx := context.Background()

	// Add a test category
	category := models.Category{
		Name:        "Test Category for Products",
		Description: "A test category for products",
	}

	categoryID, err := repo.CreateCategory(ctx, category)
	assert.NoError(t, err, "Expected no error on category creation")

	// Add a test product associated with the category
	product := &models.Product{
		Name:        "Test Product for Category",
		Price:       500000,
		Description: "A test product for category",
	}

	productRepo := NewProductRepository(dbPool)
	err = productRepo.CreateProductWithCategory(ctx, product, categoryID)
	assert.NoError(t, err, "Expected no error on product creation with category")

	// Get products by category ID
	products, err := repo.GetProductsByCategoryID(ctx, categoryID)
	assert.NoError(t, err, "Expected no error on getting products by category ID")
	assert.NotEmpty(t, products, "Expected products list to not be empty")

	// Verify the product exists in the list
	var found bool
	for _, p := range products {
		if p.ID == product.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected to find the created product in the products list")

	// Clean up
	err = productRepo.DeleteProduct(ctx, product.ID)
	assert.NoError(t, err, "Expected no error on product deletion")

	_, err = dbPool.ExecContext(ctx, "DELETE FROM categories WHERE id = $1", categoryID)
	assert.NoError(t, err, "Expected no error on category deletion")
}
