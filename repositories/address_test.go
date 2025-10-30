package repositories

import (
	"context"
	"testing"

	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
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

	// Create a test address
	addressee := "John Doe"
	address := &types.Address{
		ID:         utilities.MustGenerateIDString(),
		UserID:     user.ID,
		Addressee:  &addressee,
		Line1:      "123 Test St",
		Line2:      &addressLine2,
		City:       "Testville",
		State:      "TS",
		PostalCode: "12345",
	}

	err := repo.CreateAddress(ctx, address)
	assert.NoError(t, err, "Expected no error while creating an address")
	assert.NotEmpty(t, address.ID, "Expected address ID to be set")

	// Validate the address fields
	assert.Equal(t, user.ID, address.UserID, "Expected user ID to match")
	assert.Equal(t, addressee, *address.Addressee, "Expected addressee to match")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM addresses WHERE id = $1", address.ID)
	assert.NoError(t, err, "Expected no error on address deletion")
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on user deletion")
}

func TestCreateAddressWhenDuplicateExists(t *testing.T) {
	repo := NewAddressRepository(dbPool)
	ctx := context.Background()

	// Create a test user
	userRepo := NewUserRepository(dbPool)
	user := createUniqueTestUser(t, userRepo)

	// Nullable fields
	addressLine2 := "Apt 456"
	addressee := "John Doe"

	// Define address fields
	address := &types.Address{
		ID:         utilities.MustGenerateIDString(),
		UserID:     user.ID,
		Addressee:  &addressee,
		Line1:      "123 Test St",
		Line2:      &addressLine2,
		City:       "Testville",
		State:      "TS",
		PostalCode: "12345",
		Country:    "US",
	}

	// Create the first address
	err := repo.CreateAddress(ctx, address)
	assert.NoError(t, err, "Expected no error while creating the initial address")

	originalID := address.ID

	// Create another address with the same fields (simulating a duplicate)
	dupAddress := &types.Address{
		ID:         utilities.MustGenerateIDString(), // different ID
		UserID:     user.ID,
		Addressee:  &addressee,
		Line1:      "123 Test St",
		Line2:      &addressLine2,
		City:       "Testville",
		State:      "TS",
		PostalCode: "12345",
		Country:    "US",
	}

	err = repo.CreateAddress(ctx, dupAddress)
	assert.NoError(t, err, "Expected no error when creating a duplicate address")
	assert.Equal(t, originalID, dupAddress.ID, "Expected existing address ID to be returned for duplicate")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM addresses WHERE user_id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on address cleanup")
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on user cleanup")
}
