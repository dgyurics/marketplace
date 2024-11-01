package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/services"
)

type AuthMiddleware interface {
	AuthenticateUser(next http.Handler) http.Handler
	AuthenticateAdmin(next http.Handler) http.Handler
}

type authMiddleware struct {
	authService services.AuthService
}

func NewAuthMiddleware(authService services.AuthService) AuthMiddleware {
	return &authMiddleware{authService}
}

// verifies Authorization header token and allows access only for users.
func (a *authMiddleware) AuthenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := a.authenticateToken(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), services.UserKey, &user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// verifies Authorization header token and allows access only for admin users.
func (a *authMiddleware) AuthenticateAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := a.authenticateToken(r)
		if err != nil || !user.Admin {
			http.Error(w, "Admin access required", http.StatusForbidden)
			return
		}
		ctx := context.WithValue(r.Context(), services.UserKey, &user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extracts and validates the token, returning the user if valid.
func (a *authMiddleware) authenticateToken(r *http.Request) (models.User, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return models.User{}, errors.New("authorization header missing")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return models.User{}, errors.New("invalid token format")
	}

	user, err := a.authService.ValidateAccessToken(tokenString)
	if err != nil {
		return models.User{}, errors.New("invalid or expired token")
	}
	return user, nil
}
