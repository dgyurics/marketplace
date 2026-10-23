package routes

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/types/stripe"
	u "github.com/dgyurics/marketplace/utilities"
)

type OrderRoutes struct {
	router
	config              types.PaymentOptions
	orderService        services.OrderService
	taxService          services.TaxService
	paymentService      services.PaymentService
	cartService         services.CartService
	addressService      services.AddressService
	userService         services.UserService
	notificationService services.NotificationService
}

func NewOrderRoutes(
	config types.PaymentOptions,
	orderService services.OrderService,
	taxService services.TaxService,
	paymentService services.PaymentService,
	cartService services.CartService,
	addressService services.AddressService,
	userService services.UserService,
	notificationService services.NotificationService,
	router router) *OrderRoutes {
	return &OrderRoutes{
		router:              router,
		config:              config,
		orderService:        orderService,
		taxService:          taxService,
		paymentService:      paymentService,
		cartService:         cartService,
		addressService:      addressService,
		userService:         userService,
		notificationService: notificationService,
	}
}

func (h *OrderRoutes) GetOrderOwner(w http.ResponseWriter, r *http.Request) {
	order, err := h.orderService.GetOrderByIDAndUser(r.Context(), r.PathValue("id"))
	if errors.Is(err, types.ErrNotFound) {
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
	if errors.Is(err, types.ErrNotFound) {
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
	if errors.Is(err, types.ErrNotFound) {
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

	err := h.orderService.UpdateOrder(r.Context(), &order)
	if errors.Is(err, types.ErrNotFound) {
		u.RespondWithError(w, r, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	go h.notificationService.NotifyOrder(order.UserID, services.SubjectOrderUpdate, services.NotifyOrderUpdate, order)

	u.RespondWithJSON(w, http.StatusOK, order)
}

func (h *OrderRoutes) notifyOrderCreated(ctx context.Context, order types.Order) {
	go h.notificationService.NotifyOrder(order.UserID, services.SubjectOrderConf, services.NotifyOrderConf, order)

	admins, err := h.userService.GetAllAdmins(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to load admins for order notification", "order_id", order.ID, "error", err)
		return
	}
	for _, admin := range admins {
		go h.notificationService.NotifyOrder(admin.ID, services.SubjectOrderRecv, services.NotifyOrderRecv, order)
	}
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
		order.Amount = u.Ptr(*order.Amount + ci.UnitPrice*int64(ci.Quantity))
	}
	order.TotalAmount = u.Ptr(*order.Amount + *order.TaxAmount + *order.ShippingAmount)
}

func (h *OrderRoutes) CreateDraftOrder(w http.ResponseWriter, r *http.Request) {
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
	if errors.Is(err, types.ErrNotFound) {
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

	// Create order
	order := &types.Order{
		IdempotencyKey: &idempotencyKey,
		Address:        addr,
		Status:         u.Ptr(types.OrderDraft),
		PaymentStatus:  u.Ptr(types.PaymentStatusUnpaid),
		Amount:         u.Ptr(int64(0)),
		TaxAmount:      u.Ptr(int64(0)),
		ShippingAmount: u.Ptr(int64(0)),
		TotalAmount:    u.Ptr(int64(0)),
	}
	populateOrderFromCart(order, cart)

	err = h.orderService.CreateOrder(r.Context(), order)

	var stockErr *types.InsufficientStockError
	if errors.As(err, &stockErr) {
		u.RespondWithJSON(w, http.StatusConflict, stockErr.Items)
		return
	}
	if errors.Is(err, types.ErrConstraintViolation) {
		u.RespondWithError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusCreated, stripe.CreateOrderResponse{OrderID: order.ID})
}

func (h *OrderRoutes) CheckoutWithStripe(w http.ResponseWriter, r *http.Request) {
	if !h.config.Stripe {
		u.RespondWithError(w, r, http.StatusNotFound, "stripe payments disabled")
		return
	}

	order, err := h.orderService.GetOrderByIDAndUser(r.Context(), r.PathValue("id"))
	if errors.Is(err, types.ErrNotFound) {
		u.RespondWithError(w, r, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	if *order.Status != types.OrderDraft {
		u.RespondWithError(w, r, http.StatusConflict, "order is not open")
		return
	}

	// calculate tax
	tax, err := h.taxService.CalculateTax(r.Context(), "", order.Address, order.Items)
	if errors.Is(err, types.ErrInvalidInput) {
		u.RespondWithError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// update amounts + order
	order.TaxAmount = u.Ptr(tax)
	order.TotalAmount = u.Ptr(*order.Amount + tax + *order.ShippingAmount)
	order.PaymentMethod = u.Ptr(types.PaymentMethodStripe)
	order.Status = u.Ptr(types.OrderPending)

	err = h.orderService.UpdateOrder(r.Context(), &order)
	if errors.Is(err, types.ErrNotFound) {
		u.RespondWithError(w, r, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// create payment intent
	pi, err := h.paymentService.CreatePaymentIntent(r.Context(), order.ID, *order.TotalAmount, order.Address.Email)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	u.RespondWithJSON(w, http.StatusOK, stripe.CreateOrderResponse{
		ClientSecret: pi.ClientSecret,
		OrderID:      order.ID,
		TotalAmount:  *order.TotalAmount,
	})
}

func (h *OrderRoutes) CheckoutOnDelivery(w http.ResponseWriter, r *http.Request) {
	if !h.config.OnDelivery {
		u.RespondWithError(w, r, http.StatusBadRequest, "on delivery payments disabled")
		return
	}

	order, err := h.orderService.GetOrderByIDAndUser(r.Context(), r.PathValue("id"))
	if errors.Is(err, types.ErrNotFound) {
		u.RespondWithError(w, r, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	if *order.Status != types.OrderDraft {
		u.RespondWithError(w, r, http.StatusConflict, "order is not open")
		return
	}

	// update order
	order.PaymentMethod = u.Ptr(types.PaymentMethodOnDelivery)
	order.Status = u.Ptr(types.OrderPending)
	err = h.orderService.UpdateOrder(r.Context(), &order)
	if errors.Is(err, types.ErrNotFound) {
		u.RespondWithError(w, r, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// No payment webhook will fire for this flow, so confirm and clear here.
	h.notifyOrderCreated(r.Context(), order)

	if err := h.cartService.RemoveItems(r.Context()); err != nil {
		slog.WarnContext(r.Context(), "Error clearing cart", "user_id", order.UserID)
	}

	u.RespondWithJSON(w, http.StatusOK, stripe.CreateOrderResponse{OrderID: order.ID})
}

// RequestCheckoutOnDelivery will submit a pay-on-delivery order for approval.
func (h *OrderRoutes) RequestCheckoutOnDelivery(w http.ResponseWriter, r *http.Request) {
	u.RespondWithError(w, r, http.StatusNotImplemented, "pay-on-delivery approval requests are not implemented")
}

func (h *OrderRoutes) RegisterRoutes() {
	h.mux.Handle("POST /orders", h.secure(types.RoleGuest)(h.limit(h.CreateDraftOrder, 20, time.Hour)))
	h.mux.Handle("POST /orders/{id}/stripe/intent", h.secure(types.RoleGuest)(h.limit(h.CheckoutWithStripe, 10, time.Hour)))
	h.mux.Handle("POST /orders/{id}/pay-on-delivery", h.secure(types.RoleMember)(h.CheckoutOnDelivery))
	h.mux.Handle("POST /orders/{id}/pay-on-delivery/request", h.secure(types.RoleGuest)(h.RequestCheckoutOnDelivery))
	h.mux.Handle("PUT /orders", h.secure(types.RoleAdmin)(h.UpdateOrder))
	h.mux.Handle("GET /orders/{id}/public", http.HandlerFunc(h.GetOrderPublic))
	h.mux.Handle("GET /orders/{id}/owner", h.secure(types.RoleGuest)(h.GetOrderOwner))
	h.mux.Handle("GET /orders/{id}/admin", h.secure(types.RoleStaff)(h.GetOrderAdmin))
	h.mux.Handle("GET /orders", h.secure(types.RoleStaff)(h.GetOrders))
}
