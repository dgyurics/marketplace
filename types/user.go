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

type Address struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Country    string    `json:"country"` // FIXME this is country code
	Name       *string   `json:"name,omitempty"`
	Line1      string    `json:"line1"`
	Line2      *string   `json:"line2,omitempty"`
	City       string    `json:"city"`            // city, district, suburb, town, village
	State      *string   `json:"state,omitempty"` // state, county, province, region
	PostalCode string    `json:"postal_code"`     // zip code, postal code
	Email      string    `json:"email"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
