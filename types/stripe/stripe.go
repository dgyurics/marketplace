package stripe

type Event struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Data     *Data  `json:"data"`
	Livemode bool   `json:"livemode"`
	Created  int64  `json:"created"` // seconds elapsed since Unix epoch
}
type Data struct {
	Object PaymentIntent `json:"object"`
}

type PaymentIntent struct {
	ID           string            `json:"id"`
	Status       string            `json:"status"`
	Amount       int64             `json:"amount"`
	ClientSecret string            `json:"client_secret"`
	Currency     string            `json:"currency"`
	Metadata     map[string]string `json:"metadata"`
	Error        string            `json:"error,omitempty"`
}

type PaymentIntentResponse struct {
	ClientSecret string `json:"client_secret"`
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
	City       string `json:"city"`
	Country    string `json:"country"`
	Line1      string `json:"line1"`
	Line2      string `json:"line2,omitempty"`
	PostalCode string `json:"postal_code"`
	State      string `json:"state"`
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

// Description of the event (for example, `invoice.created` or `charge.refunded`).
type EventType string

// List of values that EventType can take
// Full list available at https://github.com/stripe/stripe-go
const (
	EventTypePaymentIntentAmountCapturableUpdated   EventType = "payment_intent.amount_capturable_updated"
	EventTypePaymentIntentCanceled                  EventType = "payment_intent.canceled"
	EventTypePaymentIntentCreated                   EventType = "payment_intent.created"
	EventTypePaymentIntentPartiallyFunded           EventType = "payment_intent.partially_funded"
	EventTypePaymentIntentPaymentFailed             EventType = "payment_intent.payment_failed"
	EventTypePaymentIntentProcessing                EventType = "payment_intent.processing"
	EventTypePaymentIntentRequiresAction            EventType = "payment_intent.requires_action"
	EventTypePaymentIntentSucceeded                 EventType = "payment_intent.succeeded"
	EventTypePaymentLinkCreated                     EventType = "payment_link.created"
	EventTypePaymentLinkUpdated                     EventType = "payment_link.updated"
	EventTypePaymentMethodAttached                  EventType = "payment_method.attached"
	EventTypePaymentMethodAutomaticallyUpdated      EventType = "payment_method.automatically_updated"
	EventTypePaymentMethodDetached                  EventType = "payment_method.detached"
	EventTypePaymentMethodUpdated                   EventType = "payment_method.updated"
	EventTypeRefundCreated                          EventType = "refund.created"
	EventTypeRefundFailed                           EventType = "refund.failed"
	EventTypeRefundUpdated                          EventType = "refund.updated"
	EventTypeTaxSettingsUpdated                     EventType = "tax.settings.updated"
	EventTypeTaxRateCreated                         EventType = "tax_rate.created"
	EventTypeTaxRateUpdated                         EventType = "tax_rate.updated"
	EventTypeTerminalReaderActionFailed             EventType = "terminal.reader.action_failed"
	EventTypeTerminalReaderActionSucceeded          EventType = "terminal.reader.action_succeeded"
	EventTypeBillingCreditBalanceTransactionCreated EventType = "billing.credit_balance_transaction.created"
	EventTypeBillingCreditGrantCreated              EventType = "billing.credit_grant.created"
	EventTypeBillingCreditGrantUpdated              EventType = "billing.credit_grant.updated"
	EventTypeBillingMeterCreated                    EventType = "billing.meter.created"
	EventTypeBillingMeterDeactivated                EventType = "billing.meter.deactivated"
	EventTypeBillingMeterReactivated                EventType = "billing.meter.reactivated"
	EventTypeBillingMeterUpdated                    EventType = "billing.meter.updated"
)
