package repositories

import (
	"context"
	"database/sql"

	"github.com/dgyurics/marketplace/types"
)

type RateLimitRepository interface {
	GetHitCount(ctx context.Context, rl *types.RateLimit) error
	RecordHit(ctx context.Context, rl *types.RateLimit) error
}

type rateLimitRepository struct {
	db *sql.DB
}

func NewRateLimitRepository(db *sql.DB) RateLimitRepository {
	return &rateLimitRepository{
		db: db,
	}
}

func (r *rateLimitRepository) RecordHit(ctx context.Context, rl *types.RateLimit) error {
	query := `
		INSERT INTO rate_limits (ip_address, path, expires_at, hit_count)
		VALUES ($1, $2, $3, 1)
		ON CONFLICT (ip_address, path) DO UPDATE
		SET hit_count = rate_limits.hit_count + 1,
		expires_at = $3
		RETURNING hit_count
	`
	return r.db.QueryRowContext(ctx, query, rl.IPAddress, rl.Path, rl.ExpiresAt).Scan(&rl.HitCount)
}

func (r *rateLimitRepository) GetHitCount(ctx context.Context, rl *types.RateLimit) error {
	query := `
		SELECT hit_count
		FROM rate_limits
		WHERE ip_address = $1 AND path = $2
	`
	err := r.db.QueryRowContext(ctx, query, rl.IPAddress, rl.Path).Scan(&rl.HitCount)
	if err == sql.ErrNoRows {
		rl.HitCount = 0
		return nil
	}
	return err
}

// Use this alongside RecordHit
// TODO use this to implement an upper bound on rate_limits table
// Without bounds, adversarial input can cause unbounded memory allocation
// getTableSize returns the size of the postgresql table in megabytes
func (r *rateLimitRepository) getTableSize(ctx context.Context, tablename string) (int64, error) {
	var sizeBytes int64
	query := `SELECT pg_total_relation_size($1)`
	if err := r.db.QueryRowContext(ctx, query, tablename).Scan(&sizeBytes); err != nil {
		return 0, err
	}
	return sizeBytes / 1024 / 1024, nil
}
