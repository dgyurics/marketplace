package routes

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/services"
)

type OrderRoutes struct {
	router
	orderService services.OrderService
}

func NewOrderRoutes(
	orderService services.OrderService,
	router router) *OrderRoutes {
	return &OrderRoutes{
		router:       router,
		orderService: orderService,
	}
}

func (h *OrderRoutes) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Parse the JSON body to extract addressID
	var requestBody struct {
		AddressID string `json:"address_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate that the addressID is provided
	if requestBody.AddressID == "" {
		http.Error(w, "Address ID is required", http.StatusBadRequest)
		return
	}
	// Create the order
	res, err := h.orderService.CreateOrder(r.Context(), requestBody.AddressID)
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
	if err := h.orderService.VerifyWebhookEventSignature(body, signature); err != nil {
		http.Error(w, "Invalid request signature", http.StatusBadRequest)
		return
	}

	// verify expected data in event
	if event.Type == "" || event.Data == nil || event.Data.Object.ID == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// save and process event
	if err := h.orderService.ProcessWebhookEvent(r.Context(), event); err != nil {
		slog.Error("Error processing webhook event", "error", err)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *OrderRoutes) GetOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.orderService.GetOrders(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)
}

func (h *OrderRoutes) RegisterRoutes() {
	h.muxRouter.HandleFunc("/orders/events", h.StripeWebhook).Methods(http.MethodPost)
	h.muxRouter.Handle("/orders", h.secure(h.CreateOrder)).Methods(http.MethodPost)
	h.muxRouter.Handle("/orders", h.secure(h.GetOrders)).Methods(http.MethodGet)
}
