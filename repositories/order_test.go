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
	order := &types.Order{
		ID:     utilities.MustGenerateIDString(),
		UserID: user.ID,
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
	assert.Equal(t, types.OrderPending, fetchedOrder.Status)
	assert.Empty(t, fetchedOrder.Items) // No items in newly created order
	assert.NotNil(t, fetchedOrder.Address)
	assert.Equal(t, addressID, fetchedOrder.Address.ID)

	// Cleanup
	dbPool.ExecContext(ctx, `DELETE FROM orders WHERE id = $1`, order.ID)
	dbPool.ExecContext(ctx, `DELETE FROM addresses WHERE id = $1`, addressID)
	dbPool.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, user.ID)
}
