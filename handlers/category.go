package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dgyurics/marketplace/middleware"
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

func RegisterCategoryHandler(
	categoryService services.CategoryService,
	router *mux.Router,
	authMiddleware middleware.AuthMiddleware) {
	handler := &categoryHandler{
		categoryService: categoryService,
		router:          router,
	}
	handler.RegisterRoutes(authMiddleware)
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
	category.ID = categoryId
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

func (h *categoryHandler) RegisterRoutes(authMiddleware middleware.AuthMiddleware) {
	h.router.HandleFunc("/categories", h.GetCategories).Methods(http.MethodGet)
	h.router.HandleFunc("/categories/{id}", h.GetCategory).Methods(http.MethodGet)
	h.router.HandleFunc("/categories/{id}/products", h.GetProductsByCategory).Methods(http.MethodGet)
	h.router.Handle("/categories", authMiddleware.AuthenticateAdmin(http.HandlerFunc(h.CreateCategory))).Methods(http.MethodPost)
	// router.HandleFunc("/categories/{id}", UpdateCategory).Methods("PUT")
	// router.HandleFunc("/categories/{id}", DeleteCategory).Methods("DELETE")
}
