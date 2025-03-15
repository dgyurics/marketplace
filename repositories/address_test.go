package repositories

import (
	"context"
	"testing"

	"github.com/dgyurics/marketplace/types"
	"github.com/stretchr/testify/assert"
)

func TestCreateAddress(t *testing.T) {
	repo := NewAddressRepository(dbPool)
	ctx := context.Background()

	// Create a test user
	userRepo := NewUserRepository(dbPool)
	user := createUniqueTestUser(t, userRepo)

	// Nullable fields
	addressLine2 := "Apt 456"
	phone := "123-456-7890"

	// Create a test address
	address := &types.Address{
		UserID:       user.ID,
		Addressee:    "John Doe",
		AddressLine1: "123 Test St",
		AddressLine2: &addressLine2,
		City:         "Testville",
		StateCode:    "TS",
		PostalCode:   "12345",
		Phone:        &phone,
	}

	err := repo.CreateAddress(ctx, address)
	assert.NoError(t, err, "Expected no error while creating an address")
	assert.NotEmpty(t, address.ID, "Expected address ID to be set")

	// Validate the address fields
	assert.Equal(t, user.ID, address.UserID, "Expected user ID to match")
	assert.Equal(t, "John Doe", address.Addressee, "Expected addressee to match")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM addresses WHERE id = $1", address.ID)
	assert.NoError(t, err, "Expected no error on address deletion")
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on user deletion")
}

func TestGetAddresses(t *testing.T) {
	repo := NewAddressRepository(dbPool)
	ctx := context.Background()

	// Create a test user
	userRepo := NewUserRepository(dbPool)
	user := createUniqueTestUser(t, userRepo)

	// Nullable fields
	addressLine2 := "Apt 456"
	phone := "123-456-7890"

	// Create multiple addresses for the user
	address1 := &types.Address{
		UserID:       user.ID,
		Addressee:    "John Doe",
		AddressLine1: "123 Test St",
		AddressLine2: &addressLine2,
		City:         "Testville",
		StateCode:    "TS",
		PostalCode:   "12345",
		Phone:        &phone,
	}
	address2 := &types.Address{
		UserID:       user.ID,
		Addressee:    "Jane Doe",
		AddressLine1: "456 Test Ave",
		City:         "Testville",
		StateCode:    "TS",
		PostalCode:   "67890",
	}

	err := repo.CreateAddress(ctx, address1)
	assert.NoError(t, err, "Expected no error while creating address1")
	err = repo.CreateAddress(ctx, address2)
	assert.NoError(t, err, "Expected no error while creating address2")

	// Retrieve addresses
	addresses, err := repo.GetAddresses(ctx, user.ID)
	assert.NoError(t, err, "Expected no error while getting addresses")
	assert.Len(t, addresses, 2, "Expected 2 addresses for the user")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM addresses WHERE user_id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on address deletion")
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on user deletion")
}

func TestRemoveAddress(t *testing.T) {
	repo := NewAddressRepository(dbPool)
	ctx := context.Background()

	// Create a test user
	userRepo := NewUserRepository(dbPool)
	user := createUniqueTestUser(t, userRepo)

	// Nullable fields
	addressLine2 := "Apt 456"
	phone := "123-456-7890"

	// Create a test address
	address := &types.Address{
		UserID:       user.ID,
		Addressee:    "John Doe",
		AddressLine1: "123 Test St",
		AddressLine2: &addressLine2,
		City:         "Testville",
		StateCode:    "TS",
		PostalCode:   "12345",
		Phone:        &phone,
	}

	err := repo.CreateAddress(ctx, address)
	assert.NoError(t, err, "Expected no error while creating an address")
	assert.NotEmpty(t, address.ID, "Expected address ID to be set")

	// Remove the address
	err = repo.RemoveAddress(ctx, user.ID, address.ID)
	assert.NoError(t, err, "Expected no error while removing the address")

	// Verify the address is deleted
	addresses, err := repo.GetAddresses(ctx, user.ID)
	assert.NoError(t, err, "Expected no error while getting addresses")
	assert.Empty(t, addresses, "Expected no addresses for the user")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on user deletion")
}
