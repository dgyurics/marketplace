package routes

import (
	"net/http"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	u "github.com/dgyurics/marketplace/utilities"
)

type TaxRoutes struct {
	router
	cartService services.CartService
	taxService  services.TaxService
}

func NewTaxRoutes(
	cartService services.CartService,
	taxService services.TaxService,
	router router) *TaxRoutes {
	return &TaxRoutes{
		router:      router,
		cartService: cartService,
		taxService:  taxService,
	}
}

// EstimateTax estimates tax for the current user's state, country, and cart items
func (h *TaxRoutes) EstimateTax(w http.ResponseWriter, r *http.Request) {
	addr := types.Address{
		Country: r.URL.Query().Get("country"),
		State:   r.URL.Query().Get("state"),
	}

	items, err := h.cartService.GetCart(r.Context())
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	taxEstimate, err := h.taxService.EstimateTax(r.Context(), addr, items)
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusNotFound, "tax data not found")
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, types.TaxEstimateResponse{TaxAmount: taxEstimate})
}

func (h *TaxRoutes) RegisterRoutes() {
	h.muxRouter.Handle("/tax/estimate", h.secure(h.EstimateTax)).Methods("GET")
}
