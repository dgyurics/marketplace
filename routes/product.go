package routes

import (
	"encoding/json"
	"net/http"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	u "github.com/dgyurics/marketplace/utilities"
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
	var product types.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request body")
		return
	}

	if err := h.productService.CreateProduct(r.Context(), &product); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusCreated, product)
}

func (h *ProductRoutes) GetProducts(w http.ResponseWriter, r *http.Request) {
	params := u.ParsePaginationParams(r, 1, 25)
	inStock := r.URL.Query().Get("in_stock") == "true"
	sortBy := types.ParseSortBy(r.URL.Query().Get("sort_by"))
	categories := r.URL.Query()["category"]
	filters := types.ProductFilter{
		Page:       params.Page,
		Limit:      params.Limit,
		InStock:    inStock,
		SortBy:     sortBy,
		Categories: categories,
	}

	products, err := h.productService.GetProducts(r.Context(), filters)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, products)
}

func (h *ProductRoutes) GetProduct(w http.ResponseWriter, r *http.Request) {
	productId, ok := mux.Vars(r)["id"]
	if !ok {
		u.RespondWithError(w, r, http.StatusBadRequest, "invalid product ID")
		return
	}
	product, err := h.productService.GetProductByID(r.Context(), productId)
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, product)
}

func (h *ProductRoutes) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	var product types.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	err := h.productService.UpdateProduct(r.Context(), product)
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	u.RespondSuccess(w)
}

func (h *ProductRoutes) RemoveProduct(w http.ResponseWriter, r *http.Request) {
	productID := mux.Vars(r)["id"]
	err := h.productService.RemoveProduct(r.Context(), productID)
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondSuccess(w)
}

func (h *ProductRoutes) RegisterRoutes() {
	h.muxRouter.HandleFunc("/products", h.GetProducts).Methods(http.MethodGet)
	h.muxRouter.HandleFunc("/products/{id}", h.GetProduct).Methods(http.MethodGet)
	h.muxRouter.Handle("/products", h.secure(types.RoleAdmin)(h.CreateProduct)).Methods(http.MethodPost)
	h.muxRouter.Handle("/products/{id}", h.secure(types.RoleAdmin)(h.RemoveProduct)).Methods(http.MethodDelete)
	h.muxRouter.Handle("/products", h.secure(types.RoleAdmin)(h.UpdateProduct)).Methods(http.MethodPut)
}
