package routes

import (
	"net/http"
	"time"

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

// CreateOrder creates a new order using the provided shipping details ID.
func (h *OrderRoutes) CreateOrder(w http.ResponseWriter, r *http.Request) {
	shippingID := r.URL.Query().Get("shipping_id")
	if shippingID == "" {
		u.RespondWithError(w, r, http.StatusBadRequest, "missing shipping_id query parameter")
		return
	}

	ord, err := h.orderService.CreateOrder(r.Context(), shippingID)
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

func (h *OrderRoutes) GetOrderOwner(w http.ResponseWriter, r *http.Request) {
	order, err := h.orderService.GetOrderByIDAndUser(r.Context(), mux.Vars(r)["id"])
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusNotFound, "order not found")
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	u.RespondWithJSON(w, http.StatusOK, order)
}

func (h *OrderRoutes) GetOrderPublic(w http.ResponseWriter, r *http.Request) {
	order, err := h.orderService.GetOrderByIDPublic(r.Context(), mux.Vars(r)["id"])
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusNotFound, "order not found")
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	u.RespondWithJSON(w, http.StatusOK, order)
}

func (h *OrderRoutes) GetOrderAdmin(w http.ResponseWriter, r *http.Request) {
	order, err := h.orderService.GetOrderByID(r.Context(), mux.Vars(r)["id"])
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusNotFound, "order not found")
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	u.RespondWithJSON(w, http.StatusOK, order)
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

// Confirm finalizes an order by calculating actual tax and generating a payment intent.
func (h *OrderRoutes) Confirm(w http.ResponseWriter, r *http.Request) {
	order, err := h.orderService.GetOrderByIDAndUser(r.Context(), mux.Vars(r)["id"])
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusNotFound, "order not found")
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// TODO calculate totals using order_items table

	tax, err := h.taxService.CalculateTax(r.Context(), order.ID, order.Address, order.Items)
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
		ID:          order.ID,
		TaxAmount:   &tax,
		TotalAmount: &totalAmount,
	}
	_, err = h.orderService.UpdateOrder(r.Context(), params)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Clear user cart after order confirmation
	if err := h.cartService.Clear(r.Context()); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, stripe.PaymentIntentResponse{ClientSecret: pi.ClientSecret})
}

// Order flow
// 1. POST /orders?shipping_id=123     → Create empty order with address
// 2. POST /orders/{id}/confirm        → Calculate tax + totals, copy cart_items -> order_items, return newly created payment intent
func (h *OrderRoutes) RegisterRoutes() {
	// Order placement
	h.muxRouter.Handle("/orders", h.secure(h.limit(h.CreateOrder, 5, time.Hour))).Methods(http.MethodPost)
	h.muxRouter.Handle("/orders/{id}/confirm", h.secure(h.limit(h.Confirm, 5, time.Hour))).Methods(http.MethodPost)
	// Order review
	h.muxRouter.HandleFunc("/orders/{id}/public", h.GetOrderPublic).Methods(http.MethodPost)
	h.muxRouter.Handle("/orders/{id}/owner", h.secure(h.GetOrderOwner)).Methods(http.MethodPost)
	h.muxRouter.Handle("/orders/{id}/admin", h.secureAdmin(h.GetOrderAdmin)).Methods(http.MethodPost)
	h.muxRouter.Handle("/orders", h.secureAdmin(h.GetOrders)).Methods(http.MethodGet)
}
