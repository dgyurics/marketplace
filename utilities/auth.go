package utilities

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTUtility struct {
	privateKey []byte
	publicKey  []byte
}

func NewJWTUtility(privateKey, publicKey []byte) JWTUtility {
	return JWTUtility{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

func (j *JWTUtility) CreateToken(userID string, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(duration).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)     // create unsigned jwt object
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(j.privateKey) // parse private key
	if err != nil {
		return "", err
	}

	tokenString, err := token.SignedString(privateKey) // sign the token with private key
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (j *JWTUtility) VerifyToken(tokenString string) (jwt.Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwt.ParseRSAPublicKeyFromPEM(j.publicKey)
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (j *JWTUtility) CreateRefreshToken() (string, error) {
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", errors.New("failed to generate refresh token")
	}

	refreshToken := hex.EncodeToString(tokenBytes)
	return refreshToken, nil
}
