package services

import (
	"context"
	"log/slog"
	"time"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
)

const (
	tolerance = time.Minute * 5 // Maximum allowed time difference between Stripe's timestamp and server time
)

type OrderService interface {
	CreateOrder(ctx context.Context) (types.Order, error)
	UpdateOrder(ctx context.Context, order types.OrderParams) (types.Order, error)
	GetOrders(ctx context.Context, page, limit int) ([]types.Order, error)
	GetOrderByID(ctx context.Context, orderID string) (types.Order, error)
	GetOrderForUser(ctx context.Context, orderID string) (types.Order, error)
	CancelStaleOrders(ctx context.Context) error
	GetPendingOrderForUser(ctx context.Context) (types.Order, error)
}

type orderService struct {
	orderRepo      repositories.OrderRepository
	cartRepo       repositories.CartRepository
	HttpClient     utilities.HTTPClient
	locConfig      types.LocaleConfig
	paymentService PaymentService
}

func NewOrderService(
	orderRepo repositories.OrderRepository,
	cartRepo repositories.CartRepository,
	paymentService PaymentService,
	locConfig types.LocaleConfig,
	httpClient utilities.HTTPClient,
) OrderService {
	if httpClient == nil {
		httpClient = utilities.NewDefaultHTTPClient(10 * time.Second)
	}
	return &orderService{
		orderRepo:      orderRepo,
		cartRepo:       cartRepo,
		HttpClient:     httpClient,
		paymentService: paymentService,
		locConfig:      locConfig,
	}
}

func (os *orderService) UpdateOrder(ctx context.Context, params types.OrderParams) (types.Order, error) {
	params.UserID = getUserID(ctx)
	return os.orderRepo.UpdateOrder(ctx, params)
}

func (os *orderService) CancelStaleOrders(ctx context.Context) error {
	// Cancel orders older than 10 minutes with status "pending"
	// TODO make this configurable to match expires_at in payments
	interval := 10 * time.Minute
	return os.orderRepo.CancelPendingOrders(ctx, interval)
}

func (os *orderService) GetOrders(ctx context.Context, page, limit int) ([]types.Order, error) {
	return os.orderRepo.GetOrders(ctx, page, limit)
}

func (os *orderService) CreateOrder(ctx context.Context) (types.Order, error) {
	var order types.Order
	order.UserID = getUserID(ctx)
	order.Currency = os.locConfig.Currency
	order.Status = types.OrderPending

	id, err := utilities.GenerateIDString()
	if err != nil {
		return order, err
	}
	order.ID = id

	if err := os.orderRepo.CreateOrder(ctx, &order); err != nil {
		slog.Debug("Error creating order", "user_id", order.UserID, "error", err)
		return order, err
	}
	slog.Info("Order created",
		"order_id", order.ID,
		"user_id", order.UserID,
		"currency", order.Currency,
		"amount", order.Amount,
	)
	return order, nil
}

func (os *orderService) GetOrderByID(ctx context.Context, orderID string) (types.Order, error) {
	return os.orderRepo.GetOrderByID(ctx, orderID)
}

func (os *orderService) GetOrderForUser(ctx context.Context, orderID string) (types.Order, error) {
	return os.orderRepo.GetOrderForUser(ctx, orderID, getUserID(ctx))
}

func (os *orderService) GetPendingOrderForUser(ctx context.Context) (types.Order, error) {
	userID := getUserID(ctx)
	return os.orderRepo.GetPendingOrder(ctx, userID)
}
