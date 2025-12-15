package types

import (
	"time"
)

type PendingUser struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CodeHash  string    `json:"-"`
	Used      bool      `json:"used"`
	ExpiresAt time.Time `json:"expires_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	Password      string    `json:"-"`
	PasswordHash  string    `json:"-"`
	Role          string    `json:"role"`
	RequiresSetup bool      `json:"requires_setup"`
	UpdatedAt     time.Time `json:"updated_at"`
	CreatedAt     time.Time `json:"created_at"`
}

func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

func (u *User) IsUser() bool {
	return u.Role == "user"
}

func (u *User) IsGuest() bool {
	return u.Role == "guest"
}
