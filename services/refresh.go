package services

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
)

// RefreshService handles the creation, validation, and revocation of refresh tokens.
type RefreshService interface {
	GenerateToken() (string, error)
	StoreToken(ctx context.Context, userID, token string) error
	VerifyToken(ctx context.Context, token string) (*types.User, error)
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
		return "", errors.New("failed to generate refresh token")
	}
	return hex.EncodeToString(token), nil
}

// StoreToken stores the refresh token in the database, associating it with a user.
func (s *refreshService) StoreToken(ctx context.Context, userID, token string) error {
	now := time.Now().UTC()
	return s.repo.StoreToken(ctx, types.RefreshToken{
		User:      &types.User{ID: userID},
		TokenHash: hashString(token, s.config.HMACSecret),
		Revoked:   false,
		ExpiresAt: now.Add(s.config.RefreshExpiry),
		CreatedAt: now,
		LastUsed:  now,
	})
}

// ValidateToken verifies the refresh token and returns the associated user if valid.
func (s *refreshService) VerifyToken(ctx context.Context, token string) (*types.User, error) {
	now := time.Now()
	tokenHash := hashString(token, s.config.HMACSecret)
	refreshToken, err := s.repo.GetToken(ctx, tokenHash)

	if err != nil {
		return nil, err
	}

	if refreshToken == nil {
		return nil, errors.New("refresh token not found")
	}

	if refreshToken.Revoked {
		return nil, errors.New("refresh token has been revoked")
	}

	if refreshToken.ExpiresAt.Before(now) {
		return nil, errors.New("refresh token has expired")
	}

	// Update the last used time
	refreshToken.LastUsed = now.UTC()
	if err := s.repo.StoreToken(ctx, *refreshToken); err != nil {
		return nil, errors.New("failed to update refresh token usage")
	}

	return refreshToken.User, nil
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

func getUserID(ctx context.Context) string {
	user, ok := ctx.Value(UserKey).(*types.User)
	if !ok {
		return ""
	}
	return user.ID
}
