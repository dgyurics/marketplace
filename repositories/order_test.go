package repositories

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/types/stripe"
	"github.com/dgyurics/marketplace/utilities"
	"github.com/stretchr/testify/assert"
)

// Helper function to insert a test address for a user
func createTestAddress(t *testing.T, db *sql.DB, userID string) string {
	ctx := context.Background()
	addressID := utilities.MustGenerateIDString()

	_, err := db.ExecContext(ctx, `
		INSERT INTO addresses (id, user_id, address_line1, city, state_code, postal_code)
		VALUES ($1, $2, '123 Test St', 'Test City', 'CA', '12345')`,
		addressID, userID)
	assert.NoError(t, err)

	return addressID
}

// Helper function to insert a product and inventory
func createTestProductAndInventory(t *testing.T, db *sql.DB, quantity int) string {
	ctx := context.Background()

	productID := utilities.MustGenerateIDString()

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

	// Create a test user
	user := createUniqueTestUser(t, userRepo)

	// Insert a test address for the user
	AddressID := createTestAddress(t, dbPool, user.ID)

	// Insert a product with inventory and add it to the cart
	productID := createTestProductAndInventory(t, dbPool, 10)
	addToCart(t, dbPool, user.ID, productID, 2)

	// Create the order
	order := &types.Order{
		ID:     utilities.MustGenerateIDString(),
		UserID: user.ID,
		Email:  "test@example.com",
		Address: &types.Address{
			ID: AddressID,
		},
	}
	err := orderRepo.CreateOrder(ctx, order)
	assert.NoError(t, err, "CreateOrder should not return an error")
	assert.NotNil(t, order, "Order should not be nil")
	assert.Equal(t, user.ID, order.UserID, "Order UserID should match")
	assert.Equal(t, AddressID, order.Address.ID, "Order AddressID should match")
	assert.Equal(t, types.OrderPending, order.Status, "Order status should be 'pending'")
	assert.EqualValues(t, 2*1000, order.Amount, "Order amount should match expected value")

	// Cleanup
	dbPool.ExecContext(ctx, `DELETE FROM order_items WHERE order_id = $1`, order.ID)
	dbPool.ExecContext(ctx, `DELETE FROM orders WHERE id = $1`, order.ID)
	dbPool.ExecContext(ctx, `DELETE FROM cart_items WHERE user_id = $1`, user.ID)
	dbPool.ExecContext(ctx, `DELETE FROM inventory WHERE product_id = $1`, productID)
	dbPool.ExecContext(ctx, `DELETE FROM products WHERE id = $1`, productID)
	dbPool.ExecContext(ctx, `DELETE FROM addresses WHERE id = $1`, AddressID)
	dbPool.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, user.ID)
}

