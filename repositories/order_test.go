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

// createTestProduct inserts a product with the given inventory and registers cleanup.
func createTestProduct(t *testing.T, inventory int) string {
	t.Helper()
	ctx := context.Background()
	productID := utilities.MustGenerateIDString()

	_, err := dbPool.ExecContext(ctx, `
		INSERT INTO products (id, name, price, summary, inventory)
		VALUES ($1, 'Test Product', 1000, 'Test product summary', $2)`,
		productID, inventory)
	assert.NoError(t, err)

	t.Cleanup(func() {
		dbPool.ExecContext(ctx, `DELETE FROM products WHERE id = $1`, productID)
	})

	return productID
}

func productInventory(t *testing.T, productID string) int {
	t.Helper()
	var inventory int
	err := dbPool.QueryRowContext(context.Background(),
		`SELECT inventory FROM products WHERE id = $1`, productID).Scan(&inventory)
	assert.NoError(t, err)
	return inventory
}

func TestReserveInventory_Success(t *testing.T) {
	ctx := context.Background()

	productA := createTestProduct(t, 10)
	productB := createTestProduct(t, 5)

	items := []types.OrderItem{
		{Product: types.Product{ID: productA}, Quantity: 3},
		{Product: types.Product{ID: productB}, Quantity: 5}, // exactly all remaining stock
	}

	tx, err := dbPool.BeginTx(ctx, nil)
	assert.NoError(t, err)
	defer tx.Rollback()

	shortages, err := reserveInventory(ctx, tx, items)
	assert.NoError(t, err)
	assert.Empty(t, shortages)
	assert.NoError(t, tx.Commit())

	assert.Equal(t, 7, productInventory(t, productA))
	assert.Equal(t, 0, productInventory(t, productB))
}

func TestReserveInventory_InsufficientStock(t *testing.T) {
	ctx := context.Background()

	productA := createTestProduct(t, 10)
	productB := createTestProduct(t, 2)
	productC := createTestProduct(t, 0)

	items := []types.OrderItem{
		{Product: types.Product{ID: productA}, Quantity: 4}, // succeeds
		{Product: types.Product{ID: productB}, Quantity: 5}, // short: only 2 available
		{Product: types.Product{ID: productC}, Quantity: 1}, // short: out of stock
	}

	tx, err := dbPool.BeginTx(ctx, nil)
	assert.NoError(t, err)
	defer tx.Rollback()

	shortages, err := reserveInventory(ctx, tx, items)
	assert.NoError(t, err)
	assert.NoError(t, tx.Commit())

	// Shortages report the requested quantity and what is actually available
	assert.Len(t, shortages, 2)
	assert.Equal(t, productB, shortages[0].Product.ID)
	assert.Equal(t, 5, shortages[0].Quantity)
	assert.Equal(t, 2, shortages[0].Inventory)
	assert.Equal(t, productC, shortages[1].Product.ID)
	assert.Equal(t, 0, shortages[1].Inventory)

	// Available items are still reserved; short items are left untouched
	assert.Equal(t, 6, productInventory(t, productA))
	assert.Equal(t, 2, productInventory(t, productB))
	assert.Equal(t, 0, productInventory(t, productC))
}

func TestReserveInventory_ProductNotFound(t *testing.T) {
	ctx := context.Background()

	items := []types.OrderItem{
		{Product: types.Product{ID: utilities.MustGenerateIDString()}, Quantity: 1},
	}

	tx, err := dbPool.BeginTx(ctx, nil)
	assert.NoError(t, err)
	defer tx.Rollback()

	shortages, err := reserveInventory(ctx, tx, items)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Nil(t, shortages)
}
