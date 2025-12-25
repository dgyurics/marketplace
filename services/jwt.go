package services

import (
	"errors"
	"time"

	"github.com/dgyurics/marketplace/types"
	"github.com/golang-jwt/jwt/v5"
)

// JWTService is an interface for generating and parsing access tokens
// In this case, an access token is a JSON Web Token (JWT)
type JWTService interface {
	GenerateToken(user types.User) (string, error)
	ParseToken(token string) (*types.User, error)
}

type jwtService struct {
	privateKey []byte
	publicKey  []byte
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

// GenerateToken generates a JWT token for a user
func (j *jwtService) GenerateToken(user types.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     now.Add(j.expiry).Unix(),
		"iat":     now.Unix(),
	}
	tokenUnsigned := jwt.NewWithClaims(jwt.SigningMethodRS256, claims) // create unsigned jwt object
	signingKey, err := jwt.ParseRSAPrivateKeyFromPEM(j.privateKey)
	if err != nil {
		return "", err
	}
	return tokenUnsigned.SignedString(signingKey) // sign the token with private key
}

// ParseToken parses a JWT token and returns the user it represents
// If the token is invalid, an error is returned
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
		user.Email = email
	}
	if role, ok := claims["role"].(string); ok {
		user.Role = types.Role(role)
	}

	return &user, nil
}
