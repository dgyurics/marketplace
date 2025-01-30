package models

import (
	"time"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Password     string    `json:"-"`
	PasswordHash string    `json:"-"`
	Admin        bool      `json:"admin"`
	UpdatedAt    time.Time `json:"updated_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type Address struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Addressee    string    `json:"addressee"`
	AddressLine1 string    `json:"address_line1"`
	AddressLine2 *string   `json:"address_line2"`
	City         string    `json:"city"`
	StateCode    string    `json:"state_code"`
	PostalCode   string    `json:"postal_code"`
	Phone        *string   `json:"phone"`
	IsDeleted    bool      `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
