package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/utilities"
	"github.com/stretchr/testify/require"

	"github.com/dgyurics/marketplace/types"
	"github.com/stretchr/testify/mock"
)

type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) CreateOrder(ctx context.Context, order *types.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderService) UpdateOrder(ctx context.Context, order *types.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderService) GetOrders(ctx context.Context, page, limit int) ([]types.Order, error) {
	args := m.Called(ctx, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.Order), args.Error(1)
}

func (m *MockOrderService) GetOrderByID(ctx context.Context, orderID string) (types.Order, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return types.Order{}, args.Error(1)
	}
	return args.Get(0).(types.Order), args.Error(1)
}

func (m *MockOrderService) GetOrderByIDPublic(ctx context.Context, orderID string) (types.Order, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return types.Order{}, args.Error(1)
	}
	return args.Get(0).(types.Order), args.Error(1)
}

func (m *MockOrderService) GetOrderByIDAndUser(ctx context.Context, orderID string) (types.Order, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return types.Order{}, args.Error(1)
	}
	return args.Get(0).(types.Order), args.Error(1)
}

type checkoutUsers struct{ services.UserService }

func (checkoutUsers) GetAllAdmins(context.Context) ([]types.User, error) {
	return nil, nil
}

type checkoutNotifications struct {
	services.NotificationService
	sent chan services.HtmlTemplate
}

func (n checkoutNotifications) NotifyOrder(_ string, _ string, template services.HtmlTemplate, _ types.Order) error {
	n.sent <- template
	return nil
}

func TestCheckoutOnDeliveryStatus(t *testing.T) {
	for _, role := range []types.Role{types.RoleGuest, types.RoleUser, types.RoleMember, types.RoleStaff, types.RoleAdmin} {
		user := &types.User{ID: "user-1", Role: role}
		status := types.OrderPending
		if !user.HasMinimumRole(types.RoleMember) {
			status = types.OrderRequested
		}
		t.Run(string(role), func(t *testing.T) {
			orders, cart := new(MockOrderService), new(MockCartService)
			notifications := checkoutNotifications{sent: make(chan services.HtmlTemplate, 1)}
			h := &OrderRoutes{
				config: types.PaymentOptions{OnDelivery: true}, orderService: orders,
				cartService: cart, userService: checkoutUsers{}, notificationService: notifications,
			}
			orders.On("GetOrderByIDAndUser", mock.Anything, "order-1").Return(types.Order{
				ID: "order-1", UserID: "user-1", Status: utilities.Ptr(types.OrderDraft),
			}, nil).Once()
			orders.On("UpdateOrder", mock.Anything, mock.MatchedBy(func(order *types.Order) bool {
				return order.ID == "order-1" && *order.Status == status && *order.PaymentMethod == types.PaymentMethodOnDelivery
			})).Return(nil).Once()
			cart.On("RemoveItems", mock.Anything).Return(nil).Once()
			r := httptest.NewRequest(http.MethodPost, "/orders/order-1/pay-on-delivery", nil)
			r.SetPathValue("id", "order-1")
			w := httptest.NewRecorder()
			r = r.WithContext(context.WithValue(r.Context(), services.UserKey, user))
			h.CheckoutOnDelivery(w, r)
			require.Equal(t, http.StatusOK, w.Code)
			require.JSONEq(t, `{"order_id":"order-1"}`, w.Body.String())
			select {
			case got := <-notifications.sent:
				require.Equal(t, services.NotifyOrderConf, got)
			case <-time.After(time.Second):
				t.Fatal("notification not sent")
			}
			orders.AssertExpectations(t)
			cart.AssertExpectations(t)
		})
	}
}

func TestCheckoutOnDeliveryDisabled(t *testing.T) {
	h := &OrderRoutes{}
	w := httptest.NewRecorder()
	h.CheckoutOnDelivery(w, httptest.NewRequest(http.MethodPost, "/", nil))
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCheckoutOnDeliveryRejectsUnavailableOrders(t *testing.T) {
	for _, tc := range []struct {
		name  string
		order types.Order
		err   error
		code  int
	}{
		{"not owned or missing", types.Order{}, types.ErrNotFound, http.StatusNotFound},
		{"already paid", types.Order{Status: utilities.Ptr(types.OrderPaid)}, nil, http.StatusConflict},
		{"already canceled", types.Order{Status: utilities.Ptr(types.OrderCanceled)}, nil, http.StatusConflict},
	} {
		t.Run(tc.name, func(t *testing.T) {
			orders := new(MockOrderService)
			h := &OrderRoutes{config: types.PaymentOptions{OnDelivery: true}, orderService: orders}
			orders.On("GetOrderByIDAndUser", mock.Anything, "order-1").Return(tc.order, tc.err).Once()
			r := httptest.NewRequest(http.MethodPost, "/", nil)
			r = r.WithContext(context.WithValue(r.Context(), services.UserKey, &types.User{ID: "user-1", Role: types.RoleGuest}))
			r.SetPathValue("id", "order-1")
			w := httptest.NewRecorder()
			h.CheckoutOnDelivery(w, r)
			require.Equal(t, tc.code, w.Code)
			orders.AssertExpectations(t)
			orders.AssertNotCalled(t, "UpdateOrder", mock.Anything, mock.Anything)
		})
	}
}
