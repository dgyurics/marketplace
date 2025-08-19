package routes

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	u "github.com/dgyurics/marketplace/utilities"
	"github.com/gorilla/mux"
)

type CartRoutes struct {
	router
	cartService  services.CartService
	orderService services.OrderService
}

func NewCartRoutes(
	cartService services.CartService,
	orderService services.OrderService,
	router router) *CartRoutes {
	return &CartRoutes{
		router:       router,
		cartService:  cartService,
		orderService: orderService,
	}
}

func (h *CartRoutes) AddItemToCart(w http.ResponseWriter, r *http.Request) {
	if err := h.cancelPendingOrder(r.Context()); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	var item types.CartItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}

	if item.Quantity <= 0 {
		u.RespondWithError(w, r, http.StatusBadRequest, "quantity must be greater than 0")
		return
	}

	item.Product.ID = mux.Vars(r)["id"]
	err := h.cartService.AddItemToCart(r.Context(), &item)
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusConflict, "insufficient stock for product")
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondSuccess(w)
}

func (h *CartRoutes) UpdateCartItem(w http.ResponseWriter, r *http.Request) {
	if err := h.cancelPendingOrder(r.Context()); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	var item types.CartItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}

	if item.Quantity <= 0 {
		u.RespondWithError(w, r, http.StatusBadRequest, "quantity must be greater than 0")
		return
	}

	item.Product.ID = mux.Vars(r)["id"]
	err := h.cartService.UpdateCartItem(r.Context(), &item)
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusConflict, "insufficient stock for product")
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, item)
}

func (h *CartRoutes) RemoveItemFromCart(w http.ResponseWriter, r *http.Request) {
	if err := h.cancelPendingOrder(r.Context()); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	vars := mux.Vars(r)
	productID := vars["id"]

	if err := h.cartService.RemoveItemFromCart(r.Context(), productID); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondSuccess(w)
}

func (h *CartRoutes) GetCart(w http.ResponseWriter, r *http.Request) {
	// Retrieve the cart for the current user
	cart, err := h.cartService.GetCart(r.Context())
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, cart)
}

// cancelPendingOrder checks for a pending order and cancels it if found.
// When a user checks out, a new order is created
func (h *CartRoutes) cancelPendingOrder(ctx context.Context) error {
	ord, err := h.orderService.GetPendingOrderForUser(ctx)
	if err == types.ErrNotFound || ord.ID == "" {
		return nil // No pending order to cancel
	}
	if err != nil {
		return err
	}
	cancel := types.OrderCanceled
	_, err = h.orderService.UpdateOrder(ctx, types.OrderParams{
		ID:     ord.ID,
		Status: &cancel,
	})
	return err
}

func (h *CartRoutes) RegisterRoutes() {
	h.muxRouter.Handle("/carts/items/{id}", h.secure(h.AddItemToCart)).Methods(http.MethodPost)
	h.muxRouter.Handle("/carts/items/{id}", h.secure(h.UpdateCartItem)).Methods(http.MethodPatch)
	h.muxRouter.Handle("/carts/items/{id}", h.secure(h.RemoveItemFromCart)).Methods(http.MethodDelete)
	h.muxRouter.Handle("/carts", h.secure(h.GetCart)).Methods(http.MethodGet)
}
