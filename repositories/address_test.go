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
		State:      utilities.String("TS"),
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

func TestGetAddress(t *testing.T) {
	repo := NewAddressRepository(dbPool)
	ctx := context.Background()

	// Create a test user
	userRepo := NewUserRepository(dbPool)
	user := createUniqueTestUser(t, userRepo)

	// Create a test address
	addressee := "Jane Smith"
	address := &types.Address{
		ID:         utilities.MustGenerateIDString(),
		UserID:     user.ID,
		Addressee:  &addressee,
		Line1:      "456 Main St",
		City:       "Hometown",
		PostalCode: "67890",
	}

	err := repo.CreateAddress(ctx, address)
	assert.NoError(t, err, "Expected no error while creating an address")

	// Get the address
	retrieved, err := repo.GetAddress(ctx, user.ID, address.ID)
	assert.NoError(t, err, "Expected no error while getting address")
	assert.Equal(t, address.ID, retrieved.ID, "Expected address ID to match")
	assert.Equal(t, addressee, *retrieved.Addressee, "Expected addressee to match")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM addresses WHERE id = $1", address.ID)
	assert.NoError(t, err, "Expected no error on address deletion")
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on user deletion")
}

func TestGetAddressNotFound(t *testing.T) {
	repo := NewAddressRepository(dbPool)
	ctx := context.Background()

	// Create a test user
	userRepo := NewUserRepository(dbPool)
	user := createUniqueTestUser(t, userRepo)

	// Try to get non-existent address
	fakeID := utilities.MustGenerateIDString()
	_, err := repo.GetAddress(ctx, user.ID, fakeID)
	assert.Equal(t, types.ErrNotFound, err, "Expected ErrNotFound for non-existent address")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on user deletion")
}

func TestUpdateAddress(t *testing.T) {
	repo := NewAddressRepository(dbPool)
	ctx := context.Background()

	// Create a test user
	userRepo := NewUserRepository(dbPool)
	user := createUniqueTestUser(t, userRepo)

	// Create a test address
	addressee := "Bob Johnson"
	address := &types.Address{
		ID:         utilities.MustGenerateIDString(),
		UserID:     user.ID,
		Addressee:  &addressee,
		Line1:      "789 Oak St",
		City:       "Oldtown",
		PostalCode: "11111",
	}

	err := repo.CreateAddress(ctx, address)
	assert.NoError(t, err, "Expected no error while creating an address")

	// Update the address
	newAddressee := "Robert Johnson"
	address.Addressee = &newAddressee
	address.City = "Newtown"
	address.PostalCode = "22222"

	err = repo.UpdateAddress(ctx, address)
	assert.NoError(t, err, "Expected no error while updating address")

	// Verify the update
	retrieved, err := repo.GetAddress(ctx, user.ID, address.ID)
	assert.NoError(t, err, "Expected no error while getting updated address")
	assert.Equal(t, newAddressee, *retrieved.Addressee, "Expected updated addressee")
	assert.Equal(t, "Newtown", retrieved.City, "Expected updated city")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM addresses WHERE id = $1", address.ID)
	assert.NoError(t, err, "Expected no error on address deletion")
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on user deletion")
}

func TestUpdateAddressNotFound(t *testing.T) {
	repo := NewAddressRepository(dbPool)
	ctx := context.Background()

	// Create a test user
	userRepo := NewUserRepository(dbPool)
	user := createUniqueTestUser(t, userRepo)

	// Try to update non-existent address
	addressee := "Nobody"
	fakeAddress := &types.Address{
		ID:         utilities.MustGenerateIDString(),
		UserID:     user.ID,
		Addressee:  &addressee,
		Line1:      "123 Fake St",
		City:       "Nowhere",
		PostalCode: "00000",
	}

	err := repo.UpdateAddress(ctx, fakeAddress)
	assert.Equal(t, types.ErrNotFound, err, "Expected ErrNotFound for non-existent address")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on user deletion")
}

func TestRemoveAddress(t *testing.T) {
	repo := NewAddressRepository(dbPool)
	ctx := context.Background()

	// Create a test user
	userRepo := NewUserRepository(dbPool)
	user := createUniqueTestUser(t, userRepo)

	// Create a test address
	addressee := "Alice Wilson"
	address := &types.Address{
		ID:         utilities.MustGenerateIDString(),
		UserID:     user.ID,
		Addressee:  &addressee,
		Line1:      "321 Pine St",
		City:       "Removetown",
		PostalCode: "33333",
	}

	err := repo.CreateAddress(ctx, address)
	assert.NoError(t, err, "Expected no error while creating an address")

	// Remove the address
	err = repo.RemoveAddress(ctx, user.ID, address.ID)
	assert.NoError(t, err, "Expected no error while removing address")

	// Verify it's gone
	_, err = repo.GetAddress(ctx, user.ID, address.ID)
	assert.Equal(t, types.ErrNotFound, err, "Expected ErrNotFound after removal")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on user deletion")
}

func TestRemoveAddressNotFound(t *testing.T) {
	repo := NewAddressRepository(dbPool)
	ctx := context.Background()

	// Create a test user
	userRepo := NewUserRepository(dbPool)
	user := createUniqueTestUser(t, userRepo)

	// Try to remove non-existent address
	fakeID := utilities.MustGenerateIDString()
	err := repo.RemoveAddress(ctx, user.ID, fakeID)
	assert.Equal(t, types.ErrNotFound, err, "Expected ErrNotFound for non-existent address")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on user deletion")
}
