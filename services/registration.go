package services

import (
	"context"
	"time"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
)

type RegistrationService interface {
	CreateCode(ctx context.Context, userID string, expiry time.Time) (string, error)
	VerifyCode(ctx context.Context, code string) (*types.User, error)
}

type registrationService struct {
	repo repositories.RegistrationRepository
}

func NewRegistrationService(repo repositories.RegistrationRepository) RegistrationService {
	return &registrationService{repo: repo}
}

func (s *registrationService) CreateCode(ctx context.Context, userID string, expiry time.Time) (string, error) {
	code, err := generateCode()
	if err != nil {
		return "", err
	}

	// store the registration code
	if err := s.repo.CreateCode(ctx, userID, code, expiry); err != nil {
		return "", err
	}

	return code, nil
}

func (s *registrationService) VerifyCode(ctx context.Context, code string) (*types.User, error) {
	return s.repo.VerifyCode(ctx, code)
}
