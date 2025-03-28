package services

import (
	"context"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
)

type ProductService interface {
	CreateProduct(ctx context.Context, product *types.Product) error
	CreateProductWithCategory(ctx context.Context, product *types.Product, categoryID string) error
	RemoveProduct(ctx context.Context, id string) error
	GetProducts(ctx context.Context, filter types.ProductFilter) ([]types.Product, error)
	GetProductByID(ctx context.Context, id string) (*types.ProductWithInventory, error)
	UpdateInventory(ctx context.Context, productID string, quantity int) error
}

type productService struct {
	repo repositories.ProductRepository
}

func NewProductService(repo repositories.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) CreateProduct(ctx context.Context, product *types.Product) error {
	return s.repo.CreateProduct(ctx, product)
}

func (s *productService) CreateProductWithCategory(ctx context.Context, product *types.Product, categoryID string) error {
	return s.repo.CreateProductWithCategory(ctx, product, categoryID)
}

func (s *productService) GetProducts(ctx context.Context, filter types.ProductFilter) ([]types.Product, error) {
	return s.repo.GetProducts(ctx, filter)
}

func (s *productService) GetProductByID(ctx context.Context, id string) (*types.ProductWithInventory, error) {
	return s.repo.GetProductByID(ctx, id)
}

func (s *productService) RemoveProduct(ctx context.Context, id string) error {
	return s.repo.DeleteProduct(ctx, id)
}

func (s *productService) UpdateInventory(ctx context.Context, productID string, quantity int) error {
	return s.repo.UpdateInventory(ctx, productID, quantity)
}
