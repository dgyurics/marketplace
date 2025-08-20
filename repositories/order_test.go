package repositories

import (
	"context"
	"database/sql"
	"testing"

	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
	"github.com/stretchr/testify/assert"
)

func TestOrderRepository_CreateOrder_Minimal(t *testing.T) {
	ctx := context.Background()

	orderRepo := NewOrderRepository(dbPool)
	userRepo := NewUserRepository(dbPool)

	// Create a test user
	user := createUniqueTestUser(t, userRepo)

	// Create a test address
	addressID := createTestAddress(t, dbPool, user.ID)

	// Create a test product
	productID := createTestProduct(t, dbPool, 5)

	// Add product to user's cart
	addToCart(t, dbPool, user.ID, productID, 1)

	order := &types.Order{
		ID:       utilities.MustGenerateIDString(),
		UserID:   user.ID,
		Currency: "usd",
		Address: &types.Address{
			ID: addressID,
		},
	}

	err := orderRepo.CreateOrder(ctx, order)
	assert.NoError(t, err, "CreateOrder should succeed")
	assert.Equal(t, user.ID, order.UserID)
	assert.Equal(t, types.OrderPending, order.Status)
	assert.Greater(t, order.Amount, int64(0))
	assert.Equal(t, order.Amount, order.TotalAmount)

	// Cleanup
	dbPool.ExecContext(ctx, `DELETE FROM order_items WHERE order_id = $1`, order.ID)
	dbPool.ExecContext(ctx, `DELETE FROM orders WHERE id = $1`, order.ID)
	dbPool.ExecContext(ctx, `DELETE FROM cart_items WHERE user_id = $1`, user.ID)
	dbPool.ExecContext(ctx, `DELETE FROM products WHERE id = $1`, productID)
	dbPool.ExecContext(ctx, `DELETE FROM addresses WHERE id = $1`, addressID)
	dbPool.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, user.ID)
}

// Helper function to insert a test address for a user
func createTestAddress(t *testing.T, db *sql.DB, userID string) string {
	ctx := context.Background()
	addressID := utilities.MustGenerateIDString()

	_, err := db.ExecContext(ctx, `
		INSERT INTO addresses (id, user_id, line1, city, state, postal_code, country)
		VALUES ($1, $2, '123 Test St', 'Test City', 'CA', '12345', 'US')`,
		addressID, userID)
	assert.NoError(t, err)

	return addressID
}

// Helper function to insert a product
func createTestProduct(t *testing.T, db *sql.DB, quantity int) string {
	ctx := context.Background()

	productID := utilities.MustGenerateIDString()

	var err error
	_, err = db.ExecContext(ctx, `
		INSERT INTO products (id, name, price, summary, inventory) 
		VALUES ($1, 'Test Product', 1000, 'Test product summary', $2)`,
		productID, quantity)
	assert.NoError(t, err)

	return productID
}

// Helper function to add an item to the cart
func addToCart(t *testing.T, db *sql.DB, userID, productID string, quantity int) {
	ctx := context.Background()

	_, err := db.ExecContext(ctx, `
		INSERT INTO cart_items (user_id, product_id, quantity, unit_price)
		VALUES ($1, $2, $3, 1000)
		ON CONFLICT (user_id, product_id) 
		DO UPDATE SET quantity = EXCLUDED.quantity`, userID, productID, quantity)
	assert.NoError(t, err)
}

func TestOrderRepository_CreateOrder_EmptyCart(t *testing.T) {
	ctx := context.Background()

	orderRepo := NewOrderRepository(dbPool)
	userRepo := NewUserRepository(dbPool)

	// Create a test user
	user := createUniqueTestUser(t, userRepo)

	// Create a test address
	addressID := createTestAddress(t, dbPool, user.ID)

	order := &types.Order{
		ID:       utilities.MustGenerateIDString(),
		UserID:   user.ID,
		Currency: "usd",
		Address: &types.Address{
			ID: addressID,
		},
	}

	err := orderRepo.CreateOrder(ctx, order)
	assert.Error(t, err, "CreateOrder should return an error when cart is empty")

	// Cleanup
	dbPool.ExecContext(ctx, `DELETE FROM addresses WHERE id = $1`, addressID)
	dbPool.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, user.ID)
}

