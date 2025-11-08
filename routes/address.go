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

type AddressRoutes struct {
	router
	addressService services.AddressService
	config         types.LocaleConfig
}

func NewAddressRoutes(
	addressService services.AddressService,
	config types.LocaleConfig,
	router router) *AddressRoutes {
	return &AddressRoutes{
		router:         router,
		addressService: addressService,
		config:         config,
	}
}

func (h *AddressRoutes) CreateAddress(w http.ResponseWriter, r *http.Request) {
	var address types.Address
	if err := json.NewDecoder(r.Body).Decode(&address); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}

	if address.Line1 == "" || address.City == "" || address.State == "" || address.PostalCode == "" {
		u.RespondWithError(w, r, http.StatusBadRequest, "missing required fields for address")
		return
	}

	if !u.PostalCodePatterns[h.config.Country].MatchString(address.PostalCode) {
		u.RespondWithError(w, r, http.StatusBadRequest, "invalid postal code format")
		return
	}

	if err := h.addressService.CreateAddress(r.Context(), &address); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusCreated, address)
}

func (h *AddressRoutes) UpdateAddress(w http.ResponseWriter, r *http.Request) {
	var address types.Address
	if err := json.NewDecoder(r.Body).Decode(&address); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request body")
		return
	}
	err := h.addressService.UpdateAddress(r.Context(), address)
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusNotFound, "Address not found")
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	u.RespondSuccess(w)
}

func (h *AddressRoutes) RemoveAddress(w http.ResponseWriter, r *http.Request) {
	addressID := mux.Vars(r)["id"]
	if err := h.addressService.RemoveAddress(r.Context(), addressID); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondSuccess(w)
}

func (h *AddressRoutes) RegisterRoutes() {
	h.muxRouter.Handle("/addresses", h.secure(h.limit(h.CreateAddress, 3, time.Hour))).Methods(http.MethodPost)
	h.muxRouter.Handle("/addresses", h.secure(h.limit(h.UpdateAddress, 10, time.Hour))).Methods(http.MethodPut)
	h.muxRouter.Handle("/addresses/{id}", h.secure(h.RemoveAddress)).Methods(http.MethodDelete)
}
