package routes

import (
	"encoding/json"
	"net/http"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
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
	var category types.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}

	if err := h.categoryService.CreateCategory(r.Context(), &category); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

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
	vars := mux.Vars(r)
	category, err := h.categoryService.GetCategoryByID(r.Context(), vars["id"])
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, category)
}

func (h *CategoryRoutes) RegisterRoutes() {
	h.muxRouter.HandleFunc("/categories", h.GetCategories).Methods(http.MethodGet)
	h.muxRouter.HandleFunc("/categories/{id}", h.GetCategory).Methods(http.MethodGet)
	h.muxRouter.Handle("/categories", h.secureAdmin(h.CreateCategory)).Methods(http.MethodPost)
	// router.HandleFunc("/categories/{id}", UpdateCategory).Methods("PATCH")
	// router.HandleFunc("/categories/{id}", DeleteCategory).Methods("DELETE")
}
