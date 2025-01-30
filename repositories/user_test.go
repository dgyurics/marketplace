package repositories

import (
	"context"
	"testing"

	"github.com/dgyurics/marketplace/models"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	repo := NewUserRepository(dbPool)
	ctx := context.Background()

	// Create a unique test user
	user := createUniqueTestUser(t, repo)

	// Validate user creation
	assert.NotEmpty(t, user.ID, "Expected user ID to be set")
	assert.NotEmpty(t, user.Email, "Expected email to be set")

	// Clean up
	_, err := dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on user deletion")
}

func TestGetUserByEmail(t *testing.T) {
	repo := NewUserRepository(dbPool)
	ctx := context.Background()

	// Create a unique test user
	user := createUniqueTestUser(t, repo)

	// Retrieve the user by email
	retrievedUser, err := repo.GetUserByEmail(ctx, user.Email)
	assert.NoError(t, err, "Expected no error on getting user by email")
	assert.NotNil(t, retrievedUser, "Expected retrieved user to not be nil")
	assert.Equal(t, user.ID, retrievedUser.ID, "Expected user ID to match")
	assert.Equal(t, user.Email, retrievedUser.Email, "Expected email to match")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on user deletion")
}

func TestGetAllUsers(t *testing.T) {
	repo := NewUserRepository(dbPool)
	ctx := context.Background()

	// Create two unique test users
	user1 := createUniqueTestUser(t, repo)
	user2 := createUniqueTestUser(t, repo)

	// Retrieve all users
	users, err := repo.GetAllUsers(ctx, 1, 10)
	assert.NoError(t, err, "Expected no error on getting all users")
	assert.True(t, len(users) >= 2, "Expected at least two users in the list")

	// Check if the created users are in the list
	var foundUser1, foundUser2 bool
	for _, u := range users {
		if u.ID == user1.ID {
			foundUser1 = true
		}
		if u.ID == user2.ID {
			foundUser2 = true
		}
	}
	assert.True(t, foundUser1, "Expected user1 to be in the list")
	assert.True(t, foundUser2, "Expected user2 to be in the list")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1 OR id = $2", user1.ID, user2.ID)
	assert.NoError(t, err, "Expected no error on user deletion")
}

func TestCreateAddress(t *testing.T) {
	repo := NewUserRepository(dbPool)
	ctx := context.Background()

	// Create a test user
	user := createUniqueTestUser(t, repo)

	// Nullable fields
	addressLine2 := "Apt 456"
	phone := "123-456-7890"

	// Create a test address
	address := &models.Address{
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
	assert.Equal(t, "123 Test St", address.AddressLine1, "Expected address line 1 to match")
	assert.Equal(t, &addressLine2, address.AddressLine2, "Expected address line 2 to match")
	assert.Equal(t, "Testville", address.City, "Expected city to match")
	assert.Equal(t, "TS", address.StateCode, "Expected state code to match")
	assert.Equal(t, "12345", address.PostalCode, "Expected postal code to match")
	assert.Equal(t, &phone, address.Phone, "Expected phone to match")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM addresses WHERE id = $1", address.ID)
	assert.NoError(t, err, "Expected no error on address deletion")
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on user deletion")
}

func TestGetAddresses(t *testing.T) {
	repo := NewUserRepository(dbPool)
	ctx := context.Background()

	// Create a test user
	user := createUniqueTestUser(t, repo)

	// Nullable fields
	addressLine2 := "Apt 456"
	phone := "123-456-7890"

	// Create multiple addresses for the user
	address1 := &models.Address{
		UserID:       user.ID,
		Addressee:    "John Doe",
		AddressLine1: "123 Test St",
		AddressLine2: &addressLine2,
		City:         "Testville",
		StateCode:    "TS",
		PostalCode:   "12345",
		Phone:        &phone,
	}
	address2 := &models.Address{
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

	// Check individual addresses
	for _, addr := range addresses {
		switch addr.ID {
		case address1.ID:
			assert.Equal(t, address1.Addressee, addr.Addressee, "Expected addressee to match")
			assert.Equal(t, address1.AddressLine1, addr.AddressLine1, "Expected address line 1 to match")
			assert.Equal(t, address1.AddressLine2, addr.AddressLine2, "Expected address line 2 to match")
			assert.Equal(t, address1.Phone, addr.Phone, "Expected phone to match")
		case address2.ID:
			assert.Equal(t, address2.Addressee, addr.Addressee, "Expected addressee to match")
			assert.Equal(t, address2.AddressLine1, addr.AddressLine1, "Expected address line 1 to match")
			assert.Nil(t, addr.AddressLine2, "Expected address line 2 to be nil")
			assert.Nil(t, addr.Phone, "Expected phone to be nil")
		default:
			t.Errorf("Unexpected address ID: %s", addr.ID)
		}
	}

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM addresses WHERE user_id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on address deletion")
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on user deletion")
}

func TestRemoveAddress(t *testing.T) {
	repo := NewUserRepository(dbPool)
	ctx := context.Background()

	// Create a test user
	user := createUniqueTestUser(t, repo)

	// Nullable fields
	addressLine2 := "Apt 456"
	phone := "123-456-7890"

	// Create a test address
	address := &models.Address{
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
