package main

import (
	"log"
	"net/http"
	"os"

	"github.com/dgyurics/marketplace/db"
	"github.com/dgyurics/marketplace/handlers"
	"github.com/dgyurics/marketplace/middleware"
	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/services"
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

	// create services
	authService := services.NewAuthService(authRepository, getKey("private.pem"), getKey("public.pem"), []byte(getEnv("HMAC_SECRET")))
	userService := services.NewUserService(userRepository)
	categoryService := services.NewCategoryService(categoryRepository)
	productService := services.NewProductService(productRepository)
	paymentService := services.NewPaymentService()
	cartService := services.NewCartService(cartRepository, paymentService)

	// Create middleware
	middleware := middleware.NewAccessControl(authService)

	// register handlers
	router := mux.NewRouter()
	handlers.RegisterUserHandler(userService, authService, router)
	handlers.RegisterCategoryHandler(categoryService, router, middleware)
	handlers.RegisterProductHandler(productService, router, middleware)
	handlers.RegisterCartHandler(cartService, router, middleware)

	log.Println("Server is running on port 8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}

// helper for fetching public/private pem keys
func getKey(filename string) []byte {
	privateKey, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading %s key: %v", filename, err)
	}
	return privateKey
}

// helper for fetching critical environment variables
func getEnv(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	log.Fatalf("%s is required", key)
	return ""
}
