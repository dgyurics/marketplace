package repositories

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper function to insert a product and inventory
func createTestProductAndInventory(t *testing.T, db *sql.DB, quantity int) string {
	ctx := context.Background()

	productID := genID()

	var err error
	_, err = db.ExecContext(ctx, `
		INSERT INTO products (id, name, price, description) 
		VALUES ($1, 'Test Product', 1000, 'Test product description')`,
		productID)
	assert.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO inventory (product_id, quantity) 
		VALUES ($1, $2)`,
		productID, quantity)
	assert.NoError(t, err)

	return productID
}

// Helper function to add an item to the cart
func addToCart(t *testing.T, db *sql.DB, userID, productID string, quantity int) {
	ctx := context.Background()

	_, err := db.ExecContext(ctx, `
		INSERT INTO carts (user_id) 
		VALUES ($1) 
		ON CONFLICT (user_id) DO NOTHING`, userID)
	assert.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO cart_items (user_id, product_id, quantity, unit_price)
		VALUES ($1, $2, $3, 1000)
		ON CONFLICT (user_id, product_id) 
		DO UPDATE SET quantity = EXCLUDED.quantity`, userID, productID, quantity)
	assert.NoError(t, err)
}

func TestOrderRepository_CreateOrder(t *testing.T) {
	ctx := context.Background()

	orderRepo := NewOrderRepository(dbPool)
	userRepo := NewUserRepository(dbPool)

	user := createUniqueTestUser(t, userRepo)
	productID := createTestProductAndInventory(t, dbPool, 10)
	addToCart(t, dbPool, user.ID, productID, 2)

	order, err := orderRepo.CreateOrder(ctx, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, user.ID, order.UserID)
	assert.Equal(t, "pending", order.Status)
	assert.EqualValues(t, 2*1000, order.Amount)

	// Cleanup
	dbPool.ExecContext(ctx, `DELETE FROM order_items WHERE order_id = $1`, order.ID)
	dbPool.ExecContext(ctx, `DELETE FROM orders WHERE id = $1`, order.ID)
	dbPool.ExecContext(ctx, `DELETE FROM cart_items WHERE user_id = $1`, user.ID)
	dbPool.ExecContext(ctx, `DELETE FROM carts WHERE user_id = $1`, user.ID)
	dbPool.ExecContext(ctx, `DELETE FROM inventory WHERE product_id = $1`, productID)
	dbPool.ExecContext(ctx, `DELETE FROM products WHERE id = $1`, productID)
	dbPool.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, user.ID)
}

func TestOrderRepository_GetOrder(t *testing.T) {}

func TestOrderRepository_UpdateOrder(t *testing.T) {
	ctx := context.Background()

	orderRepo := NewOrderRepository(dbPool)
	userRepo := NewUserRepository(dbPool)

	// 1. Create a unique test user
	user := createUniqueTestUser(t, userRepo)

	// 2. Insert a product with inventory and add it to the user's cart
	productID := createTestProductAndInventory(t, dbPool, 10)
	addToCart(t, dbPool, user.ID, productID, 2)

	// 3. Create an order for the user
	order, err := orderRepo.CreateOrder(ctx, user.ID)
	assert.NoError(t, err, "CreateOrder should not return an error")

	// 4. Mark the order as paid
	order.Status = "paid"
	err = orderRepo.UpdateOrder(ctx, order)
	assert.NoError(t, err, "UpdateOrder should not return an error")

	// 5. Validate that the order's status was updated to 'paid'
	var status string
	err = dbPool.QueryRowContext(ctx, `SELECT status FROM orders WHERE id = $1`, order.ID).Scan(&status)
	assert.NoError(t, err, "Querying the order status should not return an error")
	assert.Equal(t, "paid", status, "The order status should be updated to 'paid'")

	// 7. Cleanup
	dbPool.ExecContext(ctx, `DELETE FROM order_items WHERE order_id = $1`, order.ID)
	dbPool.ExecContext(ctx, `DELETE FROM orders WHERE id = $1`, order.ID)
	dbPool.ExecContext(ctx, `DELETE FROM carts WHERE user_id = $1`, user.ID)
	dbPool.ExecContext(ctx, `DELETE FROM inventory WHERE product_id = $1`, productID)
	dbPool.ExecContext(ctx, `DELETE FROM products WHERE id = $1`, productID)
	dbPool.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, user.ID)
}
