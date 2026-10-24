package services

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/dgyurics/marketplace/types"
)

type scheduleService struct {
	db *sql.DB
}

// ScheduleService is responsible for running tasks at intervals
type ScheduleService interface {
	Start(ctx context.Context)
}

func NewScheduleService(db *sql.DB) ScheduleService {
	return &scheduleService{
		db: db,
	}
}

// Start starts the scheduling service.
// Pass it root context to allow for clean shutdown.
func (s *scheduleService) Start(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	slog.Info("Scheduling service started")
	for {
		select {
		case <-ctx.Done():
			slog.Info("Scheduling service stopped")
			return
		case <-ticker.C:
			if s.shouldRunJob(ctx, types.StaleOrders, 15*time.Minute) {
				ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*10)
				s.removeStaleOrders(ctxTimeout)
				cancel()
			}
			if s.shouldRunJob(ctx, types.StaleRequestedOrders, 24*time.Hour) {
				ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*10)
				s.removeStaleRequestedOrders(ctxTimeout)
				cancel()
			}
			if s.shouldRunJob(ctx, types.ExpiredRateLimits, 10*time.Minute) {
				ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*10)
				s.removeStaleRateLimits(ctxTimeout)
				cancel()
			}
			if s.shouldRunJob(ctx, types.StaleCartItems, 10*time.Minute) {
				ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*10)
				s.removeStaleCartItems(ctxTimeout)
				cancel()
			}
			if s.shouldRunJob(ctx, types.ExpiredRefreshTokens, 24*time.Hour) {
				ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*10)
				s.removeExpiredRefreshTokens(ctxTimeout)
				cancel()
			}
			if s.shouldRunJob(ctx, types.ExpiredPasswordResets, 24*time.Hour) {
				ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*10)
				s.removeExpiredPasswordResets(ctxTimeout)
				cancel()
			}
			if s.shouldRunJob(ctx, types.ExpiredRegistrationCodes, 24*time.Hour) {
				ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*10)
				s.removeExpiredRegistrationCodes(ctxTimeout)
				cancel()
			}
		}
	}
}

// Users have 15 minutes to fill out payment information upon reaching the payments page
func (s *scheduleService) removeStaleOrders(ctx context.Context) {
	_, err := s.db.ExecContext(ctx, `
		WITH canceled_orders AS (
			UPDATE orders
			SET status = 'canceled', updated_at = NOW()
			WHERE
				updated_at < NOW() - INTERVAL '15 minutes' AND (
					status = 'draft' OR
					(status = 'pending' AND payment_method = 'stripe')
				)
			RETURNING id, address_id
		),
		deleted_items AS (
			DELETE FROM order_items oi
			USING canceled_orders co
			WHERE oi.order_id = co.id
			RETURNING oi.product_id, oi.quantity
		),
		restock_totals AS (
			SELECT product_id, SUM(quantity) AS quantity
			FROM deleted_items
			GROUP BY product_id
		),
		restored AS (
			UPDATE products
			SET inventory = inventory + rt.quantity
			FROM restock_totals rt
			WHERE products.id = rt.product_id
		)
		DELETE FROM addresses
		WHERE id IN (SELECT address_id FROM canceled_orders)
		AND id NOT IN (SELECT DISTINCT address_id FROM orders WHERE status != 'canceled')`)
	if err != nil {
		slog.ErrorContext(ctx, "Error canceling stale orders", "error", err)
	}
}

// Requested orders expire after seven days without an update. Their stock is already released.
func (s *scheduleService) removeStaleRequestedOrders(ctx context.Context) {
	_, err := s.db.ExecContext(ctx, `
		WITH canceled_orders AS (
			UPDATE orders
			SET status = 'canceled', updated_at = NOW()
			WHERE status = 'requested' AND updated_at < NOW() - INTERVAL '7 days'
			RETURNING id
		)
		DELETE FROM requested_items
		WHERE order_id IN (SELECT id FROM canceled_orders)`)
	if err != nil {
		slog.ErrorContext(ctx, "Error canceling stale requested orders", "error", err)
	}
}

func (s *scheduleService) removeStaleRateLimits(ctx context.Context) {
	_, err := s.db.ExecContext(ctx, `
        DELETE FROM rate_limits 
        WHERE expires_at < NOW()`)
	if err != nil {
		slog.ErrorContext(ctx, "Error removing stale rate limits", "error", err)
	}
}

func (s *scheduleService) removeStaleCartItems(ctx context.Context) {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM cart_items
		WHERE updated_at < NOW() - INTERVAL '30 days'`)
	if err != nil {
		slog.ErrorContext(ctx, "Error removing stale cart items", "error", err)
	}
}

func (s *scheduleService) removeExpiredRegistrationCodes(ctx context.Context) {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM registration_codes
		WHERE expires_at < NOW()`)
	if err != nil {
		slog.ErrorContext(ctx, "Error removing expired registration codes", "error", err)
	}
}

func (s *scheduleService) removeExpiredRefreshTokens(ctx context.Context) {
	_, err := s.db.ExecContext(ctx, `
        DELETE FROM refresh_tokens 
        WHERE expires_at < NOW() - INTERVAL '1 month'`)
	if err != nil {
		slog.ErrorContext(ctx, "Error removing expired refresh tokens", "error", err)
	}
}

func (s *scheduleService) removeExpiredPasswordResets(ctx context.Context) {
	_, err := s.db.ExecContext(ctx, `
        DELETE FROM password_reset_codes 
        WHERE expires_at < NOW() - INTERVAL '1 month'`)
	if err != nil {
		slog.ErrorContext(ctx, "Error removing expired password reset codes", "error", err)
	}
}

// shouldRunJob checks if enough time has passed since the last run and updates the timestamp
func (s *scheduleService) shouldRunJob(ctx context.Context, job types.Job, interval time.Duration) bool {
	var lastRun sql.NullTime
	err := s.db.QueryRowContext(ctx, `
        SELECT last_run_at FROM job_schedules WHERE job_name = $1
    `, job).Scan(&lastRun)

	if err == sql.ErrNoRows {
		// First time running this job
		_, err = s.db.ExecContext(ctx, `
            INSERT INTO job_schedules (job_name, last_run_at) 
            VALUES ($1, NOW())
        `, job)
		if err != nil {
			slog.ErrorContext(ctx, "failed to create job schedule", "job", job, "error", err)
			return false
		}
		return true
	}

	if err != nil {
		slog.ErrorContext(ctx, "failed to query job schedule", "job", job, "error", err)
		return false
	}

	// Check if enough time has passed
	if lastRun.Valid && time.Since(lastRun.Time) < interval {
		return false
	}

	// Update the last run time
	_, err = s.db.ExecContext(ctx, `
        UPDATE job_schedules SET last_run_at = NOW() WHERE job_name = $1
    `, job)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update job schedule", "job", job, "error", err)
		return false
	}

	return true
}
