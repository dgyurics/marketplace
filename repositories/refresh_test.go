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

	// Verify that both tokens are revoked (GetToken filters out revoked tokens, so query directly)
	var revoked1, revoked2 bool
	err = dbPool.QueryRowContext(ctx, "SELECT revoked FROM refresh_tokens WHERE token_hash = $1", refreshToken1.TokenHash).Scan(&revoked1)
	assert.NoError(t, err, "Expected no error on getting first refresh token")
	assert.True(t, revoked1, "Expected first token to be revoked")

	err = dbPool.QueryRowContext(ctx, "SELECT revoked FROM refresh_tokens WHERE token_hash = $1", refreshToken2.TokenHash).Scan(&revoked2)
	assert.NoError(t, err, "Expected no error on getting second refresh token")
	assert.True(t, revoked2, "Expected second token to be revoked")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE user_id = $1", refreshToken1.User.ID)
	assert.NoError(t, err, "Expected no error on refresh token deletion")
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on user deletion")
}
