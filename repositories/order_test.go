package repositories

import (
	"context"
	"database/sql"
	"testing"

	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
	"github.com/stretchr/testify/assert"
)

// Helper function to insert a test address for a user
func createTestAddress(t *testing.T, db *sql.DB, userID string) string {
	ctx := context.Background()
	addressID := utilities.MustGenerateIDString()

	_, err := db.ExecContext(ctx, `
		INSERT INTO addresses (id, user_id, line1, city, state, postal_code, country, email)
		VALUES ($1, $2, '123 Test St', 'Test City', 'CA', '12345', 'US', 'example@example.com')`,
		addressID, userID)
	assert.NoError(t, err)

	return addressID
}

func TestOrderRepository_GetOrder_Success(t *testing.T) {
	ctx := context.Background()

	orderRepo := NewOrderRepository(dbPool)
	userRepo := NewUserRepository(dbPool)

	// Create test user and address
	user := createUniqueTestUser(t, userRepo)
	addressID := createTestAddress(t, dbPool, user.ID)

	// Create empty order with address
	paymentMethod := types.PaymentMethodStripe
	status := types.OrderPending
	order := &types.Order{
		ID:            utilities.MustGenerateIDString(),
		UserID:        user.ID,
		PaymentMethod: &paymentMethod,
		Status:        &status,
		Address: types.Address{
			ID: addressID,
		},
	}
	err := orderRepo.CreateOrder(ctx, order)
	assert.NoError(t, err)

	// Retrieve order
	fetchedOrder, err := orderRepo.GetOrderByIDAndUser(ctx, order.ID, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, order.ID, fetchedOrder.ID)
	assert.Equal(t, user.ID, fetchedOrder.UserID)
	if assert.NotNil(t, fetchedOrder.Status) {
		assert.Equal(t, types.OrderPending, *fetchedOrder.Status)
	}
	assert.Empty(t, fetchedOrder.Items) // No items in newly created order
	assert.NotNil(t, fetchedOrder.Address)
	assert.Equal(t, addressID, fetchedOrder.Address.ID)

	// Cleanup
	dbPool.ExecContext(ctx, `DELETE FROM orders WHERE id = $1`, order.ID)
	dbPool.ExecContext(ctx, `DELETE FROM addresses WHERE id = $1`, addressID)
	dbPool.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, user.ID)
}

func TestCancelPendingOrders(t *testing.T) {
	ctx := context.Background()
	userRepo := NewUserRepository(dbPool)

	user := createUniqueTestUser(t, userRepo)
	addressID := createTestAddress(t, dbPool, user.ID)
	productID := utilities.MustGenerateIDString()
	orderID := utilities.MustGenerateIDString()

	t.Cleanup(func() {
		dbPool.ExecContext(ctx, `DELETE FROM order_items WHERE order_id = $1`, orderID)
		dbPool.ExecContext(ctx, `DELETE FROM orders WHERE id = $1`, orderID)
		dbPool.ExecContext(ctx, `DELETE FROM products WHERE id = $1`, productID)
		dbPool.ExecContext(ctx, `DELETE FROM addresses WHERE id = $1`, addressID)
		dbPool.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, user.ID)
	})

	// Product with 3 units remaining, 7 of which are reserved by a pending order
	_, err := dbPool.ExecContext(ctx, `
		INSERT INTO products (id, name, price, summary, inventory)
		VALUES ($1, 'Test Product', 1000, 'Test product summary', 3)`,
		productID)
	assert.NoError(t, err)

	_, err = dbPool.ExecContext(ctx, `
		INSERT INTO orders (id, user_id, address_id, status, payment_status)
		VALUES ($1, $2, $3, 'pending', 'unpaid')`,
		orderID, user.ID, addressID)
	assert.NoError(t, err)

	_, err = dbPool.ExecContext(ctx, `
		INSERT INTO order_items (order_id, product_id, quantity, unit_price)
		VALUES ($1, $2, 7, 1000)`,
		orderID, productID)
	assert.NoError(t, err)

	tx, err := dbPool.BeginTx(ctx, nil)
	assert.NoError(t, err)
	defer tx.Rollback()

	assert.NoError(t, cancelPendingOrders(ctx, tx, user.ID))
	assert.NoError(t, tx.Commit())

	// Order is canceled
	var status string
	err = dbPool.QueryRowContext(ctx, `SELECT status FROM orders WHERE id = $1`, orderID).Scan(&status)
	assert.NoError(t, err)
	assert.Equal(t, "canceled", status)

	// Reserved inventory is returned to stock
	var inventory int
	err = dbPool.QueryRowContext(ctx, `SELECT inventory FROM products WHERE id = $1`, productID).Scan(&inventory)
	assert.NoError(t, err)
	assert.Equal(t, 10, inventory)

	// Order items are removed so the inventory cannot be restored twice
	var itemCount int
	err = dbPool.QueryRowContext(ctx, `SELECT COUNT(*) FROM order_items WHERE order_id = $1`, orderID).Scan(&itemCount)
	assert.NoError(t, err)
	assert.Zero(t, itemCount)
}

func TestCancelPendingOrders_NoPendingOrder(t *testing.T) {
	ctx := context.Background()
	userRepo := NewUserRepository(dbPool)

	user := createUniqueTestUser(t, userRepo)
	productID := utilities.MustGenerateIDString()

	t.Cleanup(func() {
		dbPool.ExecContext(ctx, `DELETE FROM products WHERE id = $1`, productID)
		dbPool.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, user.ID)
	})

	_, err := dbPool.ExecContext(ctx, `
		INSERT INTO products (id, name, price, summary, inventory)
		VALUES ($1, 'Test Product', 1000, 'Test product summary', 5)`,
		productID)
	assert.NoError(t, err)

	tx, err := dbPool.BeginTx(ctx, nil)
	assert.NoError(t, err)
	defer tx.Rollback()

	// No pending order: should be a no-op, not an error
	assert.NoError(t, cancelPendingOrders(ctx, tx, user.ID))
	assert.NoError(t, tx.Commit())

	var inventory int
	err = dbPool.QueryRowContext(ctx, `SELECT inventory FROM products WHERE id = $1`, productID).Scan(&inventory)
	assert.NoError(t, err)
	assert.Equal(t, 5, inventory)
}
