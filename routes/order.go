package routes

import (
	"encoding/json"
	"net/http"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/types/stripe"
	u "github.com/dgyurics/marketplace/utilities"
	"github.com/gorilla/mux"
)

type OrderRoutes struct {
	router
	orderService   services.OrderService
	taxService     services.TaxService
	paymentService services.PaymentService
	cartService    services.CartService
}

func NewOrderRoutes(
	orderService services.OrderService,
	taxService services.TaxService,
	paymentService services.PaymentService,
	cartService services.CartService,
	router router) *OrderRoutes {
	return &OrderRoutes{
		router:         router,
		orderService:   orderService,
		taxService:     taxService,
		paymentService: paymentService,
		cartService:    cartService,
	}
}

func (h *OrderRoutes) CreateOrder(w http.ResponseWriter, r *http.Request) {
	ord, err := h.orderService.GetPendingOrderForUser(r.Context())
	if err == nil {
		// Pending order exists, return it
		u.RespondWithJSON(w, http.StatusOK, ord)
		return
	}

	if err != types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// No pending order, create a new one
	ord, err = h.orderService.CreateOrder(r.Context())
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, ord)
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

func (h *OrderRoutes) EstimateTax(w http.ResponseWriter, r *http.Request) {
	orderID := mux.Vars(r)["id"]
	order, err := h.orderService.GetOrderForUser(r.Context(), orderID)
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusNotFound, "order not found")
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	if order.Address == nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "order address is required for tax estimate")
		return
	}

	taxEstimate, err := h.taxService.EstimateTax(r.Context(), *order.Address, order.Items)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, types.TaxEstimateResponse{TaxAmount: taxEstimate})
}

func (h *OrderRoutes) Update(w http.ResponseWriter, r *http.Request) {
	orderID := mux.Vars(r)["id"]
	params := types.OrderParams{
		ID: orderID,
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request body")
		return
	}
	ord, err := h.orderService.UpdateOrder(r.Context(), params)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	u.RespondWithJSON(w, http.StatusOK, ord)
}

// Confirm finalizes an order by calculating actual tax and generating a payment intent.
func (h *OrderRoutes) Confirm(w http.ResponseWriter, r *http.Request) {
	orderID := mux.Vars(r)["id"]
	order, err := h.orderService.GetOrderForUser(r.Context(), orderID)
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusNotFound, "order not found")
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	tax, err := h.taxService.CalculateTax(r.Context(), order.ID, *order.Address, order.Items)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	pi, err := h.paymentService.CreatePaymentIntent(r.Context(), order.ID, order.TotalAmount+tax)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	totalAmount := order.TotalAmount + tax
	params := types.OrderParams{
		ID:                  orderID,
		TaxAmount:           &tax,
		TotalAmount:         &totalAmount,
		StripePaymentIntent: &pi,
	}
	_, err = h.orderService.UpdateOrder(r.Context(), params)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Clear user cart after order confirmation
	if err := h.cartService.ClearCart(r.Context()); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, stripe.PaymentIntentResponse{ClientSecret: pi.ClientSecret})
}

func (h *OrderRoutes) RegisterRoutes() {
	h.muxRouter.Handle("/orders", h.secure(h.CreateOrder)).Methods(http.MethodPost)
	h.muxRouter.Handle("/orders/{id}", h.secure(h.Update)).Methods(http.MethodPatch)
	h.muxRouter.Handle("/orders/{id}/confirm", h.secure(h.Confirm)).Methods(http.MethodPost)
	h.muxRouter.Handle("/orders", h.secureAdmin(h.GetOrders)).Methods(http.MethodGet)
	h.muxRouter.Handle("/orders/{id}/tax-estimate", h.secure(h.EstimateTax)).Methods(http.MethodGet)
}
