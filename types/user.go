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
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Country    string    `json:"country"`
	Addressee  *string   `json:"addressee"` // FIXME just use omit empty string
	Line1      string    `json:"line1"`
	Line2      *string   `json:"line2"`       // FIXME just use omit empty string
	City       string    `json:"city"`        // city, district, suburb, town, village
	State      string    `json:"state"`       // state, county, province, region
	PostalCode string    `json:"postal_code"` // zip code, postal code
	IsDeleted  bool      `json:"-"`           // when existing order references address, we have to soft delete
	CreatedAt  time.Time `json:"created_at"`
}
