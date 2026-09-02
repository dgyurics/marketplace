package routes

import (
	"encoding/json"
	"net/http"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	u "github.com/dgyurics/marketplace/utilities"
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
	featured := r.URL.Query().Get("featured") == "true"
	filters := types.ProductFilter{
		Page:       params.Page,
		Limit:      params.Limit,
		InStock:    inStock,
		Featured:   featured,
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
	product, err := h.productService.GetProductByID(r.Context(), r.PathValue("id"))
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
	err := h.productService.RemoveProduct(r.Context(), r.PathValue("id"))
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
	h.mux.HandleFunc("GET /products", h.GetProducts)
	h.mux.HandleFunc("GET /products/{id}", h.GetProduct)
	h.mux.Handle("POST /products", h.secure(types.RoleAdmin)(h.CreateProduct))
	h.mux.Handle("DELETE /products/{id}", h.secure(types.RoleAdmin)(h.RemoveProduct))
	h.mux.Handle("PUT /products", h.secure(types.RoleAdmin)(h.UpdateProduct))
}
