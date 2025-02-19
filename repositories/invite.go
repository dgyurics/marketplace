package repositories

import (
	"context"
	"database/sql"
	"time"
)

// InviteRepository is a repository for managing invitation codes.
type InviteRepository interface {
	StoreCode(ctx context.Context, code string, used bool) error
	GetCode(ctx context.Context, code string) (used bool, exists bool, err error)
}

type inviteRepository struct {
	db *sql.DB
}

func NewInviteRepository(db *sql.DB) InviteRepository {
	return &inviteRepository{db: db}
}

// StoreCode inserts or updates an invitation code.
func (r *inviteRepository) StoreCode(ctx context.Context, code string, used bool) error {
	var usedAt interface{}
	if used {
		usedAt = time.Now()
	} else {
		usedAt = nil
	}

	query := `
		INSERT INTO invitation_codes (code, used_at)
		VALUES ($1, $2)
		ON CONFLICT (code) DO UPDATE
		SET used_at = EXCLUDED.used_at
	`
	_, err := r.db.ExecContext(ctx, query, code, usedAt)
	return err
}

// GetCode retrieves an invite code from the database
// and returns whether it has been used and if it exists
func (r *inviteRepository) GetCode(ctx context.Context, code string) (used bool, exists bool, err error) {
	query := `SELECT used_at FROM invitation_codes WHERE code = $1`
	var usedAt sql.NullTime

	err = r.db.QueryRowContext(ctx, query, code).Scan(&usedAt)
	if err == sql.ErrNoRows {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}

	return usedAt.Valid, true, nil
}
