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

func (h *ProductRoutes) CreateProductWithCategory(w http.ResponseWriter, r *http.Request) {
	category := mux.Vars(r)["category"]
	var product types.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request body")
		return
	}

	if err := h.productService.CreateProductWithCategory(r.Context(), &product, category); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusCreated, product)
}

func (h *ProductRoutes) GetProducts(w http.ResponseWriter, r *http.Request) {
	params := u.ParsePaginationParams(r, 1, 25)

	filters := types.ProductFilter{
		Page:        params.Page,
		Limit:       params.Limit,
		InStock:     r.URL.Query().Get("in_stock") == "true",
		SortByPrice: r.URL.Query().Get("sort_by") == "price",
		SortAsc:     r.URL.Query().Get("sort_order") == "asc",
		Categories:  r.URL.Query()["category"],
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
		u.RespondWithError(w, r, http.StatusBadRequest, "Invalid ID")
		return
	}
	product, err := h.productService.GetProductByID(r.Context(), productId)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, product)
}

func (h *ProductRoutes) UpdateInventory(w http.ResponseWriter, r *http.Request) {
	productID := mux.Vars(r)["id"]
	var input struct {
		Quantity int `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err := h.productService.UpdateInventory(r.Context(), productID, input.Quantity)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondSuccess(w)
}

func (h *ProductRoutes) RemoveProduct(w http.ResponseWriter, r *http.Request) {
	productID := mux.Vars(r)["id"]
	err := h.productService.RemoveProduct(r.Context(), productID)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondSuccess(w)
}

func (h *ProductRoutes) RegisterRoutes() {
	h.muxRouter.HandleFunc("/products", h.GetProducts).Methods(http.MethodGet)
	h.muxRouter.HandleFunc("/products/{id}", h.GetProduct).Methods(http.MethodGet)
	h.muxRouter.Handle("/products", h.secureAdmin(h.CreateProduct)).Methods(http.MethodPost)
	h.muxRouter.Handle("/products/categories/{category}", h.secureAdmin(h.CreateProductWithCategory)).Methods(http.MethodPost)
	h.muxRouter.Handle("/products/{id}", h.secureAdmin(h.RemoveProduct)).Methods(http.MethodDelete)
	h.muxRouter.Handle("/products/{id}/inventory", h.secureAdmin(h.UpdateInventory)).Methods(http.MethodPut)
	// router.HandleFunc("/products/{id}", h.UpdateProduct).Methods("PUT")
}
