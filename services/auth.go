package services

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/repositories"
	"github.com/golang-jwt/jwt/v5"
)

type AuthService interface {
	GenerateAccessToken(userID string) (token string, err error)
	ValidateAccessToken(token string) (userID string, err error)
	GenerateRefreshToken() (string, error)
	ValidateRefreshToken(ctx context.Context, token string) (bool, error)
	StoreRefreshToken(ctx context.Context, userID, token string) error
	RevokeAllRefreshTokens(ctx context.Context, token string) error
}

type authService struct {
	privateKey           []byte // asymmetric key pair for signing access tokens
	publicKey            []byte // asymmetric key pair for verifying access tokens
	hmacSecret           []byte // symmetric key for hashing refresh tokens
	repo                 repositories.AuthRepository
	durationAccessToken  time.Duration
	durationRefreshToken time.Duration
}

func NewAuthService(
	repo repositories.AuthRepository,
	privateKey,
	publicKey,
	hmacSecret []byte) AuthService {
	return &authService{
		privateKey:           privateKey,
		publicKey:            publicKey,
		hmacSecret:           hmacSecret,
		repo:                 repo,
		durationAccessToken:  time.Duration(15) * time.Minute,  // 15 minutes
		durationRefreshToken: time.Duration(30*24) * time.Hour, // 30 days
	}
}

func (a *authService) GenerateAccessToken(userID string) (token string, err error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(a.durationAccessToken).Unix(),
		"iat":     time.Now().Unix(),
	}
	tokenUnsigned := jwt.NewWithClaims(jwt.SigningMethodRS256, claims) // create unsigned jwt object
	signingKey, err := jwt.ParseRSAPrivateKeyFromPEM(a.privateKey)
	if err != nil {
		return "", err
	}
	token, err = tokenUnsigned.SignedString(signingKey) // sign the token with private key
	return
}

func (a *authService) ValidateAccessToken(token string) (userID string, err error) {
	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwt.ParseRSAPublicKeyFromPEM(a.publicKey)
	})
	if err != nil {
		return "", err
	}
	if claims, ok := tokenParsed.Claims.(jwt.MapClaims); ok && tokenParsed.Valid {
		userID = claims["user_id"].(string)
		return
	}
	return userID, errors.New("invalid token")
}

func (a *authService) GenerateRefreshToken() (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", errors.New("failed to generate refresh token")
	}
	return hex.EncodeToString(token), nil
}

func (a *authService) ValidateRefreshToken(ctx context.Context, token string) (bool, error) {
	// Retrieve the refresh token from the repository
	refreshToken, err := a.repo.GetRefreshToken(ctx, token)
	if err != nil {
		return false, errors.New("invalid refresh token")
	}

	// Check if the token is revoked or expired
	if refreshToken.Revoked || refreshToken.ExpiresAt.Before(time.Now()) {
		return false, errors.New("refresh token is either revoked or expired")
	}

	// Update the last used time
	refreshToken.LastUsed = time.Now()
	if err := a.repo.StoreRefreshToken(ctx, *refreshToken); err != nil {
		return false, errors.New("failed to update refresh token usage")
	}

	return true, nil
}

func (a *authService) StoreRefreshToken(ctx context.Context, userID, token string) error {
	return a.repo.StoreRefreshToken(ctx, models.RefreshToken{
		UserID:    userID,
		TokenHash: hashRefreshToken(token, a.hmacSecret),
		ExpiresAt: time.Now().Add(a.durationRefreshToken),
		CreatedAt: time.Now(),
		Revoked:   false,
		LastUsed:  time.Now(),
	})
}

func (a *authService) RevokeAllRefreshTokens(ctx context.Context, token string) error {
	return a.repo.RevokeAllRefreshTokens(ctx, token)
}

func hashRefreshToken(token string, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil)) // return the final HMAC hash as a hexadecimal string
}

// getUserID retrieves the user ID from the context
func getUserID(ctx context.Context) string {
	userID, ok := ctx.Value("userID").(string)
	if !ok {
		return ""
	}
	return userID
}
