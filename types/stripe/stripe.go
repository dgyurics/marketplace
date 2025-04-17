package stripe

type Event struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Data     *Data  `json:"data"`
	Livemode bool   `json:"livemode"`
	Created  int64  `json:"created"` // seconds elapsed since Unix epoch
}

// TODO - add support for other webhook events
type Data struct {
	Object PaymentIntent `json:"object"`
}

type PaymentIntent struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	Amount       int64  `json:"amount"`
	ClientSecret string `json:"client_secret"`
	Currency     string `json:"currency"`
	Error        string `json:"error,omitempty"`
}

type TaxCalculationResponse struct {
	ID                 string           `json:"id"`
	Object             string           `json:"object"`
	AmountTotal        int64            `json:"amount_total"`
	ShippingCost       int64            `json:"shipping_cost"`
	TaxAmountExclusive int64            `json:"tax_amount_exclusive"`
	TaxAmountInclusive int64            `json:"tax_amount_inclusive"`
	Currency           string           `json:"currency"`
	CustomerDetails    *CustomerDetails `json:"customer_details,omitempty"`
	ExpiresAt          int64            `json:"expires_at"`
	Livemode           bool             `json:"livemode"`
	TaxBreakdown       []TaxBreakdown   `json:"tax_breakdown"`
	TaxDate            int64            `json:"tax_date"`
}

type CustomerDetails struct {
	Address            Address  `json:"address"`
	AddressSource      string   `json:"address_source"`
	IPAddress          *string  `json:"ip_address,omitempty"`
	TaxIDs             []string `json:"tax_ids"`
	TaxabilityOverride string   `json:"taxability_override"`
}

type Address struct {
	City       string  `json:"city"`
	Country    string  `json:"country"`
	Line1      string  `json:"line1"`
	Line2      *string `json:"line2,omitempty"`
	PostalCode string  `json:"postal_code"`
	State      string  `json:"state"`
}

type TaxBreakdown struct {
	Amount           int64          `json:"amount"`
	Inclusive        bool           `json:"inclusive"`
	TaxRateDetails   TaxRateDetails `json:"tax_rate_details"`
	TaxabilityReason string         `json:"taxability_reason"`
	TaxableAmount    int64          `json:"taxable_amount"`
}

type TaxRateDetails struct {
	Country           string `json:"country"`
	FlatAmount        *int64 `json:"flat_amount"`
	PercentageDecimal string `json:"percentage_decimal"`
	RateType          string `json:"rate_type"`
	State             string `json:"state"`
	TaxType           string `json:"tax_type"`
}
