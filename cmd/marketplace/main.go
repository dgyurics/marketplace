/*
Marketplace - Open Source E-commerce Platform
Created by Dennis Gyurics
https://github.com/dgyurics/marketplace
Licensed under MIT License
*/
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
)

func main() {
	// Load environment variables
	utilities.LoadEnvironment()

	// Root context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration
	config := utilities.LoadConfig()

	// Initialize logger
	utilities.InitLogger(config.Logger)

	// Initialize unique ID generator
	utilities.InitIDGenerator(config.MachineID)

	// Initialize database
	db := db.Connect(config.Database)

	// Initialize services
	services := initializeServices(db, config)

	// Start schedule service
	go services.Schedule.Start(ctx)

	// Initialize and start server
	server := initializeServer(config, services)
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
		}
	}()
	gracefulShutdown(server, cancel)
}

// initializeServer sets up the database, services, and HTTP server
func initializeServer(config types.Config, services servicesContainer) *http.Server {
	// create router
	router := mux.NewRouter()
	router.Use(middleware.RequestLoggerMiddleware) // nginx logs requests, this is used for debugging

	// create base router
	baseRouter := routes.NewRouter(router, middleware.NewAccessControl(services.JWT))

	// create routes
	routes.RegisterAllRoutes(
		routes.NewAddressRoutes(services.Address, config.Locale, baseRouter),
		routes.NewCartRoutes(services.Cart, services.Order, baseRouter),
		routes.NewCategoryRoutes(services.Category, baseRouter),
		routes.NewHealthRoutes(baseRouter),
		routes.NewImageRoutes(services.Image, baseRouter),
		routes.NewOrderRoutes(services.Order, services.Tax, services.Payment, services.Cart, baseRouter),
		routes.NewPasswordRoutes(services.Password, services.User, services.Email, services.Template, config.BaseURL, baseRouter),
		routes.NewPaymentRoutes(services.Payment, baseRouter),
		routes.NewProductRoutes(services.Product, baseRouter),
		routes.NewUserRoutes(services.User, services.Invite, services.JWT, services.Refresh, config.Auth, baseRouter),
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
	scheduleRepository := repositories.NewScheduleRepository(db)
	refreshTokenRepository := repositories.NewRefreshRepository(db)
	taxRepository := repositories.NewTaxRepository(db)
	imageRepository := repositories.NewImageRepository(db)

	// create http client required by certain services
	httpClient := utilities.NewDefaultHTTPClient(10 * time.Second) // TODO make this configurable

	// create services
	templateService, err := services.NewTemplateService(config.TemplatesDir)
	if err != nil {
		slog.Error("Failed to initialize template service", "error", err, "templatesDir", config.TemplatesDir)
		os.Exit(1)
	}
	emailService := services.NewEmailService(config.Email)
	addressService := services.NewAddressService(addressRepository, config.Locale)
	userService := services.NewUserService(userRepository)
	categoryService := services.NewCategoryService(categoryRepository)
	productService := services.NewProductService(productRepository)
	cartService := services.NewCartService(cartRepository)
	// TODO refactor this
	paymentService := services.NewPaymentService(httpClient, config.Stripe, config.Locale, emailService, templateService, userService, orderRepository, config.BaseURL)
	orderService := services.NewOrderService(orderRepository, cartRepository, paymentService, config.Locale, httpClient)
	inviteService := services.NewInviteService(inviteRepository, config.Auth.HMACSecret)
	imageService := services.NewImageService(httpClient, imageRepository, config.Image)
	passwordService := services.NewPasswordService(passwordRepository, config.Auth.HMACSecret)
	refreshService := services.NewRefreshService(refreshTokenRepository, config.Auth)
	jwtService := services.NewJWTService(config.JWT)
	scheduleService := services.NewScheduleService(orderService, scheduleRepository)
	taxService := services.NewTaxService(taxRepository, config.Stripe, config.Locale, httpClient)

	return servicesContainer{
		Address:  addressService,
		User:     userService,
		Category: categoryService,
		Product:  productService,
		Cart:     cartService,
		Order:    orderService,
		Image:    imageService,
		Invite:   inviteService,
		Password: passwordService,
		Refresh:  refreshService,
		Payment:  paymentService,
		Email:    emailService,
		JWT:      jwtService,
		Schedule: scheduleService,
		Tax:      taxService,
		Template: templateService,
	}
}

// servicesContainer holds all service dependencies
type servicesContainer struct {
	Address  services.AddressService
	User     services.UserService
	Cart     services.CartService
	Category services.CategoryService
	Email    services.EmailService
	Image    services.ImageService
	Invite   services.InviteService
	JWT      services.JWTService
	Order    services.OrderService
	Password services.PasswordService
	Payment  services.PaymentService
	Product  services.ProductService
	Refresh  services.RefreshService
	Schedule services.ScheduleService
	Tax      services.TaxService
	Template services.TemplateService
}

// gracefulShutdown handles termination signals and gracefully shuts down the server.
// It does so by waiting for all active connections to finish, or until a timeout is reached.
// If the timeout is reached, the server is forcefully shut down.
func gracefulShutdown(server *http.Server, cancel context.CancelFunc) {
	// Listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Wait for a signal
	<-stop
	slog.Info("Shutdown signal received")

	// Cancel root context
	cancel()

	// Create a context with timeout for shutdown
	ctx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	// Gracefully shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server shutdown failed", "error", err)
	} else {
		slog.Info("Server gracefully stopped")
	}
}
