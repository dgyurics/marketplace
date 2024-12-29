package routes

import (
	"encoding/json"
	"net/http"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/services"
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
	var item models.CartItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := h.cartService.AddItemToCart(r.Context(), &item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *CartRoutes) RemoveItemFromCart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["product_id"]

	err := h.cartService.RemoveItemFromCart(r.Context(), productID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *CartRoutes) GetCart(w http.ResponseWriter, r *http.Request) {
	cart, err := h.cartService.GetCart(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cart)
}

func (h *CartRoutes) RegisterRoutes() {
	h.muxRouter.Handle("/carts/items", h.secure(h.AddItemToCart)).Methods(http.MethodPost)
	h.muxRouter.Handle("/carts/items/{product_id}", h.secure(h.RemoveItemFromCart)).Methods(http.MethodDelete)
	h.muxRouter.Handle("/carts", h.secure(h.GetCart)).Methods(http.MethodGet)
}
