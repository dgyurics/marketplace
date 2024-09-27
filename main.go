package main

import (
	"log"
	"net/http"
	"os"

	"github.com/dgyurics/marketplace/db"
	"github.com/dgyurics/marketplace/handlers"
	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/utilities"
	"github.com/gorilla/mux"
)

func main() {
	// establish connection to the database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	db, err := db.Connect(dbURL)
	if err != nil {
		log.Fatal(err)
	}

	// create repositories
	userRepository := repositories.NewUserRepository(db)
	categoryRepository := repositories.NewCategoryRepository(db)
	productRepository := repositories.NewProductRepository(db)

	// create services
	userService := services.NewUserService(userRepository)
	categoryService := services.NewCategoryService(categoryRepository)
	productService := services.NewProductService(productRepository)

	// read private and public keys
	privateKey, err := os.ReadFile("private.pem")
	if err != nil {
		log.Fatal("Error reading private key:", err)
	}

	publicKey, err := os.ReadFile("public.pem")
	if err != nil {
		log.Fatal("Error reading public key:", err)
	}

	// create JWT utility
	jwtUtil := utilities.NewJWTUtility(privateKey, publicKey)

	// register handlers
	router := mux.NewRouter()
	handlers.RegisterUserHandler(userService, jwtUtil, router)
	handlers.RegisterCategoryHandler(categoryService, router)
	handlers.RegisterProductHandler(productService, router)

	log.Println("Server is running on port 8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}
