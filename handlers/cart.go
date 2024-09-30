package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/services"
	"github.com/gorilla/mux"
)

type CartHandler interface {
	AddItemToCart(w http.ResponseWriter, r *http.Request)
	RemoveItemFromCart(w http.ResponseWriter, r *http.Request)
	GetCart(w http.ResponseWriter, r *http.Request)
	Checkout(w http.ResponseWriter, r *http.Request)
}

type cartHandler struct {
	cartService services.CartService
	router      *mux.Router
}

func RegisterCartHandler(cartService services.CartService, router *mux.Router) {
	handler := &cartHandler{
		cartService: cartService,
		router:      router,
	}
	handler.registerRoutes()
}

func (h *cartHandler) AddItemToCart(w http.ResponseWriter, r *http.Request) {
	cartID := mux.Vars(r)["cart_id"]

	var item models.CartItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := h.cartService.AddItemToCart(r.Context(), cartID, &item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *cartHandler) RemoveItemFromCart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cartID := vars["cart_id"]
	productID := vars["product_id"]

	err := h.cartService.RemoveItemFromCart(r.Context(), cartID, productID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *cartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cartID := vars["cart_id"]

	cart, err := h.cartService.GetCartByID(r.Context(), cartID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cart)
}

func (h *cartHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cartID := vars["cart_id"]

	// Implement checkout logic, e.g., process payment, reduce inventory, etc.
	// For now, just clear the cart.
	err := h.cartService.ClearCart(r.Context(), cartID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Checkout completed and cart cleared",
	})
}

func (h *cartHandler) registerRoutes() {
	h.router.HandleFunc("/carts/{cart_id}/items", h.AddItemToCart).Methods("POST")
	h.router.HandleFunc("/carts/{cart_id}/items/{product_id}", h.RemoveItemFromCart).Methods("DELETE")
	h.router.HandleFunc("/carts/{cart_id}", h.GetCart).Methods("GET")
	h.router.HandleFunc("/carts/{cart_id}/checkout", h.Checkout).Methods("POST")
}
