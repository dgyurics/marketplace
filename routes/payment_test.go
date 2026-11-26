package routes

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type estimateTaxStub struct {
	services.TaxService
	tax int64
	err error
}

func (s estimateTaxStub) EstimateTax(_ context.Context, addr types.Address, _ []types.CartItem) (int64, error) {
	if addr.Country != "US" || addr.State == nil || *addr.State != "CA" {
		return 0, errors.New("incorrect address")
	}
	return s.tax, s.err
}

func TestPaymentMethods(t *testing.T) {
	for _, tc := range []struct {
		name       string
		price      int64
		tax        int64
		configured types.PaymentOptions
		expected   types.PaymentOptions
	}{
		{"zero total", 0, 0, types.PaymentOptions{Stripe: true, OnDelivery: true}, types.PaymentOptions{OnDelivery: true}},
		{"positive total", 100, 10, types.PaymentOptions{Stripe: true, OnDelivery: true}, types.PaymentOptions{Stripe: true, OnDelivery: true}},
		{"positive tax", 0, 10, types.PaymentOptions{Stripe: true}, types.PaymentOptions{Stripe: true}},
		{"stripe disabled", 100, 10, types.PaymentOptions{OnDelivery: true}, types.PaymentOptions{OnDelivery: true}},
		{"no alternative", 0, 0, types.PaymentOptions{Stripe: true}, types.PaymentOptions{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cart := new(MockCartService)
			items := []types.CartItem{{Product: types.Product{Price: &tc.price}, Quantity: 2}}
			cart.On("GetItems", mock.Anything).Return(items, nil).Once()
			h := &PaymentRoutes{cartService: cart, taxService: estimateTaxStub{tax: tc.tax}, config: types.AppMetadata{PaymentOptions: tc.configured}}
			w := httptest.NewRecorder()
			h.PaymentMethods(w, httptest.NewRequest(http.MethodGet, "/payment/methods?country=US&state=CA", nil))
			require.Equal(t, http.StatusOK, w.Code)
			var result struct {
				TaxAmount      int64                `json:"tax_amount"`
				PaymentOptions types.PaymentOptions `json:"payment_options"`
			}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
			require.Equal(t, tc.tax, result.TaxAmount)
			require.Equal(t, tc.expected, result.PaymentOptions)
			cart.AssertExpectations(t)
		})
	}
}

type calculateTaxStub struct{ services.TaxService }

func (calculateTaxStub) CalculateTax(context.Context, string, types.Address, []types.OrderItem) (int64, error) {
	return 0, nil
}

func TestStripeCheckoutRejectsZeroTotal(t *testing.T) {
	zero := int64(0)
	status := types.OrderDraft
	orders := new(MockOrderService)
	orders.On("GetOrderByIDAndUser", mock.Anything, "order-1").Return(types.Order{
		Amount: &zero, ShippingAmount: &zero, Status: &status,
	}, nil).Once()
	h := &OrderRoutes{config: types.PaymentOptions{Stripe: true}, orderService: orders, taxService: calculateTaxStub{}}
	request := httptest.NewRequest(http.MethodPost, "/orders/order-1/stripe/intent", nil)
	request.SetPathValue("id", "order-1")
	w := httptest.NewRecorder()
	h.CheckoutWithStripe(w, request)
	require.Equal(t, http.StatusBadRequest, w.Code)
	orders.AssertExpectations(t)
	orders.AssertNotCalled(t, "UpdateOrder", mock.Anything, mock.Anything)
}
