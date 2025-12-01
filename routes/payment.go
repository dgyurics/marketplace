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
	paymentService services.PaymentService
}

func NewPaymentRoutes(
	paymentService services.PaymentService,
	router router) *PaymentRoutes {
	return &PaymentRoutes{
		router:         router,
		paymentService: paymentService,
	}
}

func (h *PaymentRoutes) EventHandler(w http.ResponseWriter, r *http.Request) {
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

	if !h.paymentService.SupportedEvent(r.Context(), event) {
		u.RespondSuccess(w)
		return
	}

	if err := h.paymentService.SignatureVerifier(body, r.Header.Get("Stripe-Signature")); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error verifying signature")
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

func (h *PaymentRoutes) RegisterRoutes() {
	h.muxRouter.HandleFunc("/payment/events", h.EventHandler).Methods(http.MethodPost)
}
