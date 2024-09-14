package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/services"
	"github.com/gorilla/mux"
)

type ProductHandler interface {
	CreateProduct(w http.ResponseWriter, r *http.Request)
	GetProducts(w http.ResponseWriter, r *http.Request)
	GetProduct(w http.ResponseWriter, r *http.Request)
}

type productHandler struct {
	productService services.ProductService
	router         *mux.Router
}

func RegisterProductHandler(productService services.ProductService, router *mux.Router) {
	handler := &productHandler{
		productService: productService,
		router:         router,
	}
	handler.RegisterRoutes()
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

func (h *productHandler) RegisterRoutes() {
	h.router.HandleFunc("/products", h.CreateProduct).Methods(http.MethodPost)
	h.router.HandleFunc("/products", h.GetProducts).Methods(http.MethodGet)
	h.router.HandleFunc("/products/{id}", h.GetProduct).Methods(http.MethodGet)
	// router.HandleFunc("/products/{id}", h.UpdateProduct).Methods("PUT")
	// router.HandleFunc("/products/{id}", h.DeleteProduct).Methods("DELETE")
}
