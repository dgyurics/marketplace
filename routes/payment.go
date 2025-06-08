package routes

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/dgyurics/marketplace/services"
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

	if err := h.paymentService.SignatureVerifier(body, r.Header.Get("Stripe-Signature")); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error verifying signature")
		return
	}

	if err := verifyEventData(event); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.paymentService.EventHandler(r.Context(), event); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondSuccess(w)
}

// verifyEventData verifies all the expected fields are present in the event data.
func verifyEventData(event stripe.Event) error {
	if event.Type == "" {
		slog.Error("event.type missing")
		return errors.New("event.type is required")
	}
	if event.Data == nil {
		slog.Error("event.data missing")
		return errors.New("event.data is required")
	}
	if event.Data.Object.ID == "" {
		slog.Error("event.data.object.id missing")
		return errors.New("event.data.object.id is required")
	}
	return nil
}

func (h *PaymentRoutes) RegisterRoutes() {
	h.muxRouter.HandleFunc("/payment/events", h.EventHandler).Methods(http.MethodPost)
}
