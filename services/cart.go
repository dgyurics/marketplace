package services

import (
	"context"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/repositories"
)

type CartService interface {
	CreateCart(ctx context.Context, cart *models.Cart) error
	AddItemToCart(ctx context.Context, cartID string, item *models.CartItem) error
	GetCartByID(ctx context.Context, id string) (*models.Cart, error)
	UpdateCartItem(ctx context.Context, cartID string, item *models.CartItem) error
	RemoveItemFromCart(ctx context.Context, cartID, productID string) error
	ClearCart(ctx context.Context, cartID string) error
}

type cartService struct {
	repo repositories.CartRepository
}

func NewCartService(repo repositories.CartRepository) CartService {
	return &cartService{repo: repo}
}

func (s *cartService) CreateCart(ctx context.Context, cart *models.Cart) error {
	return s.repo.CreateCart(ctx, cart)
}

func (s *cartService) AddItemToCart(ctx context.Context, cartID string, item *models.CartItem) error {
	return s.repo.AddItemToCart(ctx, cartID, item)
}

func (s *cartService) GetCartByID(ctx context.Context, id string) (*models.Cart, error) {
	return s.repo.GetCartByID(ctx, id)
}

func (s *cartService) UpdateCartItem(ctx context.Context, cartID string, item *models.CartItem) error {
	return s.repo.UpdateCartItem(ctx, cartID, item)
}

func (s *cartService) RemoveItemFromCart(ctx context.Context, cartID, productID string) error {
	return s.repo.RemoveItemFromCart(ctx, cartID, productID)
}

func (s *cartService) ClearCart(ctx context.Context, cartID string) error {
	return s.repo.ClearCart(ctx, cartID)
}
