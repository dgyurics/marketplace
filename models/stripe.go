package models

type StripeWebhookEvent struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	Data     StripeWebhookData `json:"data"`
	Livemode bool              `json:"livemode"`
}

type StripeWebhookData struct {
	Object StripeWebhookPaymentIntent `json:"object"`
}

type StripeWebhookPaymentIntent struct {
	ID           string `json:"id"` // PaymentIntent ID should match payments.paymentIntentID
	Status       string `json:"status"`
	Amount       int    `json:"amount"`        // Amount should match payments.amount
	ClientSecret string `json:"client_secret"` // ClientSecret should match payments.clientSecret
	Currency     string `json:"currency"`      // Currency should match payments.currency
}
