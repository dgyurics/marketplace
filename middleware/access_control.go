package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/services"
)

type AccessControl interface {
	AuthenticateUser(next http.HandlerFunc) http.Handler
	AuthenticateAdmin(next http.HandlerFunc) http.Handler
}

type accessControl struct {
	authService services.AuthService
}

func NewAccessControl(authService services.AuthService) AccessControl {
	return &accessControl{authService}
}

// verifies Authorization header token and allows access only for users.
func (a *accessControl) AuthenticateUser(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := a.authenticateToken(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), services.UserKey, &user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// verifies Authorization header token and allows access only for admin users.
func (a *accessControl) AuthenticateAdmin(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := a.authenticateToken(r)
		if err != nil || !user.Admin {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		ctx := context.WithValue(r.Context(), services.UserKey, &user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extracts and validates the token, returning the user if valid.
func (a *accessControl) authenticateToken(r *http.Request) (models.User, error) {
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
