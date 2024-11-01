package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dgyurics/marketplace/middleware"
	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/services"
	"github.com/gorilla/mux"
)

type ProductHandler interface {
	CreateProduct(w http.ResponseWriter, r *http.Request)
	GetProducts(w http.ResponseWriter, r *http.Request)
	GetProduct(w http.ResponseWriter, r *http.Request)
	UpdateInventory(w http.ResponseWriter, r *http.Request)
}

type productHandler struct {
	productService services.ProductService
	router         *mux.Router
}

func RegisterProductHandler(
	productService services.ProductService,
	router *mux.Router,
	authMiddleware middleware.AccessControl) {
	handler := &productHandler{
		productService: productService,
		router:         router,
	}
	handler.RegisterRoutes(authMiddleware)
}

func (h *productHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var product models.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Retrieve optional categoryId query parameter
	categoryId := r.URL.Query().Get("categoryId")

	if categoryId != "" {
		if err := h.productService.CreateProductWithCategory(r.Context(), &product, categoryId); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if err := h.productService.CreateProduct(r.Context(), &product); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

func (h *productHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.productService.GetAllProducts(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(products)
}

func (h *productHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	productId, ok := mux.Vars(r)["id"]
	if !ok {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	product, err := h.productService.GetProductByID(r.Context(), productId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(product)
}

func (h *productHandler) UpdateInventory(w http.ResponseWriter, r *http.Request) {
	productID := mux.Vars(r)["id"]
	var input struct {
		Quantity int `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	err := h.productService.UpdateInventory(r.Context(), productID, input.Quantity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *productHandler) RegisterRoutes(authMiddleware middleware.AccessControl) {
	h.router.HandleFunc("/products", h.GetProducts).Methods(http.MethodGet)
	h.router.HandleFunc("/products/{id}", h.GetProduct).Methods(http.MethodGet)
	h.router.Handle("/products", authMiddleware.AuthenticateAdmin(h.CreateProduct)).Methods(http.MethodPost)
	h.router.Handle("/products/{id}/inventory", authMiddleware.AuthenticateAdmin(h.UpdateInventory)).Methods(http.MethodPut)
	// router.HandleFunc("/products/{id}", h.UpdateProduct).Methods("PUT")
	// router.HandleFunc("/products/{id}", h.DeleteProduct).Methods("DELETE")
}
