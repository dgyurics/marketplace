package repositories

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateProductWithImages(t *testing.T) {
	repo := NewProductRepository(dbPool)
	ctx := context.Background()

	// Create a test category for associating images with the product
	catRepo := NewCategoryRepository(dbPool)
	categoryID, _ := utilities.GenerateIDString()
	category := types.Category{
		ID:          categoryID,
		Name:        "Test Category for Images",
		Slug:        "test-category-images",
		Description: "A test category for images",
	}
	err := catRepo.CreateCategory(ctx, &category)
	require.NoError(t, err, "Expected no error on category creation")

	product := &types.Product{
		Name:        "Test Product with Images",
		Price:       100000,
		Description: "Product with images for testing",
		Details:     []byte(`{"key": "value"}`),
		Images: []types.Image{
			{ID: utilities.MustGenerateIDString(), URL: "http://example.com/image1.jpg", AltText: func(s string) *string { return &s }("Image 1"), Type: "hero", Source: "image1.jpg"},
			{ID: utilities.MustGenerateIDString(), URL: "http://example.com/image2.gif", AltText: func(s string) *string { return &s }("Image 2 animated"), Type: "thumbnail", Source: "image2.gif"},
		},
	}

	product.ID, _ = utilities.GenerateIDString()

	err = repo.CreateProduct(ctx, product, category.Slug)
	require.NoError(t, err, "Expected no error on product creation with images")
	require.NotEmpty(t, product.ID, "Expected product ID to be set")

	// Retrieve the product from the database
	retrievedProduct, err := repo.GetProductByID(ctx, product.ID)
	require.NoError(t, err, "Expected no error on fetching product by ID")
	require.NotNil(t, retrievedProduct, "Expected retrieved product to not be nil")

	var images []types.Image
	err = json.Unmarshal(retrievedProduct.Images, &images)
	require.NoError(t, err, "Expected no error on unmarshaling images")
	assert.Equal(t, 2, len(images), "Expected two images to be present")

	assert.Equal(t, "http://example.com/image1.jpg", images[0].URL, "Expected image URL to match for first image")
	assert.Equal(t, "Image 1", *images[0].AltText, "Expected alt text to match for first image")

	assert.Equal(t, "http://example.com/image2.gif", images[1].URL, "Expected image URL to match for second image")
	assert.Equal(t, "Image 2 animated", *images[1].AltText, "Expected alt text to match for second image")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM products WHERE id = $1", product.ID)
	require.NoError(t, err, "Expected no error on product deletion")

	_, err = dbPool.ExecContext(ctx, "DELETE FROM categories WHERE id = $1", categoryID)
	require.NoError(t, err, "Expected no error on category deletion")
}

func TestCreateProduct(t *testing.T) {
	repo := NewProductRepository(dbPool)
	ctx := context.Background()

	product := &types.Product{
		Name:        "Test Product with Category",
		Price:       150000,
		Description: "A test product with category description",
		Details:     []byte(`{"key": "value"}`),
	}
	product.ID, _ = utilities.GenerateIDString()

	// create category
	catRepo := NewCategoryRepository(dbPool)
	categoryID, _ := utilities.GenerateIDString()
	category := types.Category{
		ID:          categoryID,
		Name:        "Test Category",
		Slug:        "test-category-23e32",
		Description: "A test category",
	}
	err := catRepo.CreateCategory(ctx, &category)
	assert.NoError(t, err, "Expected no error on category creation")

	err = repo.CreateProduct(ctx, product, category.Slug)
	assert.NoError(t, err, "Expected no error on product creation with category")
	assert.NotEmpty(t, product.ID, "Expected product ID to be set")
	assert.Equal(t, "Test Product with Category", product.Name, "Expected product name to match")
	assert.Equal(t, "A test product with category description", product.Description, "Expected product description to match")
	expectedPrice := int64(150000)
	assert.Equal(t, expectedPrice, product.Price, "Expected product price to match")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM products WHERE id = $1", product.ID)
	assert.NoError(t, err, "Expected no error on product deletion")

	_, err = dbPool.ExecContext(ctx, "DELETE FROM categories WHERE id = $1", categoryID)
	assert.NoError(t, err, "Expected no error on category deletion")
}

func TestGetProducts(t *testing.T) {
	repo := NewProductRepository(dbPool)
	ctx := context.Background()

	// Add a test product for retrieval
	product := &types.Product{
		Name:        "Test Product for GetAll",
		Price:       200000,
		Description: "A test product for get all",
		Details:     []byte(`{"key": "value"}`),
	}
	product.ID, _ = utilities.GenerateIDString()

	// create category
	catRepo := NewCategoryRepository(dbPool)
	categoryID, _ := utilities.GenerateIDString()
	category := types.Category{
		ID:          categoryID,
		Name:        "Test Category",
		Slug:        "test-category-1111",
		Description: "A test category",
	}
	err := catRepo.CreateCategory(ctx, &category)
	assert.NoError(t, err, "Expected no error on category creation")

	err = repo.CreateProduct(ctx, product, category.Slug)
	assert.NoError(t, err, "Expected no error on product creation")
	assert.NotEmpty(t, product.ID, "Expected product ID to be set")

	// Get all products
	products, err := repo.GetProducts(ctx, types.ProductFilter{Limit: 1000, Page: 1})
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
	_, err = dbPool.ExecContext(ctx, "DELETE FROM categories WHERE id = $1", categoryID)
	assert.NoError(t, err, "Expected no error on category deletion")
}

