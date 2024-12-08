package models

type PaymentIntentRequest struct {
	Amount   Currency
	Currency string
}

type PaymentIntentResponse struct {
	ID           string `json:"id"`
	Object       string `json:"object"`
	Amount       int64  `json:"amount"`
	Currency     string `json:"currency"`
	Status       string `json:"status"`
	ClientSecret string `json:"client_secret"`
	Error        string `json:"error,omitempty"`
}

type PaymentIntent struct {
	Status         string
	AmountReceived Currency
}
