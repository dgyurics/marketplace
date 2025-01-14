package repositories

import (
	"context"
	"testing"

	"github.com/dgyurics/marketplace/models"
	"github.com/stretchr/testify/assert"
)

func TestCreateProduct(t *testing.T) {
	repo := NewProductRepository(dbPool)
	ctx := context.Background()

	product := &models.Product{
		Name:        "Test Product",
		Price:       100000,
		Description: "A test product description",
	}

	err := repo.CreateProduct(ctx, product)

	assert.NoError(t, err, "Expected no error on product creation")
	assert.NotEmpty(t, product.ID, "Expected product ID to be set")
	assert.Equal(t, "Test Product", product.Name, "Expected product name to match")
	assert.Equal(t, "A test product description", product.Description, "Expected product description to match")
	expectedPrice := int64(100000)
	assert.Equal(t, expectedPrice, product.Price, "Expected product price to match")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM products WHERE id = $1", product.ID)
	assert.NoError(t, err, "Expected no error on product deletion")
}

func TestCreateProductWithCategory(t *testing.T) {
	repo := NewProductRepository(dbPool)
	ctx := context.Background()

	product := &models.Product{
		Name:        "Test Product with Category",
		Price:       150000,
		Description: "A test product with category description",
	}

	// create category
	catRepo := NewCategoryRepository(dbPool)
	categoryID, err := catRepo.CreateCategory(ctx, models.Category{
		Name:        "Test Category",
		Description: "A test category",
	})
	assert.NoError(t, err, "Expected no error on category creation")

	err = repo.CreateProductWithCategory(ctx, product, categoryID)
	assert.NoError(t, err, "Expected no error on product creation with category")
	assert.NotEmpty(t, product.ID, "Expected product ID to be set")
	assert.Equal(t, "Test Product with Category", product.Name, "Expected product name to match")
	assert.Equal(t, "A test product with category description", product.Description, "Expected product description to match")
	expectedPrice := int64(150000)
	assert.Equal(t, expectedPrice, product.Price, "Expected product price to match")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM products WHERE id = $1", product.ID)
	assert.NoError(t, err, "Expected no error on product deletion")
}

func TestGetAllProducts(t *testing.T) {
	repo := NewProductRepository(dbPool)
	ctx := context.Background()

	// Add a test product for retrieval
	product := &models.Product{
		Name:        "Test Product for GetAll",
		Price:       200000,
		Description: "A test product for get all",
	}

	err := repo.CreateProduct(ctx, product)
	assert.NoError(t, err, "Expected no error on product creation")

	// Get all products
	products, err := repo.GetAllProducts(ctx, 1, 100)
	assert.NoError(t, err, "Expected no error on getting all products")
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
	_, err = dbPool.ExecContext(ctx, "DELETE FROM products WHERE id = $1", product.ID)
	assert.NoError(t, err, "Expected no error on product deletion")
}

func TestGetProductByID(t *testing.T) {
	repo := NewProductRepository(dbPool)
	ctx := context.Background()

	// Add a test product for retrieval
	product := &models.Product{
		Name:        "Test Product for GetByID",
		Price:       250000,
		Description: "A test product for get by ID",
	}

	err := repo.CreateProduct(ctx, product)
	assert.NoError(t, err, "Expected no error on product creation")

	// Get product by ID
	retrievedProduct, err := repo.GetProductByID(ctx, product.ID)
	assert.NoError(t, err, "Expected no error on getting product by ID")
	assert.NotNil(t, retrievedProduct, "Expected retrieved product to not be nil")
	assert.Equal(t, product.ID, retrievedProduct.ID, "Expected product ID to match")
	assert.Equal(t, product.Name, retrievedProduct.Name, "Expected product name to match")
	assert.Equal(t, product.Description, retrievedProduct.Description, "Expected product description to match")
	assert.Equal(t, product.Price, retrievedProduct.Price, "Expected product price to match")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM products WHERE id = $1", product.ID)
	assert.NoError(t, err, "Expected no error on product deletion")
}

func TestDeleteProduct(t *testing.T) {
	repo := NewProductRepository(dbPool)
	ctx := context.Background()

	// Add a test product for deletion
	product := &models.Product{
		Name:        "Test Product for Deletion",
		Price:       300000,
		Description: "A test product for deletion",
	}

	err := repo.CreateProduct(ctx, product)
	assert.NoError(t, err, "Expected no error on product creation")

	// Delete the product
	err = repo.DeleteProduct(ctx, product.ID)
	assert.NoError(t, err, "Expected no error on product deletion")

	// Verify the product no longer exists
	deletedProduct, err := repo.GetProductByID(ctx, product.ID)
	assert.Error(t, err, "Expected an error when getting a deleted product")
	assert.Nil(t, deletedProduct, "Expected deleted product to be nil")
}
