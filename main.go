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
	conPool, err := db.Connect("postgres://postgres:postgres@localhost:5432/marketplace")
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()

	userRepository := repositories.NewUserRepository(conPool)
	userService := services.NewUserService(userRepository)
	handlers.RegisterUserHandler(userService, router)

	categoryRepository := repositories.NewCategoryRepository(conPool)
	categoryService := services.NewCategoryService(categoryRepository)
	handlers.RegisterCategoryHandler(categoryService, router)

	productRepository := repositories.NewProductRepository(conPool)
	productService := services.NewProductService(productRepository)
	handlers.RegisterProductHandler(productService, router)

	log.Println("Server is running on port 8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}