func TestOrderRepository_GetOrder(t *testing.T) {
	ctx := context.Background()

	orderRepo := NewOrderRepository(dbPool)
	userRepo := NewUserRepository(dbPool)

	// 1. Create a unique test user
	user := createUniqueTestUser(t, userRepo)

	// 2. Insert a test address for the user
	AddressID := createTestAddress(t, dbPool, user.ID)

	// 3. Insert a product with inventory and add it to the user's cart
	productID := createTestProductAndInventory(t, dbPool, 10)
	addToCart(t, dbPool, user.ID, productID, 2)

	// 4. Create an order for the user
	order := &types.Order{
		ID:     utilities.MustGenerateIDString(),
		UserID: user.ID,
		Email:  "test@example.com",
		Address: &types.Address{
			ID: AddressID,
		},
	}
	err := orderRepo.CreateOrder(ctx, order)
	assert.NoError(t, err, "CreateOrder should not return an error")

	// 5. Update the order with the mocked PaymentIntentID
	mockPaymentIntent := &stripe.PaymentIntent{ID: "pi_mocked_payment_intent_id"}
	order.StripePaymentIntent = mockPaymentIntent
	err = orderRepo.UpdateOrder(ctx, order)
	assert.NoError(t, err, "UpdateOrder should not return an error")

	// 6. Test retrieving the order by ID
	retrievedOrder := &types.Order{ID: order.ID}
	err = orderRepo.GetOrder(ctx, retrievedOrder)
	assert.NoError(t, err, "GetOrder by ID should not return an error")
	assert.Equal(t, order.ID, retrievedOrder.ID, "The retrieved order ID should match")
	assert.Equal(t, order.UserID, retrievedOrder.UserID, "The retrieved order UserID should match")
	assert.Equal(t, order.Address.ID, retrievedOrder.Address.ID, "The retrieved order AddressID should match")

	// 7. Test retrieving the order by PaymentIntentID
	retrievedOrder = &types.Order{StripePaymentIntent: mockPaymentIntent}
	err = orderRepo.GetOrder(ctx, retrievedOrder)
	assert.NoError(t, err, "GetOrder by PaymentIntentID should not return an error")
	assert.Equal(t, order.ID, retrievedOrder.ID, "The retrieved order ID should match the created order's ID")
	assert.Equal(t, order.Address.ID, retrievedOrder.Address.ID, "The retrieved order AddressID should match")

	// 8. Cleanup
	dbPool.ExecContext(ctx, `DELETE FROM order_items WHERE order_id = $1`, order.ID)
	dbPool.ExecContext(ctx, `DELETE FROM orders WHERE id = $1`, order.ID)
	dbPool.ExecContext(ctx, `DELETE FROM stripe_payment_intents WHERE id = $1`, mockPaymentIntent.ID)
	dbPool.ExecContext(ctx, `DELETE FROM cart_items WHERE user_id = $1`, user.ID)
	dbPool.ExecContext(ctx, `DELETE FROM inventory WHERE product_id = $1`, productID)
	dbPool.ExecContext(ctx, `DELETE FROM products WHERE id = $1`, productID)
	dbPool.ExecContext(ctx, `DELETE FROM addresses WHERE id = $1`, AddressID)
	dbPool.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, user.ID)
}

func TestOrderRepository_GetOrder_MissingOrder(t *testing.T) {
	ctx := context.Background()

	orderRepo := NewOrderRepository(dbPool)

	// Test case 1: Missing Order by ID
	missingOrder := &types.Order{ID: "999999999999999"} // Use a valid BIGINT format
	err := orderRepo.GetOrder(ctx, missingOrder)
	assert.Error(t, err, "GetOrder by ID should return an error for a nonexistent order")
	assert.Contains(t, err.Error(), "order not found", "The error message should indicate that the order was not found")

	// Test case 2: Missing Order by PaymentIntentID
	missingOrder = &types.Order{StripePaymentIntent: &stripe.PaymentIntent{ID: "pi_missing_payment_intent_id"}}
	err = orderRepo.GetOrder(ctx, missingOrder)
	assert.Error(t, err, "GetOrder by PaymentIntentID should return an error for a nonexistent PaymentIntentID")
	assert.Contains(t, err.Error(), "order not found", "The error message should indicate that the order was not found")

	// Test case 3: No Identifiers Provided
	missingOrder = &types.Order{} // No ID, UserID, or PaymentIntentID provided
	err = orderRepo.GetOrder(ctx, missingOrder)
	assert.Error(t, err, "GetOrder should return an error if no identifiers are provided")
	assert.Contains(t, err.Error(), "missing identifier: provide order.ID or StripePaymentIntent.ID", "The error message should indicate missing identifiers")
}

