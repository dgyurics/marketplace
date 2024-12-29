package main

import (
	"log"
	"net/http"

	"github.com/dgyurics/marketplace/db"
	"github.com/dgyurics/marketplace/middleware"
	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/routes"
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
	orderRepository := repositories.NewOrderRepository(db)

	// create services
	authService := services.NewAuthService(authRepository, utilities.LoadAuthConfig())
	userService := services.NewUserService(userRepository)
	categoryService := services.NewCategoryService(categoryRepository)
	orderService := services.NewOrderService(orderRepository, utilities.LoadOrderConfig())
	productService := services.NewProductService(productRepository)
	cartService := services.NewCartService(cartRepository)

	// create router
	router := mux.NewRouter()

	// create base router which encapsulates the primary router and the access control middleware
	baseRouter := routes.NewRouter(router, middleware.NewAccessControl(authService))

	// create routes
	categoryRoutes := routes.NewCategoryRoutes(categoryService, baseRouter)
	userRoutes := routes.NewUserRoutes(userService, authService, baseRouter)
	productRoutes := routes.NewProductRoutes(productService, baseRouter)
	cartRoutes := routes.NewCartRoutes(cartService, baseRouter)
	orderRoutes := routes.NewOrderRoutes(orderService, baseRouter)

	// register routes to the main router
	routes.RegisterAllRoutes(
		userRoutes,
		categoryRoutes,
		productRoutes,
		cartRoutes,
		orderRoutes,
	)

	log.Println("Server is running on port 8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}
