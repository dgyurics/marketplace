package models

type PaymentProvider string

const (
	Stripe PaymentProvider = "stripe"
	PayPal PaymentProvider = "paypal"
)

type PaymentIntentRequest struct {
	Provider PaymentProvider
	Amount   Currency
	Currency string
	TokenID  string
}

type PaymentIntentResponse struct {
	PaymentIntentID string   `json:"payment_intent_id"`
	ClientSecret    string   `json:"client_secret,omitempty"`
	Amount          Currency `json:"amount"`
	Currency        string   `json:"currency"`
	Status          string   `json:"status"`
}

type PaymentIntent struct {
	Status         string
	AmountReceived Currency
}
