package routes

import (
	"encoding/json"
	"net/http"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/services"
	"github.com/gorilla/mux"
)

type CategoryRoutes struct {
	router
	categoryService services.CategoryService
}

func NewCategoryRoutes(
	categoryService services.CategoryService,
	router router) *CategoryRoutes {
	return &CategoryRoutes{
		router:          router,
		categoryService: categoryService,
	}
}

func (h *CategoryRoutes) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var category models.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	categoryId, err := h.categoryService.CreateCategory(r.Context(), category)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	category.ID = categoryId
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(category)
}

func (h *CategoryRoutes) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categoryService.GetAllCategories(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(categories)
}

func (h *CategoryRoutes) GetCategory(w http.ResponseWriter, r *http.Request) {
	categoryId, ok := mux.Vars(r)["id"]
	if !ok {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	category, err := h.categoryService.GetCategoryByID(r.Context(), categoryId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(category)
}

func (h *CategoryRoutes) GetProductsByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	products, err := h.categoryService.GetProductsByCategoryID(r.Context(), vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(products)
}

func (h *CategoryRoutes) RegisterRoutes() {
	h.muxRouter.HandleFunc("/categories", h.GetCategories).Methods(http.MethodGet)
	h.muxRouter.HandleFunc("/categories/{id}", h.GetCategory).Methods(http.MethodGet)
	h.muxRouter.HandleFunc("/categories/{id}/products", h.GetProductsByCategory).Methods(http.MethodGet)
	h.muxRouter.Handle("/categories", h.secureAdmin(h.CreateCategory)).Methods(http.MethodPost)
	// router.HandleFunc("/categories/{id}", UpdateCategory).Methods("PUT")
	// router.HandleFunc("/categories/{id}", DeleteCategory).Methods("DELETE")
}
