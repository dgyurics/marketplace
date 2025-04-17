package main

import (
	"context"
	"database/sql"
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
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
	"github.com/gorilla/mux"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	// Load configuration
	config := utilities.LoadConfig()

	// Initialize logger
	utilities.InitLogger(config.Logger)
	defer utilities.CloseLogger()

	// Initialize database
	db := db.Connect(config.Database)

	// Initialize services
	services := initializeServices(db, config)

	// Initialize and start server
	server := initializeServer(config, services)
	server.ListenAndServe()
	gracefulShutdown(server)
}

// initializeServer sets up the database, services, and HTTP server
func initializeServer(config types.Config, services servicesContainer) *http.Server {
	// create router
	router := mux.NewRouter()
	router.Use(middleware.RequestLoggerMiddleware)
	router.Use(middleware.CORSMiddleware(config.CORS))
	router.Use(middleware.LimitBodySizeMiddleware)

	// create base router
	baseRouter := routes.NewRouter(router, middleware.NewAccessControl(services.JWT))

	// create routes
	routes.RegisterAllRoutes(
		routes.NewAddressRoutes(services.Address, baseRouter),
		routes.NewUserRoutes(services.User, services.Invite, services.JWT, services.Refresh, config.Auth, baseRouter),
		routes.NewCategoryRoutes(services.Category, baseRouter),
		routes.NewProductRoutes(services.Product, baseRouter),
		routes.NewCartRoutes(services.Cart, baseRouter),
		routes.NewOrderRoutes(services.Order, baseRouter),
		routes.NewPasswordRoutes(services.Password, services.User, services.Email, services.Template, config.BaseURL, baseRouter),
	)

	// Create and return the server
	srvCfg := config.Server
	return &http.Server{
		Addr:           srvCfg.Addr,
		Handler:        router,
		ReadTimeout:    srvCfg.ReadTimeout,
		WriteTimeout:   srvCfg.WriteTimeout,
		IdleTimeout:    srvCfg.IdleTimeout,
		MaxHeaderBytes: srvCfg.MaxHeaderBytes,
		ErrorLog:       srvCfg.ErrorLog,
	}
}

// initializeServices creates all services and repositories
func initializeServices(db *sql.DB, config types.Config) servicesContainer {
	// create database repositories
	addressRepository := repositories.NewAddressRepository(db)
	userRepository := repositories.NewUserRepository(db)
	categoryRepository := repositories.NewCategoryRepository(db)
	productRepository := repositories.NewProductRepository(db)
	cartRepository := repositories.NewCartRepository(db)
	orderRepository := repositories.NewOrderRepository(db)
	inviteRepository := repositories.NewInviteRepository(db)
	passwordRepository := repositories.NewPasswordRepository(db)
	refreshTokenRepository := repositories.NewRefreshRepository(db)

	// create http client required by certain services
	httpClient := utilities.NewDefaultHTTPClient(10 * time.Second)

	// create services
	addressService := services.NewAddressService(addressRepository)
	userService := services.NewUserService(userRepository)
	categoryService := services.NewCategoryService(categoryRepository)
	productService := services.NewProductService(productRepository)
	cartService := services.NewCartService(cartRepository)
	orderService := services.NewOrderService(orderRepository, cartRepository, config.Order, httpClient)
	inviteService := services.NewInviteService(inviteRepository, config.Auth.HMACSecret)
	passwordService := services.NewPasswordService(passwordRepository, config.Auth.HMACSecret)
	refreshService := services.NewRefreshService(refreshTokenRepository, config.Auth)
	emailService := services.NewMailjetSender(config.Email)
	jwtService := services.NewJWTService(config.JWT)
	templateService, _ := services.NewTemplateService(config.TemplatesDir)

	return servicesContainer{
		Address:  addressService,
		User:     userService,
		Category: categoryService,
		Product:  productService,
		Cart:     cartService,
		Order:    orderService,
		Invite:   inviteService,
		Password: passwordService,
		Refresh:  refreshService,
		Email:    emailService,
		JWT:      jwtService,
		Template: templateService,
	}
}

// servicesContainer holds all service dependencies
type servicesContainer struct {
	Address  services.AddressService
	User     services.UserService
	Category services.CategoryService
	Product  services.ProductService
	Cart     services.CartService
	Order    services.OrderService
	Invite   services.InviteService
	Password services.PasswordService
	Refresh  services.RefreshService
	Email    services.EmailSender
	JWT      services.JWTService
	Template services.TemplateService
}

// gracefulShutdown handles termination signals and gracefully shuts down the server.
// It does so by waiting for all active connections to finish, or until a timeout is reached.
// If the timeout is reached, the server is forcefully shut down.
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
