package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/dgyurics/marketplace/services"
)

type AuthMiddleware interface {
	Middleware(next http.Handler) http.Handler
}

type authMiddleware struct {
	authService services.AuthService
}

func NewAuthMiddleware(authService services.AuthService) AuthMiddleware {
	return &authMiddleware{authService}
}

func (a *authMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the token from the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		// Check for Bearer token format
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		// Validate the token using AuthService
		userID, err := a.authService.ValidateAccessToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Store the user ID in the context for downstream handlers
		ctx := context.WithValue(r.Context(), "userID", userID) // TODO look into using a custom type for context key, e..g const userIDKey = contextKey("userID")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
