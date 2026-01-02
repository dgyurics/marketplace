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

	// Initialize Locale
	utilities.InitLocale(config.Country)

	// Initialize unique ID generator
	utilities.InitIDGenerator(config.MachineID)

	// Initialize database connection
	dbPool := db.Connect(config.Database)

	// Run database migrations
	if err := db.RunMigrations(dbPool); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Initialize services
	services := initializeServices(dbPool, config)

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
	// create middleware
	authorizer := middleware.NewAccessControl(services.JWT)
	rateLimit := middleware.NewRateLimit(services.RateLimit, config.RateLimit)

	// create router
	router := mux.NewRouter()
	router.Use(middleware.RequestLoggerMiddleware) // nginx logs requests, this is used for debugging
	baseRouter := routes.NewRouter(router, authorizer, rateLimit)

	// create routes
	routes.RegisterAllRoutes(
		routes.NewAddressRoutes(services.Address, services.Shipping, baseRouter),
		routes.NewShippingZoneRoutes(services.Shipping, baseRouter),
		routes.NewCartRoutes(services.Cart, services.Order, baseRouter),
		routes.NewCategoryRoutes(services.Category, baseRouter),
		routes.NewHealthRoutes(baseRouter),
		routes.NewImageRoutes(services.Image, baseRouter),
		routes.NewOrderRoutes(services.Order, services.Tax, services.Payment, services.Cart, services.Address, baseRouter),
		routes.NewPasswordRoutes(services.Password, services.User, services.Email, services.Template, config.BaseURL, baseRouter),
		routes.NewPaymentRoutes(services.Payment, baseRouter),
		routes.NewProductRoutes(services.Product, baseRouter),
		routes.NewRegisterRoutes(services.User, services.JWT, services.Refresh, services.Email, services.Template, config.BaseURL, baseRouter),
		routes.NewTaxRoutes(services.Cart, services.Tax, baseRouter),
		routes.NewUserRoutes(services.User, services.JWT, services.Refresh, config.Auth, baseRouter),
		routes.NewLocaleRoutes(baseRouter),
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
	passwordRepository := repositories.NewPasswordRepository(db)
	rateLimitRepository := repositories.NewRateLimitRepository(db)
	refreshTokenRepository := repositories.NewRefreshRepository(db)
	taxRepository := repositories.NewTaxRepository(db)
	imageRepository := repositories.NewImageRepository(db)
	shippingZoneRepository := repositories.NewShippingZoneRepository(db)

	// create http client required by certain services
	httpClient := utilities.NewDefaultHTTPClient(10 * time.Second) // TODO make this configurable

	// create services
	templateService, err := services.NewTemplateService(config.TemplatesDir)
	if err != nil {
		slog.Error("Failed to initialize template service", "error", err, "templatesDir", config.TemplatesDir)
		os.Exit(1)
	}
	scheduleService := services.NewScheduleService(db)
	emailService := services.NewEmailService(config.Email)
	addressService := services.NewAddressService(addressRepository)
	shippingZoneService := services.NewShippingZoneService(shippingZoneRepository)
	userService := services.NewUserService(userRepository)
	categoryService := services.NewCategoryService(categoryRepository)
	productService := services.NewProductService(productRepository)
	cartService := services.NewCartService(cartRepository)
	paymentService := services.NewPaymentService(httpClient, config.Payment, emailService, templateService, userService, orderRepository)
	orderService := services.NewOrderService(orderRepository, cartRepository, paymentService, httpClient)
	imageService := services.NewImageService(httpClient, imageRepository, config.Image)
	passwordService := services.NewPasswordService(passwordRepository, config.Auth.HMACSecret)
	rateLimitService := services.NewRateLimitService(rateLimitRepository)
	refreshService := services.NewRefreshService(refreshTokenRepository, config.Auth)
	jwtService := services.NewJWTService(config.JWT)
	taxService := services.NewTaxService(taxRepository, config.Payment, httpClient)

	return servicesContainer{
		Address:   addressService,
		User:      userService,
		Category:  categoryService,
		Product:   productService,
		Cart:      cartService,
		Order:     orderService,
		Image:     imageService,
		Password:  passwordService,
		RateLimit: rateLimitService,
		Refresh:   refreshService,
		Payment:   paymentService,
		Email:     emailService,
		JWT:       jwtService,
		Shipping:  shippingZoneService,
		Schedule:  scheduleService,
		Tax:       taxService,
		Template:  templateService,
	}
}

// servicesContainer holds all service dependencies
type servicesContainer struct {
	Address   services.AddressService
	User      services.UserService
	Cart      services.CartService
	Category  services.CategoryService
	Email     services.EmailService
	Image     services.ImageService
	JWT       services.JWTService
	Order     services.OrderService
	Password  services.PasswordService
	Payment   services.PaymentService
	Product   services.ProductService
	RateLimit services.RateLimitService
	Refresh   services.RefreshService
	Shipping  services.ShippingZoneService
	Schedule  services.ScheduleService
	Tax       services.TaxService
	Template  services.TemplateService
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
