package routes

import (
	"net/http"
	"time"

	"github.com/dgyurics/marketplace/middleware"
	"github.com/dgyurics/marketplace/types"
	"github.com/gorilla/mux"
)

// BaseRouter defines the common interface for all routers
type BaseRouter interface {
	RegisterRoutes()
}

// Common router structure
type router struct {
	muxRouter           *mux.Router
	authMiddleware      middleware.Authorizer
	rateLimitMiddleware middleware.RateLimit
}

func NewRouter(muxRouter *mux.Router, authMiddleware middleware.Authorizer, rateLimitMiddleware middleware.RateLimit) router {
	return router{
		muxRouter:           muxRouter,
		authMiddleware:      authMiddleware,
		rateLimitMiddleware: rateLimitMiddleware,
	}
}

// requireRole restricts endpoint access to users with the specified role or higher
func (h *router) requireRole(role types.Role) func(next http.HandlerFunc) http.HandlerFunc {
	return h.authMiddleware.RequireRole(role)
}

// restrict endpoint to authenticated users
func (h *router) secure(next http.HandlerFunc) http.HandlerFunc {
	return h.authMiddleware.AuthenticateUser(next)
}

// restrict endpoint to admin users
func (h *router) secureAdmin(next http.HandlerFunc) http.HandlerFunc {
	return h.authMiddleware.AuthenticateAdmin(next)
}

// Most common case - tracks automatically and enforces limit
func (h *router) limit(next http.HandlerFunc, limit int, expiry time.Duration) http.HandlerFunc {
	return h.rateLimitMiddleware.LimitAndRecordHit(next, limit, expiry)
}

// Special case - only guards, requires manual recordHit calls
func (h *router) guardLimit(next http.HandlerFunc, limit int) http.HandlerFunc {
	return h.rateLimitMiddleware.Limit(next, limit)
}

// Manual hit recording (pairs with guardLimit)
func (h *router) recordHit(r *http.Request, expiry time.Duration) {
	h.rateLimitMiddleware.RecordHit(r, expiry)
}

// registers all routes for the router
func RegisterAllRoutes(routes ...BaseRouter) {
	for _, route := range routes {
		route.RegisterRoutes()
	}
}
