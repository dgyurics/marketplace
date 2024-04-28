package main

import (
	"log"
	"net/http"

	"github.com/dgyurics/marketplace/db"
	"github.com/dgyurics/marketplace/handlers"
	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/services"
	"github.com/gorilla/mux"
)

func main() {
	pool, err := db.NewConnectionPool("postgres://postgres:postgres@localhost:5432/marketplace")
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()

	userRepository := repositories.NewUserRepository(pool)
	userService := services.NewUserService(userRepository)
	handlers.RegisterUserHandler(userService, router)

	categoryRepository := repositories.NewCategoryRepository(pool)
	categoryService := services.NewCategoryService(categoryRepository)
	handlers.RegisterCategoryHandler(categoryService, router)

	productRepository := repositories.NewProductRepository(pool)
	productService := services.NewProductService(productRepository)
	handlers.RegisterProductHandler(productService, router)

	log.Println("Server is running on port 8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}
