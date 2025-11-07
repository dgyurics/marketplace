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

func (h *CartRoutes) AddItem(w http.ResponseWriter, r *http.Request) {
	var item types.CartItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}

	if item.Quantity <= 0 {
		u.RespondWithError(w, r, http.StatusBadRequest, "quantity must be greater than 0")
		return
	}

	// Attempt to add item to cart
	item.Product.ID = mux.Vars(r)["id"]
	err := h.cartService.AddItem(r.Context(), &item)

	// Error handling
	if err == types.ErrConstraintViolation {
		u.RespondWithError(w, r, http.StatusConflict, "product cart constraint reached")
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondSuccess(w)
}

func (h *CartRoutes) RemoveItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["id"]

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

func (h *CartRoutes) RegisterRoutes() {
	h.muxRouter.Handle("/carts/items/{id}", h.secure(h.AddItem)).Methods(http.MethodPost)
	h.muxRouter.Handle("/carts/items/{id}", h.secure(h.RemoveItem)).Methods(http.MethodDelete)
	h.muxRouter.Handle("/carts", h.secure(h.GetItems)).Methods(http.MethodGet)
}
