package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/types/stripe"
	"github.com/dgyurics/marketplace/utilities"
)

type TaxService interface {
	CalculateTax(ctx context.Context, refID string, shippingAddress types.Address, items []types.OrderItem) (int64, error)
	EstimateTax(ctx context.Context, shippingAddress types.Address, items []types.OrderItem) (int64, error)
}

type taxService struct {
	HttpClient   utilities.HTTPClient
	repo         repositories.TaxRepository
	stripeConfig types.StripeConfig
	localeConfig types.LocaleConfig
}

func NewTaxService(
	repo repositories.TaxRepository,
	strpConfig types.StripeConfig,
	locConfig types.LocaleConfig,
	HttpClient utilities.HTTPClient,
) TaxService {
	return &taxService{
		repo:         repo,
		stripeConfig: strpConfig,
		localeConfig: locConfig,
		HttpClient:   HttpClient,
	}
}

func (s *taxService) CalculateTax(ctx context.Context, refID string, address types.Address, items []types.OrderItem) (int64, error) {
	form := url.Values{}
	form.Set("currency", s.localeConfig.Currency)

	// Customer Address
	form.Set("customer_details[address_source]", "shipping")
	form.Set("customer_details[address][country]", address.Country) // fixme this is blank
	form.Set("customer_details[address][city]", address.City)
	form.Set("customer_details[address][line1]", address.Line1)
	if line2 := address.Line2; line2 != nil && *line2 != "" {
		form.Set("customer_details[address][line2]", *line2)
	}
	form.Set("customer_details[address][state]", address.State)
	form.Set("customer_details[address][postal_code]", address.PostalCode)

	if len(items) == 0 {
		slog.Error("CalculateTax called with no items", "refID", refID)
		return 0, fmt.Errorf("no items provided for tax calculation")
	}

	// Line Items
	for i, item := range items {
		itmQty := int64(item.Quantity)
		taxCode := item.Product.TaxCode
		if taxCode == "" {
			taxCode = s.localeConfig.FallbackTaxCode
		}
		form.Set(fmt.Sprintf("line_items[%d][amount]", i), strconv.FormatInt(item.UnitPrice*itmQty, 10))
		form.Set(fmt.Sprintf("line_items[%d][quantity]", i), strconv.FormatInt(itmQty, 10))
		form.Set(fmt.Sprintf("line_items[%d][tax_behavior]", i), string(s.localeConfig.TaxBehavior))
		form.Set(fmt.Sprintf("line_items[%d][tax_code]", i), taxCode)
		form.Set(fmt.Sprintf("line_items[%d][reference]", i), fmt.Sprintf("%s:%s", refID, item.Product.ID))
	}

	url := fmt.Sprintf("%s/tax/calculations", s.stripeConfig.BaseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.stripeConfig.SecretKey))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Idempotency-Key", fmt.Sprintf("tax-calculation-%s", refID))

	resp, err := s.HttpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("stripe tax calculation failed (status %d): %s", resp.StatusCode, string(body))
	}

	var tax stripe.TaxCalculationResponse
	if err := json.NewDecoder(resp.Body).Decode(&tax); err != nil {
		return 0, err
	}

	slog.Debug("Stripe Tax Calculation Response", "response", tax)

	return tax.TaxAmountInclusive + tax.TaxAmountExclusive, nil
}

// EstimateTax estimates the tax using a combination of country (required), state (optional), and tax code (optional).
func (s *taxService) EstimateTax(ctx context.Context, shippingAddress types.Address, items []types.OrderItem) (int64, error) {
	var totalTax int64
	for _, item := range items {
		rate, err := s.repo.GetTaxRates(ctx, shippingAddress, item.Product.TaxCode)
		if err != nil {
			return 0, err
		}
		totalTax += int64(item.Quantity) * item.UnitPrice * int64(rate) / 10_000
	}
	return totalTax, nil
}
