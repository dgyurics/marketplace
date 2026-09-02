package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/types/stripe"
	u "github.com/dgyurics/marketplace/utilities"
)

type OrderRoutes struct {
	router
	orderService   services.OrderService
	taxService     services.TaxService
	paymentService services.PaymentService
	cartService    services.CartService
	addressService services.AddressService
}

func NewOrderRoutes(
	orderService services.OrderService,
	taxService services.TaxService,
	paymentService services.PaymentService,
	cartService services.CartService,
	addressService services.AddressService,
	router router) *OrderRoutes {
	return &OrderRoutes{
		router:         router,
		orderService:   orderService,
		taxService:     taxService,
		paymentService: paymentService,
		cartService:    cartService,
		addressService: addressService,
	}
}

func (h *OrderRoutes) GetOrderOwner(w http.ResponseWriter, r *http.Request) {
	order, err := h.orderService.GetOrderByIDAndUser(r.Context(), r.PathValue("id"))
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	u.RespondWithJSON(w, http.StatusOK, order)
}

func (h *OrderRoutes) GetOrderPublic(w http.ResponseWriter, r *http.Request) {
	order, err := h.orderService.GetOrderByIDPublic(r.Context(), r.PathValue("id"))
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	u.RespondWithJSON(w, http.StatusOK, order)
}

func (h *OrderRoutes) GetOrderAdmin(w http.ResponseWriter, r *http.Request) {
	order, err := h.orderService.GetOrderByID(r.Context(), r.PathValue("id"))
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	u.RespondWithJSON(w, http.StatusOK, order)
}

func (h *OrderRoutes) GetOrders(w http.ResponseWriter, r *http.Request) {
	params := u.ParsePaginationParams(r, 1, 25)
	orders, err := h.orderService.GetOrders(r.Context(), params.Page, params.Limit)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, orders)
}

func (h *OrderRoutes) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	var order types.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}

	if order.ID == "" {
		u.RespondWithError(w, r, http.StatusBadRequest, "missing order ID")
		return
	}

	if err := h.orderService.UpdateOrder(r.Context(), &order); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, order)
}

func (h *OrderRoutes) CreateOrder(w http.ResponseWriter, r *http.Request) {
	idempotencyKey := r.Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		u.RespondWithError(w, r, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	shippingID := r.URL.Query().Get("shipping_id")
	if shippingID == "" {
		u.RespondWithError(w, r, http.StatusBadRequest, "shipping_id is required")
		return
	}

	// Fetch shipping address
	addr, err := h.addressService.GetAddress(r.Context(), shippingID)
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Fetch user cart
	cart, err := h.cartService.GetItems(r.Context())
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	if len(cart) == 0 {
		u.RespondWithError(w, r, http.StatusBadRequest, "cart is empty")
		return
	}

	// Calculate tax
	tax, err := h.taxService.CalculateTax(r.Context(), "", addr, cart)
	if err == types.ErrInvalidInput {
		u.RespondWithError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Create order
	order := &types.Order{
		IdempotencyKey: &idempotencyKey,
		Address:        addr,
		TaxAmount:      tax,
	}
	populateOrderFromCart(order, cart)
	err = h.orderService.CreateOrder(r.Context(), order)

	var stockErr *types.InsufficientStockError
	if errors.As(err, &stockErr) {
		u.RespondWithJSON(w, http.StatusConflict, stockErr.Items)
		return
	}
	if err == types.ErrConstraintViolation {
		u.RespondWithError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Create payment intent
	pi, err := h.paymentService.CreatePaymentIntent(r.Context(), order.ID, order.TotalAmount, order.Address.Email)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusOK, stripe.CreateOrderResponse{ClientSecret: pi.ClientSecret, OrderID: order.ID})
}

// populateOrderFromCart builds the order items from the cart and calculates
// the subtotal (Amount) and TotalAmount (subtotal + tax + shipping).
func populateOrderFromCart(order *types.Order, cart []types.CartItem) {
	order.Items = make([]types.OrderItem, 0, len(cart))
	for _, ci := range cart {
		oi := types.OrderItem{
			Product:   ci.Product,
			Quantity:  ci.Quantity,
			UnitPrice: ci.UnitPrice,
		}
		order.Items = append(order.Items, oi)
		order.Amount = order.Amount + ci.UnitPrice*int64(ci.Quantity)
	}
	order.TotalAmount = order.Amount + order.TaxAmount + order.ShippingAmount
}

func (h *OrderRoutes) RegisterRoutes() {
	h.mux.Handle("POST /orders", h.secure(types.RoleGuest)(h.limit(h.CreateOrder, 5, time.Hour)))
	h.mux.Handle("PUT /orders", h.secure(types.RoleAdmin)(h.UpdateOrder))
	h.mux.Handle("GET /orders/{id}/public", http.HandlerFunc(h.GetOrderPublic))
	h.mux.Handle("GET /orders/{id}/owner", h.secure(types.RoleGuest)(h.GetOrderOwner))
	h.mux.Handle("GET /orders/{id}/admin", h.secure(types.RoleStaff)(h.GetOrderAdmin))
	h.mux.Handle("GET /orders", h.secure(types.RoleStaff)(h.GetOrders))
}
