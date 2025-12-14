package routes

import (
	"encoding/json"
	"net/http"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
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
	if err := h.shippingZoneService.AddShippingZone(r.Context(), &zone); err != nil {
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
	if err := h.shippingZoneService.AddExcludedShippingZone(r.Context(), &zone); err != nil {
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

func (h *ShippingZoneRoutes) RegisterRoutes() {
	h.muxRouter.Handle("/shipping-zones", h.secureAdmin(h.CreateShippingZone)).Methods("POST")
	h.muxRouter.Handle("/shipping-zones", h.secureAdmin(h.ListShippingZones)).Methods("GET")
	h.muxRouter.Handle("/shipping-zones/{id}", h.secureAdmin(h.RemoveShippingZone)).Methods("DELETE")

	h.muxRouter.Handle("/shipping-zones/excluded", h.secureAdmin(h.CreateExcludedShippingZone)).Methods("POST")
	h.muxRouter.Handle("/shipping-zones/excluded", h.secureAdmin(h.ListExcludedShippingZones)).Methods("GET")
	h.muxRouter.Handle("/shipping-zones/excluded/{id}", h.secureAdmin(h.RemoveExcludedShippingZone)).Methods("DELETE")
}
