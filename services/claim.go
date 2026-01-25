package services

import (
	"context"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
)

type ClaimService interface {
	ClaimItem(ctx context.Context, claim *types.Claim) error
}

type claimService struct {
	repo repositories.ClaimRepository
}

func NewClaimService(repo repositories.ClaimRepository) ClaimService {
	return &claimService{repo: repo}
}

func (s *claimService) ClaimItem(ctx context.Context, claim *types.Claim) (err error) {
	claim.UserID = getUserID(ctx)
	claim.ID, err = utilities.GenerateIDString()
	if err != nil {
		return err
	}

	return s.repo.ClaimItem(ctx, claim)
}
