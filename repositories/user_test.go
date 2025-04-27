package repositories

import (
	"context"
	"testing"

	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
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
	assert.Equal(t, "user", user.Role, "Expected role to be 'user'")

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
	users, err := repo.GetAllUsers(ctx, 1, 100)
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

func TestCreateGuest(t *testing.T) {
	repo := NewUserRepository(dbPool)
	ctx := context.Background()

	// Create a guest user
	guestUser := &types.User{ID: utilities.MustGenerateIDString(), Email: ""}
	err := repo.CreateGuest(ctx, guestUser)
	assert.NoError(t, err, "Expected no error while creating a guest user")
	assert.NotEmpty(t, guestUser.ID, "Expected guest user ID to be set")
	assert.Equal(t, "guest", guestUser.Role, "Expected role to be 'guest'")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", guestUser.ID)
	assert.NoError(t, err, "Expected no error on guest user deletion")
}
