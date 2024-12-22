package models

import "time"

// TODO refactor this with stripe.go models considering these
// are specific to stripe http requests
type PaymentIntentRequest struct {
	Amount   int64
	Currency string
}

type PaymentIntentResponse struct {
	ID           string `json:"id"`
	Amount       int64  `json:"amount"`
	Currency     string `json:"currency"`
	Status       string `json:"status"`
	ClientSecret string `json:"client_secret"`
	Error        string `json:"error,omitempty"`
}

type PaymentIntent struct {
	Status         string
	AmountReceived int64
}

type Payment struct {
	PaymentIntentID string    `json:"payment_intent_id"`
	ClientSecret    string    `json:"client_secret"`
	Amount          int64     `json:"amount"`
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
	OrderID         string    `json:"order_id"`
	CreatedAt       time.Time `json:"created_at"`
}
