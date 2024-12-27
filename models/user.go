package models

import (
	"time"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	Password     string    `json:"-"`
	PasswordHash string    `json:"-"`
	Admin        bool      `json:"admin"`
	UpdatedAt    time.Time `json:"updated_at"`
	CreatedAt    time.Time `json:"created_at"`
}
