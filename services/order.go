package services

import (
	"context"
	"log/slog"
	"time"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
)

type OrderService interface {
	CreateOrder(ctx context.Context, order *types.Order) error
	GetOrderByIDAndUser(ctx context.Context, orderID string) (types.Order, error)
	GetOrderByID(ctx context.Context, orderID string) (types.Order, error)
	GetOrderByIDPublic(ctx context.Context, orderID string) (types.Order, error)
	GetOrders(ctx context.Context, page, limit int) ([]types.Order, error)
}

type orderService struct {
	orderRepo      repositories.OrderRepository
	cartRepo       repositories.CartRepository
	HttpClient     utilities.HTTPClient
	paymentService PaymentService
}

func NewOrderService(
	orderRepo repositories.OrderRepository,
	cartRepo repositories.CartRepository,
	paymentService PaymentService,
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
	}
}

func (os *orderService) GetOrders(ctx context.Context, page, limit int) ([]types.Order, error) {
	return os.orderRepo.GetOrders(ctx, page, limit)
}

func (os *orderService) CreateOrder(ctx context.Context, order *types.Order) (err error) {
	order.UserID = getUserID(ctx)
	order.ID, err = utilities.GenerateIDString()
	if err != nil {
		return err
	}

	if order.TotalAmount == 0 {
		return types.ErrConstraintViolation
	}

	if err = os.orderRepo.CreateOrder(ctx, order); err != nil {
		slog.Debug("Error creating order", "user_id", order.UserID, "error", err)
		return err
	}

	slog.Info("Order created", "order_id", order.ID, "user_id", order.UserID, "total_amount", order.TotalAmount)
	return nil
}

func (os *orderService) GetOrderByID(ctx context.Context, orderID string) (types.Order, error) {
	return os.orderRepo.GetOrderByID(ctx, orderID)
}

func (os *orderService) GetOrderByIDAndUser(ctx context.Context, orderID string) (types.Order, error) {
	return os.orderRepo.GetOrderByIDAndUser(ctx, orderID, getUserID(ctx))
}

func (os *orderService) GetOrderByIDPublic(ctx context.Context, orderID string) (types.Order, error) {
	return os.orderRepo.GetOrderByIDPublic(ctx, orderID)
}
