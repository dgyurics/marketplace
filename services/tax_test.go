package services

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
	"github.com/stretchr/testify/require"
)

type taxRateStub struct {
	repositories.TaxRepository
	code *string
}

func (r *taxRateStub) GetTaxRates(_ context.Context, _ types.Address, code *string) (int32, error) {
	r.code = code
	return 1000, nil
}

func TestTaxEstimateUsesItemPriceAndTaxCode(t *testing.T) {
	for _, tc := range []struct {
		behavior               types.TaxBehavior
		price, tax, additional int64
	}{
		{types.TaxExclusive, 1000, 200, 200},
		{types.TaxInclusive, 1100, 200, 0},
		{types.TaxInclusive, 0, 0, 0},
	} {
		t.Run(fmt.Sprintf("%s-%d", tc.behavior, tc.price), func(t *testing.T) {
			repo := &taxRateStub{}
			service := NewTaxService(nil, repo, types.PaymentConfig{Tax: types.TaxConfig{Behavior: tc.behavior}})
			code := "txcd_test"
			result, err := service.EstimateTax(context.Background(), types.Address{}, []types.CartItem{{Product: types.Product{Price: utilities.Ptr(tc.price), TaxCode: &code}, Quantity: 2}})
			require.NoError(t, err)
			require.Equal(t, types.TaxEstimateResponse{TaxAmount: tc.tax, AdditionalTax: tc.additional}, result)
			require.Equal(t, &code, repo.code)
		})
	}
}

func TestCalculateTaxUsesOrderPriceAndSeparatesIncludedTax(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		require.Equal(t, "1600", r.Form.Get("line_items[0][amount]"))
		require.Equal(t, "2", r.Form.Get("line_items[0][quantity]"))
		require.Equal(t, "txcd_specific", r.Form.Get("line_items[0][tax_code]"))
		fmt.Fprint(w, `{"tax_amount_inclusive":120,"tax_amount_exclusive":30}`)
	}))
	defer server.Close()
	code := "txcd_specific"
	config := types.PaymentConfig{}
	config.Stripe.BaseURL = server.URL
	config.Tax.FallbackCode = "fallback"
	service := NewTaxService(server.Client(), nil, config)
	tax, err := service.CalculateTax(context.Background(), "ref", types.Address{}, []types.OrderItem{{Product: types.Product{ID: "p", TaxCode: &code}, UnitPrice: 800, Quantity: 2}})
	require.NoError(t, err)
	require.Equal(t, types.TaxEstimateResponse{TaxAmount: 150, AdditionalTax: 30}, tax)
}
