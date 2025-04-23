package repositories

import (
	"context"
	"database/sql"
	"fmt"
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

func (s scheduleRepository) RunJob(ctx context.Context, job types.Job, interval time.Duration) bool {
	intervalStr := fmt.Sprintf("%d seconds", int(interval.Seconds()))
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO job_schedules(job_name, last_run_at)
		VALUES ($1, NOW())
		ON CONFLICT (job_name)
		DO UPDATE SET last_run_at = NOW()
		WHERE job_schedules.last_run_at < NOW() - ($2)::INTERVAL
	`, job, intervalStr)

	if err != nil {
		slog.Error("Error updating job_schedules", "error", err)
		return false
	}

	rowsAffected, _ := res.RowsAffected()
	return rowsAffected == 1
}
