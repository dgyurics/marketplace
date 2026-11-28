package routes

import (
	"errors"
	"net/http"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	u "github.com/dgyurics/marketplace/utilities"
)

type CartRoutes struct {
	router
	cartService services.CartService
}

func NewCartRoutes(
	cartService services.CartService,
	router router) *CartRoutes {
	return &CartRoutes{
		router:      router,
		cartService: cartService,
	}
}

func (h *CartRoutes) AddItem(w http.ResponseWriter, r *http.Request) {
	err := h.cartService.AddItem(r.Context(), r.PathValue("id"))

	// Error handling
	if errors.Is(err, types.ErrConstraintViolation) {
		u.RespondWithError(w, r, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondSuccess(w)
}

func (h *CartRoutes) RemoveItem(w http.ResponseWriter, r *http.Request) {
	productID := r.PathValue("id")

	if err := h.cartService.RemoveItem(r.Context(), productID); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondSuccess(w)
}

func (h *CartRoutes) GetItems(w http.ResponseWriter, r *http.Request) {
	// Retrieve the cart for the current user
	cart, err := h.cartService.GetItems(r.Context())
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, cart)
}

func (h *CartRoutes) RemoveItems(w http.ResponseWriter, r *http.Request) {
	if err := h.cartService.RemoveItems(r.Context()); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	u.RespondSuccess(w)
}

func (h *CartRoutes) RegisterRoutes() {
	h.mux.Handle("POST /carts/items/{id}", h.secure(types.RoleGuest)(h.AddItem))
	h.mux.Handle("DELETE /carts/items/{id}", h.secure(types.RoleGuest)(h.RemoveItem))
	h.mux.Handle("GET /carts", h.secure(types.RoleGuest)(h.GetItems))
	h.mux.Handle("DELETE /carts", h.secure(types.RoleGuest)(h.RemoveItems))
}
