package types

import "time"

type Credential struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	InviteCode string `json:"invite_code"`
	ResetCode  string `json:"reset_code"`
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

type PasswordReset struct {
	ID        string    `json:"id"`
	User      *User     `json:"user"`
	CodeHash  string    `json:"code_hash"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
