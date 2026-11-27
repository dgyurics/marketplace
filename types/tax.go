package types

// TaxAmount includes all tax; AdditionalTax is the part not already in item prices.
type TaxEstimateResponse struct {
	TaxAmount     int64 `json:"tax_amount"`
	AdditionalTax int64 `json:"additional_tax"`
}
