package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
)

type MockRateLimitRepository struct {
	mock.Mock
}

func (m *MockRateLimitRepository) GetHitCount(ctx context.Context, rl *types.RateLimit) error {
	args := m.Called(ctx, rl)
	return args.Error(0)
}

func (m *MockRateLimitRepository) RecordHit(ctx context.Context, rl *types.RateLimit) error {
	args := m.Called(ctx, rl)
	return args.Error(0)
}

func (m *MockRateLimitRepository) Cleanup(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestRateLimitService_GetHitCount(t *testing.T) {
	mockRepo := new(MockRateLimitRepository)
	service := services.NewRateLimitService(mockRepo)

	rl := &types.RateLimit{
		IPAddress: "192.168.1.100",
		Path:      "/users/login",
	}

	mockRepo.On("GetHitCount", mock.Anything, rl).Return(nil)

	err := service.GetHitCount(context.Background(), rl)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestRateLimitService_RecordHit(t *testing.T) {
	mockRepo := new(MockRateLimitRepository)
	service := services.NewRateLimitService(mockRepo)

	rl := &types.RateLimit{
		IPAddress: "192.168.1.100",
		Path:      "/users/login",
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}

	mockRepo.On("RecordHit", mock.Anything, rl).Return(nil)

	err := service.RecordHit(context.Background(), rl)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestRateLimitService_Cleanup(t *testing.T) {
	mockRepo := new(MockRateLimitRepository)
	service := services.NewRateLimitService(mockRepo)

	mockRepo.On("Cleanup", mock.Anything).Return(nil)

	err := service.Cleanup(context.Background())

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestRateLimitService_Cleanup_Error(t *testing.T) {
	mockRepo := new(MockRateLimitRepository)
	service := services.NewRateLimitService(mockRepo)

	mockRepo.On("Cleanup", mock.Anything).Return(errors.New("database error"))

	err := service.Cleanup(context.Background())

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}
