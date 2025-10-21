package repositories

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/dgyurics/marketplace/types"
	"github.com/stretchr/testify/assert"
)

func TestRecordHit_FirstHit(t *testing.T) {
	repo := NewRateLimitRepository(dbPool)
	ctx := context.Background()

	ip := generateTestIP()
	path := generateTestEndpoint("first_hit")
	rl := &types.RateLimit{
		IPAddress: ip,
		Path:      path,
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}

	defer cleanupRateLimit(t, ip, path)

	err := repo.RecordHit(ctx, rl)
	assert.NoError(t, err, "Expected no error on first hit")
	assert.Equal(t, 1, rl.HitCount, "Expected hit count to be 1")
}

func TestRecordHit_MultipleHits(t *testing.T) {
	repo := NewRateLimitRepository(dbPool)
	ctx := context.Background()

	ip := generateTestIP()
	path := generateTestEndpoint("multiple_hits")
	rl := &types.RateLimit{
		IPAddress: ip,
		Path:      path,
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}

	defer cleanupRateLimit(t, ip, path)

	// Record multiple hits
	for i := 1; i <= 5; i++ {
		err := repo.RecordHit(ctx, rl)
		assert.NoError(t, err, "Expected no error on hit %d", i)
		assert.Equal(t, i, rl.HitCount, "Expected hit count to be %d", i)
	}
}

func TestGetHitCount_NoRecord(t *testing.T) {
	repo := NewRateLimitRepository(dbPool)
	ctx := context.Background()

	ip := generateTestIP()
	path := generateTestEndpoint("no_record")
	rl := &types.RateLimit{
		IPAddress: ip,
		Path:      path,
	}

	// Check hit count for non-existent record
	err := repo.GetHitCount(ctx, rl)
	assert.NoError(t, err, "Expected no error when record doesn't exist")
	assert.Equal(t, 0, rl.HitCount, "Expected hit count to be 0 when record doesn't exist")
}

func TestGetHitCount_WithinLimit(t *testing.T) {
	repo := NewRateLimitRepository(dbPool)
	ctx := context.Background()

	ip := generateTestIP()
	path := generateTestEndpoint("within_limit_check")
	rl := &types.RateLimit{
		IPAddress: ip,
		Path:      path,
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}

	defer cleanupRateLimit(t, ip, path)

	// Record some hits
	err := repo.RecordHit(ctx, rl)
	assert.NoError(t, err, "Expected no error recording hit")
	assert.Equal(t, 1, rl.HitCount, "Expected hit count to be 1")

	err = repo.RecordHit(ctx, rl)
	assert.NoError(t, err, "Expected no error recording second hit")
	assert.Equal(t, 2, rl.HitCount, "Expected hit count to be 2")

	// Check hit count
	err = repo.GetHitCount(ctx, rl)
	assert.NoError(t, err, "Expected no error getting hit count")
	assert.Equal(t, 2, rl.HitCount, "Expected hit count to be 2")
}

func TestGetHitCount_AtLimit(t *testing.T) {
	repo := NewRateLimitRepository(dbPool)
	ctx := context.Background()

	ip := generateTestIP()
	path := generateTestEndpoint("at_limit_check")
	rl := &types.RateLimit{
		IPAddress: ip,
		Path:      path,
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}

	defer cleanupRateLimit(t, ip, path)

	// Record hits up to 3
	for i := 0; i < 3; i++ {
		err := repo.RecordHit(ctx, rl)
		assert.NoError(t, err, "Expected no error recording hit %d", i+1)
		assert.Equal(t, i+1, rl.HitCount, "Expected hit count to be %d", i+1)
	}

	// Check hit count
	err := repo.GetHitCount(ctx, rl)
	assert.NoError(t, err, "Expected no error getting hit count")
	assert.Equal(t, 3, rl.HitCount, "Expected hit count to be 3")
}

