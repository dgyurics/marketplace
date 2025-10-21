package repositories

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/dgyurics/marketplace/types"
)

type ScheduleRepository interface {
	RunJob(ctx context.Context, job types.Job, interval time.Duration) bool
}

type scheduleRepository struct {
	db *sql.DB
}

func NewScheduleRepository(db *sql.DB) ScheduleRepository {
	return &scheduleRepository{
		db: db,
	}
}

// RunJob checks if a job can be run based on its last run time and the specified interval.
func (s scheduleRepository) RunJob(ctx context.Context, job types.Job, interval time.Duration) bool {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("failed to begin tx", "err", err)
		return false
	}
	defer tx.Rollback()

	var lastRun time.Time
	err = tx.QueryRowContext(ctx, `
		SELECT last_run_at
		FROM job_schedules
		WHERE job_name = $1
		FOR UPDATE
	`, job).Scan(&lastRun)
	if err == sql.ErrNoRows {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO job_schedules (job_name, last_run_at)
			VALUES ($1, NOW())
		`, job)
		if err != nil {
			slog.Error("failed to create new job_schedule", "err", err)
			return false
		}

		if err := tx.Commit(); err != nil {
			slog.Error("failed to commit transaction", "err", err)
			return false
		}
		return true
	}
	if err != nil {
		slog.Error("failed to query last_run_at", "err", err)
		return false
	}

	// If too soon, bail out
	if time.Since(lastRun) < interval {
		return false
	}

	// Otherwise, update and run job
	_, err = tx.ExecContext(ctx, `
		UPDATE job_schedules SET last_run_at = NOW()
		WHERE job_name = $1
	`, job)
	if err != nil {
		slog.Error("failed to update last_run_at", "err", err)
		return false
	}

	if err := tx.Commit(); err != nil {
		slog.Error("failed to commit transaction", "err", err)
		return false
	}

	return true
}
