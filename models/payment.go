package models

// TODO refactor this with stripe.go models considering these
// are specific to stripe http requests
type PaymentIntentRequest struct {
	Amount   Currency
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
	AmountReceived Currency
}