func TestOrderRepository_GetOrders(t *testing.T) {
	ctx := context.Background()

	orderRepo := NewOrderRepository(dbPool)
	userRepo := NewUserRepository(dbPool)

	// 1. Create a unique test user
	user := createUniqueTestUser(t, userRepo)

	// 2. Insert a test address for the user
	AddressID := createTestAddress(t, dbPool, user.ID)

	// 3. Insert products with inventory and add them to the user's cart
	productID1 := createTestProductAndInventory(t, dbPool, 10)
	addToCart(t, dbPool, user.ID, productID1, 2)

	productID2 := createTestProductAndInventory(t, dbPool, 15)
	addToCart(t, dbPool, user.ID, productID2, 3)

	// 4. Create multiple orders for the user
	order1 := &types.Order{
		ID:     utilities.MustGenerateIDString(),
		UserID: user.ID,
		Email:  "test@example.com",
		Address: &types.Address{
			ID: AddressID,
		},
	}
	err := orderRepo.CreateOrder(ctx, order1)
	assert.NoError(t, err, "CreateOrder should not return an error")
	order1.Status = types.OrderPaid
	err = orderRepo.UpdateOrder(ctx, order1)
	assert.NoError(t, err, "UpdateOrder for order1 should not return an error")

	// Add another order
	addToCart(t, dbPool, user.ID, productID1, 1)
	order2 := &types.Order{
		ID:     utilities.MustGenerateIDString(),
		UserID: user.ID,
		Email:  "test@example.com",
		Address: &types.Address{
			ID: AddressID,
		},
	}
	err = orderRepo.CreateOrder(ctx, order2)
	assert.NoError(t, err, "CreateOrder should not return an error")
	order2.Status = types.OrderShipped
	err = orderRepo.UpdateOrder(ctx, order2)
	assert.NoError(t, err, "UpdateOrder for order2 should not return an error")

	// 5. Retrieve all orders for the user
	orders, err := orderRepo.GetOrders(ctx, user.ID, 1, 10)
	assert.NoError(t, err, "GetOrders should not return an error")
	assert.Len(t, orders, 2, "GetOrders should return two orders")

	// 6. Dynamically validate the retrieved orders
	if orders[0].ID == order2.ID {
		// Validate order2
		assert.Equal(t, types.OrderShipped, orders[0].Status, "The first order's status should be 'shipped'")
		assert.Equal(t, AddressID, orders[0].Address.ID, "The first order's AddressID should match")
		assert.Equal(t, order1.ID, orders[1].ID, "The second order ID should match")
		assert.Equal(t, types.OrderPaid, orders[1].Status, "The second order's status should be 'paid'")
		assert.Equal(t, AddressID, orders[1].Address.ID, "The second order's AddressID should match")
	} else {
		// Validate order1
		assert.Equal(t, order1.ID, orders[0].ID, "The first order ID should match")
		assert.Equal(t, types.OrderPaid, orders[0].Status, "The first order's status should be 'paid'")
		assert.Equal(t, AddressID, orders[0].Address.ID, "The first order's AddressID should match")
		assert.Equal(t, order2.ID, orders[1].ID, "The second order ID should match")
		assert.Equal(t, types.OrderShipped, orders[1].Status, "The second order's status should be 'shipped'")
		assert.Equal(t, AddressID, orders[1].Address.ID, "The second order's AddressID should match")
	}

	// 7. Cleanup
	dbPool.ExecContext(ctx, `DELETE FROM order_items WHERE order_id IN ($1, $2)`, order1.ID, order2.ID)
	dbPool.ExecContext(ctx, `DELETE FROM orders WHERE id IN ($1, $2)`, order1.ID, order2.ID)
	dbPool.ExecContext(ctx, `DELETE FROM cart_items WHERE user_id = $1`, user.ID)
	dbPool.ExecContext(ctx, `DELETE FROM inventory WHERE product_id IN ($1, $2)`, productID1, productID2)
	dbPool.ExecContext(ctx, `DELETE FROM products WHERE id IN ($1, $2)`, productID1, productID2)
	dbPool.ExecContext(ctx, `DELETE FROM addresses WHERE id = $1`, AddressID)
	dbPool.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, user.ID)
}

func TestOrderRepository_GetOrders_Empty(t *testing.T) {
	ctx := context.Background()

	orderRepo := NewOrderRepository(dbPool)
	userRepo := NewUserRepository(dbPool)

	// 1. Create a unique test user
	user := createUniqueTestUser(t, userRepo)

	// 2. Retrieve orders for the user (expected to be empty)
	orders, err := orderRepo.GetOrders(ctx, user.ID, 1, 10)
	assert.NoError(t, err, "GetOrders should not return an error for a user with no orders")
	assert.Len(t, orders, 0, "GetOrders should return an empty slice for a user with no orders")

	// 3. Cleanup
	dbPool.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, user.ID)
}

