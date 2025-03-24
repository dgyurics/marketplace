package services_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	hmacSecret         = "big_secret"
	refreshTokenExpiry = 30 * 24 * time.Hour
)

type MockRefreshRepository struct {
	mock.Mock
}

func (m *MockRefreshRepository) StoreToken(ctx context.Context, refreshToken types.RefreshToken) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *MockRefreshRepository) GetToken(ctx context.Context, tokenHash string) (*types.RefreshToken, error) {
	args := m.Called(ctx, tokenHash)
	return args.Get(0).(*types.RefreshToken), args.Error(1)
}

func (m *MockRefreshRepository) RevokeTokens(ctx context.Context, tokenHash string) error {
	args := m.Called(ctx, tokenHash)
	return args.Error(0)
}

// Helper function to create an AuthService with configuration
func createRefreshService(repo *MockRefreshRepository) services.RefreshService {
	return services.NewRefreshService(repo, types.AuthConfig{
		HMACSecret: []byte(hmacSecret),
	})
}

func TestGenerateRefreshToken(t *testing.T) {
	repo := new(MockRefreshRepository)
	refreshService := createRefreshService(repo)

	refreshToken, err := refreshService.GenerateToken()
	assert.NoError(t, err, "expected no error in generating refresh token")
	assert.NotEmpty(t, refreshToken, "expected a non-empty refresh token")
	assert.Equal(t, 64, len(refreshToken), "expected refresh token to be 64 characters long")
}

func TestValidateRefreshToken(t *testing.T) {
	now := time.Now()
	repo := new(MockRefreshRepository)
	refreshService := createRefreshService(repo)

	// Mock the behavior of the repository
	refreshToken := "test_refresh_token"
	expiresAt := now.Add(24 * time.Hour)

	// Mock the repository to return a valid refresh token object
	repo.On("GetToken", mock.Anything, mock.Anything).Return(&types.RefreshToken{
		User:      &types.User{ID: "user123"},
		TokenHash: hashRefreshToken(refreshToken, []byte(hmacSecret)),
		ExpiresAt: expiresAt,
		Revoked:   false,
		LastUsed:  now.UTC(),
		CreatedAt: now.UTC(),
	}, nil)
	// Mock the repository to return no error when storing the refresh token (service will update the LastUsed field)
	repo.On("StoreToken", mock.Anything, mock.AnythingOfType("types.RefreshToken")).Return(nil)

	valid, err := refreshService.VerifyToken(context.Background(), refreshToken)
	assert.NoError(t, err, "expected no error in validating refresh token")
	assert.NotNil(t, valid, "expected a valid user object")
}

func TestStoreRefreshToken(t *testing.T) {
	repo := new(MockRefreshRepository)
	refreshService := createRefreshService(repo)

	// Mock the behavior of the repository
	token := "test_refresh_token"
	userID := "user123"

	// Expect that the StoreRefreshToken method will be called with any context and a RefreshToken struct
	repo.On("StoreToken", mock.Anything, mock.AnythingOfType("types.RefreshToken")).Return(nil)

	err := refreshService.StoreToken(context.Background(), userID, token)
	assert.NoError(t, err, "expected no error in storing refresh token")
}

func TestRevokeRefreshTokens(t *testing.T) {
	repo := new(MockRefreshRepository)
	refreshService := createRefreshService(repo)

	// Mock the behavior of the repository
	user := &types.User{
		ID:    "user123",
		Email: "user@example.com",
	}
	ctx := context.WithValue(context.Background(), services.UserKey, user)
	repo.On("RevokeTokens", mock.Anything, user.ID).Return(nil)

	err := refreshService.RevokeTokens(ctx)
	assert.NoError(t, err, "expected no error in revoking refresh tokens")
}

func hashRefreshToken(token string, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(token))                // FIXME check for error
	return hex.EncodeToString(h.Sum(nil)) // return the final HMAC hash as a hexadecimal string
}
