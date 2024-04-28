package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/services"
	"github.com/gorilla/mux"
)

type CategoryHandler interface {
	CreateCategory(w http.ResponseWriter, r *http.Request)
	GetCategories(w http.ResponseWriter, r *http.Request)
	GetCategory(w http.ResponseWriter, r *http.Request)
	GetProductsByCategory(w http.ResponseWriter, r *http.Request)
}

type categoryHandler struct {
	categoryService services.CategoryService
	router          *mux.Router
}

func RegisterCategoryHandler(categoryService services.CategoryService, router *mux.Router) {
	handler := &categoryHandler{
		categoryService: categoryService,
		router:          router,
	}
	handler.RegisterRoutes()
}

func (h *categoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
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
	category.ID = strconv.Itoa(categoryId)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(category)
}

func (h *categoryHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categoryService.GetAllCategories(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(categories)
}

func (h *categoryHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idInt, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	category, err := h.categoryService.GetCategoryByID(r.Context(), idInt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(category)
}

func (h *categoryHandler) GetProductsByCategory(w http.ResponseWriter, r *http.Request) {
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

func (h *categoryHandler) RegisterRoutes() {
	h.router.HandleFunc("/categories", h.GetCategories).Methods("GET")
	h.router.HandleFunc("/categories/{id}", h.GetCategory).Methods("GET")
	h.router.HandleFunc("/categories", h.CreateCategory).Methods("POST")
	h.router.HandleFunc("/categories/{id}/products", h.GetProductsByCategory).Methods("GET")
	// router.HandleFunc("/categories/{id}", UpdateCategory).Methods("PUT")
	// router.HandleFunc("/categories/{id}", DeleteCategory).Methods("DELETE")
}
