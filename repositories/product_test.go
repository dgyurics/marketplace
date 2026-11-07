package repositories

import (
	"context"
	"slices"
	"testing"

	"github.com/dgyurics/marketplace/types"
	util "github.com/dgyurics/marketplace/utilities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newProductFixture creates a category and returns a product ready to insert.
func newProductFixture(t *testing.T) *types.Product {
	t.Helper()
	ctx := context.Background()
	categoryID := util.MustGenerateIDString()
	category := &types.Category{
		ID:          categoryID,
		Name:        "Test Category",
		Slug:        "test-category-" + categoryID,
		Description: "A test category",
	}
	t.Cleanup(func() {
		_, err := dbPool.ExecContext(ctx, "DELETE FROM categories WHERE id = $1", category.ID)
		assert.NoError(t, err)
	})
	require.NoError(t, NewCategoryRepository(dbPool).CreateCategory(ctx, category))

	product := &types.Product{
		ID:          util.MustGenerateIDString(),
		Status:      util.Ptr(types.ProductPublished),
		Name:        "Test Product",
		Price:       util.Ptr(int64(150000)),
		Summary:     util.Ptr("A test product summary"),
		Description: util.Ptr("A test product description"),
		Details:     []byte(`{"key": "value"}`),
		Category:    category,
	}
	t.Cleanup(func() {
		_, err := dbPool.ExecContext(ctx, "DELETE FROM products WHERE id = $1", product.ID)
		assert.NoError(t, err)
	})
	return product
}

func TestCreateProduct(t *testing.T) {
	repo := NewProductRepository(dbPool)
	ctx := context.Background()
	product := newProductFixture(t)

	require.NoError(t, repo.CreateProduct(ctx, product))

	saved, err := repo.GetProductByID(ctx, product.ID)
	require.NoError(t, err)
	assert.Equal(t, product.ID, saved.ID)
	assert.Equal(t, product.Name, saved.Name)
	assert.Equal(t, product.Price, saved.Price)
	assert.Equal(t, product.Summary, saved.Summary)
	assert.Equal(t, product.Description, saved.Description)
	assert.Equal(t, product.Status, saved.Status)
	assert.JSONEq(t, string(product.Details), string(saved.Details))
	require.NotNil(t, saved.Category)
	assert.Equal(t, product.Category.ID, saved.Category.ID)
}

func TestCreateProductDefaultsStatus(t *testing.T) {
	repo := NewProductRepository(dbPool)
	product := newProductFixture(t)
	product.Status = nil

	require.NoError(t, repo.CreateProduct(context.Background(), product))

	saved, err := repo.GetProductByID(context.Background(), product.ID)
	require.NoError(t, err)
	require.NotNil(t, saved.Status)
	assert.Equal(t, types.ProductPublished, *saved.Status)
	assert.Equal(t, saved.Status, product.Status)
}

func TestGetProducts(t *testing.T) {
	repo := NewProductRepository(dbPool)
	ctx := context.Background()
	product := newProductFixture(t)
	require.NoError(t, repo.CreateProduct(ctx, product))

	products, err := repo.GetProducts(ctx, types.ProductFilter{Limit: 1000, Page: 1})
	require.NoError(t, err)
	index := slices.IndexFunc(products, func(p types.Product) bool { return p.ID == product.ID })
	require.NotEqual(t, -1, index, "created product missing from list")
	assert.Equal(t, product.Status, products[index].Status)
}

func TestGetProductsByCategory(t *testing.T) {
	repo := NewProductRepository(dbPool)
	ctx := context.Background()
	product := newProductFixture(t)
	require.NoError(t, repo.CreateProduct(ctx, product))

	products, err := repo.GetProducts(ctx, types.ProductFilter{
		Categories: []string{product.Category.Slug},
		Limit:      100,
		Page:       1,
	})
	require.NoError(t, err)
	require.Len(t, products, 1)
	assert.Equal(t, product.ID, products[0].ID)
	assert.Equal(t, product.Status, products[0].Status)
}

func TestGetProductByID(t *testing.T) {
	repo := NewProductRepository(dbPool)
	ctx := context.Background()
	product := newProductFixture(t)
	require.NoError(t, repo.CreateProduct(ctx, product))

	saved, err := repo.GetProductByID(ctx, product.ID)
	require.NoError(t, err)
	assert.Equal(t, product.ID, saved.ID)
	assert.Equal(t, product.Name, saved.Name)
	assert.Equal(t, product.Description, saved.Description)
	assert.Equal(t, product.Price, saved.Price)
	assert.Equal(t, product.Status, saved.Status)
}

func TestRemoveProduct(t *testing.T) {
	repo := NewProductRepository(dbPool)
	ctx := context.Background()
	product := newProductFixture(t)
	require.NoError(t, repo.CreateProduct(ctx, product))

	require.NoError(t, repo.RemoveProduct(ctx, product.ID))

	_, err := repo.GetProductByID(ctx, product.ID)
	assert.ErrorIs(t, err, types.ErrNotFound)

	var status types.ProductStatus
	require.NoError(t, dbPool.QueryRowContext(ctx, "SELECT status FROM products WHERE id = $1", product.ID).Scan(&status))
	assert.Equal(t, types.ProductArchived, status)
}

func TestUpdateProductPreservesOmittedStatus(t *testing.T) {
	repo := NewProductRepository(dbPool)
	ctx := context.Background()
	product := newProductFixture(t)
	product.Status = util.Ptr(types.ProductDraft)
	require.NoError(t, repo.CreateProduct(ctx, product))

	product.Status = nil
	product.Summary = util.Ptr("Updated summary")
	require.NoError(t, repo.UpdateProduct(ctx, *product))

	saved, err := repo.GetProductByID(ctx, product.ID)
	require.NoError(t, err)
	require.NotNil(t, saved.Status)
	assert.Equal(t, types.ProductDraft, *saved.Status)
	assert.Equal(t, product.Summary, saved.Summary)

	product.Status = util.Ptr(types.ProductPublished)
	require.NoError(t, repo.UpdateProduct(ctx, *product))
	saved, err = repo.GetProductByID(ctx, product.ID)
	require.NoError(t, err)
	assert.Equal(t, product.Status, saved.Status)
}
