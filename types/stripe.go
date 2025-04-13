package types

// TODO move to a separate package and remove
// Stripe prefix from all structs
type StripeEvent struct {
	ID       string      `json:"id"`
	Type     string      `json:"type"`
	Data     *StripeData `json:"data"`
	Livemode bool        `json:"livemode"`
	Created  int64       `json:"created"` // seconds elapsed since Unix epoch
}

// TODO - add support for other webhook events
type StripeData struct {
	Object StripePaymentIntent `json:"object"`
}

type StripePaymentIntent struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	Amount       int64  `json:"amount"`
	ClientSecret string `json:"client_secret"`
	Currency     string `json:"currency"`
	Error        string `json:"error,omitempty"`
}

type StripeTaxCalculationResponse struct {
	ID                 string                 `json:"id"`
	Object             string                 `json:"object"`
	AmountTotal        int64                  `json:"amount_total"`
	ShippingCost       int64                  `json:"shipping_cost"`
	TaxAmountExclusive int64                  `json:"tax_amount_exclusive"`
	TaxAmountInclusive int64                  `json:"tax_amount_inclusive"`
	Currency           string                 `json:"currency"`
	CustomerDetails    *StripeCustomerDetails `json:"customer_details,omitempty"`
	ExpiresAt          int64                  `json:"expires_at"`
	Livemode           bool                   `json:"livemode"`
	TaxBreakdown       []StripeTaxBreakdown   `json:"tax_breakdown"`
	TaxDate            int64                  `json:"tax_date"`
}

type StripeCustomerDetails struct {
	Address            StripeAddress `json:"address"`
	AddressSource      string        `json:"address_source"`
	IPAddress          *string       `json:"ip_address,omitempty"`
	TaxIDs             []string      `json:"tax_ids"`
	TaxabilityOverride string        `json:"taxability_override"`
}

type StripeAddress struct {
	City       string  `json:"city"`
	Country    string  `json:"country"`
	Line1      string  `json:"line1"`
	Line2      *string `json:"line2,omitempty"`
	PostalCode string  `json:"postal_code"`
	State      string  `json:"state"`
}

type StripeTaxBreakdown struct {
	Amount           int64                `json:"amount"`
	Inclusive        bool                 `json:"inclusive"`
	TaxRateDetails   StripeTaxRateDetails `json:"tax_rate_details"`
	TaxabilityReason string               `json:"taxability_reason"`
	TaxableAmount    int64                `json:"taxable_amount"`
}

type StripeTaxRateDetails struct {
	Country           string `json:"country"`
	FlatAmount        *int64 `json:"flat_amount"`
	PercentageDecimal string `json:"percentage_decimal"`
	RateType          string `json:"rate_type"`
	State             string `json:"state"`
	TaxType           string `json:"tax_type"`
}
