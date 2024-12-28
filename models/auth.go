package models

import "time"

type Credential struct {
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthConfig struct {
	PrivateKey           []byte        // asymmetric key pair for signing access tokens
	PublicKey            []byte        // asymmetric key pair for verifying access tokens
	HMACSecret           []byte        // symmetric key for hashing refresh tokens
	DurationAccessToken  time.Duration // duration of jwt access token
	DurationRefreshToken time.Duration // duration of refresh token
}
