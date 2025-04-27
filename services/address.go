package services

import (
	"context"
	"errors"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
)

type AddressService interface {
	CreateAddress(ctx context.Context, address *types.Address) error
	GetAddresses(ctx context.Context) ([]types.Address, error)
	RemoveAddress(ctx context.Context, addressID string) error
	UpdateAddress(ctx context.Context, address *types.Address) error
}

type addressService struct {
	repo repositories.AddressRepository
}

func NewAddressService(repo repositories.AddressRepository) AddressService {
	return &addressService{repo: repo}
}

func (s *addressService) CreateAddress(ctx context.Context, address *types.Address) error {
	var userID = getUserID(ctx)
	if address == nil || address.AddressLine1 == "" || address.City == "" || address.StateCode == "" || address.PostalCode == "" {
		return errors.New("missing required fields for address")
	}
	address.UserID = userID

	addressID, err := utilities.GenerateIDString()
	if err != nil {
		return err
	}
	address.ID = addressID
	return s.repo.CreateAddress(ctx, address)
}

func (s *addressService) UpdateAddress(ctx context.Context, address *types.Address) error {
	var userID = getUserID(ctx)
	if address == nil || address.AddressLine1 == "" || address.City == "" || address.StateCode == "" || address.PostalCode == "" {
		return errors.New("missing required fields for address")
	}
	address.UserID = userID
	return s.repo.UpdateAddress(ctx, address)
}

func (s *addressService) GetAddresses(ctx context.Context) ([]types.Address, error) {
	var userID = getUserID(ctx)
	return s.repo.GetAddresses(ctx, userID)
}

func (s *addressService) RemoveAddress(ctx context.Context, addressID string) error {
	var userID = getUserID(ctx)
	return s.repo.RemoveAddress(ctx, userID, addressID)
}
