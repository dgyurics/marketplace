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

type PaymentRoutes struct {
	router
	config         types.AppMetadata
	paymentService services.PaymentService
}

func NewPaymentRoutes(
	paymentService services.PaymentService,
	config types.AppMetadata,
	router router) *PaymentRoutes {
	return &PaymentRoutes{
		router:         router,
		config:         config,
		paymentService: paymentService,
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
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request body")
		return
	}

	if !h.paymentService.SupportedEvent(r.Context(), event) {
		u.RespondSuccess(w)
		return
	}

	err = h.paymentService.EventHandler(r.Context(), event)
	if err == types.ErrNotFound {
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
	payOpts := h.config.PaymentOptions
	usr, ok := r.Context().Value(services.UserKey).(*types.User)
	if !ok || !usr.HasMinimumRole(types.RoleMember) {
		payOpts.PayOnDelivery = false
	}

	u.RespondWithJSON(w, 200, payOpts)
}

func (h *PaymentRoutes) RegisterRoutes() {
	h.mux.Handle("POST /payment/events", http.HandlerFunc(h.EventHandler))
	h.mux.Handle("GET /payment/methods", h.authMiddleware.OptionalAuth(http.HandlerFunc(h.PaymentMethods)))
}
