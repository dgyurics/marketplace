package routes

import (
	"encoding/json"
	"net/http"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	u "github.com/dgyurics/marketplace/utilities"
	"github.com/gorilla/mux"
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

func (h *CartRoutes) AddItemToCart(w http.ResponseWriter, r *http.Request) {
	var item types.CartItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}

	if item.Quantity <= 0 {
		u.RespondWithError(w, r, http.StatusBadRequest, "quantity must be greater than 0")
		return
	}

	item.Product.ID = mux.Vars(r)["product_id"]
	if err := h.cartService.AddItemToCart(r.Context(), &item); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondSuccess(w)
}

func (h *CartRoutes) UpdateCartItem(w http.ResponseWriter, r *http.Request) {
	var item types.CartItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}

	if item.Quantity <= 0 {
		u.RespondWithError(w, r, http.StatusBadRequest, "quantity must be greater than 0")
		return
	}

	item.Product.ID = mux.Vars(r)["product_id"]
	if err := h.cartService.UpdateCartItem(r.Context(), &item); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, item)
}

func (h *CartRoutes) RemoveItemFromCart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["product_id"]

	if err := h.cartService.RemoveItemFromCart(r.Context(), productID); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondSuccess(w)
}

func (h *CartRoutes) GetCart(w http.ResponseWriter, r *http.Request) {
	cart, err := h.cartService.GetCart(r.Context())
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, cart)
}

func (h *CartRoutes) RegisterRoutes() {
	h.muxRouter.Handle("/carts/items/{product_id}", h.secure(h.AddItemToCart)).Methods(http.MethodPost)
	h.muxRouter.Handle("/carts/items/{product_id}", h.secure(h.UpdateCartItem)).Methods(http.MethodPatch)
	h.muxRouter.Handle("/carts/items/{product_id}", h.secure(h.RemoveItemFromCart)).Methods(http.MethodDelete)
	h.muxRouter.Handle("/carts", h.secure(h.GetCart)).Methods(http.MethodGet)
}
