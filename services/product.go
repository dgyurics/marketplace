package services

import (
	"context"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/repositories"
)

type ProductService interface {
	CreateProduct(ctx context.Context, product *models.Product) error
	CreateProductWithCategory(ctx context.Context, product *models.Product, categoryID string) error
	GetAllProducts(ctx context.Context) ([]models.Product, error)
	GetProductByID(ctx context.Context, id string) (*models.Product, error)
	UpdateInventory(ctx context.Context, productID string, quantity int) error
}

type productService struct {
	repo repositories.ProductRepository
}

func NewProductService(repo repositories.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) CreateProduct(ctx context.Context, product *models.Product) error {
	return s.repo.CreateProduct(ctx, product)
}

func (s *productService) CreateProductWithCategory(ctx context.Context, product *models.Product, categoryID string) error {
	return s.repo.CreateProductWithCategory(ctx, product, categoryID)
}

func (s *productService) GetAllProducts(ctx context.Context) ([]models.Product, error) {
	return s.repo.GetAllProducts(ctx)
}

func (s *productService) GetProductByID(ctx context.Context, id string) (*models.Product, error) {
	return s.repo.GetProductByID(ctx, id)
}

func (s *productService) UpdateInventory(ctx context.Context, productID string, quantity int) error {
	return s.repo.UpdateInventory(ctx, productID, quantity)
}
