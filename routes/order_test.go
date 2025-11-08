package routes

import (
	"context"

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
