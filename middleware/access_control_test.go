package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/services"
	"github.com/stretchr/testify/assert"
)

type MockAuthService struct {
	ValidateTokenFunc func(token string) (models.User, error)
}

func (m *MockAuthService) ValidateAccessToken(token string) (models.User, error) {
	return m.ValidateTokenFunc(token)
}

func (m *MockAuthService) GenerateAccessToken(user models.User) (string, error) { return "", nil }
func (m *MockAuthService) GenerateRefreshToken() (string, error)                { return "", nil }
func (m *MockAuthService) ValidateRefreshToken(ctx context.Context, token string) (models.User, error) {
	return models.User{}, nil
}
func (m *MockAuthService) StoreRefreshToken(ctx context.Context, userID, token string) error {
	return nil
}
func (m *MockAuthService) RevokeRefreshTokens(ctx context.Context) error { return nil }

func TestAuthenticateUser_ValidToken(t *testing.T) {
	mockAuthService := &MockAuthService{
		ValidateTokenFunc: func(token string) (models.User, error) {
			return models.User{ID: "123", Email: "test@example.com"}, nil
		},
	}
	auth := NewAccessControl(mockAuthService)

	// Create a test request with a valid token
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()

	// Mock next handler to verify user context
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(services.UserKey).(*models.User)
		if !ok || user.ID != "123" {
			t.Errorf("expected user ID 123 in context, got %+v", user)
		}
		w.WriteHeader(http.StatusOK)
	})

	// Call AuthenticateUser
	handler := auth.AuthenticateUser(nextHandler)
	handler.ServeHTTP(rr, req)

	// Verify the response status code
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAuthenticateUser_InvalidToken(t *testing.T) {
	mockAuthService := &MockAuthService{
		ValidateTokenFunc: func(token string) (models.User, error) {
			return models.User{}, errors.New("invalid token")
		},
	}
	auth := NewAccessControl(mockAuthService)

	// Create a test request with an invalid token
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()

	// Mock next handler to ensure it is not called
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called for invalid token")
	})

	// Call AuthenticateUser
	handler := auth.AuthenticateUser(nextHandler)
	handler.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthenticateAdmin_ValidAdminToken(t *testing.T) {
	mockAuthService := &MockAuthService{
		ValidateTokenFunc: func(token string) (models.User, error) {
			return models.User{ID: "123", Email: "admin@example.com", Admin: true}, nil
		},
	}
	auth := NewAccessControl(mockAuthService)

	// Create a test request with a valid admin token
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-admin-token")
	rr := httptest.NewRecorder()

	// Mock next handler to verify user context
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(services.UserKey).(*models.User)
		assert.True(t, ok, "expected user in context")
		assert.Equal(t, "123", user.ID)
		w.WriteHeader(http.StatusOK)
	})

	// Call AuthenticateAdmin
	handler := auth.AuthenticateAdmin(nextHandler)
	handler.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAuthenticateAdmin_NonAdminToken(t *testing.T) {
	mockAuthService := &MockAuthService{
		ValidateTokenFunc: func(token string) (models.User, error) {
			return models.User{ID: "456", Email: "user@example.com", Admin: false}, nil
		},
	}
	auth := NewAccessControl(mockAuthService)

	// Create a test request with a non-admin token
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-user-token")
	rr := httptest.NewRecorder()

	// Mock next handler to ensure it is not called
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called for non-admin user")
	})

	// Call AuthenticateAdmin
	handler := auth.AuthenticateAdmin(nextHandler)
	handler.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestAuthenticateAdmin_InvalidToken(t *testing.T) {
	mockAuthService := &MockAuthService{
		ValidateTokenFunc: func(token string) (models.User, error) {
			return models.User{}, errors.New("invalid token")
		},
	}
	auth := NewAccessControl(mockAuthService)

	// Create a test request with an invalid token
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()

	// Mock next handler to ensure it is not called
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called for invalid token")
	})

	// Call AuthenticateAdmin
	handler := auth.AuthenticateAdmin(nextHandler)
	handler.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
