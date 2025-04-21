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

	// Create a test address
	addressee := "John Doe"
	address := &types.Address{
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

func TestUpdateAddress(t *testing.T) {
	repo := NewAddressRepository(dbPool)
	ctx := context.Background()

	// Create a test user
	userRepo := NewUserRepository(dbPool)
	user := createUniqueTestUser(t, userRepo)

	// Create an initial address
	originalAddressee := "Original Name"
	originalAddressLine2 := "Suite 100"
	address := &types.Address{
		UserID:       user.ID,
		Addressee:    &originalAddressee,
		AddressLine1: "100 Original St",
		AddressLine2: &originalAddressLine2,
		City:         "Original City",
		StateCode:    "OS",
		PostalCode:   "11111",
	}

	err := repo.CreateAddress(ctx, address)
	assert.NoError(t, err, "Expected no error while creating an address")
	assert.NotEmpty(t, address.ID, "Expected address ID to be set")

	// Update the address fields
	newAddressee := "Updated Name"
	newAddressLine1 := "200 Updated Ave"
	newAddressLine2 := "Suite 200"
	newCity := "Updated City"
	newStateCode := "US"
	newPostalCode := "22222"

	address.Addressee = &newAddressee
	address.AddressLine1 = newAddressLine1
	address.AddressLine2 = &newAddressLine2
	address.City = newCity
	address.StateCode = newStateCode
	address.PostalCode = newPostalCode

	err = repo.UpdateAddress(ctx, address)
	assert.NoError(t, err, "Expected no error while updating address")

	// Retrieve addresses and find the updated one
	addresses, err := repo.GetAddresses(ctx, user.ID)
	assert.NoError(t, err, "Expected no error while retrieving addresses")

	var found *types.Address
	for i, addr := range addresses {
		if addr.ID == address.ID {
			found = &addresses[i]
			break
		}
	}
	assert.NotNil(t, found, "Expected to find the updated address")

	if found.Addressee == nil {
		assert.Fail(t, "Expected addressee not to be nil")
	} else {
		assert.Equal(t, newAddressee, *found.Addressee, "Expected addressee to be updated")
	}
	assert.Equal(t, newAddressLine1, found.AddressLine1, "Expected address_line1 to be updated")
	if found.AddressLine2 == nil {
		assert.Fail(t, "Expected address_line2 not to be nil")
	} else {
		assert.Equal(t, newAddressLine2, *found.AddressLine2, "Expected address_line2 to be updated")
	}
	assert.Equal(t, newCity, found.City, "Expected city to be updated")
	assert.Equal(t, newStateCode, found.StateCode, "Expected state_code to be updated")
	assert.Equal(t, newPostalCode, found.PostalCode, "Expected postal_code to be updated")

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

	// Create multiple addresses for the user
	addressee1 := "John Doe"
	address1 := &types.Address{
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
