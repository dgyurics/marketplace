package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
	"github.com/stretchr/testify/assert"
)

func TestStoreToken(t *testing.T) {
	repo := NewRefreshRepository(dbPool)
	ctx := context.Background()
	now := time.Now()

	// Create a unique test user
	user := createUniqueTestUser(t, NewUserRepository(dbPool))
	tokenID := utilities.MustGenerateIDString()

	// Create a refresh token
	refreshToken := types.RefreshToken{
		ID:        tokenID,
		User:      user,
		TokenHash: "testtokenhash",
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
		Revoked:   false,
		LastUsed:  now,
	}

	// Store the refresh token
	err := repo.StoreToken(ctx, refreshToken)
	assert.NoError(t, err, "Expected no error on storing refresh token")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE user_id = $1", refreshToken.User.ID)
	assert.NoError(t, err, "Expected no error on refresh token deletion")
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on user deletion")
}

func TestGetRefreshToken_GuestUser(t *testing.T) {
	repo := NewRefreshRepository(dbPool)
	ctx := context.Background()
	now := time.Now()

	// Create a unique guest user
	userRepo := NewUserRepository(dbPool)
	guestUser := createUniqueGuestUser(t, userRepo)
	tokenID := utilities.MustGenerateIDString()

	// Create a refresh token
	refreshToken := types.RefreshToken{
		ID:        tokenID,
		User:      guestUser,
		TokenHash: "testguesttokenhash",
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
		Revoked:   false,
		LastUsed:  now,
	}

	// Store the refresh token
	err := repo.StoreToken(ctx, refreshToken)
	assert.NoError(t, err, "Expected no error on storing guest refresh token")

	// Retrieve the refresh token
	retrievedToken, err := repo.GetToken(ctx, refreshToken.TokenHash)
	assert.NoError(t, err, "Expected no error on getting guest refresh token")
	assert.NotNil(t, retrievedToken, "Expected the retrieved guest token to not be nil")
	assert.Equal(t, refreshToken.User.ID, retrievedToken.User.ID, "Expected guest user ID to match")
	assert.Equal(t, refreshToken.TokenHash, retrievedToken.TokenHash, "Expected token hash to match")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE user_id = $1", refreshToken.User.ID)
	assert.NoError(t, err, "Expected no error on guest refresh token deletion")
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", guestUser.ID)
	assert.NoError(t, err, "Expected no error on guest user deletion")
}

func TestGetRefreshToken(t *testing.T) {
	repo := NewRefreshRepository(dbPool)
	ctx := context.Background()
	now := time.Now()

	// Create a unique test user
	user := createUniqueTestUser(t, NewUserRepository(dbPool))
	tokenID := utilities.MustGenerateIDString()

	// Create a refresh token
	refreshToken := types.RefreshToken{
		ID:        tokenID,
		User:      user,
		TokenHash: "testtokenhash",
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
		Revoked:   false,
		LastUsed:  now,
	}

	// Store the refresh token
	err := repo.StoreToken(ctx, refreshToken)
	assert.NoError(t, err, "Expected no error on storing refresh token")

	// Retrieve the refresh token
	retrievedToken, err := repo.GetToken(ctx, refreshToken.TokenHash)
	assert.NoError(t, err, "Expected no error on getting refresh token")
	assert.NotNil(t, retrievedToken, "Expected the retrieved token to not be nil")
	assert.Equal(t, refreshToken.User.ID, retrievedToken.User.ID, "Expected user ID to match")
	assert.Equal(t, refreshToken.TokenHash, retrievedToken.TokenHash, "Expected token hash to match")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE user_id = $1", refreshToken.User.ID)
	assert.NoError(t, err, "Expected no error on refresh token deletion")
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on user deletion")
}

func TestRevokeTokens(t *testing.T) {
	repo := NewRefreshRepository(dbPool)
	ctx := context.Background()
	now := time.Now()

	// Create a unique test user
	user := createUniqueTestUser(t, NewUserRepository(dbPool))
	tokenID := utilities.MustGenerateIDString()

	// Create two refresh tokens for the same user
	refreshToken1 := types.RefreshToken{
		ID:        tokenID,
		User:      user,
		TokenHash: "testtokenhash1",
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
		Revoked:   false,
		LastUsed:  now,
	}
	tokenID = utilities.MustGenerateIDString()
	refreshToken2 := types.RefreshToken{
		ID:        tokenID,
		User:      user,
		TokenHash: "testtokenhash2",
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
		Revoked:   false,
		LastUsed:  now,
	}

	// Store both refresh tokens
	err := repo.StoreToken(ctx, refreshToken1)
	assert.NoError(t, err, "Expected no error on storing first refresh token")
	err = repo.StoreToken(ctx, refreshToken2)
	assert.NoError(t, err, "Expected no error on storing second refresh token")

	// Revoke all refresh tokens for the user
	err = repo.RevokeTokens(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on revoking all refresh tokens")

	// Verify that both tokens are revoked
	retrievedToken1, err := repo.GetToken(ctx, refreshToken1.TokenHash)
	assert.NoError(t, err, "Expected no error on getting first refresh token")
	assert.True(t, retrievedToken1.Revoked, "Expected first token to be revoked")

	retrievedToken2, err := repo.GetToken(ctx, refreshToken2.TokenHash)
	assert.NoError(t, err, "Expected no error on getting second refresh token")
	assert.True(t, retrievedToken2.Revoked, "Expected second token to be revoked")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE user_id = $1", refreshToken1.User.ID)
	assert.NoError(t, err, "Expected no error on refresh token deletion")
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on user deletion")
}

func TestUpdateLastUsed(t *testing.T) {
	repo := NewRefreshRepository(dbPool)
	ctx := context.Background()
	now := time.Now()

	// Create a unique test user
	user := createUniqueTestUser(t, NewUserRepository(dbPool))
	tokenID := utilities.MustGenerateIDString()

	// Create a refresh token
	refreshToken := types.RefreshToken{
		ID:        tokenID,
		User:      user,
		TokenHash: "testupdatelastusedtokenhash",
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
		Revoked:   false,
		LastUsed:  now,
	}

	// Store the refresh token
	err := repo.StoreToken(ctx, refreshToken)
	assert.NoError(t, err, "Expected no error on storing refresh token")

	// Wait a bit and update LastUsed
	newLastUsed := now.Add(1 * time.Minute).UTC()
	err = repo.UpdateLastUsed(ctx, refreshToken.ID, newLastUsed)
	assert.NoError(t, err, "Expected no error on updating last used timestamp")

	// Retrieve and verify LastUsed
	retrievedToken, err := repo.GetToken(ctx, refreshToken.TokenHash)
	assert.NoError(t, err, "Expected no error on getting refresh token")
	assert.WithinDuration(t, newLastUsed, retrievedToken.LastUsed, time.Second, "Expected last used timestamp to be updated")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE user_id = $1", refreshToken.User.ID)
	assert.NoError(t, err, "Expected no error on refresh token deletion")
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on user deletion")
}
