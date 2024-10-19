package services

import (
	"context"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/repositories"
)

type CartService interface {
	CreateCart(ctx context.Context, cart *models.Cart) error
	AddItemToCart(ctx context.Context, item *models.CartItem) error
	GetCart(ctx context.Context) (*models.Cart, error)
	UpdateCartItem(ctx context.Context, item *models.CartItem) error
	RemoveItemFromCart(ctx context.Context, productID string) error
	ClearCart(ctx context.Context) error
}

type cartService struct {
	repo repositories.CartRepository
}

func NewCartService(repo repositories.CartRepository) CartService {
	return &cartService{repo: repo}
}

func (s *cartService) CreateCart(ctx context.Context, cart *models.Cart) error {
	userID := ctx.Value("userID").(string)
	return s.repo.CreateCart(ctx, userID)
}

func (s *cartService) AddItemToCart(ctx context.Context, item *models.CartItem) error {
	userID := ctx.Value("userID").(string)
	return s.repo.AddItemToCart(ctx, userID, item)
}

func (s *cartService) GetCart(ctx context.Context) (*models.Cart, error) {
	userID := ctx.Value("userID").(string)
	return s.repo.GetCart(ctx, userID)
}

func (s *cartService) UpdateCartItem(ctx context.Context, item *models.CartItem) error {
	userID := ctx.Value("userID").(string)
	return s.repo.UpdateCartItem(ctx, userID, item)
}

func (s *cartService) RemoveItemFromCart(ctx context.Context, productID string) error {
	userID := ctx.Value("userID").(string)
	return s.repo.RemoveItemFromCart(ctx, userID, productID)
}

func (s *cartService) ClearCart(ctx context.Context) error {
	userID := ctx.Value("userID").(string)
	return s.repo.ClearCart(ctx, userID)
}
