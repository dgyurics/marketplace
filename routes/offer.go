package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	u "github.com/dgyurics/marketplace/utilities"
	"github.com/gorilla/mux"
)

type OfferRoutes struct {
	router
	service services.OfferService
}

func NewOfferRoutes(
	offerService services.OfferService,
	router router) *OfferRoutes {
	return &OfferRoutes{
		router:  router,
		service: offerService,
	}
}

func (h *OfferRoutes) CreateOffer(w http.ResponseWriter, r *http.Request) {
	productID := mux.Vars(r)["id"]
	var offer types.Offer
	if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request body")
		return
	}
	offer.Product = types.Product{
		ID: productID,
	}

	// PickupNotes cannot be blank
	if offer.PickupNotes == "" {
		u.RespondWithError(w, r, http.StatusBadRequest, "Pickup notes is required")
		return
	}

	// Create offer
	err := h.service.CreateOffer(r.Context(), &offer)
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusNotFound, err.Error())
		return
	}
	if err == types.ErrConstraintViolation {
		u.RespondWithError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondSuccess(w)
}

func (h *OfferRoutes) UpdateOffer(w http.ResponseWriter, r *http.Request) {
	action := mux.Vars(r)["action"]
	offer := types.Offer{
		ID: mux.Vars(r)["id"],
	}

	// Update offer status based on action
	switch action {
	case "pending":
		offer.Status = types.OfferPending
	case "accepted":
		offer.Status = types.OfferAccepted
	case "rejected":
		offer.Status = types.OfferRejected
	case "canceled":
		offer.Status = types.OfferCanceled
	case "completed":
		offer.Status = types.OfferCompleted
	default:
		u.RespondWithError(w, r, http.StatusBadRequest, "invalid action")
		return
	}

	// Update offer status
	err := h.service.UpdateOffer(r.Context(), &offer)
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
}

func (h *OfferRoutes) GetOfferByID(w http.ResponseWriter, r *http.Request) {
	offer, err := h.service.GetOfferByID(r.Context(), mux.Vars(r)["id"])
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	u.RespondWithJSON(w, http.StatusOK, offer)
}

func (h *OfferRoutes) GetOfferByProductID(w http.ResponseWriter, r *http.Request) {
	offers, err := h.service.GetOffersByProductID(r.Context(), mux.Vars(r)["id"])
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	u.RespondWithJSON(w, http.StatusOK, offers)
}

func (h *OfferRoutes) GetOffers(w http.ResponseWriter, r *http.Request) {
	offers, err := h.service.GetOffers(r.Context())
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	u.RespondWithJSON(w, http.StatusOK, offers)
}

func (h *OfferRoutes) RegisterRoutes() {
	h.muxRouter.Handle("/offers/items/{id}", h.secure(types.RoleMember)(h.limit(h.CreateOffer, 5, time.Hour))).Methods(http.MethodPost)
	h.muxRouter.Handle("/offers/{id}/{action}", h.secure(types.RoleAdmin)(h.UpdateOffer)).Methods(http.MethodPut)
	h.muxRouter.Handle("/offers/{id}", h.secure(types.RoleAdmin)(h.GetOfferByID)).Methods(http.MethodGet)
	h.muxRouter.Handle("/offers/items/{id}", h.secure(types.RoleMember)(h.GetOfferByProductID)).Methods(http.MethodGet)
	h.muxRouter.Handle("/offers", h.secure(types.RoleAdmin)(h.GetOffers)).Methods(http.MethodGet)
}
