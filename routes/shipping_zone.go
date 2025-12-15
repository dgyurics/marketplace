package routes

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
	u "github.com/dgyurics/marketplace/utilities"
	"github.com/gorilla/mux"
)

type ShippingZoneRoutes struct {
	router
	shippingZoneService services.ShippingZoneService
}

func NewShippingZoneRoutes(
	shippingZoneService services.ShippingZoneService,
	router router) *ShippingZoneRoutes {
	return &ShippingZoneRoutes{
		router:              router,
		shippingZoneService: shippingZoneService,
	}
}

func (h *ShippingZoneRoutes) CreateShippingZone(w http.ResponseWriter, r *http.Request) {
	var zone types.ShippingZone
	if err := json.NewDecoder(r.Body).Decode(&zone); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}

	if err := validateShippingZone(zone); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	err := h.shippingZoneService.AddShippingZone(r.Context(), &zone)
	if err == types.ErrUniqueConstraintViolation {
		u.RespondWithError(w, r, http.StatusConflict, err.Error())
		return
	}
	if err == types.ErrConstraintViolation {
		u.RespondWithError(w, r, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondSuccess(w)
}

func (h *ShippingZoneRoutes) ListShippingZones(w http.ResponseWriter, r *http.Request) {
	shippingZones, err := h.shippingZoneService.GetShippingZones(r.Context())
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, shippingZones)
}

func (h *ShippingZoneRoutes) RemoveShippingZone(w http.ResponseWriter, r *http.Request) {
	zoneID := mux.Vars(r)["id"]
	if err := h.shippingZoneService.RemoveShippingZone(r.Context(), zoneID); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondSuccess(w)
}

func (h *ShippingZoneRoutes) CreateExcludedShippingZone(w http.ResponseWriter, r *http.Request) {
	var zone types.ExcludedShippingZone
	if err := json.NewDecoder(r.Body).Decode(&zone); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}

	if err := validateExcludedShippingZone(zone); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	err := h.shippingZoneService.AddExcludedShippingZone(r.Context(), &zone)
	if err == types.ErrUniqueConstraintViolation {
		u.RespondWithError(w, r, http.StatusConflict, err.Error())
		return
	}
	if err == types.ErrConstraintViolation {
		u.RespondWithError(w, r, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondSuccess(w)
}

func (h *ShippingZoneRoutes) ListExcludedShippingZones(w http.ResponseWriter, r *http.Request) {
	excludedZones, err := h.shippingZoneService.GetExcludedShippingZones(r.Context())
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, excludedZones)
}

func (h *ShippingZoneRoutes) RemoveExcludedShippingZone(w http.ResponseWriter, r *http.Request) {
	zoneID := mux.Vars(r)["id"]
	if err := h.shippingZoneService.RemoveExcludedShippingZone(r.Context(), zoneID); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondSuccess(w)
}

func validateShippingZone(zone types.ShippingZone) error {
	if zone.Country != utilities.Locale.CountryCode {
		return errors.New("invalid country code")
	}

	if zone.State != nil && u.ValidateState(zone.Country, *zone.State) != nil {
		return errors.New("invalid state")
	}

	if zone.PostalCode != nil && u.ValidatePostalCode(zone.Country, *zone.PostalCode) != nil {
		return errors.New("invalid postal code")
	}

	return nil
}

func validateExcludedShippingZone(zone types.ExcludedShippingZone) error {
	if zone.Country != utilities.Locale.CountryCode {
		return errors.New("invalid country code")
	}

	if u.ValidatePostalCode(zone.Country, zone.PostalCode) != nil {
		return errors.New("invalid postal code")
	}

	return nil
}

func (h *ShippingZoneRoutes) RegisterRoutes() {
	h.muxRouter.Handle("/shipping-zones", h.secureAdmin(h.CreateShippingZone)).Methods("POST")
	h.muxRouter.Handle("/shipping-zones", h.secureAdmin(h.ListShippingZones)).Methods("GET")
	h.muxRouter.Handle("/shipping-zones/{id}", h.secureAdmin(h.RemoveShippingZone)).Methods("DELETE")

	h.muxRouter.Handle("/shipping-zones/excluded", h.secureAdmin(h.CreateExcludedShippingZone)).Methods("POST")
	h.muxRouter.Handle("/shipping-zones/excluded", h.secureAdmin(h.ListExcludedShippingZones)).Methods("GET")
	h.muxRouter.Handle("/shipping-zones/excluded/{id}", h.secureAdmin(h.RemoveExcludedShippingZone)).Methods("DELETE")
}
