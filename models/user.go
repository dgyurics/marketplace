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
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
