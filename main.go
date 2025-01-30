package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dgyurics/marketplace/db"
	"github.com/dgyurics/marketplace/middleware"
	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/routes"
	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/utilities"
	"github.com/gorilla/mux"
)

func main() {
	// Initialize logger
	utilities.InitLogger(utilities.LoadLoggerConfig())
	defer utilities.CloseLogger()

	// Initialize and start server
	server := initializeServer()
	server.ListenAndServe()
	gracefulShutdown(server)
}

// initializeServer sets up the database, services, and HTTP server
func initializeServer() *http.Server {
	// connect to database
	db := db.Connect()

	// create database repositories
	authRepository := repositories.NewAuthRepository(db)
	userRepository := repositories.NewUserRepository(db)
	categoryRepository := repositories.NewCategoryRepository(db)
	productRepository := repositories.NewProductRepository(db)
	cartRepository := repositories.NewCartRepository(db)
	orderRepository := repositories.NewOrderRepository(db)

	// create services
	authService := services.NewAuthService(authRepository, utilities.LoadAuthConfig())
	userService := services.NewUserService(userRepository)
	categoryService := services.NewCategoryService(categoryRepository)
	orderService := services.NewOrderService(orderRepository, cartRepository, utilities.LoadOrderConfig(), nil)
	productService := services.NewProductService(productRepository)
	cartService := services.NewCartService(cartRepository)

	// create router
	router := mux.NewRouter()

	// add middleware
	router.Use(middleware.LimitBodySizeMiddleware)
	router.Use(middleware.LoggingMiddleware)

	// create base router which encapsulates the primary router and access control middleware
	baseRouter := routes.NewRouter(router, middleware.NewAccessControl(authService))

	// create routes
	categoryRoutes := routes.NewCategoryRoutes(categoryService, baseRouter)
	userRoutes := routes.NewUserRoutes(userService, authService, baseRouter)
	productRoutes := routes.NewProductRoutes(productService, baseRouter)
	cartRoutes := routes.NewCartRoutes(cartService, baseRouter)
	orderRoutes := routes.NewOrderRoutes(orderService, baseRouter)

	// register routes with main router
	routes.RegisterAllRoutes(
		userRoutes,
		categoryRoutes,
		productRoutes,
		cartRoutes,
		orderRoutes,
	)

	// Create and return the HTTP server
	server := &http.Server{
		Addr:           ":8000",
		Handler:        router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 0, // DefaultMaxHeaderBytes used if 0
		ErrorLog:       log.New(&utilities.ErrorLog{}, "", 0),
	}
	slog.Info("Server initialized", "port", 8000)
	return server
}

// gracefulShutdown handles termination signals and gracefully shuts down the server
func gracefulShutdown(server *http.Server) {
	// Listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Wait for a signal
	<-stop
	slog.Info("Shutdown signal received")

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Gracefully shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server shutdown failed", "error", err.Error())
	} else {
		slog.Info("Server gracefully stopped")
	}
}
