package services

import (
	"context"
	"testing"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
	"github.com/stretchr/testify/require"
)

type borrowProductRepository struct {
	repositories.ProductRepository
	product types.Product
}

func (r *borrowProductRepository) UpdateProduct(_ context.Context, product types.Product) error {
	r.product = product
	return nil
}

func TestBorrowOnlyProductIsFree(t *testing.T) {
	repo := &borrowProductRepository{}
	service := NewProductService(repo)
	err := service.UpdateProduct(context.Background(), types.Product{
		BorrowOnly:  true,
		Price:       utilities.Ptr(int64(1000)),
		MemberPrice: utilities.Ptr(int64(500)),
	})
	require.NoError(t, err)
	require.Equal(t, int64(0), *repo.product.Price)
	require.Nil(t, repo.product.MemberPrice)
}
