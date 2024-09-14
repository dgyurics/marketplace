package services

import (
	"context"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/repositories"
)

type CategoryService interface {
	CreateCategory(ctx context.Context, category models.Category) (string, error)
	GetAllCategories(ctx context.Context) ([]models.Category, error)
	GetCategoryByID(ctx context.Context, id string) (*models.Category, error)
	GetProductsByCategoryID(ctx context.Context, id string) ([]models.Product, error)
}

type categoryService struct {
	repo repositories.CategoryRepository
}

func NewCategoryService(repo repositories.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) CreateCategory(ctx context.Context, category models.Category) (string, error) {
	return s.repo.CreateCategory(ctx, category)
}

func (s *categoryService) GetAllCategories(ctx context.Context) ([]models.Category, error) {
	return s.repo.GetAllCategories(ctx)
}

func (s *categoryService) GetCategoryByID(ctx context.Context, id string) (*models.Category, error) {
	return s.repo.GetCategoryByID(ctx, id)
}

func (s *categoryService) GetProductsByCategoryID(ctx context.Context, id string) ([]models.Product, error) {
	return s.repo.GetProductsByCategoryID(ctx, id)
}
