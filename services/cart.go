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
	RemoveItems(ctx context.Context) error
}

type cartService struct {
	cartRepo repositories.CartRepository
	userRepo repositories.UserRepository
}

func NewCartService(cartRepo repositories.CartRepository, userRepo repositories.UserRepository) CartService {
	return &cartService{cartRepo: cartRepo, userRepo: userRepo}
}

func (s *cartService) AddItem(ctx context.Context, item *types.CartItem) error {
	return s.cartRepo.AddItem(ctx, getUserID(ctx), item)
}

// GetItems retrieves the items in the user's cart
// If the user has a minimum role of "member", a member price will be applied
// if the product has a member price defined
func (s *cartService) GetItems(ctx context.Context) ([]types.CartItem, error) {
	userID := getUserID(ctx)
	items, err := s.cartRepo.GetItems(ctx, userID)
	if err != nil || len(items) == 0 {
		return items, err
	}
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.HasMinimumRole(types.RoleMember) {
		for i := range items {
			if items[i].Product.MemberPrice != nil {
				items[i].Product.Price = items[i].Product.MemberPrice
			}
		}
	}
	return items, nil
}

func (s *cartService) RemoveItem(ctx context.Context, productID string) error {
	return s.cartRepo.RemoveItem(ctx, getUserID(ctx), productID)
}

func (s *cartService) RemoveItems(ctx context.Context) error {
	return s.cartRepo.RemoveItems(ctx, getUserID(ctx))
}
