package services

import (
	"context"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
)

type CartService interface {
	AddItemToCart(ctx context.Context, item *types.CartItem) error
	GetCart(ctx context.Context) ([]types.CartItem, error)
	UpdateCartItem(ctx context.Context, item *types.CartItem) error
	RemoveItemFromCart(ctx context.Context, productID string) error
	ClearCart(ctx context.Context) error
}

type cartService struct {
	cartRepo repositories.CartRepository
}

func NewCartService(cartRepo repositories.CartRepository) CartService {
	return &cartService{cartRepo: cartRepo}
}

func (s *cartService) AddItemToCart(ctx context.Context, item *types.CartItem) error {
	// FIXME if user has no cart, create one (similar to GetOrCreateCart)
	return s.cartRepo.AddItemToCart(ctx, getUserID(ctx), item)
}

func (s *cartService) GetCart(ctx context.Context) ([]types.CartItem, error) {
	return s.cartRepo.GetCart(ctx, getUserID(ctx))
}

func (s *cartService) UpdateCartItem(ctx context.Context, item *types.CartItem) error {
	return s.cartRepo.UpdateCartItem(ctx, getUserID(ctx), item)
}

func (s *cartService) RemoveItemFromCart(ctx context.Context, productID string) error {
	return s.cartRepo.RemoveItemFromCart(ctx, getUserID(ctx), productID)
}

func (s *cartService) ClearCart(ctx context.Context) error {
	return s.cartRepo.ClearCart(ctx, getUserID(ctx))
}
