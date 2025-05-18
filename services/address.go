package services

import (
	"context"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
)

type AddressService interface {
	CreateAddress(ctx context.Context, address *types.Address) error
	GetAddresses(ctx context.Context) ([]types.Address, error)
	RemoveAddress(ctx context.Context, addressID string) error
}

type addressService struct {
	repo   repositories.AddressRepository
	config types.LocaleConfig
}

func NewAddressService(repo repositories.AddressRepository, config types.LocaleConfig) AddressService {
	return &addressService{
		config: config,
		repo:   repo,
	}
}

func (s *addressService) CreateAddress(ctx context.Context, address *types.Address) error {
	var userID = getUserID(ctx)
	address.UserID = userID
	address.Country = s.config.Country

	// if a duplicate address is found, addressID will not be used
	addressID, err := utilities.GenerateIDString()
	if err != nil {
		return err
	}
	address.ID = addressID
	return s.repo.CreateAddress(ctx, address)
}

func (s *addressService) GetAddresses(ctx context.Context) ([]types.Address, error) {
	var userID = getUserID(ctx)
	return s.repo.GetAddresses(ctx, userID)
}

func (s *addressService) RemoveAddress(ctx context.Context, addressID string) error {
	var userID = getUserID(ctx)
	return s.repo.RemoveAddress(ctx, userID, addressID)
}
