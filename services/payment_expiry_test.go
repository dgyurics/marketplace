package services

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgyurics/marketplace/types"
	"github.com/stretchr/testify/require"
)

func TestCancelPaymentIntent(t *testing.T) {
	for _, tc := range []struct {
		name          string
		cancelStatus  int
		currentStatus string
		wantError     bool
	}{
		{"canceled", http.StatusOK, "", false},
		{"already canceled", http.StatusBadRequest, "canceled", false},
		{"already succeeded", http.StatusBadRequest, "succeeded", true},
		{"processing", http.StatusBadRequest, "processing", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodPost:
					require.Equal(t, "/payment_intents/pi_test/cancel", r.URL.Path)
					w.WriteHeader(tc.cancelStatus)
				case http.MethodGet:
					require.Equal(t, "/payment_intents/pi_test", r.URL.Path)
					fmt.Fprintf(w, `{"status":%q}`, tc.currentStatus)
				default:
					t.Fatalf("unexpected method %s", r.Method)
				}
			}))
			defer server.Close()
			config := types.PaymentConfig{}
			config.Stripe.BaseURL = server.URL
			service := NewPaymentService(server.Client(), config, nil, nil, nil, nil)
			err := service.CancelPaymentIntent(context.Background(), "pi_test")
			if tc.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
