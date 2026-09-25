package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
)

type Authorizer interface {
	RequireRole(role types.Role) func(next http.HandlerFunc) http.HandlerFunc
	OptionalAuth(next http.HandlerFunc) http.HandlerFunc
}

type authorizer struct {
	jwtService services.JWTService
}

func NewAccessControl(jwtService services.JWTService) Authorizer {
	return &authorizer{jwtService}
}

// RequireRole authenticates a user.
// Upon successful authentication, checks if the user has a role equal to or higher than the specified.
// The role hierarchy is defined in types.Role, where higher roles have more privileges.
func (a *authorizer) RequireRole(role types.Role) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			usr, err := a.authenticateToken(r)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if !usr.HasMinimumRole(role) {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			ctx := context.WithValue(r.Context(), services.UserKey, &usr)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}

// OptionalAuth attaches the user to the context when a valid token is present.
// Requests without a token, or with an invalid one, proceed unauthenticated.
func (a *authorizer) OptionalAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		usr, err := a.authenticateToken(r)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), services.UserKey, &usr)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// authenticateToken checks the Authorization header for a token,
// and validates it using the authService. If the token is valid,
// the user is returned. If the token is invalid, an error is returned.
func (a *authorizer) authenticateToken(r *http.Request) (types.User, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return types.User{}, errors.New("authorization header missing")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return types.User{}, errors.New("invalid token format")
	}

	usr, err := a.jwtService.ParseToken(tokenString)
	if err != nil {
		return types.User{}, fmt.Errorf("invalid or expired token: %w", err)
	}
	return *usr, nil
}
