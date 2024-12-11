package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/services"
	"github.com/gorilla/mux"
)

type PaymentHandler interface {
	StripeWebhook(w http.ResponseWriter, r *http.Request)
}

type paymentHandler struct {
	paymentService services.PaymentService
	router         *mux.Router
}

func RegisterPaymentHandler(
	paymentService services.PaymentService,
	router *mux.Router,
) {
	handler := &paymentHandler{
		paymentService: paymentService,
		router:         router,
	}
	handler.RegisterRoutes()
}

func (h *paymentHandler) StripeWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536) // limit request body to 64KB // TODO do this globally
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	var event models.StripeWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// verify signature
	signature := r.Header.Get("Stripe-Signature")
	if err := h.paymentService.VerifyWebhookSignature(body, signature); err != nil {
		http.Error(w, "Invalid request signature", http.StatusBadRequest)
		return
	}

	// TODO store and process event
	// h.paymentService.ProcessWebhookEvent(r.Context(), event)

	w.WriteHeader(http.StatusOK)
}

func (h *paymentHandler) RegisterRoutes() {
	h.router.HandleFunc("/stripe/webhook", h.StripeWebhook).Methods(http.MethodPost)
}
