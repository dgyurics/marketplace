package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPaymentMethods(t *testing.T) {
	zero, positive := int64(0), int64(100)
	for _, tc := range []struct {
		name       string
		items      []types.CartItem
		configured types.PaymentOptions
		expected   types.PaymentOptions
	}{
		{"all free", []types.CartItem{{Product: types.Product{Price: &zero}, Quantity: 2}}, types.PaymentOptions{Stripe: true, OnDelivery: true}, types.PaymentOptions{OnDelivery: true}},
		{"mixed prices", []types.CartItem{{Product: types.Product{Price: &zero}, Quantity: 1}, {Product: types.Product{Price: &positive}, Quantity: 1}}, types.PaymentOptions{Stripe: true, OnDelivery: true}, types.PaymentOptions{Stripe: true, OnDelivery: true}},
		{"stripe disabled", []types.CartItem{{Product: types.Product{Price: &positive}, Quantity: 1}}, types.PaymentOptions{OnDelivery: true}, types.PaymentOptions{OnDelivery: true}},
		{"no alternative", []types.CartItem{{Product: types.Product{Price: &zero}, Quantity: 1}}, types.PaymentOptions{Stripe: true}, types.PaymentOptions{}},
		{"empty cart", []types.CartItem{}, types.PaymentOptions{Stripe: true}, types.PaymentOptions{}},
		{"missing price", []types.CartItem{{Product: types.Product{}, Quantity: 1}}, types.PaymentOptions{Stripe: true}, types.PaymentOptions{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cart := new(MockCartService)
			if tc.configured.Stripe {
				cart.On("GetItems", mock.Anything).Return(tc.items, nil).Once()
			}
			h := &PaymentRoutes{cartService: cart, config: types.AppMetadata{PaymentOptions: tc.configured}}
			w := httptest.NewRecorder()
			h.PaymentMethods(w, httptest.NewRequest(http.MethodGet, "/payment/methods", nil))
			require.Equal(t, http.StatusOK, w.Code)
			var result types.PaymentOptions
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
			require.Equal(t, tc.expected, result)
			cart.AssertExpectations(t)
			if !tc.configured.Stripe {
				cart.AssertNotCalled(t, "GetItems", mock.Anything)
			}
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
