package repositories

import (
	"context"
	"testing"

	"github.com/dgyurics/marketplace/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateProduct(t *testing.T) {
	t.Parallel()
	repo := NewProductRepository(dbPool)
	ctx := context.Background()

	product := &types.Product{
		Name:        "Test Product",
		Price:       100000,
		Description: "A test product description",
	}

	err := repo.CreateProduct(ctx, product)
	require.NoError(t, err, "Failed to create product")
	require.NotEmpty(t, product.ID, "Product ID should not be empty")

	// Verify product exists
	storedProduct, err := repo.GetProductByID(ctx, product.ID)
	require.NoError(t, err, "Failed to fetch created product")
	assert.Equal(t, product.Name, storedProduct.Name)
	assert.Equal(t, product.Description, storedProduct.Description)
	assert.Equal(t, product.Price, storedProduct.Price)

	// Ensure images field is properly unmarshaled (since we are using a materialized view)
	assert.NotNil(t, storedProduct.Images, "Images should not be nil")

	// Cleanup
	_, err = dbPool.ExecContext(ctx, "DELETE FROM products WHERE id = $1", product.ID)
	require.NoError(t, err, "Failed to delete test product")
}

func TestCreateProductWithImages(t *testing.T) {
	repo := NewProductRepository(dbPool)
	ctx := context.Background()

	// Create a test category for associating images with the product
	catRepo := NewCategoryRepository(dbPool)
	categoryID, err := catRepo.CreateCategory(ctx, types.Category{
		Name:        "Test Category for Images",
		Description: "A test category for images",
	})
	require.NoError(t, err, "Expected no error on category creation")

	product := &types.Product{
		Name:        "Test Product with Images",
		Price:       100000,
		Description: "Product with images for testing",
		Images: []types.Image{
			{ImageURL: "http://example.com/image1.jpg", Animated: false, AltText: func(s string) *string { return &s }("Image 1")},
			{ImageURL: "http://example.com/image2.gif", Animated: true, AltText: func(s string) *string { return &s }("Image 2 animated")},
		},
	}

	err = repo.CreateProductWithCategory(ctx, product, categoryID)
	require.NoError(t, err, "Expected no error on product creation with images")
	require.NotEmpty(t, product.ID, "Expected product ID to be set")

	// Retrieve the product from the database
	retrievedProduct, err := repo.GetProductByID(ctx, product.ID)
	require.NoError(t, err, "Expected no error on fetching product by ID")
	require.NotNil(t, retrievedProduct, "Expected retrieved product to not be nil")

	// Verify that the images were inserted correctly
	assert.Len(t, retrievedProduct.Images, 2, "Expected two images to be associated with the product")

	assert.Equal(t, "http://example.com/image1.jpg", retrievedProduct.Images[0].ImageURL, "Expected image URL to match for first image")
	assert.Equal(t, false, retrievedProduct.Images[0].Animated, "Expected animated flag to be false for first image")
	assert.Equal(t, "Image 1", *retrievedProduct.Images[0].AltText, "Expected alt text to match for first image")

	assert.Equal(t, "http://example.com/image2.gif", retrievedProduct.Images[1].ImageURL, "Expected image URL to match for second image")
	assert.Equal(t, true, retrievedProduct.Images[1].Animated, "Expected animated flag to be true for second image")
	assert.Equal(t, "Image 2 animated", *retrievedProduct.Images[1].AltText, "Expected alt text to match for second image")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM products WHERE id = $1", product.ID)
	require.NoError(t, err, "Expected no error on product deletion")
}

func TestCreateProductWithCategory(t *testing.T) {
	repo := NewProductRepository(dbPool)
	ctx := context.Background()

	product := &types.Product{
		Name:        "Test Product with Category",
		Price:       150000,
		Description: "A test product with category description",
	}

	// create category
	catRepo := NewCategoryRepository(dbPool)
	categoryID, err := catRepo.CreateCategory(ctx, types.Category{
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

func TestGetProducts(t *testing.T) {
	repo := NewProductRepository(dbPool)
	ctx := context.Background()

	// Add a test product for retrieval
	product := &types.Product{
		Name:        "Test Product for GetAll",
		Price:       200000,
		Description: "A test product for get all",
	}

	err := repo.CreateProduct(ctx, product)
	assert.NoError(t, err, "Expected no error on product creation")

	// Get all products
	products, err := repo.GetProducts(ctx, types.ProductFilter{Limit: 100, Page: 1})
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
	product := &types.Product{
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
	product := &types.Product{
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

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM products WHERE id = $1", product.ID)
	assert.NoError(t, err, "Expected no error on product deletion")
}
