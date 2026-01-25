package routes

import (
	"encoding/json"
	"net/http"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	u "github.com/dgyurics/marketplace/utilities"
	"github.com/gorilla/mux"
)

type ClaimRoutes struct {
	router
	claimService services.ClaimService
}

func NewClaimRoutes(
	claimService services.ClaimService,
	router router) *ClaimRoutes {
	return &ClaimRoutes{
		router:       router,
		claimService: claimService,
	}
}

// ClaimItem allows a member to claim an item marked as free.
func (h *ClaimRoutes) ClaimItem(w http.ResponseWriter, r *http.Request) {
	var claim types.Claim
	if err := json.NewDecoder(r.Body).Decode(&claim); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}

	// PickupNotes cannot be blank
	if claim.PickupNotes == "" {
		u.RespondWithError(w, r, http.StatusBadRequest, "Pickup notes is required")
		return
	}

	claim.Product.ID = mux.Vars(r)["id"]
	err := h.claimService.ClaimItem(r.Context(), &claim)

	// Error handling
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusNotFound, err.Error())
		return
	}
	if err == types.ErrConstraintViolation {
		u.RespondWithError(w, r, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondSuccess(w)
}

func (h *ClaimRoutes) RegisterRoutes() {
	h.muxRouter.Handle("/claims/items/{id}", h.secure(types.RoleMember)(h.ClaimItem)).Methods(http.MethodPost)
}
