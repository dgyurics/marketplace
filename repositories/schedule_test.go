package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/dgyurics/marketplace/types"
	"github.com/stretchr/testify/assert"
)

func TestRunJob_FirstRun(t *testing.T) {
	ctx := context.Background()
	jobName := generateTestJob("first_run")
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
	jobName := generateTestJob("too_soon")
	interval := 10 * time.Second

	repo := NewScheduleRepository(dbPool)

	// Cleanup after test
	defer func() {
		_, _ = dbPool.ExecContext(ctx, "DELETE FROM job_schedules WHERE job_name = $1", jobName)
	}()

	// Insert initial job entry with recent last_run_at
	_, _ = dbPool.ExecContext(ctx, `
		INSERT INTO job_schedules (job_name, last_run_at)
		VALUES ($1, NOW())
		ON CONFLICT (job_name) DO UPDATE SET last_run_at = NOW()
	`, jobName)

	// Run again immediately should return false
	ran := repo.RunJob(ctx, jobName, interval)
	assert.False(t, ran)
}

func TestRunJob_AfterInterval(t *testing.T) {
	ctx := context.Background()
	jobName := generateTestJob("after_interval")
	interval := 1 * time.Second

	repo := NewScheduleRepository(dbPool)

	// Cleanup after test
	defer func() {
		_, _ = dbPool.ExecContext(ctx, "DELETE FROM job_schedules WHERE job_name = $1", jobName)
	}()

	// Insert initial job entry with older last_run_at
	_, _ = dbPool.ExecContext(ctx, `
		INSERT INTO job_schedules (job_name, last_run_at)
		VALUES ($1, NOW() - INTERVAL '2 seconds')
		ON CONFLICT (job_name) DO UPDATE SET last_run_at = NOW() - INTERVAL '2 seconds'
	`, jobName)

	// Wait for interval to pass (should already be enough but wait extra)
	time.Sleep(100 * time.Millisecond)

	// Run should succeed
	ran := repo.RunJob(ctx, jobName, interval)
	assert.True(t, ran)
}

func TestRunJob_Concurrent(t *testing.T) {
	ctx := context.Background()
	jobName := generateTestJob("concurrent")
	interval := 10 * time.Second

	repo := NewScheduleRepository(dbPool)

	// Cleanup after test
	defer func() {
		_, _ = dbPool.ExecContext(ctx, "DELETE FROM job_schedules WHERE job_name = $1", jobName)
	}()

	const numGoroutines = 10
	results := make(chan bool, numGoroutines)

	// Insert an initial entry so all goroutines race for it
	_, _ = dbPool.ExecContext(ctx, `
		INSERT INTO job_schedules (job_name, last_run_at)
		VALUES ($1, NOW() - INTERVAL '1 hour')
		ON CONFLICT (job_name) DO UPDATE SET last_run_at = NOW() - INTERVAL '1 hour'
	`, jobName)

	// Start concurrent runners
	for i := 0; i < numGoroutines; i++ {
		go func() {
			success := repo.RunJob(ctx, jobName, interval)
			results <- success
		}()
	}

	successCount := 0
	for i := 0; i < numGoroutines; i++ {
		if <-results {
			successCount++
		}
	}

	assert.Equal(t, 1, successCount, "only one goroutine should succeed")
}

func generateTestJob(name string) types.Job {
	return types.Job(fmt.Sprintf("test_%s_%d", name, time.Now().UnixNano()))
}
