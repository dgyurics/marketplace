package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	"github.com/stretchr/testify/assert"
)

// MockJWTService simulates JWTService behavior for testing
type MockJWTService struct {
	ParseTokenFunc func(token string) (*types.User, error)
}

func (m *MockJWTService) GenerateToken(user types.User) (string, error) {
	return "", errors.New("not implemented")
}

func (m *MockJWTService) ParseToken(token string) (*types.User, error) {
	if m.ParseTokenFunc != nil {
		return m.ParseTokenFunc(token)
	}
	return nil, errors.New("invalid token")
}

func TestAuthenticateUser_ValidToken(t *testing.T) {
	mockJWTService := &MockJWTService{
		ParseTokenFunc: func(token string) (*types.User, error) {
			return &types.User{ID: "123", Email: "test@example.com"}, nil
		},
	}
	auth := NewAccessControl(mockJWTService)

	// Create a test request with a valid token
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()

	// Mock next handler to verify user context
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(services.UserKey).(*types.User)
		assert.True(t, ok, "expected user to be stored in context")
		assert.NotNil(t, user, "user should not be nil")
		assert.Equal(t, "123", user.ID, "expected user ID to be 123")
		w.WriteHeader(http.StatusOK)
	})

	// Call AuthenticateUser
	handler := auth.RequireRole(types.RoleGuest)(nextHandler)
	handler.ServeHTTP(rr, req)

	// Verify the response status code
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAuthenticateUser_InvalidToken(t *testing.T) {
	mockJWTService := &MockJWTService{
		ParseTokenFunc: func(token string) (*types.User, error) {
			return nil, errors.New("invalid token")
		},
	}
	auth := NewAccessControl(mockJWTService)

	// Create a test request with an invalid token
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()

	// Mock next handler to ensure it is not called
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called for invalid token")
	})

	// Call AuthenticateUser
	handler := auth.RequireRole(types.RoleUser)(nextHandler)
	handler.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthenticateAdmin_ValidAdminToken(t *testing.T) {
	mockJWTService := &MockJWTService{
		ParseTokenFunc: func(token string) (*types.User, error) {
			return &types.User{ID: "123", Email: "admin@example.com", Role: "admin"}, nil
		},
	}
	auth := NewAccessControl(mockJWTService)

	// Create a test request with a valid admin token
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-admin-token")
	rr := httptest.NewRecorder()

	// Mock next handler to verify user context
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(services.UserKey).(*types.User)
		assert.True(t, ok, "expected user in context")
		assert.NotNil(t, user, "user should not be nil")
		assert.Equal(t, "123", user.ID)
		assert.True(t, user.Role == types.RoleAdmin, "user should be admin")
		w.WriteHeader(http.StatusOK)
	})

	// Call AuthenticateAdmin
	handler := auth.RequireRole(types.RoleAdmin)(nextHandler)
	handler.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAuthenticateAdmin_NonAdminToken(t *testing.T) {
	mockJWTService := &MockJWTService{
		ParseTokenFunc: func(token string) (*types.User, error) {
			return &types.User{ID: "456", Email: "user@example.com", Role: "user"}, nil
		},
	}
	auth := NewAccessControl(mockJWTService)

	// Create a test request with a non-admin token
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-user-token")
	rr := httptest.NewRecorder()

	// Mock next handler to ensure it is not called
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called for non-admin user")
	})

	// Call AuthenticateAdmin
	handler := auth.RequireRole(types.RoleAdmin)(nextHandler)
	handler.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestAuthenticateAdmin_InvalidToken(t *testing.T) {
	mockJWTService := &MockJWTService{
		ParseTokenFunc: func(token string) (*types.User, error) {
			return nil, errors.New("invalid token")
		},
	}
	auth := NewAccessControl(mockJWTService)

	// Create a test request with an invalid token
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()

	// Mock next handler to ensure it is not called
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called for invalid token")
	})

	// Call AuthenticateAdmin
	handler := auth.RequireRole(types.RoleAdmin)(nextHandler)
	handler.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthenticateUser_GuestUser(t *testing.T) {
	mockJWTService := &MockJWTService{
		ParseTokenFunc: func(token string) (*types.User, error) {
			return &types.User{ID: "789", Role: "guest"}, nil
		},
	}
	auth := NewAccessControl(mockJWTService)

	// Create a test request with a guest token
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer guest-token")
	rr := httptest.NewRecorder()

	// Mock next handler to verify guest user context
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(services.UserKey).(*types.User)
		assert.True(t, ok, "expected user to be stored in context")
		assert.NotNil(t, user, "user should not be nil")
		assert.Equal(t, "789", user.ID, "expected user ID to be 789")
		assert.Equal(t, types.RoleGuest, user.Role, "expected role to be 'guest'")
		w.WriteHeader(http.StatusOK)
	})

	// Call AuthenticateUser
	handler := auth.RequireRole(types.RoleGuest)(nextHandler)
	handler.ServeHTTP(rr, req)

	// Verify the response status code
	assert.Equal(t, http.StatusOK, rr.Code)
}
