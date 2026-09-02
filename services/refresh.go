package services

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
)

// RefreshService handles the creation, validation, and revocation of refresh tokens.
type RefreshService interface {
	GenerateToken() (string, error)
	StoreToken(ctx context.Context, userID, token string) error
	GetToken(ctx context.Context, token string) (types.RefreshToken, error)
	RevokeTokens(ctx context.Context) error
}

type refreshService struct {
	repo   repositories.RefreshRepository
	config types.AuthConfig
}

// NewRefreshService creates a new RefreshService instance.
func NewRefreshService(repo repositories.RefreshRepository, config types.AuthConfig) RefreshService {
	return &refreshService{
		repo:   repo,
		config: config,
	}
}

// GenerateToken creates a new random refresh token.
func (s *refreshService) GenerateToken() (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	return hex.EncodeToString(token), nil
}

// StoreToken creates a new refresh token and stores it in the database, associating it with a user.
func (s *refreshService) StoreToken(ctx context.Context, userID, token string) error {
	now := time.Now().UTC()
	tokenID, err := utilities.GenerateIDString()
	if err != nil {
		return err
	}
	return s.repo.StoreToken(ctx, types.RefreshToken{
		ID:        tokenID,
		User:      &types.User{ID: userID},
		TokenHash: hashString(token, s.config.HMACSecret),
		ExpiresAt: now.Add(s.config.RefreshExpiry),
	})
}

func (s *refreshService) GetToken(ctx context.Context, token string) (types.RefreshToken, error) {
	tokenHash := hashString(token, s.config.HMACSecret)
	return s.repo.GetToken(ctx, tokenHash)
}

// RevokeTokens revokes all refresh tokens for the authenticated user.
func (s *refreshService) RevokeTokens(ctx context.Context) error {
	var userID = getUserID(ctx)
	return s.repo.RevokeTokens(ctx, userID)
}

func hashString(token string, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(token))                // FIXME check for error
	return hex.EncodeToString(h.Sum(nil)) // return the final HMAC hash as a hexadecimal string
}

// TODO: move somewhere else (util?)
func getUserID(ctx context.Context) string {
	user, ok := ctx.Value(UserKey).(*types.User)
	if !ok {
		return ""
	}
	return user.ID
}
