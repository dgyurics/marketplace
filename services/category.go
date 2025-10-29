package services

import (
	"context"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
)

type CategoryService interface {
	CreateCategory(ctx context.Context, category *types.Category) error
	UpdateCategory(ctx context.Context, category types.Category) error
	GetAllCategories(ctx context.Context) ([]types.Category, error)
	GetCategoryByID(ctx context.Context, id string) (*types.Category, error)
	RemoveCategory(ctx context.Context, id string) error
}

type categoryService struct {
	repo repositories.CategoryRepository
}

func NewCategoryService(repo repositories.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) CreateCategory(ctx context.Context, category *types.Category) error {
	categoryID, err := utilities.GenerateIDString()
	if err != nil {
		return err
	}
	category.ID = categoryID
	return s.repo.CreateCategory(ctx, category)
}

func (s *categoryService) GetAllCategories(ctx context.Context) ([]types.Category, error) {
	return s.repo.GetAllCategories(ctx)
}

func (s *categoryService) UpdateCategory(ctx context.Context, category types.Category) error {
	return s.repo.UpdateCategory(ctx, category)
}

func (s *categoryService) GetCategoryByID(ctx context.Context, id string) (*types.Category, error) {
	return s.repo.GetCategoryByID(ctx, id)
}

func (s *categoryService) RemoveCategory(ctx context.Context, id string) error {
	return s.repo.RemoveCategory(ctx, id)
}
