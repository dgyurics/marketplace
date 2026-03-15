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

type PurchaseIntentRoutes struct {
	router
	service services.PurchaseIntentService
}

func NewPurchaseIntentRoutes(
	claimService services.PurchaseIntentService,
	router router) *PurchaseIntentRoutes {
	return &PurchaseIntentRoutes{
		router:  router,
		service: claimService,
	}
}

func (h *PurchaseIntentRoutes) CreatePurchaseIntent(w http.ResponseWriter, r *http.Request) {
	productID := mux.Vars(r)["id"]
	var purchaseIntent types.PurchaseIntent
	if err := json.NewDecoder(r.Body).Decode(&purchaseIntent); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request body")
		return
	}
	purchaseIntent.Product = types.Product{
		ID: productID,
	}

	// PickupNotes cannot be blank
	if purchaseIntent.PickupNotes == "" {
		u.RespondWithError(w, r, http.StatusBadRequest, "Pickup notes is required")
		return
	}

	// Create purchase intent
	err := h.service.CreatePurchaseIntent(r.Context(), &purchaseIntent)
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

func (h *PurchaseIntentRoutes) UpdatePurchaseIntent(w http.ResponseWriter, r *http.Request) {
	action := mux.Vars(r)["action"]
	purchaseIntent := types.PurchaseIntent{
		ID: mux.Vars(r)["id"],
	}

	// Update purchase intent status based on action
	switch action {
	case "pending":
		purchaseIntent.Status = types.PurchaseIntentPending
	case "accepted":
		purchaseIntent.Status = types.PurchaseIntentAccepted
	case "rejected":
		purchaseIntent.Status = types.PurchaseIntentRejected
	case "canceled":
		purchaseIntent.Status = types.PurchaseIntentCanceled
	case "completed":
		purchaseIntent.Status = types.PurchaseIntentCompleted
	default:
		u.RespondWithError(w, r, http.StatusBadRequest, "invalid action")
		return
	}

	// Update purchase intent status
	err := h.service.UpdatePurchaseIntent(r.Context(), &purchaseIntent)
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
}

func (h *PurchaseIntentRoutes) GetPurchaseIntentByID(w http.ResponseWriter, r *http.Request) {
	purchaseIntent, err := h.service.GetPurchaseIntentByID(r.Context(), mux.Vars(r)["id"])
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	u.RespondWithJSON(w, http.StatusOK, purchaseIntent)
}

func (h *PurchaseIntentRoutes) GetPurchaseIntentByProductID(w http.ResponseWriter, r *http.Request) {
	purchaseIntents, err := h.service.GetPurchaseIntentsByProductID(r.Context(), mux.Vars(r)["id"])
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	u.RespondWithJSON(w, http.StatusOK, purchaseIntents)
}

func (h *PurchaseIntentRoutes) GetPurchaseIntents(w http.ResponseWriter, r *http.Request) {
	purchaseIntents, err := h.service.GetPurchaseIntents(r.Context())
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	u.RespondWithJSON(w, http.StatusOK, purchaseIntents)
}

func (h *PurchaseIntentRoutes) RegisterRoutes() {
	h.muxRouter.Handle("/purchase-intents/items/{id}", h.secure(types.RoleMember)(h.limit(h.CreatePurchaseIntent, 5, time.Hour))).Methods(http.MethodPost)
	h.muxRouter.Handle("/purchase-intents/{id}/{action}", h.secure(types.RoleAdmin)(h.UpdatePurchaseIntent)).Methods(http.MethodPut)
	h.muxRouter.Handle("/purchase-intents/{id}", h.secure(types.RoleAdmin)(h.GetPurchaseIntentByID)).Methods(http.MethodGet)
	h.muxRouter.Handle("/purchase-intents/items/{id}", h.secure(types.RoleMember)(h.GetPurchaseIntentByProductID)).Methods(http.MethodGet)
	h.muxRouter.Handle("/purchase-intents", h.secure(types.RoleAdmin)(h.GetPurchaseIntents)).Methods(http.MethodGet)
}
