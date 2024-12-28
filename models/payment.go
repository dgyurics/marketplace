package models

import "time"

type PaymentIntent struct {
	ID           string `json:"id"`
	Amount       int64  `json:"amount"`
	Currency     string `json:"currency"`
	Status       string `json:"status"`
	ClientSecret string `json:"client_secret"`
	Error        string `json:"error,omitempty"`
}

type Payment struct {
	OrderID         string    `json:"order_id"`
	PaymentIntentID string    `json:"payment_intent_id"`
	ClientSecret    string    `json:"client_secret"`
	Amount          int64     `json:"amount"`
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
