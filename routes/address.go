package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
	u "github.com/dgyurics/marketplace/utilities"
	"github.com/gorilla/mux"
)

type AddressRoutes struct {
	router
	addressService      services.AddressService
	shippingZoneService services.ShippingZoneService
}

func NewAddressRoutes(
	addressService services.AddressService,
	shippingZoneService services.ShippingZoneService,
	router router) *AddressRoutes {
	return &AddressRoutes{
		router:              router,
		addressService:      addressService,
		shippingZoneService: shippingZoneService,
	}
}

func (h *AddressRoutes) CreateAddress(w http.ResponseWriter, r *http.Request) {
	var address types.Address
	if err := json.NewDecoder(r.Body).Decode(&address); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}

	if err := validateAddress(address); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// check if the address provided can be shipped to
	isShippable, err := h.shippingZoneService.IsShippable(r.Context(), &address)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	if !isShippable {
		u.RespondWithError(w, r, http.StatusUnprocessableEntity, "cannot ship to the provided address")
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

	if address.ID == "" {
		u.RespondWithError(w, r, http.StatusBadRequest, "missing address ID")
		return
	}

	if err := validateAddress(address); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// check if the address provided can be shipped to
	isShippable, err := h.shippingZoneService.IsShippable(r.Context(), &address)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	if !isShippable {
		u.RespondWithError(w, r, http.StatusUnprocessableEntity, "cannot ship to the provided address")
		return
	}

	err = h.addressService.UpdateAddress(r.Context(), &address)
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, address)
}

func (h *AddressRoutes) RemoveAddress(w http.ResponseWriter, r *http.Request) {
	addressID := mux.Vars(r)["id"]
	if err := h.addressService.RemoveAddress(r.Context(), addressID); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondSuccess(w)
}

func validateAddress(address types.Address) error {
	// case-sensitive
	if address.Country != utilities.Locale.CountryCode {
		return errors.New("invalid country code")
	}

	if err := u.ValidateState(address.Country, u.StringValue(address.State, "")); err != nil {
		return err
	}

	if err := u.ValidatePostalCode(address.Country, address.PostalCode); err != nil {
		return err
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
	h.muxRouter.Handle("/addresses", h.secure(types.RoleGuest)(h.limit(h.CreateAddress, 3, time.Hour))).Methods(http.MethodPost)
	h.muxRouter.Handle("/addresses", h.secure(types.RoleGuest)(h.limit(h.UpdateAddress, 10, time.Hour))).Methods(http.MethodPut)
	h.muxRouter.Handle("/addresses/{id}", h.secure(types.RoleGuest)(h.limit(h.RemoveAddress, 3, time.Hour))).Methods(http.MethodDelete)
}
