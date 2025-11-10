package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
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

	if err := h.validateAddress(address); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, err.Error())
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
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}

	err := h.addressService.UpdateAddress(r.Context(), address)
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

func (h *AddressRoutes) RemoveAddress(w http.ResponseWriter, r *http.Request) {
	addressID := mux.Vars(r)["id"]
	if err := h.addressService.RemoveAddress(r.Context(), addressID); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondSuccess(w)
}

func (h *AddressRoutes) validateAddress(address types.Address) error {
	if !strings.EqualFold(address.Country, h.config.Country) {
		return errors.New("invalid country code")
	}

	if !u.ValidateState(address.Country, u.StringValue(address.State, "")) {
		return errors.New("invalid state")
	}

	if !u.ValidatePostalCode(address.Country, address.PostalCode) {
		return errors.New("invalid postal code format")
	}

	if address.Line1 == "" {
		return errors.New("missing address line1")
	}

	if address.City == "" {
		return errors.New("missing address city")
	}

	return nil
}

func (h *AddressRoutes) RegisterRoutes() {
	h.muxRouter.Handle("/addresses", h.secure(h.limit(h.CreateAddress, 3, time.Hour))).Methods(http.MethodPost)
	h.muxRouter.Handle("/addresses", h.secure(h.limit(h.UpdateAddress, 10, time.Hour))).Methods(http.MethodPut)
	h.muxRouter.Handle("/addresses/{id}", h.secure(h.RemoveAddress)).Methods(http.MethodDelete)
}
