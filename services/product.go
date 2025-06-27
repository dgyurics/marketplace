package services

import (
	"context"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
)

type ProductService interface {
	CreateProduct(ctx context.Context, product *types.Product, categorySlug string) error
	GetProducts(ctx context.Context, filter types.ProductFilter) ([]types.Product, error)
	GetProductByID(ctx context.Context, id string) (*types.ProductWithInventory, error)
	UpdateInventory(ctx context.Context, productID string, quantity int) error
	RemoveProduct(ctx context.Context, id string) error
}

type productService struct {
	repo repositories.ProductRepository
}

func NewProductService(repo repositories.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func generateImageIDs(images []types.Image) error {
	for idx := range images {
		imgID, err := utilities.GenerateIDString()
		if err != nil {
			return err
		}
		images[idx].ID = imgID
	}
	return nil
}

func (s *productService) CreateProduct(ctx context.Context, product *types.Product, categorySlug string) error {
	productID, err := utilities.GenerateIDString()
	if err != nil {
		return err
	}
	product.ID = productID
	if err = generateImageIDs(product.Images); err != nil {
		return err
	}
	return s.repo.CreateProduct(ctx, product, categorySlug)
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
