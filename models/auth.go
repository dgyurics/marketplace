package models

import "time"

type Credential struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	InviteCode string `json:"invite_code"`
}

type AuthConfig struct {
	PrivateKey           []byte        // asymmetric key pair for signing access tokens
	PublicKey            []byte        // asymmetric key pair for verifying access tokens
	HMACSecret           []byte        // symmetric key for hashing refresh tokens
	DurationAccessToken  time.Duration // duration of jwt access token
	DurationRefreshToken time.Duration // duration of refresh token
}

type RefreshToken struct {
	ID        string    `json:"id"`
	User      *User     `json:"user"`
	TokenHash string    `json:"token_hash"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
	LastUsed  time.Time `json:"last_used"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
