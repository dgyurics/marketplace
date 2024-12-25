package models

import "time"

type RefreshToken struct {
	ID        string    `json:"id"`
	User      *User     `json:"user"`
	TokenHash string    `json:"token_hash"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	Revoked   bool      `json:"revoked"`
	LastUsed  time.Time `json:"last_used"`
}
