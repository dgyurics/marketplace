package repositories

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/dgyurics/marketplace/db"
	"github.com/dgyurics/marketplace/models"
	"github.com/stretchr/testify/assert"
)

var dbPool *sql.DB

// Setup the PostgreSQL connection
func TestMain(m *testing.M) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	// Connect to the database
	var err error
	dbPool, err = db.Connect(dbURL)
	if err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	// Run tests
	code := m.Run()
	dbPool.Close()
	os.Exit(code)
}

func TestCreateProduct(t *testing.T) {
	repo := NewProductRepository(dbPool)
	ctx := context.Background()

	product := &models.Product{
		Name:        "Test Product",
		Price:       models.Currency{Amount: 1000},
		Description: "A test product description",
	}

	err := repo.CreateProduct(ctx, product)

	assert.NoError(t, err, "Expected no error on product creation")
	assert.NotEmpty(t, product.ID, "Expected product ID to be set")
	assert.Equal(t, "Test Product", product.Name, "Expected product name to match")
	assert.Equal(t, "A test product description", product.Description, "Expected product description to match")
	expectedPrice := models.Currency{Amount: 1000}
	assert.Equal(t, expectedPrice, product.Price, "Expected product price to match")

	// Clean up
	err = repo.DeleteProduct(ctx, product.ID)
	assert.NoError(t, err, "Expected no error on product deletion")
}

func TestCreateProductWithCategory(t *testing.T) {
	repo := NewProductRepository(dbPool)
	ctx := context.Background()

	product := &models.Product{
		Name:        "Test Product with Category",
		Price:       models.Currency{Amount: 1500},
		Description: "A test product with category description",
	}

	// Reference category ID creating in init.sql
	categoryID := "3d6f0c4a-75bf-4b9b-9f12-003d6f2f9a1f"

	err := repo.CreateProductWithCategory(ctx, product, categoryID)

	assert.NoError(t, err, "Expected no error on product creation with category")
	assert.NotEmpty(t, product.ID, "Expected product ID to be set")
	assert.Equal(t, "Test Product with Category", product.Name, "Expected product name to match")
	assert.Equal(t, "A test product with category description", product.Description, "Expected product description to match")
	expectedPrice := models.Currency{Amount: 1500}
	assert.Equal(t, expectedPrice, product.Price, "Expected product price to match")

	// Clean up
	err = repo.DeleteProduct(ctx, product.ID)
	assert.NoError(t, err, "Expected no error on product deletion")
}

func TestGetAllProducts(t *testing.T) {
	repo := NewProductRepository(dbPool)
	ctx := context.Background()

	// Add a test product for retrieval
	product := &models.Product{
		Name:        "Test Product for GetAll",
		Price:       models.Currency{Amount: 2000},
		Description: "A test product for get all",
	}

	err := repo.CreateProduct(ctx, product)
	assert.NoError(t, err, "Expected no error on product creation")

	// Get all products
	products, err := repo.GetAllProducts(ctx)
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
	err = repo.DeleteProduct(ctx, product.ID)
	assert.NoError(t, err, "Expected no error on product deletion")
}

func TestGetProductByID(t *testing.T) {
	repo := NewProductRepository(dbPool)
	ctx := context.Background()

	// Add a test product for retrieval
	product := &models.Product{
		Name:        "Test Product for GetByID",
		Price:       models.Currency{Amount: 2500},
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
	err = repo.DeleteProduct(ctx, product.ID)
	assert.NoError(t, err, "Expected no error on product deletion")
}

func TestDeleteProduct(t *testing.T) {
	repo := NewProductRepository(dbPool)
	ctx := context.Background()

	// Add a test product for deletion
	product := &models.Product{
		Name:        "Test Product for Deletion",
		Price:       models.Currency{Amount: 3000},
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
