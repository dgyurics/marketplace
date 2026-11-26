package repositories

import (
	"context"
	"testing"

	"github.com/dgyurics/marketplace/types"
	"github.com/stretchr/testify/assert"
)

func TestGetTaxRates_CanceledContext(t *testing.T) {
	repo := NewTaxRepository(dbPool)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.GetTaxRates(ctx, types.Address{Country: "US"}, nil)
	assert.ErrorIs(t, err, context.Canceled)
}
