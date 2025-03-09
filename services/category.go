package services

import (
	"context"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
)

type CategoryService interface {
	CreateCategory(ctx context.Context, category types.Category) (string, error)
	GetAllCategories(ctx context.Context) ([]types.Category, error)
	GetCategoryByID(ctx context.Context, id string) (*types.Category, error)
}

type categoryService struct {
	repo repositories.CategoryRepository
}

func NewCategoryService(repo repositories.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) CreateCategory(ctx context.Context, category types.Category) (string, error) {
	return s.repo.CreateCategory(ctx, category)
}

func (s *categoryService) GetAllCategories(ctx context.Context) ([]types.Category, error) {
	return s.repo.GetAllCategories(ctx)
}

func (s *categoryService) GetCategoryByID(ctx context.Context, id string) (*types.Category, error) {
	return s.repo.GetCategoryByID(ctx, id)
}
