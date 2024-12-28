package services

import (
	"context"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/repositories"
)

type CartService interface {
	AddItemToCart(ctx context.Context, item *models.CartItem) error
	GetCart(ctx context.Context) (*models.Cart, error)
	UpdateCartItem(ctx context.Context, item *models.CartItem) error
	RemoveItemFromCart(ctx context.Context, productID string) error
	ClearCart(ctx context.Context) error
}

type cartService struct {
	cartRepo repositories.CartRepository
}

func NewCartService(cartRepo repositories.CartRepository) CartService {
	return &cartService{cartRepo: cartRepo}
}

func (s *cartService) AddItemToCart(ctx context.Context, item *models.CartItem) error {
	return s.cartRepo.AddItemToCart(ctx, getUserID(ctx), item)
}

func (s *cartService) GetCart(ctx context.Context) (*models.Cart, error) {
	return s.cartRepo.GetOrCreateCart(ctx, getUserID(ctx))
}

func (s *cartService) UpdateCartItem(ctx context.Context, item *models.CartItem) error {
	return s.cartRepo.UpdateCartItem(ctx, getUserID(ctx), item)
}

func (s *cartService) RemoveItemFromCart(ctx context.Context, productID string) error {
	return s.cartRepo.RemoveItemFromCart(ctx, getUserID(ctx), productID)
}

func (s *cartService) ClearCart(ctx context.Context) error {
	return s.cartRepo.ClearCart(ctx, getUserID(ctx))
}
