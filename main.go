package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgyurics/marketplace/db"
	"github.com/dgyurics/marketplace/handlers"
	"github.com/dgyurics/marketplace/middleware"
	"github.com/dgyurics/marketplace/models"
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
	paymentRepository := repositories.NewPaymentRepository(db)
	orderRepository := repositories.NewOrderRepository(db)

	// create services
	authService := services.NewAuthService(authRepository, loadAuthServiceConfig())
	userService := services.NewUserService(userRepository)
	categoryService := services.NewCategoryService(categoryRepository)
	productService := services.NewProductService(productRepository)
	paymentService := services.NewPaymentService(paymentRepository, orderRepository, getEnv("ENVIRONMENT"), getEnv("STRIPE_BASE_URL"), getEnv("STRIPE_SECRET_KEY"), getEnv("STRIPE_WEBHOOK_SIGNING_SECRET"))
	cartService := services.NewCartService(cartRepository, orderRepository, paymentService)

	// Create middleware
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

func loadAuthServiceConfig() models.AuthServiceConfig {
	accessTokenDuration, err := time.ParseDuration(getEnv("DURATION_ACCESS_TOKEN"))
	if err != nil {
		log.Fatalf("Invalid access token duration: %v", err)
	}

	refreshTokenDuration, err := time.ParseDuration(getEnv("DURATION_REFRESH_TOKEN"))
	if err != nil {
		log.Fatalf("Invalid refresh token duration: %v", err)
	}

	return models.AuthServiceConfig{
		PrivateKey:           getKey("private.pem"),
		PublicKey:            getKey("public.pem"),
		HMACSecret:           []byte(getEnv("HMAC_SECRET")),
		DurationAccessToken:  accessTokenDuration,
		DurationRefreshToken: refreshTokenDuration,
	}
}
