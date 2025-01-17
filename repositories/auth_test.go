package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/dgyurics/marketplace/models"
	"github.com/stretchr/testify/assert"
)

func TestStoreRefreshToken(t *testing.T) {
	repo := NewAuthRepository(dbPool)
	ctx := context.Background()
	now := time.Now()

	// Create a unique test user
	user := createUniqueTestUser(t, NewUserRepository(dbPool))

	// Create a refresh token
	refreshToken := models.RefreshToken{
		User:      user,
		TokenHash: "testtokenhash",
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
		Revoked:   false,
		LastUsed:  now,
	}

	// Store the refresh token
	err := repo.StoreRefreshToken(ctx, refreshToken)
	assert.NoError(t, err, "Expected no error on storing refresh token")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE user_id = $1", refreshToken.User.ID)
	assert.NoError(t, err, "Expected no error on refresh token deletion")
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on user deletion")
}

func TestGetRefreshToken(t *testing.T) {
	repo := NewAuthRepository(dbPool)
	ctx := context.Background()
	now := time.Now()

	// Create a unique test user
	user := createUniqueTestUser(t, NewUserRepository(dbPool))

	// Create a refresh token
	refreshToken := models.RefreshToken{
		User:      user,
		TokenHash: "testtokenhash",
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
		Revoked:   false,
		LastUsed:  now,
	}

	// Store the refresh token
	err := repo.StoreRefreshToken(ctx, refreshToken)
	assert.NoError(t, err, "Expected no error on storing refresh token")

	// Retrieve the refresh token
	retrievedToken, err := repo.GetRefreshToken(ctx, refreshToken.TokenHash)
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

func TestRevokeRefreshTokens(t *testing.T) {
	repo := NewAuthRepository(dbPool)
	ctx := context.Background()
	now := time.Now()

	// Create a unique test user
	user := createUniqueTestUser(t, NewUserRepository(dbPool))

	// Create two refresh tokens for the same user
	refreshToken1 := models.RefreshToken{
		User:      user,
		TokenHash: "testtokenhash1",
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
		Revoked:   false,
		LastUsed:  now,
	}
	refreshToken2 := models.RefreshToken{
		User:      user,
		TokenHash: "testtokenhash2",
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
		Revoked:   false,
		LastUsed:  now,
	}

	// Store both refresh tokens
	err := repo.StoreRefreshToken(ctx, refreshToken1)
	assert.NoError(t, err, "Expected no error on storing first refresh token")
	err = repo.StoreRefreshToken(ctx, refreshToken2)
	assert.NoError(t, err, "Expected no error on storing second refresh token")

	// Revoke all refresh tokens for the user
	err = repo.RevokeRefreshTokens(ctx, user.ID)
	assert.NoError(t, err, "Expected no error on revoking all refresh tokens")

	// Verify that both tokens are revoked
	retrievedToken1, err := repo.GetRefreshToken(ctx, refreshToken1.TokenHash)
	assert.NoError(t, err, "Expected no error on getting first refresh token")
	assert.True(t, retrievedToken1.Revoked, "Expected first token to be revoked")

	retrievedToken2, err := repo.GetRefreshToken(ctx, refreshToken2.TokenHash)
	assert.NoError(t, err, "Expected no error on getting second refresh token")
	assert.True(t, retrievedToken2.Revoked, "Expected second token to be revoked")

	// Clean up
	_, err = dbPool.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE user_id = $1", refreshToken1.User.ID)
	assert.NoError(t, err, "Expected no error on refresh token deletion")
	_, err = dbPool.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	assert.NoError(t, err, "Expected no error on user deletion")
}
