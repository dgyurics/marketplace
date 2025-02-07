package routes

import (
	"encoding/json"
	"net/http"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/services"
	u "github.com/dgyurics/marketplace/utilities"
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
		u.RespondWithError(w, r, http.StatusBadRequest, "Invalid request payload")
		return
	}

	categoryId, err := h.categoryService.CreateCategory(r.Context(), category)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	category.ID = categoryId
	u.RespondWithJSON(w, http.StatusCreated, category)
}

func (h *CategoryRoutes) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categoryService.GetAllCategories(r.Context())
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, categories)
}

func (h *CategoryRoutes) GetCategory(w http.ResponseWriter, r *http.Request) {
	categoryId, ok := mux.Vars(r)["id"]
	if !ok {
		u.RespondWithError(w, r, http.StatusBadRequest, "Invalid ID")
		return
	}
	category, err := h.categoryService.GetCategoryByID(r.Context(), categoryId)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, category)
}

func (h *CategoryRoutes) GetProductsByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	products, err := h.categoryService.GetProductsByCategoryID(r.Context(), vars["id"])
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, products)
}

func (h *CategoryRoutes) RegisterRoutes() {
	h.muxRouter.HandleFunc("/categories", h.GetCategories).Methods(http.MethodGet)
	h.muxRouter.HandleFunc("/categories/{id}", h.GetCategory).Methods(http.MethodGet)
	h.muxRouter.HandleFunc("/categories/{id}/products", h.GetProductsByCategory).Methods(http.MethodGet)
	h.muxRouter.Handle("/categories", h.secureAdmin(h.CreateCategory)).Methods(http.MethodPost)
	// router.HandleFunc("/categories/{id}", UpdateCategory).Methods("PUT")
	// router.HandleFunc("/categories/{id}", DeleteCategory).Methods("DELETE")
}
