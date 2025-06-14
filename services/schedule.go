package services

import (
	"context"
	"log/slog"
	"time"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
)

// FIXME replace orderService with servicesContainer,
// this	 would allow schedule service to handle many other
// operations like sending emails, etc.
type scheduleService struct {
	orderSrv  OrderService
	schedRepo repositories.ScheduleRepository
}

// ScheduleService is responsible for running tasks at a specified interval
type ScheduleService interface {
	Start(ctx context.Context)
}

func NewScheduleService(orderSrv OrderService, schedRepo repositories.ScheduleRepository) ScheduleService {
	return &scheduleService{
		orderSrv:  orderSrv,
		schedRepo: schedRepo,
	}
}

// TODO cleanup expired refresh tokens
// TODO cleanup expired password reset codes
// TODO cleanup unused addresses

// Start starts the scheduling service.
// Pass it root context to allow for clean shutdown.
func (s *scheduleService) Start(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Minute) // TODO make this configurable
	defer ticker.Stop()

	slog.Info("Scheduling service started")
	for {
		select {
		case <-ctx.Done():
			slog.Info("Scheduling service stopped")
			return
		case <-ticker.C:
			// run job every 10 minutes
			if s.schedRepo.RunJob(ctx, types.StaleOrders, 10*time.Minute) {
				ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*10)
				s.orderSrv.CancelStaleOrders(ctxTimeout)
				cancel()
			}
		}
	}
}
