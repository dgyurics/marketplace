package routes

import (
	"encoding/json"
	"net/http"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	u "github.com/dgyurics/marketplace/utilities"
	"github.com/gorilla/mux"
)

type AddressRoutes struct {
	router
	userService services.AddressService
}

func NewAddressRoutes(addressService services.AddressService, router router) *AddressRoutes {
	return &AddressRoutes{
		router:      router,
		userService: addressService,
	}
}

func (h *AddressRoutes) GetAddresses(w http.ResponseWriter, r *http.Request) {
	addresses, err := h.userService.GetAddresses(r.Context())
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, addresses)
}

func (h *AddressRoutes) CreateAddress(w http.ResponseWriter, r *http.Request) {
	var address types.Address
	if err := json.NewDecoder(r.Body).Decode(&address); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}

	if err := h.userService.CreateAddress(r.Context(), &address); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusCreated, address)
}

func (h *AddressRoutes) UpdateAddress(w http.ResponseWriter, r *http.Request) {
	var address types.Address
	if err := json.NewDecoder(r.Body).Decode(&address); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}

	if err := h.userService.UpdateAddress(r.Context(), &address); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, address)
}

func (h *AddressRoutes) RemoveAddress(w http.ResponseWriter, r *http.Request) {
	addressID := mux.Vars(r)["id"]
	if err := h.userService.RemoveAddress(r.Context(), addressID); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondSuccess(w)
}

func (h *AddressRoutes) RegisterRoutes() {
	h.muxRouter.Handle("/addresses", h.secure(h.CreateAddress)).Methods(http.MethodPost)
	h.muxRouter.Handle("/addresses", h.secure(h.UpdateAddress)).Methods(http.MethodPut)
	h.muxRouter.Handle("/addresses", h.secure(h.GetAddresses)).Methods(http.MethodGet)
	h.muxRouter.Handle("/addresses/{id}", h.secure(h.RemoveAddress)).Methods(http.MethodDelete)
}
