package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/services"
)

type OrderRoutes struct {
	router
	paymentService services.OrderService
}

func NewOrderRoutes(
	paymentService services.OrderService,
	router router) *OrderRoutes {
	return &OrderRoutes{
		router:         router,
		paymentService: paymentService,
	}
}

func (h *OrderRoutes) CreateOrder(w http.ResponseWriter, r *http.Request) {
	res, err := h.paymentService.CreateOrder(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if res.Error != "" {
		http.Error(w, res.Error, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *OrderRoutes) StripeWebhook(w http.ResponseWriter, r *http.Request) {
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
	if err := h.paymentService.VerifyWebhookEventSignature(body, signature); err != nil {
		http.Error(w, "Invalid request signature", http.StatusBadRequest)
		return
	}

	// save and process event
	if err := h.paymentService.ProcessWebhookEvent(r.Context(), event); err != nil {
		// TODO log error
		fmt.Printf("Error processing webhook event: %v\n", err)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *OrderRoutes) RegisterRoutes() {
	h.muxRouter.HandleFunc("/orders/events", h.StripeWebhook).Methods(http.MethodPost)
	h.muxRouter.Handle("/orders", h.secure(h.CreateOrder)).Methods(http.MethodPost)
}