func TestOrderRepository_CreateStripeEvent(t *testing.T) {
	ctx := context.Background()

	orderRepo := NewOrderRepository(dbPool)

	// 1. Define a test Stripe event
	event := stripe.Event{
		ID:   "evt_test_123",
		Type: "payment_intent.succeeded",
		Data: &stripe.Data{
			Object: stripe.PaymentIntent{
				ID:           "pi_test_123",
				Status:       "succeeded",
				Amount:       1000,
				ClientSecret: "secret_test_123",
				Currency:     "usd",
			},
		},
		Livemode: false,
		Created:  time.Now().UTC().Unix(), // Use UTC for event creation time
	}

	// 2. Call CreateStripeEvent
	err := orderRepo.CreateStripeEvent(ctx, event)
	assert.NoError(t, err, "CreateStripeEvent should not return an error")

	// 3. Validate the event was stored correctly
	var eventType string
	var payload []byte
	var processedAt time.Time

	err = dbPool.QueryRowContext(ctx, `
		SELECT event_type, payload, processed_at
		FROM stripe_events
		WHERE id = $1`, event.ID).Scan(&eventType, &payload, &processedAt)
	assert.NoError(t, err, "The event should exist in the database")
	assert.Equal(t, event.Type, eventType, "Event type should match")
	assert.JSONEq(t, `
		{
			"id": "pi_test_123",
			"status": "succeeded",
			"amount": 1000,
			"client_secret": "",
			"currency": "usd"
		}`, string(payload), "Payload should match the event data object")

	// Convert expected time to UTC before comparison
	expectedProcessedAt := time.Unix(event.Created, 0).UTC()
	assert.WithinDuration(t, expectedProcessedAt, processedAt, time.Second, "Processed timestamp should match")

	// 4. Cleanup
	dbPool.ExecContext(ctx, `DELETE FROM stripe_events WHERE id = $1`, event.ID)
}

func TestOrderRepository_PopulateOrderItems(t *testing.T) {
	ctx := context.Background()

	orderRepo := NewOrderRepository(dbPool)
	userRepo := NewUserRepository(dbPool)

	// 1. Create a unique test user
	user := createUniqueTestUser(t, userRepo)

	// 2. Insert a test address for the user
	AddressID := createTestAddress(t, dbPool, user.ID)

	// 3. Insert products with inventory and add them to the user's cart
	productID1 := createTestProductAndInventory(t, dbPool, 10)
	addToCart(t, dbPool, user.ID, productID1, 2)

	productID2 := createTestProductAndInventory(t, dbPool, 15)
	addToCart(t, dbPool, user.ID, productID2, 3)

	// 4. Create an order for the user
	order := &types.Order{
		ID:     utilities.MustGenerateIDString(),
		UserID: user.ID,
		Email:  "test@example.com",
		Address: &types.Address{
			ID: AddressID,
		},
	}
	err := orderRepo.CreateOrder(ctx, order)
	assert.NoError(t, err, "CreateOrder should not return an error")

	// 5. Prepare orders for PopulateOrderItems
	orders := []types.Order{*order}

	// 6. Call PopulateOrderItems
	err = orderRepo.PopulateOrderItems(ctx, &orders)
	assert.NoError(t, err, "PopulateOrderItems should not return an error")

	// 7. Validate populated order items
	assert.Len(t, orders[0].Items, 2, "Order should have 2 items")
	// Validate items without assuming order
	itemMap := make(map[string]int)
	for _, item := range orders[0].Items {
		itemMap[item.ProductID] = item.Quantity
	}
	assert.Equal(t, 2, itemMap[productID1], "Expected quantity for productID1 to be 2")
	assert.Equal(t, 3, itemMap[productID2], "Expected quantity for productID2 to be 3")

	// 8. Cleanup
	dbPool.ExecContext(ctx, `DELETE FROM order_items WHERE order_id = $1`, order.ID)
	dbPool.ExecContext(ctx, `DELETE FROM orders WHERE id = $1`, order.ID)
	dbPool.ExecContext(ctx, `DELETE FROM cart_items WHERE user_id = $1`, user.ID)
	dbPool.ExecContext(ctx, `DELETE FROM inventory WHERE product_id IN ($1, $2)`, productID1, productID2)
	dbPool.ExecContext(ctx, `DELETE FROM products WHERE id IN ($1, $2)`, productID1, productID2)
	dbPool.ExecContext(ctx, `DELETE FROM addresses WHERE id = $1`, AddressID)
	dbPool.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, user.ID)
}
