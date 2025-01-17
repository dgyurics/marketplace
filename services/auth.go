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

type contextKey string

const UserKey contextKey = "user"

type AuthService interface {
	GenerateAccessToken(user models.User) (token string, err error)
	ValidateAccessToken(token string) (user models.User, err error)
	GenerateRefreshToken() (string, error)
	ValidateRefreshToken(ctx context.Context, token string) (models.User, error)
	StoreRefreshToken(ctx context.Context, userID, token string) error
	RevokeRefreshTokens(ctx context.Context) error
}

type authService struct {
	repo                 repositories.AuthRepository
	privateKey           []byte        // asymmetric key pair for signing access tokens
	publicKey            []byte        // asymmetric key pair for verifying access tokens
	hmacSecret           []byte        // symmetric key for hashing refresh tokens
	durationAccessToken  time.Duration // duration of jwt access token
	durationRefreshToken time.Duration // duration of refresh token
}

func NewAuthService(
	repo repositories.AuthRepository,
	config models.AuthConfig) AuthService {
	return &authService{
		repo:                 repo,
		privateKey:           config.PrivateKey,
		publicKey:            config.PublicKey,
		hmacSecret:           config.HMACSecret,
		durationAccessToken:  config.DurationAccessToken,
		durationRefreshToken: config.DurationRefreshToken,
	}
}

func (a *authService) GenerateAccessToken(user models.User) (token string, err error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"phone":   user.Phone,
		"admin":   user.Admin,
		"exp":     now.Add(a.durationAccessToken).Unix(),
		"iat":     now.Unix(),
	}
	tokenUnsigned := jwt.NewWithClaims(jwt.SigningMethodRS256, claims) // create unsigned jwt object
	signingKey, err := jwt.ParseRSAPrivateKeyFromPEM(a.privateKey)
	if err != nil {
		return "", err
	}
	token, err = tokenUnsigned.SignedString(signingKey) // sign the token with private key
	return
}

func (a *authService) ValidateAccessToken(token string) (user models.User, err error) {
	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwt.ParseRSAPublicKeyFromPEM(a.publicKey)
	})
	if err != nil {
		return models.User{}, err
	}
	if claims, ok := tokenParsed.Claims.(jwt.MapClaims); ok && tokenParsed.Valid {
		user = models.User{
			ID:    claims["user_id"].(string),
			Email: claims["email"].(string),
			Phone: claims["phone"].(string),
			Admin: claims["admin"].(bool),
		}
		return
	}
	return models.User{}, errors.New("invalid token")
}

func (a *authService) GenerateRefreshToken() (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", errors.New("failed to generate refresh token")
	}
	return hex.EncodeToString(token), nil
}

func (a *authService) ValidateRefreshToken(ctx context.Context, token string) (usr models.User, err error) {
	now := time.Now()
	// Retrieve the refresh token from the repository
	tokenHash := hashRefreshToken(token, a.hmacSecret)
	refreshToken, err := a.repo.GetRefreshToken(ctx, tokenHash)
	if err != nil || refreshToken == nil {
		return usr, errors.New("invalid refresh token")
	}

	if refreshToken.Revoked || refreshToken.ExpiresAt.Before(now) {
		return usr, errors.New("refresh token is either revoked or expired")
	}

	// Update the last used time
	refreshToken.LastUsed = now.UTC()
	if err := a.repo.StoreRefreshToken(ctx, *refreshToken); err != nil {
		return usr, errors.New("failed to update refresh token usage")
	}

	// Return the user associated with the refresh token
	return *refreshToken.User, nil
}

func (a *authService) StoreRefreshToken(ctx context.Context, userID, token string) error {
	now := time.Now().UTC()
	return a.repo.StoreRefreshToken(ctx, models.RefreshToken{
		User:      &models.User{ID: userID},
		TokenHash: hashRefreshToken(token, a.hmacSecret),
		Revoked:   false,
		ExpiresAt: now.Add(a.durationRefreshToken),
		CreatedAt: now,
		LastUsed:  now,
	})
}

func (a *authService) RevokeRefreshTokens(ctx context.Context) error {
	var userID = getUserID(ctx)
	return a.repo.RevokeRefreshTokens(ctx, userID)
}

func hashRefreshToken(token string, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(token))                // FIXME check for error
	return hex.EncodeToString(h.Sum(nil)) // return the final HMAC hash as a hexadecimal string
}

func getUserID(ctx context.Context) string {
	user, ok := ctx.Value(UserKey).(*models.User)
	if !ok {
		return ""
	}
	return user.ID
}
