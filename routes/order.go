package routes

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/services"
	u "github.com/dgyurics/marketplace/utilities"
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
		u.RespondWithError(w, r, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate that the addressID is provided
	if requestBody.AddressID == "" {
		u.RespondWithError(w, r, http.StatusBadRequest, "Address ID is required")
		return
	}
	// Create the order
	res, err := h.orderService.CreateOrder(r.Context(), requestBody.AddressID)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	if res.Error != "" {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, res)
}

func (h *OrderRoutes) StripeWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	var event models.StripeWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// verify signature
	signature := r.Header.Get("Stripe-Signature")
	if err := h.orderService.VerifyWebhookEventSignature(body, signature); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "Invalid request signature")
		return
	}

	// verify expected data in event
	if event.Type == "" || event.Data == nil || event.Data.Object.ID == "" {
		u.RespondWithError(w, r, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// save and process event
	if err := h.orderService.ProcessWebhookEvent(r.Context(), event); err != nil {
		slog.Error("Error processing webhook event", "error", err)
	}

	u.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *OrderRoutes) GetOrders(w http.ResponseWriter, r *http.Request) {
	params := u.ParsePaginationParams(r, 1, 25)
	orders, err := h.orderService.GetOrders(r.Context(), params.Page, params.Limit)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, orders)
}

func (h *OrderRoutes) RegisterRoutes() {
	h.muxRouter.HandleFunc("/orders/events", h.StripeWebhook).Methods(http.MethodPost)
	h.muxRouter.Handle("/orders", h.secure(h.CreateOrder)).Methods(http.MethodPost)
	h.muxRouter.Handle("/orders", h.secure(h.GetOrders)).Methods(http.MethodGet)
}
