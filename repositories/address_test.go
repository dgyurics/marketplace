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
		ID:           utilities.MustGenerateIDString(),
		UserID:       user.ID,
		Addressee:    &addressee,
		AddressLine1: "123 Test St",
		AddressLine2: &addressLine2,
		City:         "Testville",
		StateCode:    "TS",
		PostalCode:   "12345",
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
		ID:           utilities.MustGenerateIDString(),
		UserID:       user.ID,
		Addressee:    &addressee,
		AddressLine1: "123 Test St",
		AddressLine2: &addressLine2,
		City:         "Testville",
		StateCode:    "TS",
		PostalCode:   "12345",
		CountryCode:  "US",
	}

	// Create the first address
	err := repo.CreateAddress(ctx, address)
	assert.NoError(t, err, "Expected no error while creating the initial address")

	originalID := address.ID

	// Create another address with the same fields (simulating a duplicate)
	dupAddress := &types.Address{
		ID:           utilities.MustGenerateIDString(), // different ID
		UserID:       user.ID,
		Addressee:    &addressee,
		AddressLine1: "123 Test St",
		AddressLine2: &addressLine2,
		City:         "Testville",
		StateCode:    "TS",
		PostalCode:   "12345",
		CountryCode:  "US",
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

func TestGetAddresses(t *testing.T) {
	repo := NewAddressRepository(dbPool)
	ctx := context.Background()

	// Create a test user
	userRepo := NewUserRepository(dbPool)
	user := createUniqueTestUser(t, userRepo)

	// Nullable fields
	addressLine2 := "Apt 456"

	// Create multiple addresses for the user
	addressee1 := "John Doe"
	address1 := &types.Address{
		ID:           utilities.MustGenerateIDString(),
		UserID:       user.ID,
		Addressee:    &addressee1,
		AddressLine1: "123 Test St",
		AddressLine2: &addressLine2,
		City:         "Testville",
		StateCode:    "TS",
		PostalCode:   "12345",
	}
	addressee2 := "Jane Doe"
	address2 := &types.Address{
		ID:           utilities.MustGenerateIDString(),
		UserID:       user.ID,
		Addressee:    &addressee2,
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

	// Create a test address
	addressee := "John Doe"
	address := &types.Address{
		ID:           utilities.MustGenerateIDString(),
		UserID:       user.ID,
		Addressee:    &addressee,
		AddressLine1: "123 Test St",
		AddressLine2: &addressLine2,
		City:         "Testville",
		StateCode:    "TS",
		PostalCode:   "12345",
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

func TestGetAddressWithEmptyFields(t *testing.T) {
	repo := NewAddressRepository(dbPool)
	ctx := context.Background()

	// Create a test user
	userRepo := NewUserRepository(dbPool)
	user := createUniqueTestUser(t, userRepo)

	// Create an address with empty addressee, and nil for address_line2
	addressee := ""
	address := &types.Address{
		ID:           utilities.MustGenerateIDString(),
		UserID:       user.ID,
		Addressee:    &addressee, // empty addressee
		AddressLine1: "789 Test Blvd",
		AddressLine2: nil, // nil address_line2
		City:         "Emptyville",
		StateCode:    "EM",
		PostalCode:   "00000",
	}

	err := repo.CreateAddress(ctx, address)
	assert.NoError(t, err, "Expected no error while creating an address with empty fields")
	assert.NotEmpty(t, address.ID, "Expected address ID to be set")

	// Retrieve addresses for the user
	addresses, err := repo.GetAddresses(ctx, user.ID)
	assert.NoError(t, err, "Expected no error while retrieving addresses")

	// Find the created address in the list
	var found *types.Address
	for i, addr := range addresses {
		if addr.ID == address.ID {
			found = &addresses[i]
			break
		}
	}
	assert.NotNil(t, found, "Expected to find the created address")

	// Validate that the fields are as expected
	if found.Addressee == nil {
		assert.Fail(t, "Expected addressee not to be nil")
	} else {
		assert.Equal(t, "", *found.Addressee, "Expected empty addressee")
	}
	assert.Nil(t, found.AddressLine2, "Expected address_line2 to be nil")

	// Clean up: remove created address and user
	_, err = dbPool.ExecContext(ctx, "DELETE FROM addresses WHERE id = $1", address.ID)
	assert.NoError(t, err, "Expected no error on address deletion")
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on user deletion")
}
