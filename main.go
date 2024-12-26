package main

import (
	"log"
	"net/http"

	"github.com/dgyurics/marketplace/db"
	"github.com/dgyurics/marketplace/handlers"
	"github.com/dgyurics/marketplace/middleware"
	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/utilities"
	"github.com/gorilla/mux"
)

func main() {
	// connect to database
	db := db.Connect()

	// create repositories
	authRepository := repositories.NewAuthRepository(db)
	userRepository := repositories.NewUserRepository(db)
	categoryRepository := repositories.NewCategoryRepository(db)
	productRepository := repositories.NewProductRepository(db)
	cartRepository := repositories.NewCartRepository(db)
	paymentRepository := repositories.NewPaymentRepository(db)
	orderRepository := repositories.NewOrderRepository(db)

	// create services
	authService := services.NewAuthService(authRepository, utilities.LoadAuthConfig())
	userService := services.NewUserService(userRepository)
	categoryService := services.NewCategoryService(categoryRepository)
	paymentService := services.NewPaymentService(paymentRepository, orderRepository, utilities.LoadPaymentConfig())
	productService := services.NewProductService(productRepository)
	cartService := services.NewCartService(cartRepository, orderRepository, paymentService)

	// create middleware
	middleware := middleware.NewAccessControl(authService)

	// register handlers
	router := mux.NewRouter()
	handlers.RegisterUserHandler(userService, authService, router)
	handlers.RegisterCategoryHandler(categoryService, router, middleware)
	handlers.RegisterProductHandler(productService, router, middleware)
	handlers.RegisterCartHandler(cartService, router, middleware)
	handlers.RegisterPaymentHandler(paymentService, router)

	log.Println("Server is running on port 8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}
