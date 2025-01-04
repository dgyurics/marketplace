package models

type PaymentIntent struct {
	ID           string `json:"id"`
	Amount       int64  `json:"amount"`
	Currency     string `json:"currency"`
	Status       string `json:"status"`
	ClientSecret string `json:"client_secret"`
	Error        string `json:"error,omitempty"`
}

type StripeWebhookEvent struct {
	ID       string             `json:"id"`
	Type     string             `json:"type"`
	Data     *StripeWebhookData `json:"data"`
	Livemode bool               `json:"livemode"`
	Created  int64              `json:"created"` // seconds elapsed since Unix epoch
}

// TODO - add support for other webhook events
type StripeWebhookData struct {
	Object StripeWebhookPaymentIntent `json:"object"`
}

type StripeWebhookPaymentIntent struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	Amount       int64  `json:"amount"`
	ClientSecret string `json:"client_secret"`
	Currency     string `json:"currency"`
}
