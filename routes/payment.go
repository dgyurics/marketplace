package routes

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/types/stripe"
	u "github.com/dgyurics/marketplace/utilities"
)

type PaymentRoutes struct {
	router
	config         types.AppMetadata
	paymentService services.PaymentService
	cartService    services.CartService
	taxService     services.TaxService
}

func NewPaymentRoutes(
	paymentService services.PaymentService,
	cartService services.CartService,
	taxService services.TaxService,
	config types.AppMetadata,
	router router) *PaymentRoutes {
	return &PaymentRoutes{
		router:         router,
		config:         config,
		paymentService: paymentService,
		cartService:    cartService,
		taxService:     taxService,
	}
}

func (h *PaymentRoutes) EventHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.paymentService.SignatureVerifier(body, r.Header.Get("Stripe-Signature")); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error verifying signature")
		return
	}

	var event stripe.Event
	if err := json.Unmarshal(body, &event); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}

	if !h.paymentService.SupportedEvent(r.Context(), event) {
		u.RespondSuccess(w)
		return
	}

	err = h.paymentService.EventHandler(r.Context(), event)
	if errors.Is(err, types.ErrNotFound) {
		u.RespondWithError(w, r, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondSuccess(w)
}

func (h *PaymentRoutes) PaymentMethods(w http.ResponseWriter, r *http.Request) {
	addr := types.Address{Country: r.URL.Query().Get("country")}
	if state := r.URL.Query().Get("state"); state != "" {
		addr.State = &state
	}

	items, err := h.cartService.GetItems(r.Context())
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	var subtotal int64
	for _, item := range items {
		if item.Product.Price == nil {
			u.RespondWithError(w, r, http.StatusBadRequest, "cart item has no price")
			return
		}
		subtotal += *item.Product.Price * int64(item.Quantity)
	}
	tax, err := h.taxService.EstimateTax(r.Context(), addr, items)
	if errors.Is(err, types.ErrNotFound) {
		u.RespondWithError(w, r, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	options := h.config.PaymentOptions
	options.Stripe = options.Stripe && subtotal+tax > 0
	u.RespondWithJSON(w, http.StatusOK, struct {
		TaxAmount      int64                `json:"tax_amount"`
		PaymentOptions types.PaymentOptions `json:"payment_options"`
	}{TaxAmount: tax, PaymentOptions: options})
}

func (h *PaymentRoutes) RegisterRoutes() {
	h.mux.Handle("POST /payment/events", http.HandlerFunc(h.EventHandler))
	h.mux.Handle("GET /payment/methods", h.secure(types.RoleGuest)(h.PaymentMethods))
}
