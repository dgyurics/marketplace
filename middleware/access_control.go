package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
)

type Authorizer interface {
	AuthenticateUser(next http.HandlerFunc) http.HandlerFunc
	AuthenticateAdmin(next http.HandlerFunc) http.HandlerFunc
	RequireRole(role types.Role) func(next http.HandlerFunc) http.HandlerFunc
}

type authorizer struct {
	jwtService services.JWTService
}

func NewAccessControl(jwtService services.JWTService) *authorizer {
	return &authorizer{jwtService}
}

// RequireRole authenticates a user.
// Upon successful authentication, checks if the user has a role equal to or higher than the specified.
// The role hierarchy is defined in types.Role, where higher roles have more privileges.
func (a *authorizer) RequireRole(role types.Role) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := a.authenticateToken(r)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if !user.HasMinimumRole(role) {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			ctx := context.WithValue(r.Context(), services.UserKey, &user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthenticateUser verifies the Authorization header.
// If the token is valid, the user is stored in the request context.
// If the token is invalid, or does not exist, a 401 Unauthorized response is returned.
func (a *authorizer) AuthenticateUser(next http.HandlerFunc) http.HandlerFunc {
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

// AuthenticateAdmin verifies the Authorization header.
// If the token is valid and the user is an admin, the user is stored in the request context.
// If the token is invalid, or does not exist, a 401 Unauthorized response is returned.
// If the user is not an admin, a 403 Forbidden response is returned.
func (a *authorizer) AuthenticateAdmin(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := a.authenticateToken(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if !user.IsAdmin() {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		ctx := context.WithValue(r.Context(), services.UserKey, &user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
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

	user, err := a.jwtService.ParseToken(tokenString)
	if err != nil {
		return types.User{}, errors.New("invalid or expired token")
	}
	return *user, nil
}
