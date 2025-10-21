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
	orderSrv   OrderService
	rateLimSrv RateLimitService
	schedRepo  repositories.ScheduleRepository
}

// ScheduleService is responsible for running tasks at a specified interval
type ScheduleService interface {
	Start(ctx context.Context)
}

func NewScheduleService(orderSrv OrderService, rateLimSrv RateLimitService, schedRepo repositories.ScheduleRepository) ScheduleService {
	return &scheduleService{
		orderSrv:   orderSrv,
		rateLimSrv: rateLimSrv,
		schedRepo:  schedRepo,
	}
}

// TODO cleanup expired refresh tokens
// TODO cleanup expired password reset codes
// TODO cleanup unused addresses

// FIXME refactor/redesign the service

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
			// TODO refactor
		case <-ticker.C:
			// remove stale orders every 10 minutes
			if s.schedRepo.RunJob(ctx, types.StaleOrders, 10*time.Minute) {
				ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*10)
				if err := s.orderSrv.CancelStaleOrders(ctxTimeout); err != nil {
					slog.Error("Error canceling stale orders", "error", err)
				}
				cancel()
			}
			// remove stale rate limits every 10 minutes
			if s.schedRepo.RunJob(ctx, types.StaleRateLimits, 10*time.Minute) {
				ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*10)
				if err := s.rateLimSrv.Cleanup(ctxTimeout); err != nil {
					slog.Error("Error purging expired rate limit entries", "error", err)
				}
				cancel()
			}
		}
	}
}