func TestGetProductsByCategory(t *testing.T) {
	repo := NewProductRepository(dbPool)
	ctx := context.Background()

	// Create a test category
	catRepo := NewCategoryRepository(dbPool)
	categoryID, _ := utilities.GenerateIDString()
	category := types.Category{
		ID:          categoryID,
		Name:        "Test Category for Filtering",
		Slug:        "test-category-filtering",
		Description: "Category for testing product filtering",
	}
	err := catRepo.CreateCategory(ctx, &category)
	require.NoError(t, err, "Expected no error on category creation")

	// Create a product linked to that category
	product := &types.Product{
		Name:        "Test Product with Category",
		Price:       200000,
		Description: "A test product with a category",
		Details:     []byte(`{"key": "value"}`),
	}
	product.ID, _ = utilities.GenerateIDString()

	err = repo.CreateProduct(ctx, product, category.Slug)
	require.NoError(t, err, "Expected no error on product creation with category")
	require.NotEmpty(t, product.ID, "Expected product ID to be set")

	// Get products filtered by the category slug
	products, err := repo.GetProducts(ctx, types.ProductFilter{
		Categories: []string{"test-category-filtering"},
		Limit:      100,
		Page:       1,
	})
	require.NoError(t, err, "Expected no error on getting products by category")
	require.NotEmpty(t, products, "Expected products list to not be empty")

	// Verify that the product exists in the list
	var found bool
	for _, p := range products {
		if p.ID == product.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected to find the created product in the filtered products list")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM products WHERE id = $1", product.ID)
	assert.NoError(t, err, "Expected no error on product deletion")

	_, err = dbPool.ExecContext(ctx, "DELETE FROM categories WHERE id = $1", categoryID)
	assert.NoError(t, err, "Expected no error on category deletion")
}

func TestGetProductByID(t *testing.T) {
	repo := NewProductRepository(dbPool)
	ctx := context.Background()

	// Add a test product for retrieval
	product := &types.Product{
		Name:        "Test Product for GetByID",
		Price:       250000,
		Description: "A test product for get by ID",
		Details:     []byte(`{"key": "value"}`),
	}
	product.ID, _ = utilities.GenerateIDString()

	// create category
	catRepo := NewCategoryRepository(dbPool)
	categoryID, _ := utilities.GenerateIDString()
	category := types.Category{
		ID:          categoryID,
		Name:        "Test Category",
		Slug:        "test-category-0000",
		Description: "A test category",
	}
	err := catRepo.CreateCategory(ctx, &category)
	assert.NoError(t, err, "Expected no error on category creation")

	err = repo.CreateProduct(ctx, product, category.Slug)
	assert.NoError(t, err, "Expected no error on product creation")

	// Get product by ID
	retrievedProduct, err := repo.GetProductByID(ctx, product.ID)
	assert.NoError(t, err, "Expected no error on getting product by ID")
	assert.NotNil(t, retrievedProduct, "Expected retrieved product to not be nil")
	assert.NotEmpty(t, retrievedProduct.ID, "Expected retrieved product ID to not be empty")
	assert.Equal(t, product.ID, retrievedProduct.ID, "Expected product ID to match")
	assert.Equal(t, product.Name, retrievedProduct.Name, "Expected product name to match")
	assert.Equal(t, product.Description, retrievedProduct.Description, "Expected product description to match")
	assert.Equal(t, product.Price, retrievedProduct.Price, "Expected product price to match")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM products WHERE id = $1", product.ID)
	assert.NoError(t, err, "Expected no error on product deletion")
	_, err = dbPool.ExecContext(ctx, "DELETE FROM categories WHERE id = $1", categoryID)
	assert.NoError(t, err, "Expected no error on category deletion")
}

func TestDeleteProduct(t *testing.T) {
	repo := NewProductRepository(dbPool)
	ctx := context.Background()

	// Add a test product for deletion
	product := &types.Product{
		Name:        "Test Product for Deletion",
		Price:       300000,
		Description: "A test product for deletion",
		Details:     []byte(`{"key": "value"}`),
	}
	product.ID, _ = utilities.GenerateIDString()

	// create category
	catRepo := NewCategoryRepository(dbPool)
	categoryID, _ := utilities.GenerateIDString()
	category := types.Category{
		ID:          categoryID,
		Name:        "Test Category",
		Slug:        "test-category-033003",
		Description: "A test category",
	}
	err := catRepo.CreateCategory(ctx, &category)
	assert.NoError(t, err, "Expected no error on category creation")

	err = repo.CreateProduct(ctx, product, category.Slug)
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
	_, err = dbPool.ExecContext(ctx, "DELETE FROM categories WHERE id = $1", categoryID)
	assert.NoError(t, err, "Expected no error on category deletion")
}
