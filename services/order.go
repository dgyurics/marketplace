package services

import (
	"context"
	"log/slog"
	"sync"
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
	GetOrder(ctx context.Context, orderID string) (types.Order, error)
	CancelStaleOrders(ctx context.Context)
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

func (os *orderService) CancelStaleOrders(ctx context.Context) {
	interval := 30 * time.Minute // Cancel orders older than 30 minutes with status "pending"
	paymentIntentIDs, err := os.orderRepo.CancelPendingOrders(ctx, interval)
	if err != nil {
		slog.Error("Error canceling stale orders", "error", err)
	}

	var wg sync.WaitGroup
	for _, id := range paymentIntentIDs {
		wg.Add(1)
		go func(pID string) {
			defer wg.Done()
			slog.Debug("Cancelling payment intent", "id", id)
			if err := os.paymentService.CancelPaymentIntent(ctx, pID); err != nil {
				slog.Error("Error canceling payment intent", "id", pID)
			}
		}(id)
	}
	wg.Wait()
}

// GetOrders retrieves a list of orders for the current user, with pagination support.
func (os *orderService) GetOrders(ctx context.Context, page, limit int) ([]types.Order, error) {
	var userID = getUserID(ctx)
	orders, err := os.orderRepo.GetOrders(ctx, userID, page, limit)
	if err != nil {
		return nil, err
	}
	return orders, nil
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

func (os *orderService) GetOrder(ctx context.Context, orderID string) (types.Order, error) {
	return os.orderRepo.GetOrder(ctx, orderID, getUserID(ctx))
}

func (os *orderService) GetPendingOrderForUser(ctx context.Context) (types.Order, error) {
	userID := getUserID(ctx)
	return os.orderRepo.GetPendingOrder(ctx, userID)
}
