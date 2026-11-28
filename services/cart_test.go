package services

import (
	"context"
	"errors"
	"testing"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
	"github.com/stretchr/testify/require"
)

type cartPriceRepository struct {
	repositories.CartRepository
	items []types.CartItem
}

func (r *cartPriceRepository) GetItems(_ context.Context, _ string) ([]types.CartItem, error) {
	return append([]types.CartItem(nil), r.items...), nil
}

type cartPriceUserRepository struct {
	repositories.UserRepository
	user *types.User
	err  error
}

func (r *cartPriceUserRepository) GetUserByID(_ context.Context, _ string) (*types.User, error) {
	return r.user, r.err
}

func TestCartMemberPricing(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserKey, &types.User{ID: "user", Role: types.RoleMember})
	for _, role := range []types.Role{types.RoleGuest, types.RoleUser, types.RoleMember, types.RoleStaff, types.RoleAdmin} {
		for _, memberPrice := range []*int64{nil, utilities.Ptr(int64(0)), utilities.Ptr(int64(800))} {
			cartRepo := &cartPriceRepository{items: []types.CartItem{{Product: types.Product{Price: utilities.Ptr(int64(1000)), MemberPrice: memberPrice}}}}
			userRepo := &cartPriceUserRepository{user: &types.User{Role: role}}
			cart := NewCartService(cartRepo, userRepo)
			items, err := cart.GetItems(ctx)
			require.NoError(t, err)
			require.Len(t, items, 1)
			expected := int64(1000)
			if userRepo.user.HasMinimumRole(types.RoleMember) && memberPrice != nil {
				expected = *memberPrice
			}
			require.Equal(t, expected, *items[0].Product.Price)
			if userRepo.user.HasMinimumRole(types.RoleMember) {
				require.Equal(t, memberPrice, items[0].Product.MemberPrice)
			} else {
				require.Nil(t, items[0].Product.MemberPrice)
			}

			// A changed database role takes effect even if the token still says member.
			userRepo.user.Role = types.RoleUser
			items, err = cart.GetItems(ctx)
			require.NoError(t, err)
			require.Equal(t, int64(1000), *items[0].Product.Price)
			require.Nil(t, items[0].Product.MemberPrice)
		}
	}
}

func TestCartPricingFailsWhenUserLookupFails(t *testing.T) {
	cartRepo := &cartPriceRepository{items: []types.CartItem{{}}}
	userRepo := &cartPriceUserRepository{err: errors.New("user lookup failed")}
	_, err := NewCartService(cartRepo, userRepo).GetItems(context.Background())
	require.Error(t, err)
}
