package services

import (
	"context"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
)

type CartService interface {
	AddItem(ctx context.Context, item *types.CartItem) error
	GetItems(ctx context.Context) ([]types.CartItem, error)
	RemoveItem(ctx context.Context, productID string) error
	Clear(ctx context.Context) error
}

type cartService struct {
	cartRepo repositories.CartRepository
}

func NewCartService(cartRepo repositories.CartRepository) CartService {
	return &cartService{cartRepo: cartRepo}
}

func (s *cartService) AddItem(ctx context.Context, item *types.CartItem) error {
	return s.cartRepo.AddItem(ctx, getUserID(ctx), item)
}

func (s *cartService) GetItems(ctx context.Context) ([]types.CartItem, error) {
	return s.cartRepo.GetItems(ctx, getUserID(ctx))
}

func (s *cartService) RemoveItem(ctx context.Context, productID string) error {
	return s.cartRepo.RemoveItem(ctx, getUserID(ctx), productID)
}

func (s *cartService) Clear(ctx context.Context) error {
	return s.cartRepo.Clear(ctx, getUserID(ctx))
}
