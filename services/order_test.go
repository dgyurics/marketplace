package services

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/dgyurics/marketplace/types"
	"github.com/stretchr/testify/mock"
)

// mockOrderRepo implements the OrderRepository interface for testing
type mockOrderRepo struct {
	mock.Mock
}

// MockHTTPClient implements the utilities.HTTPClient interface for testing
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	resp, _ := args.Get(0).(*http.Response)
	return resp, args.Error(1)
}

func contextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserKey, &types.User{ID: userID})
}

func TestGetOrders_RepoError(t *testing.T) {
	mockRepo := new(mockOrderRepo)
	svc := &orderService{
		orderRepo: mockRepo,
	}

	ctx := contextWithUserID(context.Background(), "user-123")
	mockRepo.On("GetOrders", ctx, "user-123", 1, 10).Return(nil, errors.New("db error"))

	result, err := svc.GetOrders(ctx, 1, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}

	mockRepo.AssertExpectations(t)
}

func TestGetOrder_Success(t *testing.T) {
	mockRepo := new(mockOrderRepo)
	svc := &orderService{
		orderRepo: mockRepo,
	}

	userID := "user-123"
	orderID := "order-456"
	expectedOrder := types.Order{
		ID:     orderID,
		UserID: userID,
		Email:  "test@example.com",
	}

	ctx := contextWithUserID(context.Background(), userID)
	mockRepo.On("GetOrder", ctx, orderID, userID).Return(expectedOrder, nil)

	result, err := svc.GetOrder(ctx, orderID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ID != expectedOrder.ID {
		t.Errorf("expected order ID %s, got %s", expectedOrder.ID, result.ID)
	}
	if result.UserID != expectedOrder.UserID {
		t.Errorf("expected user ID %s, got %s", expectedOrder.UserID, result.UserID)
	}
	if result.Email != expectedOrder.Email {
		t.Errorf("expected email %s, got %s", expectedOrder.Email, result.Email)
	}

	mockRepo.AssertExpectations(t)
}

func (m *mockOrderRepo) GetOrder(ctx context.Context, orderID, userID string) (types.Order, error) {
	args := m.Called(ctx, orderID, userID)
	if v := args.Get(0); v != nil {
		return v.(types.Order), args.Error(1)
	}
	return types.Order{}, args.Error(1)
}

func TestGetOrder_NotFound(t *testing.T) {
	mockRepo := new(mockOrderRepo)
	svc := &orderService{
		orderRepo: mockRepo,
	}

	userID := "user-123"
	orderID := "missing-order"
	ctx := contextWithUserID(context.Background(), userID)

	mockRepo.On("GetOrder", ctx, orderID, userID).Return(types.Order{}, types.ErrNotFound)

	result, err := svc.GetOrder(ctx, orderID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != types.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if result.ID != "" {
		t.Errorf("expected empty order ID, got %s", result.ID)
	}

	mockRepo.AssertExpectations(t)
}

func TestUpdateOrder_Success(t *testing.T) {
	mockRepo := new(mockOrderRepo)
	svc := &orderService{
		orderRepo: mockRepo,
	}

	userID := "user-123"
	ctx := contextWithUserID(context.Background(), userID)

	params := types.OrderParams{
		ID:        "order-789",
		Email:     strPtr("newemail@example.com"),
		TaxAmount: int64Ptr(500),
	}

	expected := types.Order{
		ID:    "order-789",
		Email: "newemail@example.com",
	}

	paramsWithUser := params
	paramsWithUser.UserID = userID

	mockRepo.On("UpdateOrder", ctx, paramsWithUser).Return(expected, nil)

	result, err := svc.UpdateOrder(ctx, params)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ID != expected.ID {
		t.Errorf("expected order ID %s, got %s", expected.ID, result.ID)
	}
	if result.Email != expected.Email {
		t.Errorf("expected email %s, got %s", expected.Email, result.Email)
	}

	mockRepo.AssertExpectations(t)
}

func TestUpdateOrder_Error(t *testing.T) {
	mockRepo := new(mockOrderRepo)
	svc := &orderService{
		orderRepo: mockRepo,
	}

	userID := "user-123"
	ctx := contextWithUserID(context.Background(), userID)

	params := types.OrderParams{
		ID:    "order-789",
		Email: strPtr("fail@example.com"),
	}

	paramsWithUser := params
	paramsWithUser.UserID = userID

	mockRepo.On("UpdateOrder", ctx, paramsWithUser).Return(types.Order{}, errors.New("update failed"))

	_, err := svc.UpdateOrder(ctx, params)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "update failed" {
		t.Fatalf("expected 'update failed', got %v", err)
	}

	mockRepo.AssertExpectations(t)
}

func strPtr(s string) *string { return &s }
func int64Ptr(i int64) *int64 { return &i }

func (m *mockOrderRepo) GetOrders(ctx context.Context, userID string, page, limit int) ([]types.Order, error) {
	args := m.Called(ctx, userID, page, limit)
	var orders []types.Order
	if v := args.Get(0); v != nil {
		orders = v.([]types.Order)
	}
	return orders, args.Error(1)
}

func (m *mockOrderRepo) CreateOrder(ctx context.Context, order *types.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *mockOrderRepo) UpdateOrder(ctx context.Context, params types.OrderParams) (types.Order, error) {
	args := m.Called(ctx, params)
	if v := args.Get(0); v != nil {
		return v.(types.Order), args.Error(1)
	}
	return types.Order{}, args.Error(1)
}

func (m *mockOrderRepo) CancelPendingOrders(ctx context.Context, interval time.Duration) ([]string, error) {
	return nil, nil
}
