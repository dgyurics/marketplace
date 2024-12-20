package models

type StripeWebhookEvent struct {
	ID       string             `json:"id"`
	Type     string             `json:"type"`
	Data     *StripeWebhookData `json:"data"`
	Livemode bool               `json:"livemode"`
	Created  int64              `json:"created"` // Time at which the object was created. Measured in seconds since the Unix epoch.
}

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
