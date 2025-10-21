package services

import (
	"context"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
)

type RateLimitService interface {
	GetHitCount(ctx context.Context, rl *types.RateLimit) error
	RecordHit(ctx context.Context, rl *types.RateLimit) error
	Cleanup(ctx context.Context) error
}

type rateLimitService struct {
	repo repositories.RateLimitRepository
}

func NewRateLimitService(repo repositories.RateLimitRepository) RateLimitService {
	return &rateLimitService{
		repo: repo,
	}
}

func (s *rateLimitService) Cleanup(ctx context.Context) error {
	return s.repo.Cleanup(ctx)
}

func (s *rateLimitService) GetHitCount(ctx context.Context, rl *types.RateLimit) error {
	return s.repo.GetHitCount(ctx, rl)
}

func (s *rateLimitService) RecordHit(ctx context.Context, rl *types.RateLimit) error {
	return s.repo.RecordHit(ctx, rl)
}