func TestGetHitCount_ExceedLimit(t *testing.T) {
	repo := NewRateLimitRepository(dbPool)
	ctx := context.Background()

	ip := generateTestIP()
	path := generateTestEndpoint("exceed_limit_check")
	rl := &types.RateLimit{
		IPAddress: ip,
		Path:      path,
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}

	defer cleanupRateLimit(t, ip, path)

	// Record 4 hits
	for i := 0; i < 4; i++ {
		err := repo.RecordHit(ctx, rl)
		assert.NoError(t, err, "Expected no error recording hit %d", i+1)
		assert.Equal(t, i+1, rl.HitCount, "Expected hit count to be %d", i+1)
	}

	// Check hit count
	err := repo.GetHitCount(ctx, rl)
	assert.NoError(t, err, "Expected no error getting hit count")
	assert.Equal(t, 4, rl.HitCount, "Expected hit count to be 4")
}

func TestCleanup_ExpiredRecords(t *testing.T) {
	repo := NewRateLimitRepository(dbPool)
	ctx := context.Background()

	ip := generateTestIP()
	path := generateTestEndpoint("cleanup_expired")
	rl := &types.RateLimit{
		IPAddress: ip,
		Path:      path,
		ExpiresAt: time.Now().UTC().Add(time.Millisecond), // Very short expiry
	}

	defer cleanupRateLimit(t, ip, path)

	// Create a record with short expiry
	err := repo.RecordHit(ctx, rl)
	assert.NoError(t, err, "Expected no error creating record")

	// Wait for expiry
	time.Sleep(10 * time.Millisecond)

	// Cleanup expired records
	err = repo.Cleanup(ctx)
	assert.NoError(t, err, "Expected no error during cleanup")

	// Record should be gone (hit count should be 0)
	err = repo.GetHitCount(ctx, rl)
	assert.NoError(t, err, "Expected no error getting hit count")
	assert.Equal(t, 0, rl.HitCount, "Expected record to be cleaned up")
}

func TestCleanup_RecentRecords(t *testing.T) {
	repo := NewRateLimitRepository(dbPool)
	ctx := context.Background()

	ip := generateTestIP()
	path := generateTestEndpoint("cleanup_recent")
	rl := &types.RateLimit{
		IPAddress: ip,
		Path:      path,
		ExpiresAt: time.Now().UTC().Add(time.Hour), // Long expiry
	}

	defer cleanupRateLimit(t, ip, path)

	// Create a recent record
	err := repo.RecordHit(ctx, rl)
	assert.NoError(t, err, "Expected no error creating record")
	assert.Equal(t, 1, rl.HitCount, "Expected hit count to be 1 after recording")

	// Verify record exists before cleanup
	rl.HitCount = 0 // Reset
	err = repo.GetHitCount(ctx, rl)
	assert.NoError(t, err, "Expected no error getting hit count before cleanup")
	assert.Equal(t, 1, rl.HitCount, "Expected record to exist before cleanup")

	// Cleanup (should not delete recent record with 1-hour expiry)
	err = repo.Cleanup(ctx)
	assert.NoError(t, err, "Expected no error during cleanup")

	// Record should still exist after cleanup
	rl.HitCount = 0 // Reset to verify it gets updated
	err = repo.GetHitCount(ctx, rl)
	assert.NoError(t, err, "Expected no error getting hit count after cleanup")
	assert.Equal(t, 1, rl.HitCount, "Expected recent record to still exist after cleanup")
}

func TestCleanup_EmptyTable(t *testing.T) {
	repo := NewRateLimitRepository(dbPool)
	ctx := context.Background()

	// Cleanup non-existent records should not error
	err := repo.Cleanup(ctx)
	assert.NoError(t, err, "Expected no error cleaning up empty table")
}

// Helper function to generate unique test IP addresses
func generateTestIP() string {
	return fmt.Sprintf("192.168.%d.%d",
		time.Now().UnixNano()%255+1,
		time.Now().UnixNano()%254+1)
}

// Helper function to generate unique test endpoints
func generateTestEndpoint(prefix string) string {
	return fmt.Sprintf("/test/%s/%d", prefix, time.Now().UnixNano())
}

// Helper function to clean up test data
func cleanupRateLimit(t *testing.T, ip, path string) {
	ctx := context.Background()

	// Parse IP to ensure it's valid for INET type
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		t.Logf("Warning: invalid IP address for cleanup: %s", ip)
		return
	}

	_, err := dbPool.ExecContext(ctx,
		"DELETE FROM rate_limits WHERE ip_address = $1 AND path = $2",
		ip, path)
	if err != nil {
		t.Logf("Warning: failed to cleanup rate limit record: %v", err)
	}
}
