package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/dgyurics/marketplace/types"
	"github.com/stretchr/testify/assert"
)

func TestRunJob_FirstRun(t *testing.T) {
	ctx := context.Background()
	jobName := types.StaleOrders
	interval := 10 * time.Second

	repo := NewScheduleRepository(dbPool)

	// Cleanup after test
	defer func() {
		_, _ = dbPool.ExecContext(ctx, "DELETE FROM job_schedules WHERE job_name = $1", jobName)
	}()

	// RunJob should insert new job and return true
	ran := repo.RunJob(ctx, jobName, interval)
	assert.True(t, ran)
}

func TestRunJob_TooSoon(t *testing.T) {
	ctx := context.Background()
	jobName := types.StaleOrders
	interval := 10 * time.Second

	repo := NewScheduleRepository(dbPool)

	// Cleanup after test
	defer func() {
		_, _ = dbPool.ExecContext(ctx, "DELETE FROM job_schedules WHERE job_name = $1", jobName)
	}()

	// First run inserts the job and returns true
	ran := repo.RunJob(ctx, jobName, interval)
	assert.True(t, ran)

	// Run again immediately should return false
	ran = repo.RunJob(ctx, jobName, interval)
	assert.False(t, ran)
}

func TestRunJob_AfterInterval(t *testing.T) {
	ctx := context.Background()
	jobName := types.StaleOrders
	interval := 1 * time.Second

	repo := NewScheduleRepository(dbPool)

	// Cleanup after test
	defer func() {
		_, _ = dbPool.ExecContext(ctx, "DELETE FROM job_schedules WHERE job_name = $1", jobName)
	}()

	// First run inserts the job and returns true
	ran := repo.RunJob(ctx, jobName, interval)
	assert.True(t, ran)

	// Wait for interval to pass
	time.Sleep(interval + 100*time.Millisecond)

	// Run again should update and return true
	ran = repo.RunJob(ctx, jobName, interval)
	assert.True(t, ran)
}
