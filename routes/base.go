package routes

import (
	"net/http"

	"github.com/dgyurics/marketplace/middleware"
	"github.com/gorilla/mux"
)

// BaseRouter defines the common interface for all routers
type BaseRouter interface {
	RegisterRoutes()
}

// Common router structure
type router struct {
	muxRouter      *mux.Router
	authMiddleware middleware.Authorizer
}

func NewRouter(muxRouter *mux.Router, authMiddleware middleware.Authorizer) router {
	return router{
		muxRouter:      muxRouter,
		authMiddleware: authMiddleware,
	}
}

// restrict endpoint to authenticated users
func (h *router) secure(next http.HandlerFunc) http.Handler {
	return h.authMiddleware.AuthenticateUser(next)
}

// restrict endpoint to admin users
func (h *router) secureAdmin(next http.HandlerFunc) http.Handler {
	return h.authMiddleware.AuthenticateAdmin(next)
}

// registers all routes for the router
func RegisterAllRoutes(routes ...BaseRouter) {
	for _, route := range routes {
		route.RegisterRoutes()
	}
}
