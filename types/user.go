package types

import (
	"time"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Password     string    `json:"-"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	UpdatedAt    time.Time `json:"updated_at"`
	CreatedAt    time.Time `json:"created_at"`
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
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	CountryCode  string    `json:"country_code"` // TODO rename to country ISO 3166-1 alpha-2
	Addressee    *string   `json:"addressee"`
	AddressLine1 string    `json:"address_line1"` // TODO rename to line1 Address line 1 (Street address/PO Box/Company name)
	AddressLine2 *string   `json:"address_line2"` // TODO rename to line2 Address line 2 (Apartment/Suite/Unit/Building)
	City         string    `json:"city"`          // City/District/Suburb/Town/Village
	StateCode    string    `json:"state_code"`    // TODO rename to state ISO 3166-2
	PostalCode   string    `json:"postal_code"`   // ZIP or postal code
	IsDeleted    bool      `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
