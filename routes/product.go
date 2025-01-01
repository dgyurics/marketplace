package routes

import (
	"encoding/json"
	"net/http"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/services"
	"github.com/gorilla/mux"
)

type ProductRoutes struct {
	router
	productService services.ProductService
}

func NewProductRoutes(
	productService services.ProductService,
	router router) *ProductRoutes {
	return &ProductRoutes{
		router:         router,
		productService: productService,
	}
}

func (h *ProductRoutes) CreateProduct(w http.ResponseWriter, r *http.Request) {
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

func (h *ProductRoutes) GetProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.productService.GetAllProducts(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(products)
}

func (h *ProductRoutes) GetProduct(w http.ResponseWriter, r *http.Request) {
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

func (h *ProductRoutes) UpdateInventory(w http.ResponseWriter, r *http.Request) {
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

func (h *ProductRoutes) RemoveProduct(w http.ResponseWriter, r *http.Request) {
	productID := mux.Vars(r)["id"]
	err := h.productService.RemoveProduct(r.Context(), productID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductRoutes) RegisterRoutes() {
	h.muxRouter.HandleFunc("/products", h.GetProducts).Methods(http.MethodGet)
	h.muxRouter.HandleFunc("/products/{id}", h.GetProduct).Methods(http.MethodGet)
	h.muxRouter.Handle("/products", h.secureAdmin(h.CreateProduct)).Methods(http.MethodPost)
	h.muxRouter.Handle("/products/{id}", h.secureAdmin(h.RemoveProduct)).Methods(http.MethodDelete)
	h.muxRouter.Handle("/products/{id}/inventory", h.secureAdmin(h.UpdateInventory)).Methods(http.MethodPut)
	// router.HandleFunc("/products/{id}", h.UpdateProduct).Methods("PUT")
}
