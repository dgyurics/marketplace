package services

import (
	"context"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
)

type ShippingZoneService interface {
	// Check shipping availability
	IsShippable(ctx context.Context, address *types.Address) (bool, error)

	// Manage shipping zones
	AddShippingZone(ctx context.Context, zone *types.ShippingZone) error
	RemoveShippingZone(ctx context.Context, zoneID string) error
	GetShippingZones(ctx context.Context) ([]types.ShippingZone, error)

	// Manage restricted zones
	AddExcludedShippingZone(ctx context.Context, zone *types.ExcludedShippingZone) error
	RemoveExcludedShippingZone(ctx context.Context, zoneID string) error
	GetExcludedShippingZones(ctx context.Context) ([]types.ExcludedShippingZone, error)
}

type shippingZoneService struct {
	repo repositories.ShippingZoneRepository
}

func NewShippingZoneService(repo repositories.ShippingZoneRepository) ShippingZoneService {
	return &shippingZoneService{
		repo: repo,
	}
}

func (s *shippingZoneService) IsShippable(ctx context.Context, address *types.Address) (bool, error) {
	return s.repo.IsShippable(ctx, address)
}

func (s *shippingZoneService) AddShippingZone(ctx context.Context, zone *types.ShippingZone) error {
	zoneID, err := utilities.GenerateIDString()
	if err != nil {
		return err
	}
	zone.ID = zoneID
	return s.repo.AddShippingZone(ctx, zone)
}

func (s *shippingZoneService) RemoveShippingZone(ctx context.Context, zoneID string) error {
	return s.repo.RemoveShippingZone(ctx, zoneID)
}

func (s *shippingZoneService) GetShippingZones(ctx context.Context) ([]types.ShippingZone, error) {
	return s.repo.GetShippingZones(ctx)
}

func (s *shippingZoneService) AddExcludedShippingZone(ctx context.Context, zone *types.ExcludedShippingZone) error {
	zoneID, err := utilities.GenerateIDString()
	if err != nil {
		return err
	}
	zone.ID = zoneID
	return s.repo.AddExcludedShippingZone(ctx, zone)
}

func (s *shippingZoneService) RemoveExcludedShippingZone(ctx context.Context, zoneID string) error {
	return s.repo.RemoveExcludedShippingZone(ctx, zoneID)
}

func (s *shippingZoneService) GetExcludedShippingZones(ctx context.Context) ([]types.ExcludedShippingZone, error) {
	return s.repo.GetExcludedShippingZones(ctx)
}
