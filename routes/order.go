package routes

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/types/stripe"
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
		Currency  string `json:"currency"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request body")
		return
	}

	// Validate addressID is provided
	if requestBody.AddressID == "" {
		u.RespondWithError(w, r, http.StatusBadRequest, "Address ID is required")
		return
	}

	// Validate that the currency is provided
	if requestBody.Currency == "" {
		u.RespondWithError(w, r, http.StatusBadRequest, "Currency is required")
		return
	}

	// Create the order
	order := types.Order{
		Address:  &types.Address{ID: requestBody.AddressID},
		Currency: requestBody.Currency,
	}
	if err := h.orderService.CreateOrder(r.Context(), &order); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, order)
}

func (h *OrderRoutes) StripeWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	var event stripe.Event
	if err := json.Unmarshal(body, &event); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request body")
		return
	}

	// verify signature
	signature := r.Header.Get("Stripe-Signature")
	if err := h.orderService.VerifyStripeEventSignature(body, signature); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "Invalid request signature")
		return
	}

	// verify expected data in event
	if event.Type == "" {
		u.RespondWithError(w, r, http.StatusBadRequest, "event.type is required")
		return
	}

	if event.Data == nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "event.data is required")
		return
	}

	if event.Data.Object.ID == "" {
		u.RespondWithError(w, r, http.StatusBadRequest, "event.data.object.id is required")
		return
	}

	// save and process event
	if err := h.orderService.ProcessStripeEvent(r.Context(), event); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
	}

	u.RespondSuccess(w)
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
