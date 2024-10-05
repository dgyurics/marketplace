package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/dgyurics/marketplace/models"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	repo := NewUserRepository(dbPool)
	ctx := context.Background()

	// Create a test user
	user := &models.User{
		Email:        "testuser@example.com",
		Phone:        "1234567890",
		PasswordHash: "hashedpassword",
	}

	// Create the user
	err := repo.CreateUser(ctx, user)
	assert.NoError(t, err, "Expected no error on user creation")
	assert.NotEmpty(t, user.ID, "Expected user ID to be set")
	assert.Equal(t, "testuser@example.com", user.Email, "Expected email to match")
	assert.Equal(t, "1234567890", user.Phone, "Expected phone to match")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on user deletion")
}

func TestGetUserByPhone(t *testing.T) {
	repo := NewUserRepository(dbPool)
	ctx := context.Background()

	// Create a test user
	user := &models.User{
		Email:        "testuser@example.com",
		Phone:        "1234567890",
		PasswordHash: "hashedpassword",
	}

	// Insert the test user
	err := repo.CreateUser(ctx, user)
	assert.NoError(t, err, "Expected no error on user creation")

	// Retrieve the user by phone number
	retrievedUser, err := repo.GetUserByPhone(ctx, "1234567890")
	assert.NoError(t, err, "Expected no error on getting user by phone")
	assert.NotNil(t, retrievedUser, "Expected retrieved user to not be nil")
	assert.Equal(t, user.ID, retrievedUser.ID, "Expected user ID to match")
	assert.Equal(t, user.Email, retrievedUser.Email, "Expected email to match")
	assert.Equal(t, user.Phone, retrievedUser.Phone, "Expected phone to match")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on user deletion")
}

func TestGetUserByEmail(t *testing.T) {
	repo := NewUserRepository(dbPool)
	ctx := context.Background()

	// Create a test user
	user := &models.User{
		Email:        "testuser@example.com",
		Phone:        "1234567890",
		PasswordHash: "hashedpassword",
	}

	// Insert the test user
	err := repo.CreateUser(ctx, user)
	assert.NoError(t, err, "Expected no error on user creation")

	// Retrieve the user by email
	retrievedUser, err := repo.GetUserByEmail(ctx, "testuser@example.com")
	assert.NoError(t, err, "Expected no error on getting user by email")
	assert.NotNil(t, retrievedUser, "Expected retrieved user to not be nil")
	assert.Equal(t, user.ID, retrievedUser.ID, "Expected user ID to match")
	assert.Equal(t, user.Email, retrievedUser.Email, "Expected email to match")
	assert.Equal(t, user.Phone, retrievedUser.Phone, "Expected phone to match")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on user deletion")
}

func TestGetAllUsers(t *testing.T) {
	repo := NewUserRepository(dbPool)
	ctx := context.Background()

	// Create two test users
	user1 := &models.User{
		Email:        "testuser1@example.com",
		Phone:        "1234567891",
		PasswordHash: "hashedpassword1",
	}
	user2 := &models.User{
		Email:        "testuser2@example.com",
		Phone:        "1234567892",
		PasswordHash: "hashedpassword2",
	}

	// Insert the test users
	err := repo.CreateUser(ctx, user1)
	assert.NoError(t, err, "Expected no error on user1 creation")
	err = repo.CreateUser(ctx, user2)
	assert.NoError(t, err, "Expected no error on user2 creation")

	// Retrieve all users
	users, err := repo.GetAllUsers(ctx)
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

func TestStoreRefreshToken(t *testing.T) {
	repo := NewUserRepository(dbPool)
	ctx := context.Background()

	// Create a test user
	user := &models.User{
		Email:        "testuser@example.com",
		Phone:        "1234567890",
		PasswordHash: "hashedpassword",
	}

	// Insert the test user
	err := repo.CreateUser(ctx, user)
	assert.NoError(t, err, "Expected no error on user creation")

	// Create a refresh token
	refreshToken := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: "testtokenhash",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		Revoked:   false,
		LastUsed:  time.Now(),
	}

	// Store the refresh token
	err = repo.StoreRefreshToken(ctx, refreshToken)
	assert.NoError(t, err, "Expected no error on storing refresh token")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE user_id = $1", refreshToken.UserID)
	assert.NoError(t, err, "Expected no error on refresh token deletion")
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on user deletion")
}
