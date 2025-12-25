package routes

import (
	"net/http"
	"time"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/types/stripe"
	u "github.com/dgyurics/marketplace/utilities"
	"github.com/gorilla/mux"
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
	order, err := h.orderService.GetOrderByIDAndUser(r.Context(), mux.Vars(r)["id"])
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
	order, err := h.orderService.GetOrderByIDPublic(r.Context(), mux.Vars(r)["id"])
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
	order, err := h.orderService.GetOrderByID(r.Context(), mux.Vars(r)["id"])
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

// CreateOrder handles order creation
// It fetches the shipping address and cart items, calculates tax,
// creates the order, generates a payment intent, and clears the cart.
// The final step is handled by the /payment/events webhook, which marks the order as paid
// upon successful payment. After this, the order is ready for fulfillment.
func (h *OrderRoutes) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Fetch shipping address
	addr, err := h.addressService.GetAddress(r.Context(), r.URL.Query().Get("shipping_id"))
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
		Address:   addr,
		TaxAmount: tax,
	}
	calculateOrderFromCart(order, cart)
	err = h.orderService.CreateOrder(r.Context(), order)
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

	// Respond with client secret and order ID for payment processing
	u.RespondWithJSON(w, http.StatusOK, stripe.CreateOrderResponse{ClientSecret: pi.ClientSecret, OrderID: order.ID})
}

func calculateOrderFromCart(order *types.Order, cart []types.CartItem) {
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
	h.muxRouter.Handle("/orders", h.secure(types.RoleGuest)(h.limit(h.CreateOrder, 5, time.Hour))).Methods(http.MethodPost)
	h.muxRouter.HandleFunc("/orders/{id}/public", h.GetOrderPublic).Methods(http.MethodPost)
	h.muxRouter.Handle("/orders/{id}/owner", h.secure(types.RoleGuest)(h.GetOrderOwner)).Methods(http.MethodPost)
	// FIXME rename endpoint now that we have staff + admin
	h.muxRouter.Handle("/orders/{id}/admin", h.secure(types.RoleStaff)(h.GetOrderAdmin)).Methods(http.MethodPost)
	h.muxRouter.Handle("/orders", h.secure(types.RoleStaff)(h.GetOrders)).Methods(http.MethodGet)
}