func TestOrderRepository_GetOrder_Success(t *testing.T) {
	ctx := context.Background()

	orderRepo := NewOrderRepository(dbPool)
	userRepo := NewUserRepository(dbPool)

	// Create test user and address
	user := createUniqueTestUser(t, userRepo)
	addressID := createTestAddress(t, dbPool, user.ID)

	// Create product
	productID := createTestProduct(t, dbPool, 2)
	addToCart(t, dbPool, user.ID, productID, 1)

	// Create order
	order := &types.Order{
		ID:       utilities.MustGenerateIDString(),
		UserID:   user.ID,
		Currency: "usd",
		Address: &types.Address{
			ID: addressID,
		},
	}
	err := orderRepo.CreateOrder(ctx, order)
	assert.NoError(t, err)

	_, err = orderRepo.UpdateOrder(ctx, types.OrderParams{
		ID:        order.ID,
		UserID:    user.ID,
		AddressID: &addressID,
	})
	assert.NoError(t, err)

	// Retrieve order
	fetchedOrder, err := orderRepo.GetOrder(ctx, order.ID, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, order.ID, fetchedOrder.ID)
	assert.Equal(t, user.ID, fetchedOrder.UserID)
	assert.Equal(t, types.OrderPending, fetchedOrder.Status)
	assert.NotEmpty(t, fetchedOrder.Items)
	assert.NotNil(t, fetchedOrder.Address)
	assert.Equal(t, addressID, fetchedOrder.Address.ID)

	// Cleanup
	dbPool.ExecContext(ctx, `DELETE FROM order_items WHERE order_id = $1`, order.ID)
	dbPool.ExecContext(ctx, `DELETE FROM orders WHERE id = $1`, order.ID)
	dbPool.ExecContext(ctx, `DELETE FROM cart_items WHERE user_id = $1`, user.ID)
	dbPool.ExecContext(ctx, `DELETE FROM products WHERE id = $1`, productID)
	dbPool.ExecContext(ctx, `DELETE FROM addresses WHERE id = $1`, addressID)
	dbPool.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, user.ID)
}

func TestOrderRepository_UpdateOrder(t *testing.T) {
	ctx := context.Background()

	orderRepo := NewOrderRepository(dbPool)
	userRepo := NewUserRepository(dbPool)

	// Create a test user and address
	user := createUniqueTestUser(t, userRepo)
	addressID := createTestAddress(t, dbPool, user.ID)

	// Create a product and add to cart
	productID := createTestProduct(t, dbPool, 3)
	addToCart(t, dbPool, user.ID, productID, 2)

	// Create initial order
	order := &types.Order{
		ID:       utilities.MustGenerateIDString(),
		UserID:   user.ID,
		Currency: "usd",
		Address: &types.Address{
			ID: addressID,
		},
	}
	err := orderRepo.CreateOrder(ctx, order)
	assert.NoError(t, err)

	// Attempt to update the order's status
	status := types.OrderPaid
	params := types.OrderParams{
		ID:     order.ID,
		UserID: user.ID,
		Status: &status,
	}
	updatedOrder, err := orderRepo.UpdateOrder(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, status, updatedOrder.Status)

	// Cleanup
	dbPool.ExecContext(ctx, `DELETE FROM order_items WHERE order_id = $1`, order.ID)
	dbPool.ExecContext(ctx, `DELETE FROM orders WHERE id = $1`, order.ID)
	dbPool.ExecContext(ctx, `DELETE FROM cart_items WHERE user_id = $1`, user.ID)
	dbPool.ExecContext(ctx, `DELETE FROM products WHERE id = $1`, productID)
	dbPool.ExecContext(ctx, `DELETE FROM addresses WHERE id = $1`, addressID)
	dbPool.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, user.ID)
}
