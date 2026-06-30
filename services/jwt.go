package services

import (
	"errors"
	"time"

	"github.com/dgyurics/marketplace/types"
	"github.com/golang-jwt/jwt/v5"
)

// JWTService generates and validates access tokens using RS256 (RSA + SHA-256).
//
// RS256 is an asymmetric algorithm: the private key signs tokens and the public
// key verifies them. This separation means verification can happen in services
// that never hold the signing secret, unlike HMAC (HS256) where the same shared
// secret both signs and verifies — any service that can verify can also forge.
type JWTService interface {
	GenerateToken(user types.User) (string, error)
	ParseToken(token string) (*types.User, error)
}

type jwtService struct {
	privateKey []byte // PEM-encoded RSA private key used for signing
	publicKey  []byte // PEM-encoded RSA public key used for verification
	expiry     time.Duration
}

// NewJWTService returns an implementation of JWTService
func NewJWTService(config types.JWTConfig) JWTService {
	return &jwtService{
		privateKey: config.PrivateKey,
		publicKey:  config.PublicKey,
		expiry:     config.Expiry,
	}
}

// GenerateToken creates a signed JWT containing the user's ID, email, and role.
// The token is signed with the RSA private key so that any holder of the
// corresponding public key can verify authenticity without being able to
// mint new tokens.
func (j *jwtService) GenerateToken(user types.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     now.Add(j.expiry).Unix(),
		"iat":     now.Unix(),
	}
	tokenUnsigned := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signingKey, err := jwt.ParseRSAPrivateKeyFromPEM(j.privateKey)
	if err != nil {
		return "", err
	}
	return tokenUnsigned.SignedString(signingKey)
}

// ParseToken verifies the token signature using the RSA public key, checks
// expiration, and extracts the embedded user claims.
func (j *jwtService) ParseToken(token string) (*types.User, error) {
	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwt.ParseRSAPublicKeyFromPEM(j.publicKey)
	})
	if err != nil {
		return nil, err
	}
	if !tokenParsed.Valid {
		return nil, errors.New("invalid token")
	}

	// Extract claims
	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token")
	}
	user := types.User{}
	if id, ok := claims["user_id"].(string); ok {
		user.ID = id
	}
	if email, ok := claims["email"].(string); ok {
		user.Email = &email
	}
	if role, ok := claims["role"].(string); ok {
		user.Role = types.Role(role)
	}

	return &user, nil
}
