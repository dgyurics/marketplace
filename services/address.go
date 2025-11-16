package services

import (
	"context"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
)

type AddressService interface {
	CreateAddress(ctx context.Context, address *types.Address) error
	UpdateAddress(ctx context.Context, address *types.Address) error
	GetAddress(ctx context.Context, addressID string) (types.Address, error)
	RemoveAddress(ctx context.Context, addressID string) error
}

type addressService struct {
	repo repositories.AddressRepository
}

func NewAddressService(repo repositories.AddressRepository) AddressService {
	return &addressService{
		repo: repo,
	}
}

func (s *addressService) CreateAddress(ctx context.Context, address *types.Address) error {
	var userID = getUserID(ctx)
	address.UserID = userID
	address.Country = utilities.Locale.CountryCode
	addressID, err := utilities.GenerateIDString()
	if err != nil {
		return err
	}
	address.ID = addressID
	return s.repo.CreateAddress(ctx, address)
}

func (s *addressService) GetAddress(ctx context.Context, addressID string) (types.Address, error) {
	var userID = getUserID(ctx)
	return s.repo.GetAddress(ctx, userID, addressID)
}

func (s *addressService) UpdateAddress(ctx context.Context, address *types.Address) error {
	address.UserID = getUserID(ctx)
	return s.repo.UpdateAddress(ctx, address)
}

func (s *addressService) RemoveAddress(ctx context.Context, addressID string) error {
	var userID = getUserID(ctx)
	return s.repo.RemoveAddress(ctx, userID, addressID)
}
