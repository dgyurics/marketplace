package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dgyurics/marketplace/types"
	"github.com/stretchr/testify/assert"
)

// MockRateLimitService simulates RateLimitService behavior for testing
type MockRateLimitService struct {
	GetHitCountFunc func(ctx context.Context, rl *types.RateLimit) error
	RecordHitFunc   func(ctx context.Context, rl *types.RateLimit) error
}

func (m *MockRateLimitService) GetHitCount(ctx context.Context, rl *types.RateLimit) error {
	if m.GetHitCountFunc != nil {
		return m.GetHitCountFunc(ctx, rl)
	}
	return nil
}

func (m *MockRateLimitService) RecordHit(ctx context.Context, rl *types.RateLimit) error {
	if m.RecordHitFunc != nil {
		return m.RecordHitFunc(ctx, rl)
	}
	return nil
}

func (m *MockRateLimitService) PurgeExpiredEntries(ctx context.Context) error {
	return nil
}

func (m *MockRateLimitService) Cleanup(ctx context.Context) error {
	return nil
}

func TestLimit_WithinLimit(t *testing.T) {
	mockService := &MockRateLimitService{
		GetHitCountFunc: func(ctx context.Context, rl *types.RateLimit) error {
			rl.HitCount = 3 // Below limit
			return nil
		},
	}
	rateLimit := NewRateLimit(mockService)

	req := httptest.NewRequest(http.MethodPost, "/users/login", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	rr := httptest.NewRecorder()

	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := rateLimit.Limit(nextHandler, 5)
	handler.ServeHTTP(rr, req)

	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestLimit_LimitReached(t *testing.T) {
	mockService := &MockRateLimitService{
		GetHitCountFunc: func(ctx context.Context, rl *types.RateLimit) error {
			rl.HitCount = 6 // Above limit
			return nil
		},
	}
	rateLimit := NewRateLimit(mockService)

	req := httptest.NewRequest(http.MethodPost, "/users/login", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	rr := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Next handler should not be called when limit is exceeded")
	})

	handler := rateLimit.Limit(nextHandler, 5)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
}

func TestLimit_ServiceError(t *testing.T) {
	mockService := &MockRateLimitService{
		GetHitCountFunc: func(ctx context.Context, rl *types.RateLimit) error {
			return errors.New("database error")
		},
	}
	rateLimit := NewRateLimit(mockService)

	req := httptest.NewRequest(http.MethodPost, "/users/login", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	rr := httptest.NewRecorder()

	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := rateLimit.Limit(nextHandler, 5)
	handler.ServeHTTP(rr, req)

	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestLimitAndRecordHit_Success(t *testing.T) {
	getHitCountCalled := false
	recordHitCalled := false
	mockService := &MockRateLimitService{
		GetHitCountFunc: func(ctx context.Context, rl *types.RateLimit) error {
			getHitCountCalled = true
			rl.HitCount = 3 // Below limit
			return nil
		},
		RecordHitFunc: func(ctx context.Context, rl *types.RateLimit) error {
			recordHitCalled = true
			return nil
		},
	}
	rateLimit := NewRateLimit(mockService)

	req := httptest.NewRequest(http.MethodPost, "/users/login", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	rr := httptest.NewRecorder()

	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := rateLimit.LimitAndRecordHit(nextHandler, 5, time.Hour)
	handler.ServeHTTP(rr, req)

	assert.True(t, getHitCountCalled)
	assert.True(t, recordHitCalled)
	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestLimitAndRecordHit_ExceedsLimit(t *testing.T) {
	mockService := &MockRateLimitService{
		GetHitCountFunc: func(ctx context.Context, rl *types.RateLimit) error {
			rl.HitCount = 6 // Above limit
			return nil
		},
	}
	rateLimit := NewRateLimit(mockService)

	req := httptest.NewRequest(http.MethodPost, "/users/login", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	rr := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Next handler should not be called when limit is exceeded")
	})

	handler := rateLimit.LimitAndRecordHit(nextHandler, 5, time.Hour)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
}

func TestRecordHit_Success(t *testing.T) {
	recordCalled := false
	mockService := &MockRateLimitService{
		RecordHitFunc: func(ctx context.Context, rl *types.RateLimit) error {
			recordCalled = true
			return nil
		},
	}
	rateLimit := NewRateLimit(mockService)

	req := httptest.NewRequest(http.MethodPost, "/users/login", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	rateLimit.RecordHit(req, time.Hour)

	assert.True(t, recordCalled)
}

func TestGetClientIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 192.168.1.1")
	req.RemoteAddr = "127.0.0.1:12345"

	ip := getClientIP(req)
	assert.Equal(t, "203.0.113.1", ip)
}

func TestGetClientIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.200:8080"

	ip := getClientIP(req)
	assert.Equal(t, "192.168.1.200", ip)
}
